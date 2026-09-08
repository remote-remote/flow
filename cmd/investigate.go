package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/notes"
	"github.com/spf13/cobra"
)

const noOpenUsage = "print the resolved note path instead of opening $EDITOR"

// Flags shared by both scaffolding commands. They take what the caller already
// knows rather than inferring it, so flow stays usable without an agent.
type investigationFlags struct {
	noOpen  bool
	scratch string
	scopes  []string
	project string
	task    string
}

func (f *investigationFlags) register(cmd *cobra.Command, scoped bool) {
	cmd.Flags().BoolVar(&f.noOpen, "no-open", false, noOpenUsage)
	cmd.Flags().StringVar(&f.scratch, "scratch", "", "scratch directory holding oracle evidence")
	cmd.Flags().StringArrayVar(&f.scopes, "scope", nil, "oracle scope name (repeatable)")
	cmd.Flags().StringVar(&f.task, "task", "", "task this investigation is for")
	if !scoped {
		cmd.Flags().StringVar(&f.project, "project", "", "project this investigation is for")
	}
}

func (f *investigationFlags) apply(inv notes.Investigation) notes.Investigation {
	inv.Scratch = f.scratch
	inv.Scopes = f.scopes
	inv.Task = f.task
	if f.project != "" {
		inv.Project = f.project
	}
	return inv
}

var investigateFlags investigationFlags

var investigateCmd = &cobra.Command{
	Use:   "investigate <title...>",
	Short: "Scaffold an investigation note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigured()
		if err != nil || cfg == nil {
			return err
		}
		inv := investigateFlags.apply(notes.NewInvestigation(strings.Join(args, " ")))
		return notes.OpenInvestigation(cfg, inv, investigateFlags.noOpen)
	},
}

// loadConfigured returns a nil config with a nil error when flow is unconfigured,
// having already told the user how to fix that.
func loadConfigured() (*config.Config, error) {
	cfg, err := config.Load()
	if errors.Is(err, config.ErrNotConfigured) {
		fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
		return nil, nil
	}
	return cfg, err
}

func init() {
	investigateFlags.register(investigateCmd, false)
}
