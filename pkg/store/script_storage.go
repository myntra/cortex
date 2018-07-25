package store

import (
	"fmt"
	"sync"
)

type scriptStorage struct {
	mu sync.RWMutex
	m  map[string][]byte
}

func (s *scriptStorage) addScript(id string, script []byte) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[id]; ok {
		return fmt.Errorf("script name already exists. script name must be unique")
	}

	s.m[id] = script

	return nil
}

func (s *scriptStorage) updateScript(id string, script []byte) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[id]; !ok {
		return fmt.Errorf("script name not found. can't update")
	}

	s.m[id] = script
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

func (s *scriptStorage) getScript(id string) []byte {

	s.mu.Lock()
	defer s.mu.Unlock()

	var b []byte
	if _, ok := s.m[id]; !ok {
		return b
	}
	b = s.m[id]
	return b
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

func (s *scriptStorage) clone() map[string][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	var scripts map[string][]byte
	for k, v := range s.m {
		scripts[k] = v
	}
	return scripts
}

func (s *scriptStorage) restore(m map[string][]byte) {
	s.m = m
}
