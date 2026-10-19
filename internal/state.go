package internal

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type State struct {
	mu          sync.Mutex
	path        string
	alertedPath string
	feeds       map[string]map[string]time.Time
	alerted     map[string]map[string]bool
}

func NewState(path, alertedPath string) (*State, error) {
	s := &State{path: path, alertedPath: alertedPath, feeds: make(map[string]map[string]time.Time), alerted: make(map[string]map[string]bool)}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &s.feeds); err != nil {
			log.Printf("state: cannot parse %s, starting fresh: %v", path, err)
			s.feeds = make(map[string]map[string]time.Time)
		}
	}
	data, err = os.ReadFile(alertedPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &s.alerted); err != nil {
			log.Printf("state: cannot parse %s, starting fresh: %v", alertedPath, err)
			s.alerted = make(map[string]map[string]bool)
		}
		return s, nil
	}
	for feed, ids := range s.feeds {
		s.alerted[feed] = make(map[string]bool, len(ids))
		for id := range ids {
			s.alerted[feed][id] = true
		}
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

func (s *State) Alerted(feed, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alerted[feed][id]
}

func (s *State) Replace(feed string, seen map[string]time.Time, alerted map[string]bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feeds[feed] = seen
	s.alerted[feed] = alerted
	data, err := json.Marshal(s.feeds)
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}
	data, err = json.Marshal(s.alerted)
	if err != nil {
		return err
	}
	return os.WriteFile(s.alertedPath, data, 0644)
}
