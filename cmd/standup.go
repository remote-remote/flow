package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/notes"
	"github.com/remote-remote/flow/internal/standup"
	"github.com/spf13/cobra"
)

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate standup from yesterday's work",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			if errors.Is(err, config.ErrNotConfigured) {
				fmt.Println("Flow is not configured yet. Run `flow config` to set up.")
				return nil
			}
			return err
		}

		fmt.Print("Gathering standup data...")
		data := standup.Aggregate(cfg, time.Now())
		fmt.Print("\r                         \r")

		md := standup.Format(data)
		fmt.Print(md)

		if err := clipboard.WriteAll(md); err != nil {
			fmt.Printf("\n(could not copy to clipboard: %v)\n", err)
		} else {
			fmt.Println("\nCopied to clipboard!")
		}

		if err := notes.AppendStandup(cfg, md); err != nil {
			fmt.Printf("(could not append to daily note: %v)\n", err)
		} else {
			fmt.Println("Appended to daily note.")
		}

		// Open yesterday's and today's daily notes in a split
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		if now.Weekday() == time.Monday {
			yesterday = now.AddDate(0, 0, -3)
		}

		todayPath, err := config.DailyNotePath(cfg.VaultPath, now)
		if err != nil {
			return err
		}
		yesterdayPath, err := config.DailyNotePath(cfg.VaultPath, yesterday)
		if err != nil {
			return err
		}

		// Ensure both files exist
		for _, p := range []string{todayPath, yesterdayPath} {
			if _, err := os.Stat(p); os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					return err
				}
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		c := exec.Command(editor, "-O", yesterdayPath, todayPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
