package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"mime"
	"net/http"
	"net/url"
	"path"
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

func (c *Cache) Get(feed, title string) (Release, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, r := range c.entries[feed] {
		if r.Title == title {
			return r, true
		}
	}
	return Release{}, false
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

func imageType(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return ""
	}
	return mime.TypeByExtension(path.Ext(u.Path))
}

func FeedHandler(cache *Cache, imageBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		releases := cache.All()
		var sb strings.Builder
		sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		sb.WriteString(`<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/">`)
		sb.WriteString(`<title>Release Feed</title>`)
		sb.WriteString(fmt.Sprintf(`<updated>%s</updated>`, time.Now().Format(time.RFC3339)))
		for _, rel := range releases {
			body := rel.Content
			src := ""
			if rel.Image != "" {
				src = rel.Image
				if !strings.HasPrefix(src, "http") {
					src = imageBaseURL + "/image/" + src
				}
				body = fmt.Sprintf(`<img src="%s" width="48" height="48" style="margin-bottom:8px;"/><br/>`, src) + body
			}
			sb.WriteString(`<entry>`)
			sb.WriteString(fmt.Sprintf(`<title>%s</title>`, xmlEscape("["+rel.FeedName+"] "+rel.Title)))
			sb.WriteString(fmt.Sprintf(`<link href="%s"/>`, xmlEscape(imageBaseURL+"/release/"+url.PathEscape(rel.FeedName)+"/"+url.PathEscape(rel.Title))))
			sb.WriteString(fmt.Sprintf(`<id>%s</id>`, xmlEscape(rel.ID)))
			sb.WriteString(fmt.Sprintf(`<updated>%s</updated>`, rel.PublishedAt.Format(time.RFC3339)))
			if src != "" {
				sb.WriteString(fmt.Sprintf(`<media:thumbnail url="%s"/>`, xmlEscape(src)))
				if t := imageType(src); t != "" {
					sb.WriteString(fmt.Sprintf(`<link rel="enclosure" type="%s" href="%s"/>`, t, xmlEscape(src)))
				}
			}
			sb.WriteString(`<content type="html"><![CDATA[`)
			sb.WriteString(strings.ReplaceAll(body, "]]>", "]]]]><![CDATA[>"))
			sb.WriteString(`]]></content>`)
			sb.WriteString(`</entry>`)
		}
		sb.WriteString(`</feed>`)
		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		w.Write([]byte(sb.String()))
	}
}

func ReleaseHandler(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel, ok := cache.Get(r.PathValue("name"), r.PathValue("title"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="color-scheme" content="light dark"><title>%s</title><style>body{max-width:44rem;margin:0 auto;padding:1rem;font-family:system-ui,sans-serif;line-height:1.5}img{max-width:100%%;height:auto}pre{overflow-x:auto}</style></head><body><article><h1>%s</h1><p><a href="%s">View on GitHub</a></p>%s</article></body></html>`,
			html.EscapeString("["+rel.FeedName+"] "+rel.Title),
			html.EscapeString("["+rel.FeedName+"] "+rel.Title),
			html.EscapeString(rel.URL),
			rel.Content)
	}
}