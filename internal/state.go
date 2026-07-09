package internal

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

const maxStateSize = 1000

type State struct {
	mu   sync.Mutex
	path string
	seen map[string]time.Time
}

func NewState(path string) (*State, error) {
	s := &State{path: path, seen: make(map[string]time.Time)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.seen); err != nil {
		log.Printf("state: cannot parse %s, starting fresh: %v", path, err)
		s.seen = make(map[string]time.Time)
	}
	return s, nil
}

func (s *State) Empty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.seen) == 0
}

func (s *State) FirstSeen(id string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.seen[id]
	return t, ok
}

func (s *State) Mark(id string, published time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[id]; ok {
		return nil
	}
	s.seen[id] = published
	if len(s.seen) > maxStateSize {
		oldest := ""
		var oldestT time.Time
		for k, v := range s.seen {
			if oldest == "" || v.Before(oldestT) {
				oldest, oldestT = k, v
			}
		}
		delete(s.seen, oldest)
	}
	data, err := json.Marshal(s.seen)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
