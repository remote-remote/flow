package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work [identifier]",
	Short: "Pick a task to work on",
	Long:  "Browse projects and tasks, self-assign, and open the task note.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		var identifier string

		if len(args) == 1 {
			identifier = args[0]
		} else {
			// Pick a project
			projects, err := linear.Projects()
			if err != nil {
				return fmt.Errorf("failed to fetch projects: %w", err)
			}
			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			project := tui.PickProject(projects)
			if project == nil {
				return nil
			}

			// Pick an issue from that project
			issues, err := linear.ProjectIssues(project.Name)
			if err != nil {
				return fmt.Errorf("failed to fetch issues: %w", err)
			}
			if len(issues) == 0 {
				fmt.Println("No issues in this project.")
				return nil
			}

			picked := tui.PickIssue(issues)
			if picked == nil {
				return nil
			}
			identifier = picked.Identifier
		}

		// Start the issue (self-assign + In Progress + checkout branch)
		fmt.Printf("Starting %s...\n", identifier)
		if gitWorktreeDirty() {
			// Start without checkout, warn about dirty worktree
			if err := linear.StartIssue(identifier); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not start issue: %v\n", err)
			}
			fmt.Println("Worktree is dirty — commit or stash to checkout the branch.")
		} else {
			if err := linear.StartIssueWithCheckout(identifier); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not start issue: %v\n", err)
			}
		}

		// Fetch full details and open task note
		issue, err := linear.IssueByIdentifier(identifier)
		if err != nil {
			return fmt.Errorf("failed to fetch issue: %w", err)
		}

		return notes.OpenTask(cfg, issue)
	},
}

func gitWorktreeDirty() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return true // assume dirty if we can't check
	}
	return strings.TrimSpace(string(out)) != ""
}
