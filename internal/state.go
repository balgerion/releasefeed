package internal

import (
	"encoding/json"
	"os"
	"sync"
)

type State struct {
	mu   sync.Mutex
	path string
	seen map[string]bool
}

func NewState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := &State{path: path, seen: make(map[string]bool)}
	return s, json.Unmarshal(data, &s.seen)
}

func (s *State) Seen(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[id]
}

func (s *State) Mark(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[id] = true
	data, err := json.Marshal(s.seen)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}