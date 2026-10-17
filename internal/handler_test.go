package internal

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFeedEnclosure(t *testing.T) {
	cache := NewCache()
	cache.Set("Vaultwarden", []Release{{ID: "1", FeedName: "Vaultwarden", Title: "1.37.3", PublishedAt: time.Now(), Image: "https://icons.balgeriada.com/vaultwarden.webp"}})
	cache.Set("Hyprland", []Release{{ID: "2", FeedName: "Hyprland", Title: "v0.55.4", PublishedAt: time.Now(), Image: "hyprland.svg"}})
	cache.Set("Plain", []Release{{ID: "3", FeedName: "Plain", Title: "v1", PublishedAt: time.Now()}})
	rec := httptest.NewRecorder()
	FeedHandler(cache, "https://feed.example")(rec, httptest.NewRequest("GET", "/feed", nil))
	out := rec.Body.String()
	for _, want := range []string{
		`<link rel="enclosure" type="image/webp" href="https://icons.balgeriada.com/vaultwarden.webp"/>`,
		`<link rel="enclosure" type="image/svg+xml" href="https://feed.example/image/hyprland.svg"/>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if strings.Count(out, `rel="enclosure"`) != 2 {
		t.Errorf("release without image must not get an enclosure: %s", out)
	}
}

func TestReleasesFilter(t *testing.T) {
	cache := NewCache()
	cache.Set("Renovate", []Release{{FeedName: "Renovate", Title: "45.0.0", Content: "BREAKING CHANGES: node 24", PublishedAt: time.Now()}})
	cache.Set("Forgejo", []Release{{FeedName: "Forgejo", Title: "v13.0.1", Content: "bug fixes", PublishedAt: time.Now()}})
	for query, want := range map[string]int{"": 2, "?feed=renovate": 1, "?q=breaking": 1, "?feed=Forgejo&q=breaking": 0} {
		rec := httptest.NewRecorder()
		ReleasesHandler(cache)(rec, httptest.NewRequest("GET", "/api/releases"+query, nil))
		var got []apiRelease
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if len(got) != want {
			t.Errorf("%q: got %d releases, want %d", query, len(got), want)
		}
	}
}
