package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var projectNote = &cobra.Command{
	Use:   "project [name...]",
	Short: "Open a project note",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		name := strings.Join(args, " ")

		if name == "" {
			result := tui.ProjectPicker()
			if result.Err != nil {
				return result.Err
			}
			name = result.ProjectName
		}

		if name == "" {
			return nil
		}

		return notes.OpenProject(cfg, name)
	},
}
