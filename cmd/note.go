package cmd

import (
	tui "github.com/remote-remote/flow/internal/tui"

	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Work with notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		tui.Menu("note")
		return nil
	},
}

func init() {
	noteCmd.AddCommand(branchNote)
	noteCmd.AddCommand(dailyNote)
}
