package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/myntra/cortex/pkg/rules"
	"github.com/sethgrid/pester"
)

// NewBucket creates a new Bucket
func NewBucket(rule rules.Rule) *Bucket {
	return &Bucket{
		flushWait:    rule.Dwell,
		dwellResetAt: time.Now(),
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
		Rule:         rule,
	}
}

//go:generate msgp

// Bucket contains the rule for a collection of events and the events
type Bucket struct {
	Rule         rules.Rule `json:"rule"`
	Events       []*Event   `json:"events"`
	FlushLock    bool       `json:"flush_lock"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedAt    time.Time  `json:"created_at"`
	dwellResetAt time.Time
	flushWait    uint64
}

// AddEvent to the bucket
func (rb *Bucket) AddEvent(event *Event) {
	glog.Info("add event ==>  ", event)
	rb.Events = append(rb.Events, event)
	rb.updateDwell()
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

// GetDwellDuration converts dwell(ms) to time.Duration
func (rb *Bucket) getDwellDuration() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.Dwell)
}

// getDwellDeadlineDuration converts dwell_deadline(ms) to time.Duration
func (rb *Bucket) getDwellDeadlineDuration() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.DwellDeadline)
}

// getMaxDwell converts max_dwell(ms) to time.Duration
func (rb *Bucket) getMaxDwell() time.Duration {
	return time.Millisecond * time.Duration(rb.Rule.MaxDwell)
}

// CanFlush returns if the bucket can be evicted from the db
func (rb *Bucket) CanFlush() bool {
	return time.Since(rb.CreatedAt) >= time.Millisecond*time.Duration(rb.flushWait)
}

// CanFlushIn returns time left for flush
func (rb *Bucket) CanFlushIn() time.Duration {
	return time.Millisecond*time.Duration(rb.flushWait) - time.Since(rb.CreatedAt)
}

// UpdateDwell updates flush waiting duration
func (rb *Bucket) updateDwell() {
	glog.Infof("updateDwell ")
	timeSinceDwellReset := time.Since(rb.dwellResetAt)

	glog.Infof("updateDwell %v %v %v %v", timeSinceDwellReset, rb.getDwellDuration(), rb.getMaxDwell(), rb.getDwellDeadlineDuration())
	if (timeSinceDwellReset + rb.getDwellDuration()) >= rb.getMaxDwell() {
		rb.UpdatedAt = time.Now()
		return
	}

	if timeSinceDwellReset >= rb.getDwellDeadlineDuration() {
		glog.Info("updateDwell flushwait + dwell")
		rb.dwellResetAt = time.Now()
		rb.flushWait = rb.flushWait + rb.Rule.Dwell
	}

	rb.UpdatedAt = time.Now()
}
