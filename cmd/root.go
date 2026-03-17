// Package cmd is the CLI layer.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var tuiSelection string

var rootCommand = &cobra.Command{
	Use:   "flow",
	Short: "A devtool for keeping flow in a terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// First-run: check if configured
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Welcome to Flow! Let's set up your config first.")
				return tui.ConfigWizard()
			}
			return err
		}
		_ = cfg

		t := ""
		if len(args) > 0 {
			t = args[0]
		}

		tuiSelection = tui.Menu(t)
		return nil
	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if tuiSelection != "" {
		rootCommand.SetArgs(strings.Split(tuiSelection, ":"))
		if err := rootCommand.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCommand.AddCommand(noteCmd)
	rootCommand.AddCommand(configCmd)
	rootCommand.AddCommand(standupCmd)
}
