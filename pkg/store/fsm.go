package store

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/myntra/aggo/pkg/event"
)

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
