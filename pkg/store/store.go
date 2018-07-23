package store

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/myntra/aggo/pkg/event"
	"github.com/myntra/aggo/pkg/js"
	"github.com/myntra/aggo/pkg/util"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op       string       `json:"op"` // stash or evict
	Rule     *event.Rule  `json:"rule,omitempty"`
	RuleID   string       `json:"ruleID,omitempty"`
	Event    *event.Event `json:"event,omitempty"`
	ScriptID string       `json:"script_id,omitempty"`
	Script   []byte       `json:"script,omitempty"`
}

type defaultStore struct {
	opt             *util.Config
	raft            *raft.Raft
	eventStorage    *eventStorage
	scriptStorage   *scriptStorage
	postBucketQueue chan *event.RuleBucket
}

func newStore(opt *util.Config) (*defaultStore, error) {

	store := &defaultStore{
		eventStorage: &eventStorage{
			m:               make(map[string]*event.RuleBucket),
			flusherChan:     make(chan string),
			quitFlusherChan: make(chan struct{}),
		},
		scriptStorage: &scriptStorage{
			m: make(map[string][]byte),
		},
		opt:             opt,
		postBucketQueue: make(chan *event.RuleBucket, 100),
	}
	store.open()
	return store, nil
}

func (d *defaultStore) open() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(d.opt.NodeID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", d.opt.GetBindAddr())
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(d.opt.GetBindAddr(), addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(d.opt.Dir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore

	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(d.opt.Dir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	// Instantiate the Raft systemd.
	ra, err := raft.NewRaft(config, (*fsm)(d), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	d.raft = ra

	// bootstrap single node configuration
	if d.opt.JoinAddr == "" {
		fmt.Printf("starting %v in a single node cluster \n", d.opt.NodeID)
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		f := ra.BootstrapCluster(configuration)
		if f.Error() != nil {
			return f.Error()
		}

		// since in bootstrap mode, block until leadership is attained.
	loop:
		for {
			select {
			case leader := <-d.raft.LeaderCh():
				glog.Info("isLeader ", leader)
				if leader {
					break loop
				}
			}
		}
	}

	go d.flusher()

	return nil
}

func (d *defaultStore) close() error {
	d.eventStorage.quitFlusherChan <- struct{}{}
	f := d.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}
	glog.Flush()
	return nil
}

func (d *defaultStore) poster() {
	for {
		select {
		case rb := <-d.postBucketQueue:
			glog.Infof("received bucket %+v", rb)
			go func(rb *event.RuleBucket) {
				script := d.getScript(rb.Rule.ScriptID)
				if len(script) == 0 {
					util.RetryPost(rb, rb.Rule.HookEndpoint, rb.Rule.HookRetry)
					return
				}

				result := js.Execute(script, rb)
				if result == nil {
					util.RetryPost(rb, rb.Rule.HookEndpoint, rb.Rule.HookRetry)
					return
				}
				util.RetryPost(result, rb.Rule.HookEndpoint, rb.Rule.HookRetry)

			}(rb)

		}
	}
}

func (d *defaultStore) flusher() {
	if !d.opt.DisablePostHook {
		go d.poster()
	}

loop:
	for {
		select {
		case ruleID := <-d.eventStorage.flusherChan:
			rb := d.eventStorage.getRuleBucket(ruleID)
			if rb == nil {
				glog.Errorf("unexpected err ruleID %v not found", ruleID)
				return
			}

			go func() {

				if d.opt.DisablePostHook {
					glog.Infof("post bucket to hook %v ", rb)
				}

				d.postBucketQueue <- rb

			}()

			err := d.flushRule(ruleID)
			if err != nil {
				glog.Errorf("error flushing %v", err)
			}

		case <-d.eventStorage.quitFlusherChan:
			break loop
		}

	}
}

func (d *defaultStore) applyCMD(cmd *command) error {
	if d.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	f := d.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (d *defaultStore) stash(event *event.Event) error {
	return d.applyCMD(&command{
		Op:    "stash",
		Event: event,
	})
}

func (d *defaultStore) addRule(rule *event.Rule) error {

	if rule.WaitWindow == 0 || rule.WaitWindowThreshold == 0 || rule.MaxWaitWindow == 0 {
		rule.WaitWindow = d.opt.DefaultWaitWindow
		rule.WaitWindowThreshold = d.opt.DefaultWaitWindowThreshold
		rule.MaxWaitWindow = d.opt.DefaultMaxWaitWindow
	}

	return d.applyCMD(&command{
		Op:   "add_rule",
		Rule: rule,
	})
}

func (d *defaultStore) addScript(id string, script []byte) error {
	return d.applyCMD(&command{
		Op:       "add_script",
		ScriptID: id,
		Script:   script,
	})
}

func (d *defaultStore) updateScript(id string, script []byte) error {
	return d.applyCMD(&command{
		Op:       "update_script",
		ScriptID: id,
		Script:   script,
	})
}

func (d *defaultStore) removeScript(id string) error {
	return d.applyCMD(&command{
		Op:       "remove_script",
		ScriptID: id,
	})
}

func (d *defaultStore) getScripts() []string {
	return d.scriptStorage.getScripts()
}

func (d *defaultStore) getScript(id string) []byte {
	return d.scriptStorage.getScript(id)
}

func (d *defaultStore) removeRule(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "remove_rule",
		RuleID: ruleID,
	})
}

func (d *defaultStore) flushRule(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "flush_rule",
		RuleID: ruleID,
	})
}

func (d *defaultStore) getRules() []*event.Rule {
	return d.eventStorage.getRules()
}

func (d *defaultStore) getRule(ruleID string) *event.Rule {
	return d.eventStorage.getRule(ruleID)
}

func (d *defaultStore) acceptJoin(nodeID, addr string) error {
	glog.Infof("received join request for remote node %s at %s", nodeID, addr)

	configFuture := d.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		glog.Infof("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				glog.Infof("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := d.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := d.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	glog.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil

}

func (d *defaultStore) acceptLeave(nodeID string) error {

	glog.Infof("received leave request for remote node %s", nodeID)

	cf := d.raft.GetConfiguration()

	if err := cf.Error(); err != nil {
		glog.Infof("failed to get raft configuration")
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) {
			f := d.raft.RemoveServer(server.ID, 0, 0)
			if err := f.Error(); err != nil {
				glog.Infof("failed to remove server %s", nodeID)
				return err
			}

			glog.Infof("node %s left successfully", nodeID)
			return nil
		}
	}

	glog.Infof("node %s not exists in raft group", nodeID)

	return nil

}
