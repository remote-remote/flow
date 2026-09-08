package cmd

import (
	"strings"

	"github.com/remote-remote/flow/internal/notes"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Work with projects",
}

var projectPlanFlags investigationFlags

var projectPlanCmd = &cobra.Command{
	Use:   "plan <name...>",
	Short: "Scaffold an investigation into an existing project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigured()
		if err != nil || cfg == nil {
			return err
		}
		inv := projectPlanFlags.apply(notes.NewProjectPlan(strings.Join(args, " ")))
		return notes.OpenInvestigation(cfg, inv, projectPlanFlags.noOpen)
	},
}

func init() {
	projectPlanFlags.register(projectPlanCmd, true)
	projectCmd.AddCommand(projectPlanCmd)
}
