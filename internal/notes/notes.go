// Package notes owns the vault layout: where each kind of note lives, how it is
// created from a template, and how it gets handed back to the caller.
package notes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a title to a filename-friendly slug.
func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// createFromTemplate writes path from the named template unless it already exists.
func createFromTemplate(vaultPath, path, name string, vars map[string]string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, source := render(vaultPath, name, vars)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	reportTemplate(path, source)
	return nil
}

// insertUnderHeading splices payload in directly below heading, so a section
// reads newest-first. A missing heading gets the section appended to the end.
// The heading must be a line of its own, so a "### Tasks" nested further down
// cannot capture an insert meant for "## Tasks".
func insertUnderHeading(content, heading, payload string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimRight(line, " \t\r") != heading {
			continue
		}
		spliced := make([]string, 0, len(lines)+1)
		spliced = append(spliced, lines[:i+1]...)
		spliced = append(spliced, payload)
		spliced = append(spliced, lines[i+1:]...)
		return strings.Join(spliced, "\n")
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + heading + "\n" + payload + "\n"
}

// ensureDaily returns the daily note path for now, creating the note if missing.
func ensureDaily(cfg *config.Config, now time.Time) (string, error) {
	path, err := config.DailyNotePath(cfg.VaultPath, now)
	if err != nil {
		return "", err
	}
	content, source := RenderDailyTemplate(cfg.VaultPath, now)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", err
		}
		reportTemplate(path, source)
	}
	return path, nil
}

// appendToDaily inserts payload under heading in today's daily note. A non-empty
// dedupeKey makes the write idempotent when that string is already present.
func appendToDaily(cfg *config.Config, heading, payload, dedupeKey string) error {
	dailyPath, err := ensureDaily(cfg, time.Now())
	if err != nil {
		return err
	}

	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return err
	}
	if dedupeKey != "" && strings.Contains(string(data), dedupeKey) {
		return nil
	}

	return os.WriteFile(dailyPath, []byte(insertUnderHeading(string(data), heading, payload)), 0o644)
}

// deliver hands a note back to the caller: $EDITOR normally, or the resolved path
// on stdout when noOpen is set, since an agent shelling out would hang in vim.
func deliver(path string, noOpen bool) error {
	if noOpen {
		fmt.Println(path)
		return nil
	}
	return openInEditor(path)
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
