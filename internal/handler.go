package internal

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	mu      sync.RWMutex
	entries map[string][]Release
}

func NewCache() *Cache {
	return &Cache{entries: make(map[string][]Release)}
}

func (c *Cache) Set(name string, releases []Release) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[name] = releases
}

func (c *Cache) All() []Release {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var all []Release
	for _, releases := range c.entries {
		all = append(all, releases...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].PublishedAt.After(all[j].PublishedAt)
	})
	return all
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Content atomContent `xml:"content"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",innerxml"`
}

func escapeContent(s string) string {
	var buf bytes.Buffer
	for _, line := range strings.Split(s, "\n") {
		xml.EscapeText(&buf, []byte(line))
		buf.WriteString("<br/>")
	}
	return buf.String()
}

func FeedHandler(cache *Cache, imageBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		releases := cache.All()
		entries := make([]atomEntry, 0, len(releases))
		for _, rel := range releases {
			body := ""
			if rel.Image != "" {
				body = `<img src="` + imageBaseURL + `/static/` + rel.Image + `" style="max-height:48px;margin-bottom:8px;"/><br/>`
			}
			body += escapeContent(rel.Content)
			entries = append(entries, atomEntry{
				Title:   "[" + rel.FeedName + "] " + rel.Title,
				Link:    atomLink{Href: rel.URL},
				ID:      rel.ID,
				Updated: rel.PublishedAt.Format(time.RFC3339),
				Content: atomContent{Type: "html", Value: body},
			})
		}
		af := atomFeed{
			XMLNS:   "http://www.w3.org/2005/Atom",
			Title:   "Release Feed",
			Updated: time.Now().Format(time.RFC3339),
			Entries: entries,
		}
		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		w.Write([]byte(xml.Header))
		xml.NewEncoder(w).Encode(af)
	}
}