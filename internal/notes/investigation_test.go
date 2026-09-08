package notes

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

// captureStdout runs fn with os.Stdout replaced by a pipe and returns what it wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	w.Close()

	out, _ := io.ReadAll(r)
	return string(out)
}

func TestInvestigationPath(t *testing.T) {
	got, err := NewInvestigation("How does auth work?").Path("/vault")
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := filepath.Join("/vault", "Notes", "how-does-auth-work.md"); got != want {
		t.Errorf("investigation path = %q, want %q", got, want)
	}

	got, err = NewProjectPlan("Agentic Flow").Path("/vault")
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := filepath.Join("/vault", "Projects", "Agentic Flow", "plan.md"); got != want {
		t.Errorf("plan path = %q, want %q", got, want)
	}

	if _, err := NewInvestigation("???").Path("/vault"); err == nil {
		t.Error("a title that slugifies to nothing should be rejected")
	}
}

func TestOpenInvestigation(t *testing.T) {
	vault := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{VaultPath: vault}

	inv := NewInvestigation("How does auth work?")
	inv.Scratch = "/state/agents/scratch/api/main-0af01e"
	inv.Scopes = []string{"session", "middleware"}
	inv.Task = "ENG-42"

	out := captureStdout(t, func() {
		if err := OpenInvestigation(cfg, inv, true); err != nil {
			t.Fatalf("OpenInvestigation: %v", err)
		}
	})

	notePath, _ := inv.Path(vault)
	if strings.TrimSpace(out) != notePath {
		t.Errorf("--no-open printed %q, want the note path %q", strings.TrimSpace(out), notePath)
	}

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("note not created: %v", err)
	}
	note := string(data)
	for _, want := range []string{
		`title: "How does auth work?"`,
		`scratch: "/state/agents/scratch/api/main-0af01e"`,
		`scopes: ["session", "middleware"]`,
		`task: "ENG-42"`,
		"## Log",
	} {
		if !strings.Contains(note, want) {
			t.Errorf("note missing %q\ngot:\n%s", want, note)
		}
	}

	// Cross-linked into today's daily note, once.
	dailyPath, _ := config.DailyNotePath(vault, time.Now())
	daily, err := os.ReadFile(dailyPath)
	if err != nil {
		t.Fatalf("daily note not created: %v", err)
	}
	wikilink := "[[Notes/how-does-auth-work|How does auth work?]]"
	if !strings.Contains(string(daily), wikilink) {
		t.Errorf("daily note missing %q\ngot:\n%s", wikilink, daily)
	}

	// Re-running reopens the same note rather than scaffolding a second one.
	captureStdout(t, func() {
		if err := OpenInvestigation(cfg, inv, true); err != nil {
			t.Fatalf("second OpenInvestigation: %v", err)
		}
	})
	daily, _ = os.ReadFile(dailyPath)
	if strings.Count(string(daily), wikilink) != 1 {
		t.Errorf("wikilink duplicated in daily note")
	}
}

func TestOpenProjectPlanCreatesProjectNote(t *testing.T) {
	vault := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{VaultPath: vault}

	inv := NewProjectPlan("Agentic Flow")
	captureStdout(t, func() {
		if err := OpenInvestigation(cfg, inv, true); err != nil {
			t.Fatalf("OpenInvestigation: %v", err)
		}
	})

	if _, err := os.Stat(ProjectNotePath(vault, "Agentic Flow")); err != nil {
		t.Errorf("project note not created alongside the plan: %v", err)
	}

	planPath, _ := inv.Path(vault)
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("plan note not created: %v", err)
	}
	if want := `project: "Agentic Flow"`; !strings.Contains(string(data), want) {
		t.Errorf("plan note missing %q\ngot:\n%s", want, data)
	}
}
