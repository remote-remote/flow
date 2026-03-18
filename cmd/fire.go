package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/remind"
	"github.com/spf13/cobra"
)

// _fire is a hidden command used by remind to fire a popup after sleeping.
var fireCmd = &cobra.Command{
	Use:    "_fire <unix_timestamp> <message>",
	Hidden: true,
	Args:   cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ts, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		fireAt := time.Unix(ts, 0)
		message := strings.Join(args[1:], " ")

		// Sleep until fire time
		remaining := time.Until(fireAt)
		if remaining > 0 {
			time.Sleep(remaining)
		}

		// Check if tmux is running
		if err := exec.Command("tmux", "info").Run(); err != nil {
			// tmux not running, exit silently
			return nil
		}

		// Find our own executable for the popup command
		self, err := selfExecutable()
		if err != nil {
			return err
		}

		// Fire tmux popup
		popupCmd := fmt.Sprintf("%s _popup reminder %q", self, message)
		exec.Command("tmux", "display-popup", "-E", "-w", "60", "-h", "10", popupCmd).Run()

		// Clean up: remove ourselves from the reminders list
		cleanupFiredReminder(message)

		return nil
	},
}

// _popup renders content inside a tmux popup.
var popupCmd = &cobra.Command{
	Use:    "_popup",
	Hidden: true,
}

var popupReminderCmd = &cobra.Command{
	Use:    "reminder <message>",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := args[0]
		fmt.Printf("\n  ⏰ Reminder\n\n  %s\n\n", message)
		fmt.Println("  [d] Dismiss")
		fmt.Println("  [1] Snooze 5m")
		fmt.Println("  [2] Snooze 10m")
		fmt.Println("  [3] Snooze 15m")
		fmt.Print("\n  > ")

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			return spawnReminder(5*time.Minute, message)
		case "2":
			return spawnReminder(10*time.Minute, message)
		case "3":
			return spawnReminder(15*time.Minute, message)
		default:
			// dismiss
			return nil
		}
	},
}

func spawnReminder(duration time.Duration, message string) error {
	fireAt := time.Now().Add(duration)
	self, err := os.Executable()
	if err != nil {
		return err
	}

	proc := exec.Command(self, "_fire", strconv.FormatInt(fireAt.Unix(), 10), message)
	proc.SysProcAttr = sysProcAttr()
	proc.Stdout = nil
	proc.Stderr = nil
	proc.Stdin = nil
	if err := proc.Start(); err != nil {
		return err
	}

	id, err := remind.Add(proc.Process.Pid, message, fireAt)
	if err != nil {
		return err
	}
	proc.Process.Release()

	fmt.Printf("Reminder #%d set: %q in %s (at %s)\n",
		id, message, remind.FormatDuration(duration), fireAt.Format("15:04"))
	return nil
}

func init() {
	popupCmd.AddCommand(popupReminderCmd)
}

func selfExecutable() (string, error) {
	return os.Executable()
}

func cleanupFiredReminder(message string) {
	reminders, err := remind.Load()
	if err != nil {
		return
	}
	pid := pidSelf()
	filtered := reminders[:0]
	for _, r := range reminders {
		if r.PID == pid || (r.Message == message && time.Until(r.FireAt) <= 0) {
			continue
		}
		filtered = append(filtered, r)
	}
	remind.Save(filtered)
}

func pidSelf() int {
	return os.Getpid()
}
