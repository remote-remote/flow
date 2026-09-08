package notes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
)

// OpenTask opens or creates a task note for the given Linear issue, cross-links it
// to today's daily note, and hands it to the caller.
func OpenTask(cfg *config.Config, issue *linear.Issue, noOpen bool) error {
	taskPath := TaskNotePath(cfg.VaultPath, issue)

	if err := createFromTemplate(cfg.VaultPath, taskPath, "task", taskVars(issue)); err != nil {
		return err
	}

	if issue.Project != nil {
		if err := ensureProject(cfg.VaultPath, issue.Project.Name); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create project note: %v\n", err)
		}
	}

	if err := crossLinkToDaily(cfg, issue); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not cross-link to daily note: %v\n", err)
	}

	return deliver(taskPath, noOpen)
}

// OpenExistingTask opens a task note that already exists on disk.
func OpenExistingTask(path string, noOpen bool) error {
	return deliver(path, noOpen)
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

func taskVars(issue *linear.Issue) map[string]string {
	project := ""
	if issue.Project != nil {
		project = issue.Project.Name
	}
	return map[string]string{
		"title":      issue.Title,
		"linear_id":  issue.Identifier,
		"linear_url": issue.URL,
		"status":     issue.State.Name,
		"project":    project,
	}
}

func crossLinkToDaily(cfg *config.Config, issue *linear.Issue) error {
	wikilink := taskWikilink(issue)
	return appendToDaily(cfg, "## Tasks", "- "+wikilink, wikilink)
}
