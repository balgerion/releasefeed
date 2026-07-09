package internal

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

type FeedStatus struct {
	Name        string    `json:"name"`
	LastSuccess time.Time `json:"last_success"`
	Entries     int       `json:"entries"`
	LastError   string    `json:"last_error,omitempty"`
	Failures    int       `json:"consecutive_failures"`
}

type StatusBoard struct {
	mu    sync.Mutex
	feeds map[string]*FeedStatus
}

func NewStatusBoard() *StatusBoard {
	return &StatusBoard{feeds: make(map[string]*FeedStatus)}
}

func (b *StatusBoard) get(name string) *FeedStatus {
	if b.feeds[name] == nil {
		b.feeds[name] = &FeedStatus{Name: name}
	}
	return b.feeds[name]
}

func (b *StatusBoard) Success(name string, entries int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.get(name)
	s.LastSuccess = time.Now()
	s.Entries = entries
	s.LastError = ""
	s.Failures = 0
}

func (b *StatusBoard) Failure(name string, err error) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.get(name)
	s.LastError = err.Error()
	s.Failures++
	return s.Failures
}

func StatusHandler(b *StatusBoard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b.mu.Lock()
		list := make([]FeedStatus, 0, len(b.feeds))
		for _, s := range b.feeds {
			list = append(list, *s)
		}
		b.mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}
