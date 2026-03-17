package cmd

import (
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.ConfigWizard()
	},
}
