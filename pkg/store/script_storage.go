package store

import (
	"fmt"
	"sync"
)

type scriptStorage struct {
	mu sync.Mutex
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
