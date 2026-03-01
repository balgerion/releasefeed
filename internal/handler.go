package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
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

func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

func FeedHandler(cache *Cache, imageBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		releases := cache.All()
		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		sb.WriteString(`<feed xmlns="http://www.w3.org/2005/Atom">`)
		sb.WriteString(`<title>Release Feed</title>`)
		sb.WriteString(fmt.Sprintf(`<updated>%s</updated>`, time.Now().Format(time.RFC3339)))
		for _, rel := range releases {
			body := rel.Content
			if rel.Image != "" {
				src := imageBaseURL + "/image/" + rel.Image
				body = fmt.Sprintf(`<img src="%s" width="48" height="48" style="margin-bottom:8px;"/><br/>`, src) + body
			}
			sb.WriteString(`<entry>`)
			sb.WriteString(fmt.Sprintf(`<title>%s</title>`, xmlEscape("["+rel.FeedName+"] "+rel.Title)))
			sb.WriteString(fmt.Sprintf(`<link href="%s"/>`, xmlEscape(rel.URL)))
			sb.WriteString(fmt.Sprintf(`<id>%s</id>`, xmlEscape(rel.ID)))
			sb.WriteString(fmt.Sprintf(`<updated>%s</updated>`, rel.PublishedAt.Format(time.RFC3339)))
			sb.WriteString(`<content type="html"><![CDATA[`)
			sb.WriteString(body)
			sb.WriteString(`]]></content>`)
			sb.WriteString(`</entry>`)
		}
		sb.WriteString(`</feed>`)
		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		w.Write([]byte(sb.String()))
	}
}