package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cnf/structhash"

	"github.com/fnproject/cloudevent"
	"github.com/sethgrid/pester"
)

// Event wraps cloudevent.CloudEvent
type Event struct {
	*cloudevent.CloudEvent
}

// Hash returns md5 hash string of the type
func (e *Event) Hash() (string, error) {

	data := new(cloudevent.CloudEvent)
	data.CloudEventsVersion = e.CloudEventsVersion
	data.ContentType = e.ContentType
	data.Data = e.Data
	//data.EventID = e.EventID
	data.EventType = e.EventType
	data.EventTypeVersion = e.EventTypeVersion
	data.Extensions = e.Extensions
	data.SchemaURL = e.SchemaURL
	data.Source = e.Source

	return structhash.Hash(data, 1)
}

// FromRequest parses the event from request
func FromRequest(req *http.Request) (*Event, error) {
	ce := &cloudevent.CloudEvent{}
	err := ce.FromRequest(req)
	return &Event{ce}, err
}

// RuleBucket contains the rule for a collection of events and the events
type RuleBucket struct {
	Rule   *Rule    `json:"rule"`
	Bucket []*Event `json:"events"`
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
	ID           string   `json:"id,omitempty"`
	HookEndpoint string   `json:"hook_endpoint"`
	HookRetry    int      `json:"hook_retry"`
	EventTypes   []string `json:"event_types"`
	WaitWindow   uint64   `json:"wait_window"`
}
