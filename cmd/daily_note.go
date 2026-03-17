package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dailyNote = &cobra.Command{
	Use:   "daily",
	Short: "Open daily note",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("daily note command")
		return nil
	},
}

func init() {}
