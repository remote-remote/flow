package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
)

func TestTaskNotePath(t *testing.T) {
	got := TaskNotePath("/vault", "ENG-123")
	want := filepath.Join("/vault", "Tasks", "ENG-123.md")
	if got != want {
		t.Errorf("TaskNotePath = %q, want %q", got, want)
	}
}

func TestCrossLinkToDaily(t *testing.T) {
	vault := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	cfg := &config.Config{VaultPath: vault}
	issue := &linear.Issue{
		Identifier: "ENG-42",
		Title:      "Fix the thing",
		URL:        "https://linear.app/eng-42",
		State:      linear.IssueState{Name: "In Progress", Category: "started"},
	}

	// Create daily note first
	now := time.Now()
	dailyPath, _ := config.DailyNotePath(vault, now)
	os.MkdirAll(filepath.Dir(dailyPath), 0o755)
	os.WriteFile(dailyPath, []byte(RenderDailyTemplate(now)), 0o644)

	// Cross-link
	if err := crossLinkToDaily(cfg, issue); err != nil {
		t.Fatalf("crossLinkToDaily: %v", err)
	}

	data, _ := os.ReadFile(dailyPath)
	content := string(data)

	wikilink := "[[Tasks/ENG-42|ENG-42: Fix the thing]]"
	if !strings.Contains(content, wikilink) {
		t.Errorf("daily note missing wikilink %q\ncontent:\n%s", wikilink, content)
	}

	// Second call should not duplicate
	if err := crossLinkToDaily(cfg, issue); err != nil {
		t.Fatalf("crossLinkToDaily second call: %v", err)
	}
	data, _ = os.ReadFile(dailyPath)
	if strings.Count(string(data), wikilink) != 1 {
		t.Errorf("wikilink duplicated in daily note")
	}
}

func TestRenderTaskTemplate(t *testing.T) {
	issue := &linear.Issue{
		Identifier: "ENG-99",
		Title:      "Build the widget",
		URL:        "https://linear.app/eng-99",
		State:      linear.IssueState{Name: "Todo", Category: "unstarted"},
	}

	got := renderTaskTemplate(issue)
	checks := []string{
		`title: "Build the widget"`,
		"linear_id: ENG-99",
		"# ENG-99: Build the widget",
		"## Notes",
		"## Log",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("task template missing %q\ngot:\n%s", want, got)
		}
	}
}
