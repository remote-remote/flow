// Package cmd is the CLI layer.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

// package variable for capturing args
var tuiSelection string

// Our root command
var rootCommand = &cobra.Command{
	Use:   "flow",
	Short: "A devtool for keeping flow in a terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Welcome to Flow! Let's set up your config first.")
				return tui.ConfigWizard()
			}
			return err
		}

		t := ""
		if len(args) > 0 {
			t = args[0]
		}

		result := tui.Menu(t)

		if result.Err != nil {
			return result.Err
		}

		// Handle inline results that completed inside the TUI
		switch result.Action {
		case "work:done":
			if result.WorkResult != nil {
				if result.WorkResult.Dirty {
					fmt.Println("Worktree is dirty — commit or stash to checkout the branch.")
				}
				return notes.OpenTask(cfg, result.WorkResult.Issue)
			}
		case "note:task:done":
			if issue, ok := result.Issue.(*linear.Issue); ok {
				return notes.OpenTask(cfg, issue)
			}
		case "remind:done":
			if result.RemindResult != nil {
				return spawnReminder(result.RemindResult.Duration, result.RemindResult.Message)
			}
		case "standup":
			tuiSelection = "standup"
		case "config":
			tuiSelection = "config"
		case "note:daily":
			tuiSelection = "note:daily"
		}

		return nil
	},
}

// Execute is the main entrypoint for the CLI
func Execute() {
	// try running rootCommand's Execute
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// if there are subcommands we fire that
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
	rootCommand.AddCommand(workCmd)
	rootCommand.AddCommand(remindCmd)
	rootCommand.AddCommand(fireCmd)
	rootCommand.AddCommand(popupCmd)
}
