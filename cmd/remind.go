package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/remind"
	"github.com/spf13/cobra"
)

var remindCmd = &cobra.Command{
	Use:   "remind <duration|time> <message>",
	Short: "Set a reminder that pops up in tmux",
	Long:  "Set a timer or clock time. When it fires, a tmux popup shows the message.\nExamples: flow remind 30m \"check deployment\"\n          flow remind 1h30m \"standup\"\n          flow remind 3:30pm \"team sync\"\n          flow remind 15:04 \"deploy window\"",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fireAt, duration, err := remind.ParseTimeOrDuration(args[0])
		if err != nil {
			return err
		}
		_ = fireAt // spawnReminder computes fireAt from duration

		message := strings.Join(args[1:], " ")

		if err := spawnReminder(duration, message); err != nil {
			return err
		}
		return nil
	},
}

var remindListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show active reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		reminders, err := remind.Active()
		if err != nil {
			return err
		}
		if len(reminders) == 0 {
			fmt.Println("No active reminders.")
			return nil
		}
		for _, r := range reminders {
			remaining := time.Until(r.FireAt)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Printf("#%d  %s  (fires in %s at %s)\n",
				r.ID, r.Message, remind.FormatDuration(remaining), r.FireAt.Format("15:04"))
		}
		return nil
	},
}

var remindCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a reminder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid reminder ID: %s", args[0])
		}
		if err := remind.Cancel(id); err != nil {
			return err
		}
		fmt.Printf("Reminder #%d cancelled.\n", id)
		return nil
	},
}

var remindClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Cancel all reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := remind.CancelAll(); err != nil {
			return err
		}
		fmt.Println("All reminders cleared.")
		return nil
	},
}

func init() {
	remindCmd.AddCommand(remindListCmd)
	remindCmd.AddCommand(remindCancelCmd)
	remindCmd.AddCommand(remindClearCmd)
}
