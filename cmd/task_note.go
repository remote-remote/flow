package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var taskNote = &cobra.Command{
	Use:   "task",
	Short: "Open a task note",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("task note command — coming in Phase 2")
		return nil
	},
}
