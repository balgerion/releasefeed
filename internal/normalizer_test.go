package internal

import (
	"strings"
	"testing"
	"time"
)

func TestExcludeSections(t *testing.T) {
	names := []string{"sponsors", "contributors"}
	cases := []struct{ name, in, gone, kept string }{
		{"plain h2", `<h2>Changes</h2><p>fix</p><h2>New Contributors</h2><ul><li>@x</li></ul>`, "@x", "fix"},
		{"h3 inside excluded h2", `<h2>Sponsors</h2><h3>Gold</h3><p>acme</p><h2>Fixes</h2><p>bug</p>`, "acme", "bug"},
		{"bold paragraph heading", `<h2>Changes</h2><p>fix</p><p><strong>Sponsors</strong></p><ul><li>acme</li></ul>`, "acme", "fix"},
		{"multiline heading", "<h2>\nSponsors\n</h2><p>acme</p><h2>Fixes</h2><p>bug</p>", "acme", "bug"},
	}
	for _, c := range cases {
		out := excludeSectionsHTML(c.in, names)
		if strings.Contains(strings.ToLower(out), c.gone) {
			t.Errorf("%s: %q leaked: %s", c.name, c.gone, out)
		}
		if !strings.Contains(out, c.kept) {
			t.Errorf("%s: %q wrongly removed: %s", c.name, c.kept, out)
		}
	}
}

func TestCachePinsFirstSeenTime(t *testing.T) {
	c := NewCache()
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c.Set("f", []Release{{ID: "a", PublishedAt: first}})
	c.Set("f", []Release{{ID: "a", PublishedAt: first.Add(48 * time.Hour)}})
	if got := c.All()[0].PublishedAt; !got.Equal(first) {
		t.Errorf("timestamp not pinned: %v", got)
	}
}
