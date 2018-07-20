package store

import (
	"bytes"
	"sync"
	"time"

	"github.com/myntra/aggo/pkg/event"
)

type storage struct {
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

func (s *storage) stash(event *event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: efficient regex matching for rule bucket

	for ruleID, ruleBucket := range s.m {
		for _, eventTypePattern := range ruleBucket.Rule.EventTypes {
			// add event to all matching rule buckets
			if isMatch(event.EventType, eventTypePattern) {
				// has rule bucket been initialized ?
				if len(s.m[ruleID].Bucket) == 0 {
					// start a flusher for this rule
					if s.m[ruleID].Touch == nil {
						s.m[ruleID].Touch = make(chan struct{})
						// flusher routine
						go flush(ruleID, s.m[ruleID], s.flusherChan)
					}
				}
				// dedup, reschedule flusher(sliding wait window), frequency count
				dup := false
				for _, existingEvent := range s.m[ruleID].Bucket {
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

				s.m[ruleID].Bucket = append(s.m[ruleID].Bucket, event)
			}
		}
	}
}

func (s *storage) getRule(ruleID string) *event.RuleBucket {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rb *event.RuleBucket
	var ok bool
	if rb, ok = s.m[ruleID]; !ok {
		return nil
	}
	return rb
}

func (s *storage) flushRule(ruleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[ruleID].Bucket = nil
}

func (s *storage) addRule(rule *event.Rule) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[rule.ID]; ok {
		// rule id already exists
		return false
	}

	ruleBucket := &event.RuleBucket{
		Rule: rule,
	}

	s.m[rule.ID] = ruleBucket

	return true
}

func (s *storage) removeRule(ruleID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[ruleID]; !ok {
		// rule id does not exist
		return false
	}

	delete(s.m, ruleID)

	return true
}

func (s *storage) getRules() []*event.Rule {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rules []*event.Rule
	for _, v := range s.m {
		rules = append(rules, v.Rule)
	}
	return rules
}

func (s *storage) clone() map[string]*event.RuleBucket {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := make(map[string]*event.RuleBucket)
	for k, v := range s.m {
		clone[k] = v
	}
	return clone
}

func (s *storage) restore(m map[string]*event.RuleBucket) {
	s.m = m
}
