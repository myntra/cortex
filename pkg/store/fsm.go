package store

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"
)

type fsm defaultStore

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "stash":
		return f.applyStash(c.RuleID, c.Event)
	case "add_rule":
		return f.applyAddRule(c.Rule)
	case "update_rule":
		return f.applyUpdateRule(c.Rule)
	case "remove_rule":
		return f.applyRemoveRule(c.RuleID)
	case "flush_bucket":
		return f.applyFlushBucket(c.RuleID)
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

func (f *fsm) applyStash(ruleID string, event *events.Event) interface{} {
	return f.bucketStorage.stash(ruleID, event)
}

func (f *fsm) applyAddRule(rule *rules.Rule) interface{} {
	return f.bucketStorage.rs.addRule(rule)
}

func (f *fsm) applyUpdateRule(rule *rules.Rule) interface{} {
	return f.bucketStorage.rs.updateRule(rule)
}

func (f *fsm) applyRemoveRule(ruleID string) interface{} {
	return f.bucketStorage.rs.removeRule(ruleID)
}

func (f *fsm) applyFlushBucket(ruleID string) interface{} {
	return f.bucketStorage.es.flushBucket(ruleID)
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	buckets := f.bucketStorage.es.clone()
	rules := f.bucketStorage.rs.clone()
	scripts := f.scriptStorage.clone()

	return &fsmSnapShot{
		data: &db{
			buckets: buckets,
			rules:   rules,
			scripts: scripts,
		}}, nil
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

	f.bucketStorage.es.restore(data.buckets)
	f.bucketStorage.rs.restore(data.rules)
	f.scriptStorage.restore(data.scripts)

	return nil
}
