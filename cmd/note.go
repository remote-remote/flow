package cmd

import (
	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Work with notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := tui.Menu("note")

		switch result.Action {
		case "note:task:done":
			if issue, ok := result.Issue.(*linear.Issue); ok {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				return notes.OpenTask(cfg, issue)
			}
		case "note:daily":
			return dailyNote.RunE(cmd, nil)
		}

		return nil
	},
}

func init() {
	noteCmd.AddCommand(taskNote)
	noteCmd.AddCommand(dailyNote)
}
