package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"

	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/util"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op       string        `json:"op"` // stash or evict
	Rule     *rules.Rule   `json:"rule,omitempty"`
	RuleID   string        `json:"ruleID,omitempty"`
	Event    *events.Event `json:"event,omitempty"`
	ScriptID string        `json:"script_id,omitempty"`
	Script   []byte        `json:"script,omitempty"`
}

type defaultStore struct {
	opt *config.Config

	raft            *raft.Raft
	scriptStorage   *scriptStorage
	bucketStorage   *bucketStorage
	postBucketQueue chan *events.Bucket
	quitFlusherChan chan struct{}
}

func newStore(opt *config.Config) (*defaultStore, error) {

	store := &defaultStore{
		scriptStorage: &scriptStorage{
			m: make(map[string][]byte),
		},
		bucketStorage: &bucketStorage{
			es: &eventStorage{
				m: make(map[string]*events.Bucket),
			},
			rs: &ruleStorage{
				m: make(map[string]*rules.Rule),
			},
		},
		opt:             opt,
		quitFlusherChan: make(chan struct{}),
		postBucketQueue: make(chan *events.Bucket, 100),
	}
	store.open()
	return store, nil
}

func (d *defaultStore) poster() {
	for {
		select {
		case rb := <-d.postBucketQueue:
			glog.Infof("received bucket %+v", rb)
			go func(rb *events.Bucket) {
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

	ticker := time.NewTicker(time.Second)
loop:
	for {
		select {
		case <-ticker.C:
			glog.Info("rule flusher ==> ticker called")
			for ruleID, rb := range d.bucketStorage.es.clone() {
				glog.Info("rule flusher ==> ", ruleID, rb.CanFlush())
				if rb.CanFlush() {
					if d.opt.DisablePostHook {
						glog.Infof("post bucket to hook %v ", rb)
						go func() {
							d.postBucketQueue <- rb
						}()
					}

					err := d.flushBucket(ruleID)
					if err != nil {
						glog.Errorf("error flushing %v", err)
					}
				}
			}

		case <-d.quitFlusherChan:
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

func (d *defaultStore) matchAndStash(event *events.Event) error {
	for _, rule := range d.getRules() {
		go d.match(rule, event)
	}
	return nil
}

func (d *defaultStore) match(rule *rules.Rule, event *events.Event) error {
	if util.PatternMatch(event.EventType, rule.EventTypes) {
		go d.stash(rule.ID, event)
	}
	return nil
}

func (d *defaultStore) stash(ruleID string, event *events.Event) error {
	return d.applyCMD(&command{
		Op:     "stash",
		RuleID: ruleID,
		Event:  event,
	})
}

func (d *defaultStore) addRule(rule *rules.Rule) error {

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

func (d *defaultStore) updateRule(rule *rules.Rule) error {
	return d.applyCMD(&command{
		Op:   "update_rule",
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

func (d *defaultStore) removeRule(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "remove_rule",
		RuleID: ruleID,
	})
}

func (d *defaultStore) flushBucket(ruleID string) error {
	return d.applyCMD(&command{
		Op:     "flush_bucket",
		RuleID: ruleID,
	})
}

func (d *defaultStore) getScripts() []string {
	return d.scriptStorage.getScripts()
}

func (d *defaultStore) getScript(id string) []byte {
	return d.scriptStorage.getScript(id)
}

func (d *defaultStore) getRules() []*rules.Rule {
	return d.bucketStorage.rs.getRules()
}

func (d *defaultStore) getRule(ruleID string) *rules.Rule {
	return d.bucketStorage.rs.getRule(ruleID)
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
