package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sethgrid/pester"
)

// RuleBucket contains the rule for a collection of events and the events
type RuleBucket struct {
	Rule   *Rule         `json:"rule"`
	Bucket []*Event      `json:"events"`
	Touch  chan struct{} `json:"-"` // a channel used to expand the waitwindow
}

// Post posts rulebucket to the configured hook endpoint
func (rb *RuleBucket) Post() error {

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(rb)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", rb.Rule.HookEndpoint, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "application/json")

	client := pester.New()
	client.MaxRetries = rb.Rule.HookRetry

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return fmt.Errorf("invalid status code return from %v endpoint", rb.Rule.HookEndpoint)
	}

	return nil
}

// Rule is the array of related service events
type Rule struct {
	ID                  string   `json:"id"`
	ScriptID            string   `json:"script_id"`             // javascript script which is called before hookEndPoint is called.
	HookEndpoint        string   `json:"hook_endpoint"`         // endpoint which accepts a POST json objects
	HookRetry           int      `json:"hook_retry"`            // number of retries while attempting to post
	EventTypes          []string `json:"event_types"`           // a list of event types to look for. a regex pattern is also allowed
	WaitWindow          uint64   `json:"wait_window"`           // wait duration in milliseconds for events to arrive
	WaitWindowThreshold uint64   `json:"wait_window_threshold"` // wait duration threshold after which arriving events expand the wait window
	MaxWaitWindow       uint64   `json:"max_wait_window"`       // maximum wait duration until which WaitWindow can expand
}
