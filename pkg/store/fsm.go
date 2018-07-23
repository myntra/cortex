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
	case "add_script":
		return f.applyAddScript(c.ScriptID, c.Script)
	case "update_script":
		return f.applyUpdateScript(c.ScriptID, c.Script)
	case "remove_script":
		return f.applyRemoveScript(c.ScriptID)
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
	ruleBuckets := f.eventStorage.clone()
	return &fsmSnapShot{data: &db{ruleBuckets: ruleBuckets, scripts: make(map[string][]byte)}}, nil
}

func (f *fsm) applyAddScript(id string, script []byte) interface{} {
	return f.scriptStorage.addScript(id, script)
}

func (f *fsm) applyUpdateScript(id string, script []byte) interface{} {
	return f.scriptStorage.updateScript(id, script)
}

func (f *fsm) applyRemoveScript(id string) interface{} {
	return f.scriptStorage.removeScript(id)
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var data db

	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		return err
	}

	f.eventStorage.restore(data.ruleBuckets)
	f.scriptStorage.restore(data.scripts)

	return nil
}
