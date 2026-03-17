package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var issueIDRe = regexp.MustCompile(`[A-Z]+-\d+`)

var taskNote = &cobra.Command{
	Use:   "task [identifier]",
	Short: "Open a task note for a Linear issue",
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
		} else if id := identifierFromBranch(); id != "" {
			identifier = id
		} else {
			issues, err := linear.AssignedIssues()
			if err != nil {
				return fmt.Errorf("failed to fetch issues: %w", err)
			}
			if len(issues) == 0 {
				fmt.Println("No assigned issues found.")
				return nil
			}
			picked := tui.PickIssue(issues)
			if picked == nil {
				return nil
			}
			identifier = picked.Identifier
		}

		issue, err := linear.IssueByIdentifier(identifier)
		if err != nil {
			return fmt.Errorf("failed to fetch issue: %w", err)
		}

		return notes.OpenTask(cfg, issue)
	},
}

func identifierFromBranch() string {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	return issueIDRe.FindString(branch)
}
