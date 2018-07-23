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

	return f.eventStorage.stash(event)
}

func (f *fsm) applyAddRule(rule *event.Rule) interface{} {
	return f.eventStorage.addRule(rule)
}

func (f *fsm) applyRemoveRule(ruleID string) interface{} {
	return f.eventStorage.removeRule(ruleID)
}

func (f *fsm) applyFlushRule(ruleID string) interface{} {
	return f.eventStorage.flushRule(ruleID)
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	db := f.eventStorage.clone()
	return &fsmSnapShot{db: db}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	db := make(map[string]*event.RuleBucket)

	if err := json.NewDecoder(rc).Decode(&db); err != nil {
		return err
	}

	f.eventStorage.restore(db)

	return nil
}
