package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/remote-remote/flow/internal/linear"
)

// A missing note with a known identifier must be created and printed without
// needing a TTY, the same as when the note already exists.
func TestTaskNoteNoOpenCreatesNoteWithoutTTY(t *testing.T) {
	home := t.TempDir()
	vault := filepath.Join(home, "vault")
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "flow.yaml"), []byte("vault_path: "+vault+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig := issueByIdentifier
	t.Cleanup(func() { issueByIdentifier = orig; taskNoteNoOpen = false })
	issueByIdentifier = func(id string) (*linear.Issue, error) {
		return &linear.Issue{Identifier: id, Title: "Fix it", URL: "https://linear.app/x/issue/" + id}, nil
	}

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	origStdin, origStdout := os.Stdin, os.Stdout
	r, w, _ := os.Pipe()
	os.Stdin, os.Stdout = devNull, w
	t.Cleanup(func() { os.Stdin, os.Stdout = origStdin, origStdout })

	rootCommand.SetArgs([]string{"note", "task", "eng-42", "--no-open"})
	runErr := rootCommand.Execute()
	w.Close()
	os.Stdin, os.Stdout = origStdin, origStdout
	out, _ := io.ReadAll(r)

	if runErr != nil {
		t.Fatalf("execute: %v", runErr)
	}
	want := filepath.Join(vault, "Tasks", "ENG-42.md")
	if !strings.Contains(string(out), want) {
		t.Fatalf("stdout %q does not contain %q", out, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("note not created: %v", err)
	}
}
