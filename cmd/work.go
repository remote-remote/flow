package cmd

import (
	"errors"
	"fmt"

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

		var issue *linear.Issue
		var dirty bool

		if len(args) == 1 {
			identifier := args[0]
			dirty = tui.GitDirty()
			result := tui.StartIssueResult(identifier, dirty)
			if result.Err() != nil {
				return result.Err()
			}
			issue = result.Issue()
		} else {
			result, err := tui.RunWorkFlow()
			if err != nil {
				return err
			}
			if result == nil {
				return nil
			}
			issue = result.Issue
			dirty = result.Dirty
		}

		if dirty {
			fmt.Println("Worktree is dirty — commit or stash to checkout the branch.")
		}

		return notes.OpenTask(cfg, issue)
	},
}

