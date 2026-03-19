package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
	tui "github.com/remote-remote/flow/internal/tui"
	"github.com/spf13/cobra"
)

var issueIDRe = regexp.MustCompile(`(?i)[A-Z]+-\d+`)

var taskNote = &cobra.Command{
	Use:   "task [identifier]",
	Short: "Open a task note for a Linear issue",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		// Try to resolve identifier from args or branch
		var identifier string
		if len(args) == 1 {
			identifier = strings.ToUpper(args[0])
		} else {
			identifier = identifierFromBranch()
		}

		// If we have an identifier and the note already exists, skip the API call
		if identifier != "" {
			notePath := notes.TaskNotePath(cfg.VaultPath, identifier)
			if _, err := os.Stat(notePath); err == nil {
				return notes.OpenExistingTask(notePath)
			}
		}

		var issue *linear.Issue

		if identifier != "" {
			issue, err = tui.RunTaskPickerForIdentifier(identifier)
		} else {
			issue, err = tui.RunTaskPicker()
		}

		if err != nil {
			return err
		}
		if issue == nil {
			return nil
		}

		return notes.OpenTask(cfg, issue)
	},
}

func identifierFromBranch() string {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	return strings.ToUpper(issueIDRe.FindString(branch))
}
