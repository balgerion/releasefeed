package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAlertedMigrationAndPersistence(t *testing.T) {
	dir := t.TempDir()
	seenPath, alertedPath := filepath.Join(dir, "state.json"), filepath.Join(dir, "alerted.json")
	if err := os.WriteFile(seenPath, []byte(`{"qui":{"v1.33.0":"2026-10-19T08:34:26Z"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	st, err := NewState(seenPath, alertedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Alerted("qui", "v1.33.0") {
		t.Error("a release known before alerted.json existed must count as alerted")
	}
	if st.Alerted("qui", "v1.34.0") {
		t.Error("an unknown release must not count as alerted")
	}
	now := time.Now()
	if err := st.Replace("qui", map[string]time.Time{"v1.33.0": now, "v1.34.0": now}, map[string]bool{"v1.33.0": true, "v1.34.0": false}); err != nil {
		t.Fatal(err)
	}
	st, err = NewState(seenPath, alertedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Alerted("qui", "v1.33.0") || st.Alerted("qui", "v1.34.0") {
		t.Error("alerted.json must be read back as written, not migrated again")
	}
}
