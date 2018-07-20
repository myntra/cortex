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
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

// Store is the raft backed json storage
type Store interface {
	Stash(event *event.Event) error // stash to an aggregate map by event type
	AddRule(rule *event.Rule) error // rule which correlates aggregated events
	RemoveRule(ruleID string) error
	FlushRule(ruleID string) error
	GetRules() []*event.Rule
	Join(addr string) error
	Leave(id string) error
}

type command struct {
	Op     string       `json:"op"` // stash or evict
	Rule   *event.Rule  `json:"rule,omitempty"`
	RuleID string       `json:"ruleID,omitempty"`
	Event  *event.Event `json:"event,omitempty"`
}

type defaultStore struct {
	opt     *Config
	raft    *raft.Raft
	storage *storage
}

func newStore(opt *Config) (*defaultStore, error) {

	store := &defaultStore{
		storage: &storage{
			m:               make(map[string]*event.RuleBucket),
			flusherChan:     make(chan string),
			quitFlusherChan: make(chan struct{}),
		},
		opt: opt,
	}
	store.open()
	return store, nil
}

func (d *defaultStore) open() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(d.opt.NodeID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", d.opt.BindAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(d.opt.BindAddr, addr, 3, 10*time.Second, os.Stderr)
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

	// single node configuration
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
	} else {
		d.Join(d.opt.JoinAddr)
	}

	go d.flusher()

	return nil
}

func (d *defaultStore) close() error {
	d.storage.quitFlusherChan <- struct{}{}
	f := d.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}
	glog.Flush()
	return nil
}

func (d *defaultStore) flusher() {
loop:
	for {
		select {
		case ruleID := <-d.storage.flusherChan:
			rb := d.storage.getRule(ruleID)
			if rb == nil {
				glog.Errorf("unexpected err ruleID %v not found", ruleID)
				return
			}
			err := rb.Post()
			if err != nil {
				b, err2 := json.Marshal(rb)
				glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, string(b), err2)
			}
			err = d.FlushRule(ruleID)
			glog.Errorf("error flushing %v", err)

		case <-d.storage.quitFlusherChan:
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

func (d *defaultStore) Stash(event *event.Event) error {
	return d.applyCMD(&command{
		Op:    "stash",
		Event: event,
	})
}

func (d *defaultStore) AddRule(rule *event.Rule) error {

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

func (d *defaultStore) RemoveRule(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "remove_rule",
		RuleID: ruleID,
	})
}

func (d *defaultStore) FlushRule(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "flush_rule",
		RuleID: ruleID,
	})
}

func (d *defaultStore) GetRules() []*event.Rule {
	return d.storage.getRules()
}

func (d *defaultStore) Join(addr string) error {
	glog.Infof("received join request for remote node %s at %s", d.opt.NodeID, addr)

	configFuture := d.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		glog.Infof("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(d.opt.NodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(d.opt.NodeID) {
				glog.Infof("node %s at %s already member of cluster, ignoring join request", d.opt.NodeID, addr)
				return nil
			}

			future := d.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", d.opt.NodeID, addr, err)
			}
		}
	}

	f := d.raft.AddVoter(raft.ServerID(d.opt.NodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	glog.Infof("node %s at %s joined successfully", d.opt.NodeID, addr)
	return nil
}

func (d *defaultStore) Leave(nodeID string) error {

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

			glog.Infof("node %s leaved successfully", nodeID)
			return nil
		}
	}

	glog.Infof("node %s not exists in raft group", nodeID)

	return nil

}
