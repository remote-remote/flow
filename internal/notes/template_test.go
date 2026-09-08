package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderDailyTemplate(t *testing.T) {
	got, source := RenderDailyTemplate(t.TempDir(), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC))

	if source != "built-in" {
		t.Errorf("source = %q, want built-in", source)
	}

	checks := []string{
		"date: 2026-03-16",
		"# Monday, March 16, 2026",
		"## Tasks",
		"## Notes",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("template missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestRenderPrefersVaultTemplate(t *testing.T) {
	vault := t.TempDir()
	path := TemplatePath(vault, "daily")
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("# {{date_long}}\n\n## Shipped\n"), 0o644)

	got, source := RenderDailyTemplate(vault, time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC))

	if source != filepath.Join("_templates", "daily.md") {
		t.Errorf("source = %q, want the vault template", source)
	}
	if want := "# Monday, March 16, 2026\n\n## Shipped\n"; got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRenderLeavesUnknownPlaceholders(t *testing.T) {
	vault := t.TempDir()
	path := TemplatePath(vault, "project")
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("# {{name}} — {{nope}}\n"), 0o644)

	got, _ := render(vault, "project", map[string]string{"name": "Flow"})
	if want := "# Flow — {{nope}}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderProjectSections(t *testing.T) {
	got, _ := render(t.TempDir(), "project", map[string]string{"name": "Flow"})

	checks := []string{
		`title: "Flow"`,
		"linear_project_id:",
		"## Overview",
		"## Approach",
		"## Rejected Alternatives",
		"## Tasks",
		"## Out of scope",
		"## Open",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("project template missing %q\ngot:\n%s", want, got)
		}
	}
	if strings.Contains(got, "## Notes") {
		t.Errorf("project template should no longer carry ## Notes\ngot:\n%s", got)
	}
}
