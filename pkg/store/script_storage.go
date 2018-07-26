package store

import (
	"fmt"
	"sync"

	"github.com/myntra/cortex/pkg/js"
)

type scriptStorage struct {
	mu sync.RWMutex
	m  map[string]*js.Script
}

func (s *scriptStorage) addScript(script *js.Script) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[script.ID]; ok {
		return fmt.Errorf("script name already exists. script name must be unique")
	}

	s.m[script.ID] = script

	return nil
}

func (s *scriptStorage) updateScript(script *js.Script) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[script.ID]; !ok {
		return fmt.Errorf("script name not found. can't update")
	}

	s.m[script.ID] = script
	return nil
}

func (s *scriptStorage) removeScript(id string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[id]; !ok {
		return fmt.Errorf("script name not found. can't remove")
	}

	delete(s.m, id)

	return nil
}

func (s *scriptStorage) getScript(id string) *js.Script {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[id]; !ok {
		return nil
	}

	return s.m[id]
}

func (s *scriptStorage) getScripts() []string {

	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []string

	for k := range s.m {
		ids = append(ids, k)
	}

	return ids
}

func (s *scriptStorage) clone() map[string]*js.Script {
	s.mu.Lock()
	defer s.mu.Unlock()
	scripts := make(map[string]*js.Script)
	for k, v := range s.m {
		scripts[k] = v
	}
	return scripts
}

func (s *scriptStorage) restore(m map[string]*js.Script) {
	s.m = m
}
