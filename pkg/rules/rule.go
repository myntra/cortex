package rules

import (
	"fmt"

	"github.com/myntra/cortex/pkg/matcher"
)

//go:generate msgp

// Rule is the array of related service events
type Rule struct {
	Title             string   `json:"title"`
	ID                string   `json:"id"`
	ScriptID          string   `json:"script_id"`           // javascript script which is called before hookEndPoint is called.
	HookEndpoint      string   `json:"hook_endpoint"`       // endpoint which accepts a POST json objects
	HookRetry         int      `json:"hook_retry"`          // number of retries while attempting to post
	EventTypePatterns []string `json:"event_type_patterns"` // a list of event types to look for. wildcards are allowed.
	Dwell             uint64   `json:"dwell"`               // dwell duration in milliseconds for events to arrive
	DwellDeadline     uint64   `json:"dwell_deadline"`      // dwell duration threshold after which arriving events expand the dwell window
	MaxDwell          uint64   `json:"max_dwell"`           // maximum dwell duration including expansion
	Regexes           []string `json:"regexes,omitempty"`   // generated regex string array from event types
	Disabled          bool     `json:"disabled,omitempty"`  // if the rule is disabled
}

// Validate rule data
func (r *Rule) Validate() error {

	for _, pattern := range r.EventTypePatterns {
		m, err := matcher.New(pattern)
		if err != nil {
			return fmt.Errorf("invalid event type pattern %v,  err: %v", pattern, err)
		}

		r.Regexes = append(r.Regexes, m.GetRegexString())
	}

	return nil
}

// HasMatching checks whether the rule has a matching event type pattern
func (r *Rule) HasMatching(eventType string) bool {
	if r.Disabled {
		return false
	}
	for _, regexStr := range r.Regexes {
		m := matcher.NewCompile(regexStr)
		if m.HasMatches(eventType) {
			return true
		}
	}
	return false
}

// PublicRule is used to create, update a request and is returned as a response
type PublicRule struct {
	Title             string   `json:"title"`
	ID                string   `json:"id"`
	ScriptID          string   `json:"script_id"`           // javascript script which is called before hookEndPoint is called.
	HookEndpoint      string   `json:"hook_endpoint"`       // endpoint which accepts a POST json objects
	HookRetry         int      `json:"hook_retry"`          // number of retries while attempting to post
	EventTypePatterns []string `json:"event_type_patterns"` // a list of event types to look for. wildcards are allowed.
	Dwell             uint64   `json:"dwell"`               // dwell duration in milliseconds for events to arrive
	DwellDeadline     uint64   `json:"dwell_deadline"`      // dwell duration threshold after which arriving events expand the dwell window
	MaxDwell          uint64   `json:"max_dwell"`           // maximum dwell duration including expansion
	Disabled          bool     `json:"disabled,omitempty"`  // if the rule is disabled
}

// NewFromPublic creates a rule from a public rule
func NewFromPublic(r *PublicRule) *Rule {
	return &Rule{
		Title:             r.Title,
		ID:                r.ID,
		ScriptID:          r.ScriptID,
		HookEndpoint:      r.HookEndpoint,
		HookRetry:         r.HookRetry,
		EventTypePatterns: r.EventTypePatterns,
		Dwell:             r.Dwell,
		DwellDeadline:     r.DwellDeadline,
		MaxDwell:          r.MaxDwell,
		Disabled:          r.Disabled,
	}
}

// NewFromPrivate creates public rule from a private rule
func NewFromPrivate(r *Rule) *PublicRule {
	return &PublicRule{
		Title:             r.Title,
		ID:                r.ID,
		ScriptID:          r.ScriptID,
		HookEndpoint:      r.HookEndpoint,
		HookRetry:         r.HookRetry,
		EventTypePatterns: r.EventTypePatterns,
		Dwell:             r.Dwell,
		DwellDeadline:     r.DwellDeadline,
		MaxDwell:          r.MaxDwell,
		Disabled:          r.Disabled,
	}
}
