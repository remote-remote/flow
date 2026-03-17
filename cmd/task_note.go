package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

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

		apiKey, err := config.GetSecret("linear-api-key")
		if err != nil {
			fmt.Println("Linear API key not configured. Run `flow config` to set up.")
			return nil
		}

		client := linear.NewClient(apiKey)
		ctx := context.Background()

		var issue *linear.Issue

		if len(args) == 1 {
			// Direct identifier lookup
			issue, err = client.IssueByIdentifier(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch issue: %w", err)
			}
			if issue == nil {
				return fmt.Errorf("issue %s not found", args[0])
			}
		} else {
			// Interactive picker
			issues, err := client.AssignedIssues(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch issues: %w", err)
			}
			if len(issues) == 0 {
				fmt.Println("No assigned issues found.")
				return nil
			}
			issue = tui.PickIssue(issues)
			if issue == nil {
				return nil // cancelled
			}
		}

		return notes.OpenTask(cfg, issue)
	},
}
