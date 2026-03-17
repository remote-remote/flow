package cmd

import (
	"errors"
	"fmt"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/notes"
	"github.com/spf13/cobra"
)

var dailyNote = &cobra.Command{
	Use:   "daily",
	Short: "Open today's daily note",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}
		return notes.OpenDaily(cfg)
	},
}
