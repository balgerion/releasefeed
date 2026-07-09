package internal

import (
	"encoding/json"
	"os"
	"sync"
)

const maxStateSize = 1000

type State struct {
	mu   sync.Mutex
	path string
	seen map[string]bool
	keys []string
}

func NewState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		data = []byte("[]")
	} else if err != nil {
		return nil, err
	}
	s := &State{path: path, seen: make(map[string]bool)}
	if err := json.Unmarshal(data, &s.keys); err != nil {
		return nil, err
	}
	for _, k := range s.keys {
		s.seen[k] = true
	}
	return s, nil
}

func (s *State) Empty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.keys) == 0
}

func (s *State) Seen(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[id]
}

func (s *State) Mark(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen[id] {
		return nil
	}
	s.keys = append(s.keys, id)
	s.seen[id] = true
	if len(s.keys) > maxStateSize {
		removed := s.keys[0]
		s.keys = s.keys[1:]
		delete(s.seen, removed)
	}
	data, err := json.Marshal(s.keys)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}