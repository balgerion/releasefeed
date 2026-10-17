package internal

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type State struct {
	mu    sync.Mutex
	path  string
	feeds map[string]map[string]time.Time
}

func NewState(path string) (*State, error) {
	s := &State{path: path, feeds: make(map[string]map[string]time.Time)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.feeds); err != nil {
		log.Printf("state: cannot parse %s, starting fresh: %v", path, err)
		s.feeds = make(map[string]map[string]time.Time)
	}
	return s, nil
}

func (s *State) Known(feed string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.feeds[feed]
	return ok
}

func (s *State) FirstSeen(feed, id string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.feeds[feed][id]
	return t, ok
}

func (s *State) Replace(feed string, seen map[string]time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feeds[feed] = seen
	data, err := json.Marshal(s.feeds)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
