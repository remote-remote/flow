package notes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
)

// OpenTask opens or creates a task note for the given Linear issue, cross-links it
// to today's daily note, and opens it in $EDITOR.
func OpenTask(cfg *config.Config, issue *linear.Issue) error {
	taskPath := TaskNotePath(cfg.VaultPath, issue.Identifier)

	// Create task note if it doesn't exist
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(taskPath), 0o755); err != nil {
			return err
		}
		content := renderTaskTemplate(issue)
		if err := os.WriteFile(taskPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Cross-link to today's daily note
	if err := crossLinkToDaily(cfg, issue); err != nil {
		// Non-fatal: don't block opening the task note
		fmt.Fprintf(os.Stderr, "warning: could not cross-link to daily note: %v\n", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, taskPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenExistingTask opens a task note that already exists on disk.
func OpenExistingTask(path string) error {
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

func TaskNotePath(vaultPath, identifier string) string {
	return filepath.Join(vaultPath, "Tasks", identifier+".md")
}

func renderTaskTemplate(issue *linear.Issue) string {
	return fmt.Sprintf(`---
title: "%s"
linear_id: %s
linear_url: %s
status: %s
tags: [task]
---
# %s: %s

## Notes

## Log
`, issue.Title, issue.Identifier, issue.URL, issue.State.Name,
		issue.Identifier, issue.Title)
}

func crossLinkToDaily(cfg *config.Config, issue *linear.Issue) error {
	now := time.Now()
	dailyPath, err := config.DailyNotePath(cfg.VaultPath, now)
	if err != nil {
		return err
	}

	// Create daily note if it doesn't exist
	if _, err := os.Stat(dailyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dailyPath), 0o755); err != nil {
			return err
		}
		content := RenderDailyTemplate(now)
		if err := os.WriteFile(dailyPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Read daily note
	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return err
	}

	wikilink := fmt.Sprintf("[[Tasks/%s|%s: %s]]", issue.Identifier, issue.Identifier, issue.Title)

	// Don't add duplicate links
	if strings.Contains(string(data), wikilink) {
		return nil
	}

	// Find ## Tasks section and append the wikilink
	content := string(data)
	tasksIdx := strings.Index(content, "## Tasks")
	if tasksIdx == -1 {
		// Append a Tasks section
		content += "\n## Tasks\n- " + wikilink + "\n"
	} else {
		// Find end of the ## Tasks line
		afterTasks := tasksIdx + len("## Tasks")
		// Find the next line
		nextLine := strings.Index(content[afterTasks:], "\n")
		if nextLine == -1 {
			content += "\n- " + wikilink + "\n"
		} else {
			insertAt := afterTasks + nextLine + 1
			content = content[:insertAt] + "- " + wikilink + "\n" + content[insertAt:]
		}
	}

	return os.WriteFile(dailyPath, []byte(content), 0o644)
}
