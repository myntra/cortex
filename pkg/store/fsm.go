package store

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/executions"
	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/rules"
	"github.com/tinylib/msgp/msgp"
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
		return f.applyAddScript(c.Script)
	case "update_script":
		return f.applyUpdateScript(c.Script)
	case "remove_script":
		return f.applyRemoveScript(c.ScriptID)
	case "add_record":
		return f.applyAddRecord(c.Record)
	case "remove_record":
		return f.applyRemoveRecord(c.RecordID)
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

func (f *fsm) applyAddScript(script *js.Script) interface{} {
	return f.scriptStorage.addScript(script)
}

func (f *fsm) applyUpdateScript(script *js.Script) interface{} {
	return f.scriptStorage.updateScript(script)
}

func (f *fsm) applyRemoveScript(id string) interface{} {
	return f.scriptStorage.removeScript(id)
}

func (f *fsm) applyAddRecord(r *executions.Record) interface{} {
	return f.executionStorage.add(r)
}

func (f *fsm) applyRemoveRecord(id string) interface{} {
	return f.executionStorage.remove(id)
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	glog.Info("snapshot =>")

	rules := f.bucketStorage.rs.clone()
	scripts := f.scriptStorage.clone()
	history := f.executionStorage.clone()
	return &fsmSnapShot{
		msgs: &Messages{
			Rules:   rules,
			Scripts: scripts,
			Records: history,
		}}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	glog.Info("restore <=")
	defer rc.Close()
	var msgs Messages

	bts, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}

	left, err := msgs.UnmarshalMsg(bts)

	if len(left) > 0 {
		return fmt.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		return err
	}
	if len(left) > 0 {
		return fmt.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}

	f.bucketStorage.rs.restore(msgs.Rules)
	f.scriptStorage.restore(msgs.Scripts)
	f.executionStorage.restore(msgs.Records)

	return nil
}
