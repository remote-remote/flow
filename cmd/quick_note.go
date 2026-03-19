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

var quickNote = &cobra.Command{
	Use:   "quick [title...]",
	Short: "Create and open a quick note",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		title := strings.Join(args, " ")

		// No title provided — show TUI prompt
		if title == "" {
			result := tui.QuickNotePrompt()
			if result.Err != nil {
				return result.Err
			}
			title = result.QuickNoteTitle
		}

		return notes.OpenQuick(cfg, title)
	},
}
