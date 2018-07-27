package store

import (
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/executions"
	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/rules"
)

//go:generate msgp

// Command is the container for a raft command
type Command struct {
	Op       string             `json:"op"` // stash or evict
	Rule     *rules.Rule        `json:"rule,omitempty"`
	RuleID   string             `json:"ruleID,omitempty"`
	Event    *events.Event      `json:"event,omitempty"`
	ScriptID string             `json:"script_id,omitempty"`
	Script   *js.Script         `json:"script,omitempty"`
	Record   *executions.Record `json:"record,omitempty"`
	RecordID string             `json:"record_id,omitempty"`
}
