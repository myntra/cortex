package store

import (
	"github.com/myntra/cortex/pkg/executions"
	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/rules"
)

// MessageType of the data entry
type MessageType uint8

const (
	// RuleType denotes the rules.Rule type
	RuleType MessageType = 0
	// ScriptType denotes the script type
	ScriptType = 1
	// RecordType denotes the executions.Record type
	RecordType = 2
)

//go:generate msgp

// Messages store entries to the underlying storage
type Messages struct {
	Rules   map[string]*rules.Rule        `json:"rules"`
	Records map[string]*executions.Record `json:"records"`
	Scripts map[string]*js.Script         `json:"script"`
}
