package store

import (
	"fmt"
	"sync"

	"github.com/myntra/cortex/pkg/rules"
)

type ruleStorage struct {
	mu sync.Mutex
	m  map[string]*rules.Rule // [ruleID]
}

func (r *ruleStorage) getRule(ruleID string) *rules.Rule {
	r.mu.Lock()
	defer r.mu.Unlock()
	var rule *rules.Rule
	var ok bool
	if rule, ok = r.m[ruleID]; !ok {
		return nil
	}
	return rule
}

func (r *ruleStorage) addRule(rule *rules.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.m[rule.ID]; ok {
		return fmt.Errorf("rule id already exists")
	}

	r.m[rule.ID] = rule
	return nil
}

func (r *ruleStorage) updateRule(rule *rules.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.m[rule.ID]; !ok {
		return fmt.Errorf("rule id does not exist")
	}

	r.m[rule.ID] = rule

	return nil
}

func (r *ruleStorage) removeRule(ruleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.m[ruleID]; !ok {
		return fmt.Errorf("rule id does not exist")
	}

	delete(r.m, ruleID)

	return nil
}

func (r *ruleStorage) getRules() []*rules.Rule {
	r.mu.Lock()
	defer r.mu.Unlock()
	var rules []*rules.Rule
	for _, rule := range r.m {
		rules = append(rules, rule)
	}
	return rules
}

func (r *ruleStorage) clone() map[string]*rules.Rule {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := make(map[string]*rules.Rule)
	for k, v := range r.m {
		clone[k] = v
	}
	return clone
}

func (r *ruleStorage) restore(m map[string]*rules.Rule) {
	r.m = m
}
