package store

import (
	"encoding/json"
	"fmt"
	"io"

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
	records := f.executionStorage.clone()

	return &fsmSnapShot{
		persisters: f.persisters,
		messages: &Messages{
			Rules:   rules,
			Scripts: scripts,
			Records: records,
		}}, nil
}

type restorer func(messages *Messages, reader *msgp.Reader) error

func (f *fsm) Restore(rc io.ReadCloser) error {
	glog.Info("restore <=")
	defer rc.Close()

	// body, _ := ioutil.ReadAll(rc)
	// glog.Infoln(string(body))

	messages := &Messages{
		Rules:   make(map[string]*rules.Rule),
		Scripts: make(map[string]*js.Script),
		Records: make(map[string]*executions.Record),
	}

	msgpReader := msgp.NewReader(rc)

	msgType := make([]byte, 1)
	for {
		// Read the message type
		_, err := msgpReader.Read(msgType)
		if err == io.EOF {
			glog.Infof("err => %v", err)
			break
		} else if err != nil {
			glog.Error(err)
			return err
		}

		// Decode
		msg := MessageType(msgType[0])
		glog.Infof("resotre, messageType %+v\n", msg)
		if fn := f.restorers[msg]; fn != nil {
			if err := fn(messages, msgpReader); err != nil {
				glog.Error(err)
				return err
			}
		} else {
			glog.Error(fmt.Errorf("Unrecognized msg type %d", msg))
			return fmt.Errorf("Unrecognized msg type %d", msg)
		}

	}

	f.bucketStorage.rs.restore(messages.Rules)
	f.scriptStorage.restore(messages.Scripts)
	f.executionStorage.restore(messages.Records)

	return nil
}

func restoreRules(messages *Messages, reader *msgp.Reader) error {
	var rule rules.Rule
	err := rule.DecodeMsg(reader)
	if err != nil {
		glog.Error(err)
		return err
	}

	glog.Infof("restoreRules %+v\n", rule)

	if &rule == nil {
		return fmt.Errorf("restored rule nil")
	}

	rulePtr := &rule
	err = rulePtr.Validate()
	if err != nil {
		return err
	}

	messages.Rules[rule.ID] = rulePtr
	return nil
}

func restoreScripts(messages *Messages, reader *msgp.Reader) error {
	var script js.Script
	err := script.DecodeMsg(reader)
	if err != nil {
		glog.Error(err)
		return err
	}

	glog.Infof("restoreScripts %+v\n", script)

	if &script == nil {
		return fmt.Errorf("restored script nil")
	}

	messages.Scripts[script.ID] = &script
	return nil
}

func restoreRecords(messages *Messages, reader *msgp.Reader) error {
	var record executions.Record
	err := record.DecodeMsg(reader)
	if err != nil {
		glog.Error(err)
		return err
	}

	glog.Infof("restoreRecords %+v\n", record)

	if &record == nil {
		return fmt.Errorf("restored record nil")
	}

	messages.Records[record.ID] = &record

	return nil
}
