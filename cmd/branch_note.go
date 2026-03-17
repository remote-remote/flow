package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var branchNote = &cobra.Command{
	Use:   "branch",
	Short: "Open branch note",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("branch note command")
		return nil
	},
}

func init() {}
