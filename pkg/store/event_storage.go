package store

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/myntra/cortex/pkg/event"
)

type eventStorage struct {
	mu              sync.Mutex
	m               map[string]*event.RuleBucket // [ruleID]
	flusherChan     chan string
	quitFlusherChan chan struct{}
}

func isMatch(eventType, pattern string) bool {
	return eventType == pattern
}

func flush(ruleID string, rb *event.RuleBucket, flusherChan chan string) {

	flusherStartedAt := time.Now()
	tickerStartedAt := time.Now()
	waitWindowDuration := time.Duration(rb.Rule.WaitWindow) * time.Millisecond
	ticker := time.NewTicker(waitWindowDuration)

loop:
	for {
		select {
		case <-ticker.C:
			flusherChan <- ruleID
			break loop
		case <-rb.Touch:

			tickerRanDuration := time.Since(tickerStartedAt)
			// if ticker has run less than the wait window threshold duration, do nothing
			if tickerRanDuration < time.Millisecond*time.Duration(rb.Rule.WaitWindowThreshold) {
				continue
			}

			flusherRunDuration := time.Since(flusherStartedAt)
			// if flusher running duration is greater than max waiting window, do nothing
			if flusherRunDuration >= time.Millisecond*time.Duration(rb.Rule.MaxWaitWindow) {
				continue
			}

			// else reset the ticker
			ticker = time.NewTicker(waitWindowDuration)
			tickerStartedAt = time.Now()
		}
	}
}

func (e *eventStorage) stash(event *event.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// TODO: efficient regex matching for rule bucket

	for ruleID, ruleBucket := range e.m {
		for _, eventTypePattern := range ruleBucket.Rule.EventTypes {
			// add event to all matching rule buckets
			if isMatch(event.EventType, eventTypePattern) {
				// has rule bucket been initialized ?
				if len(e.m[ruleID].Bucket) == 0 {
					// start a flusher for this rule
					if e.m[ruleID].Touch == nil {
						e.m[ruleID].Touch = make(chan struct{})
						// flusher routine
						go flush(ruleID, e.m[ruleID], e.flusherChan)
					}
				}
				// dedup, reschedule flusher(sliding wait window), frequency count
				dup := false
				for _, existingEvent := range e.m[ruleID].Bucket {
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
					continue
				}

				e.m[ruleID].Bucket = append(e.m[ruleID].Bucket, event)
			}
		}
	}

	return nil
}

func (e *eventStorage) getRule(ruleID string) *event.Rule {
	e.mu.Lock()
	defer e.mu.Unlock()
	var rb *event.RuleBucket
	var ok bool
	if rb, ok = e.m[ruleID]; !ok {
		return nil
	}
	return rb.Rule
}

func (e *eventStorage) getRuleBucket(ruleID string) *event.RuleBucket {
	e.mu.Lock()
	defer e.mu.Unlock()
	var rb *event.RuleBucket
	var ok bool
	if rb, ok = e.m[ruleID]; !ok {
		return nil
	}
	return rb
}

func (e *eventStorage) flushRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.m[ruleID].Bucket = nil
	return nil
}

func (e *eventStorage) addRule(rule *event.Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.m[rule.ID]; ok {
		return fmt.Errorf("rule id already exists")
	}

	ruleBucket := &event.RuleBucket{
		Rule: rule,
	}

	e.m[rule.ID] = ruleBucket

	return nil
}

func (e *eventStorage) removeRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.m[ruleID]; !ok {
		return fmt.Errorf("rule id does not exist")
	}

	delete(e.m, ruleID)

	return nil
}

func (e *eventStorage) getRules() []*event.Rule {
	e.mu.Lock()
	defer e.mu.Unlock()
	var rules []*event.Rule
	for _, v := range e.m {
		rules = append(rules, v.Rule)
	}
	return rules
}

func (e *eventStorage) clone() map[string]*event.RuleBucket {
	e.mu.Lock()
	defer e.mu.Unlock()
	clone := make(map[string]*event.RuleBucket)
	for k, v := range e.m {
		clone[k] = v
	}
	return clone
}

func (e *eventStorage) restore(m map[string]*event.RuleBucket) {
	e.m = m
}
