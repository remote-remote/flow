// Package cmd is the CLI layer.
package cmd

import (
	"fmt"
	"os"
	"strings"

	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var tuiSelection string

var rootCommand = &cobra.Command{
	Use:   "flow",
	Short: "A devtool for keeping flow in a terminal.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var t string

		fmt.Printf("RunE")

		if len(args) == 0 {
			fmt.Printf("no args")
			t = ""
		} else {
			fmt.Printf("some args %s", args[0])
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
}
