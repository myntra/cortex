package store

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"
)

type eventStorage struct {
	mu sync.RWMutex
	m  map[string]*events.Bucket // [ruleID]
}

func (e *eventStorage) stash(rule rules.Rule, event *events.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	glog.Infof("stash event ==>  %+v", event)
	ruleID := rule.ID
	if _, ok := e.m[ruleID]; !ok {
		bucket := events.NewBucket(rule)
		bucket.Events = append(bucket.Events, event)
		e.m[ruleID] = bucket
		return nil
	}

	// dedup, reschedule flusher(sliding wait window), frequency count
	dup := false
	for _, existingEvent := range e.m[ruleID].Events {
		// check if source is equal
		if existingEvent.Source == event.Source {
			// check if equal hash
			if bytes.Equal(existingEvent.Hash(), event.Hash()) {
				dup = true
			}
		}
	}
	// is a duplicate event, skip appending event to bucket
	if dup {
		return nil
	}
	// update event
	e.m[ruleID].AddEvent(event)

	return nil
}

func (e *eventStorage) flushLock(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.m[ruleID]; !ok {
		return fmt.Errorf("bucket with id %v not found", ruleID)
	}

	// update flush lock
	bucket := e.m[ruleID]
	bucket.FlushLock = true
	e.m[ruleID] = bucket

	return nil
}

func (e *eventStorage) flushBucket(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.m[ruleID]; !ok {
		return fmt.Errorf("bucket with id %v not found", ruleID)
	}

	delete(e.m, ruleID)
	return nil
}

func (e *eventStorage) bucketExists(ruleID string) bool {
	_, ok := e.m[ruleID]
	return ok
}

func (e *eventStorage) getBucket(ruleID string) *events.Bucket {
	e.mu.Lock()
	defer e.mu.Unlock()
	var rb *events.Bucket
	var ok bool
	if rb, ok = e.m[ruleID]; !ok {
		return nil
	}
	return rb
}

func (e *eventStorage) clone() map[string]*events.Bucket {
	e.mu.Lock()
	defer e.mu.Unlock()
	clone := make(map[string]*events.Bucket)
	for k, v := range e.m {
		clone[k] = v
	}
	return clone
}

func (e *eventStorage) restore(m map[string]*events.Bucket) {
	e.m = m
}
