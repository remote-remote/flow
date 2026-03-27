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
	Use:    "_fire <unix_timestamp> <duration_secs> <message>",
	Hidden: true,
	Args:   cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ts, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		durSecs, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return err
		}
		fireAt := time.Unix(ts, 0)
		message := strings.Join(args[2:], " ")

		// Sleep until fire time
		remaining := time.Until(fireAt)
		if remaining > 0 {
			time.Sleep(remaining)
		}

		// Check if tmux is running
		if err := exec.Command("tmux", "info").Run(); err != nil {
			return nil
		}

		self, err := selfExecutable()
		if err != nil {
			return err
		}

		// Fire tmux popup with duration info
		popupCmd := fmt.Sprintf("%s _popup reminder %d %q", self, durSecs, message)
		exec.Command("tmux", "display-popup", "-E", "-w", "50", "-h", "16", popupCmd).Run()

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
	Use:    "reminder <duration_secs> <message>",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		durSecs, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		origDur := time.Duration(durSecs) * time.Second
		message := args[1]

		// Box drawing popup
		w := 46
		border := strings.Repeat("─", w)
		pad := func(s string) string {
			// pad to width
			vis := len(s)
			if vis < w {
				return s + strings.Repeat(" ", w-vis)
			}
			return s
		}

		fmt.Printf("  ┌%s┐\n", border)
		fmt.Printf("  │%s│\n", pad(""))
		fmt.Printf("  │%s│\n", pad("  ⏰  "+message))
		fmt.Printf("  │%s│\n", pad(""))
		fmt.Printf("  │%s│\n", pad("  ─── Actions ───"))
		fmt.Printf("  │%s│\n", pad(""))
		fmt.Printf("  │%s│\n", pad("  d  dismiss"))
		fmt.Printf("  │%s│\n", pad("  1  snooze 5m"))
		fmt.Printf("  │%s│\n", pad("  2  snooze 10m"))
		fmt.Printf("  │%s│\n", pad("  3  snooze 15m"))
		fmt.Printf("  │%s│\n", pad("  s  snooze custom..."))
		if origDur > 0 {
			fmt.Printf("  │%s│\n", pad(fmt.Sprintf("  r  repeat (%s)", remind.FormatDuration(origDur))))
		}
		fmt.Printf("  │%s│\n", pad(""))
		fmt.Printf("  └%s┘\n", border)
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
		case "s":
			fmt.Print("  duration (e.g. 30m, 1h, 2h30m): ")
			var input string
			fmt.Scanln(&input)
			d, err := time.ParseDuration(input)
			if err != nil {
				fmt.Printf("  invalid duration: %s\n", input)
				return nil
			}
			return spawnReminder(d, message)
		case "r":
			if origDur > 0 {
				return spawnReminder(origDur, message)
			}
			return nil
		default:
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

	durSecs := strconv.FormatInt(int64(duration.Seconds()), 10)
	proc := exec.Command(self, "_fire", strconv.FormatInt(fireAt.Unix(), 10), durSecs, message)
	proc.SysProcAttr = sysProcAttr()
	proc.Stdout = nil
	proc.Stderr = nil
	proc.Stdin = nil
	if err := proc.Start(); err != nil {
		return err
	}

	id, err := remind.Add(proc.Process.Pid, message, fireAt, duration)
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
	remind.WithLock(func() error {
		reminders, err := remind.Load()
		if err != nil {
			return err
		}
		pid := pidSelf()
		filtered := reminders[:0]
		for _, r := range reminders {
			if r.PID == pid || (r.Message == message && time.Until(r.FireAt) <= 0) {
				continue
			}
			filtered = append(filtered, r)
		}
		return remind.Save(filtered)
	})
}

func pidSelf() int {
	return os.Getpid()
}
