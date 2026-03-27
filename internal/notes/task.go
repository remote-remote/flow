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
	taskPath := TaskNotePath(cfg.VaultPath, issue)

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

	// Ensure project note exists if issue has a project
	if issue.Project != nil {
		projPath := ProjectNotePath(cfg.VaultPath, issue.Project.Name)
		if _, err := os.Stat(projPath); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(projPath), 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not create project note: %v\n", err)
			} else {
				content := renderProjectTemplate(issue.Project.Name)
				os.WriteFile(projPath, []byte(content), 0o644)
			}
		}
	}

	// Cross-link to today's daily note
	if err := crossLinkToDaily(cfg, issue); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not cross-link to daily note: %v\n", err)
	}

	return openInEditor(taskPath)
}

// OpenExistingTask opens a task note that already exists on disk.
func OpenExistingTask(path string) error {
	return openInEditor(path)
}

// TaskNotePath returns the path for a task note.
// With project: {vault}/Projects/{project}/Tasks/{identifier}.md
// Without:     {vault}/Tasks/{identifier}.md
func TaskNotePath(vaultPath string, issue *linear.Issue) string {
	if issue.Project != nil {
		return filepath.Join(vaultPath, "Projects", issue.Project.Name, "Tasks", issue.Identifier+".md")
	}
	return filepath.Join(vaultPath, "Tasks", issue.Identifier+".md")
}

// TaskNotePathByID returns the path for a task note by identifier alone.
// Checks project-based paths first, falls back to flat Tasks/ directory.
func TaskNotePathByID(vaultPath, identifier string) string {
	// Check if it exists under any project
	projectsDir := filepath.Join(vaultPath, "Projects")
	entries, err := os.ReadDir(projectsDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			candidate := filepath.Join(projectsDir, e.Name(), "Tasks", identifier+".md")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	// Fall back to flat path
	return filepath.Join(vaultPath, "Tasks", identifier+".md")
}

// taskWikilink returns the relative wikilink for a task note.
func taskWikilink(issue *linear.Issue) string {
	label := issue.Identifier + ": " + issue.Title
	if issue.Project != nil {
		return fmt.Sprintf("[[Projects/%s/Tasks/%s|%s]]", issue.Project.Name, issue.Identifier, label)
	}
	return fmt.Sprintf("[[Tasks/%s|%s]]", issue.Identifier, label)
}

func renderTaskTemplate(issue *linear.Issue) string {
	project := ""
	if issue.Project != nil {
		project = issue.Project.Name
	}
	return fmt.Sprintf(`---
title: "%s"
linear_id: %s
linear_url: %s
status: %s
project: "%s"
tags: [task]
---
# %s: %s

## Notes

## Log
`, issue.Title, issue.Identifier, issue.URL, issue.State.Name,
		project, issue.Identifier, issue.Title)
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

	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return err
	}

	wikilink := taskWikilink(issue)

	// Don't add duplicate links
	if strings.Contains(string(data), wikilink) {
		return nil
	}

	// Find ## Tasks section and append the wikilink
	content := string(data)
	tasksIdx := strings.Index(content, "## Tasks")
	if tasksIdx == -1 {
		content += "\n## Tasks\n- " + wikilink + "\n"
	} else {
		afterTasks := tasksIdx + len("## Tasks")
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
