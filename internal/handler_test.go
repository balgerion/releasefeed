package internal

import (
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
