package standup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

func TestTaskLinksFromDaily(t *testing.T) {
	vault := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	cfg := &config.Config{VaultPath: vault}
	date := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)

	dailyPath, _ := config.DailyNotePath(vault, date)
	os.MkdirAll(filepath.Dir(dailyPath), 0o755)
	content := `---
date: 2026-03-16
---
# Monday

## Tasks
- [[Tasks/ENG-42|ENG-42: Fix the thing]]
- [[Tasks/ENG-99|ENG-99: Build widget]]

## Notes
`
	os.WriteFile(dailyPath, []byte(content), 0o644)

	links := taskLinksFromDaily(cfg, date)
	if len(links) != 2 {
		t.Fatalf("got %d links, want 2", len(links))
	}
	if links[0] != "ENG-42: Fix the thing" {
		t.Errorf("link[0] = %q, want %q", links[0], "ENG-42: Fix the thing")
	}
}

func TestRecentTaskNotes(t *testing.T) {
	vault := t.TempDir()
	tasksDir := filepath.Join(vault, "Tasks")
	os.MkdirAll(tasksDir, 0o755)

	content := `---
title: "Fix the thing"
---
`
	os.WriteFile(filepath.Join(tasksDir, "ENG-42.md"), []byte(content), 0o644)

	cfg := &config.Config{VaultPath: vault}
	since := time.Now().Add(-1 * time.Hour)
	items := recentTaskNotes(cfg, since)

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0] != "ENG-42: Fix the thing" {
		t.Errorf("item = %q, want %q", items[0], "ENG-42: Fix the thing")
	}
}

func TestReadTaskTitle(t *testing.T) {
	f := filepath.Join(t.TempDir(), "test.md")
	os.WriteFile(f, []byte("---\ntitle: \"My Task\"\n---\n"), 0o644)

	got := readTaskTitle(f)
	if got != "My Task" {
		t.Errorf("readTaskTitle = %q, want %q", got, "My Task")
	}
}
