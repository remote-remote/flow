package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyNotePath_Fallback(t *testing.T) {
	vault := t.TempDir()
	date := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)

	path, err := DailyNotePath(vault, date)
	if err != nil {
		t.Fatalf("DailyNotePath: %v", err)
	}

	want := filepath.Join(vault, "2026-03-16.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestDailyNotePath_CoreConfig(t *testing.T) {
	vault := t.TempDir()
	obsDir := filepath.Join(vault, ".obsidian")
	os.MkdirAll(obsDir, 0o755)

	config := `{"folder": "Daily", "format": "YYYY-MM-DD"}`
	os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(config), 0o644)

	date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	path, err := DailyNotePath(vault, date)
	if err != nil {
		t.Fatalf("DailyNotePath: %v", err)
	}

	want := filepath.Join(vault, "Daily", "2026-01-05.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestDailyNotePath_PeriodicNotes(t *testing.T) {
	vault := t.TempDir()
	pluginDir := filepath.Join(vault, ".obsidian", "plugins", "periodic-notes")
	os.MkdirAll(pluginDir, 0o755)

	config := `{"daily": {"folder": "Journal", "format": "YYYY-MM-DD"}}`
	os.WriteFile(filepath.Join(pluginDir, "data.json"), []byte(config), 0o644)

	date := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	path, err := DailyNotePath(vault, date)
	if err != nil {
		t.Fatalf("DailyNotePath: %v", err)
	}

	want := filepath.Join(vault, "Journal", "2026-12-25.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestMomentToGoConversion(t *testing.T) {
	tests := []struct {
		moment string
		date   time.Time
		want   string
	}{
		{"YYYY-MM-DD", time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), "2026-03-16"},
		{"DD-MM-YYYY", time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), "16-03-2026"},
		{"YYYY/MM/DD", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), "2026/01/05"},
	}

	for _, tt := range tests {
		goFmt := momentToGo.Replace(tt.moment)
		got := tt.date.Format(goFmt)
		if got != tt.want {
			t.Errorf("format(%q) = %q, want %q", tt.moment, got, tt.want)
		}
	}
}
