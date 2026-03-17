package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/standup"
	"github.com/spf13/cobra"
)

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate standup from yesterday's work",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		data := standup.Aggregate(cfg, time.Now())

		md := standup.Format(data)
		fmt.Print(md)

		if err := clipboard.WriteAll(md); err != nil {
			fmt.Printf("\n(could not copy to clipboard: %v)\n", err)
		} else {
			fmt.Println("\nCopied to clipboard!")
		}

		return nil
	},
}
