package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/myntra/cortex/pkg/rules"
	"github.com/sethgrid/pester"
)

// NewBucket creates a new Bucket
func NewBucket(rule rules.Rule) *Bucket {
	return &Bucket{
		flushWait: rule.WaitWindow,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		Rule:      rule,
	}
}

// Bucket contains the rule for a collection of events and the events
type Bucket struct {
	Rule      rules.Rule `json:"rule"`
	Events    []*Event   `json:"events"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedAt time.Time  `json:"created_at"`
	flushWait uint64
}

// AddEvent to the bucket
func (rb *Bucket) AddEvent(event *Event) {
	rb.Events = append(rb.Events, event)
	rb.updateWaitWindow()
}

// Post posts rulebucket to the configured hook endpoint
func (rb *Bucket) Post() error {

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

// GetWaitWindowDuration converts wait_window(ms) to time.Duration
func (rb *Bucket) getWaitWindowDuration() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.WaitWindow)
}

// getWaitWindowThresholdDuration converts wait_window_threshold(ms) to time.Duration
func (rb *Bucket) getWaitWindowThresholdDuration() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.WaitWindowThreshold)
}

// getMaxWaitWindow converts max_wait_window(ms) to time.Duration
func (rb *Bucket) getMaxWaitWindow() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.MaxWaitWindow)
}

// CanFlush returns if the bucket can be evicted from the db
func (rb *Bucket) CanFlush() bool {
	return time.Since(rb.UpdatedAt) >= time.Millisecond*time.Duration(rb.flushWait)
}

// UpdateWaitWindow updates flush waiting duration
func (rb *Bucket) updateWaitWindow() {
	timeSinceCreated := time.Since(rb.CreatedAt)

	if timeSinceCreated >= rb.getMaxWaitWindow() {
		rb.UpdatedAt = time.Now()
		return
	}

	timeSinceLastEventAdded := time.Since(rb.UpdatedAt)

	if timeSinceLastEventAdded >= rb.getWaitWindowThresholdDuration() {
		rb.flushWait = rb.flushWait + rb.Rule.WaitWindow
	}

	rb.UpdatedAt = time.Now()
}
