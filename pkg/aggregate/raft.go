package aggregate

import (
	"encoding/json"
	"fmt"
	"io"
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
}

type command struct {
	Op     string       `json:"op"` // stash or evict
	Rule   *event.Rule  `json:"rule,omitempty"`
	RuleID string       `json:"ruleID,omitempty"`
	Event  *event.Event `json:"event,omitempty"`
}

type defaultStore struct {
	opt     *options
	raft    *raft.Raft
	storage *storage
}

type options struct {
	dir  string
	bind string
	id   string
	join string
}

func newStore(opt *options) (*defaultStore, error) {

	store := &defaultStore{
		storage: &storage{
			m:           make(map[string]*event.RuleBucket),
			flusherChan: make(chan string),
		},
		opt: opt,
	}
	store.open()
	return store, nil
}

func (d *defaultStore) open() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(d.opt.id)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", d.opt.bind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(d.opt.bind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(d.opt.dir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore

	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(d.opt.dir, "raft.db"))
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

	if d.opt.join == "" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	} else {
		d.join()
	}

	go d.flusher()

	return nil
}

func (d *defaultStore) flusher() {
	for {
		select {
		case ruleID := <-d.storage.flusherChan:
			err := d.FlushRule(ruleID)
			glog.Errorf("error flushing %v", err)

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

func (d *defaultStore) join() error {
	glog.Infof("received join request for remote node %s at %s", d.opt.id, d.opt.bind)

	configFuture := d.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		glog.Infof("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(d.opt.id) || srv.Address == raft.ServerAddress(d.opt.join) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(d.opt.join) && srv.ID == raft.ServerID(d.opt.id) {
				glog.Infof("node %s at %s already member of cluster, ignoring join request", d.opt.id, d.opt.join)
				return nil
			}

			future := d.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", d.opt.id, d.opt.join, err)
			}
		}
	}

	f := d.raft.AddVoter(raft.ServerID(d.opt.id), raft.ServerAddress(d.opt.join), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	glog.Infof("node %s at %s joined successfully", d.opt.id, d.opt.join)
	return nil
}

type fsm defaultStore

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "stash":
		return f.applyStash(c.Event)
	case "add_rule":
		return f.applyAddRule(c.Rule)
	case "remove_rule":
		return f.applyRemoveRule(c.RuleID)
	case "flush_rule":
		return f.applyFlushRule(c.RuleID)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}

}

func (f *fsm) applyStash(event *event.Event) interface{} {
	f.storage.stash(event)
	return nil
}

func (f *fsm) applyAddRule(rule *event.Rule) interface{} {
	ok := f.storage.addRule(rule)
	if !ok {
		return fmt.Errorf("error adding rule %v", rule)
	}
	return nil
}

func (f *fsm) applyRemoveRule(ruleID string) interface{} {
	ok := f.storage.removeRule(ruleID)
	if !ok {
		return fmt.Errorf("error removing rule %v", ruleID)
	}
	return nil
}

func (f *fsm) applyFlushRule(ruleID string) interface{} {
	f.storage.flushRule(ruleID)
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	db := f.storage.clone()
	return &fsmSnapShot{db: db}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	db := make(map[string]*event.RuleBucket)

	if err := json.NewDecoder(rc).Decode(&db); err != nil {
		return err
	}

	f.storage.restore(db)

	return nil
}

type fsmSnapShot struct {
	db map[string]*event.RuleBucket
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.db)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapShot) Release() {}
