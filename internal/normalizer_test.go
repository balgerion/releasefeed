package internal

import (
	"os"
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

func TestRewritePrivateImages(t *testing.T) {
	in := `<img src="https://private-user-images.githubusercontent.com/34356590/611930431-daba6e6c-5df3-4e5b-b7a1-c737ca265f0e.webp?jwt=eyJ0.abc-123_x" alt="banner">`
	out := Normalize(Release{Content: in}, Feed{})
	want := `<img src="https://github.com/user-attachments/assets/daba6e6c-5df3-4e5b-b7a1-c737ca265f0e" alt="banner">`
	if out.Content != want {
		t.Errorf("got: %s", out.Content)
	}
}

func TestStatePersistsFirstSeenTime(t *testing.T) {
	path := t.TempDir() + "/state.json"
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	st, err := NewState(path, path+".alerted")
	if err != nil {
		t.Fatal(err)
	}
	if st.Known("Forgejo") {
		t.Error("new state should not know any feed")
	}
	if err := st.Replace("Forgejo", map[string]time.Time{"a": first}, map[string]bool{"a": true}); err != nil {
		t.Fatal(err)
	}

	st2, err := NewState(path, path+".alerted")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := st2.FirstSeen("Forgejo", "a")
	if !ok || !got.Equal(first) {
		t.Errorf("first-seen time not persisted: %v %v", got, ok)
	}
	if _, ok := st2.FirstSeen("Renovate", "a"); ok || st2.Known("Renovate") {
		t.Error("ids must not leak between feeds")
	}
}

func TestStateStartsFreshFromOldFormat(t *testing.T) {
	path := t.TempDir() + "/state.json"
	if err := os.WriteFile(path, []byte(`{"a":"2026-01-01T00:00:00Z"}`), 0644); err != nil {
		t.Fatal(err)
	}
	st, err := NewState(path, path+".alerted")
	if err != nil {
		t.Fatal(err)
	}
	if st.Known("a") {
		t.Error("old flat state must not be read as feeds")
	}
}
