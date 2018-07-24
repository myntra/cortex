package store

import (
	"sync"

	"github.com/myntra/cortex/pkg/executions"
)

type executionStorage struct {
	mu sync.Mutex
	m  map[string]*executions.Record
}

func (e *executionStorage) add(r *executions.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.m[r.ID] = r
	return nil
}

func (e *executionStorage) remove(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.m, id)
	return nil
}

func (e *executionStorage) getRecords(ruleID string) []*executions.Record {
	e.mu.Lock()
	defer e.mu.Unlock()

	var exs []*executions.Record
	for _, record := range e.m {
		if record.Bucket.Rule.ID == ruleID {
			exs = append(exs, record)
		}
	}
	return exs
}

func (e *executionStorage) getRecordsCount(ruleID string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	count := 0
	for _, record := range e.m {
		if record.Bucket.Rule.ID == ruleID {
			count++
		}
	}

	return count
}

func (e *executionStorage) getTotalRecordsCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.m)
}

func (e *executionStorage) flush(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.m, id)
}

func (e *executionStorage) clone() map[string]*executions.Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	clone := make(map[string]*executions.Record)
	for k, v := range e.m {
		clone[k] = v
	}
	return clone
}

func (e *executionStorage) restore(m map[string]*executions.Record) {
	e.m = m
}
