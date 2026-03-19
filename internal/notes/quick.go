package notes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a title to a URL-friendly slug.
func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// OpenQuick creates a quick note in the vault and opens it in $EDITOR.
// If title is empty, a timestamp-based filename is used.
func OpenQuick(cfg *config.Config, title string) error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	var filename, displayTitle string
	if title == "" {
		displayTitle = now.Format("15:04 Note")
		filename = dateStr + "-" + now.Format("1504") + ".md"
	} else {
		displayTitle = title
		filename = dateStr + "-" + slugify(title) + ".md"
	}

	notePath := filepath.Join(cfg.VaultPath, "Notes", filename)

	// Create note if it doesn't exist
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
			return err
		}
		content := renderQuickTemplate(dateStr, displayTitle)
		if err := os.WriteFile(notePath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Cross-link to today's daily note
	if err := crossLinkQuickToDaily(cfg, filename, displayTitle); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not cross-link to daily note: %v\n", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, notePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func renderQuickTemplate(date, title string) string {
	return fmt.Sprintf(`---
date: %s
title: "%s"
tags: [note]
---
# %s

`, date, title, title)
}

func crossLinkQuickToDaily(cfg *config.Config, filename, title string) error {
	now := time.Now()
	dailyPath, err := config.DailyNotePath(cfg.VaultPath, now)
	if err != nil {
		return err
	}

	// Create daily note if it doesn't exist
	if _, err := os.Stat(dailyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dailyPath), 0o755); err != nil {
			return err
		}
		content := RenderDailyTemplate(now)
		if err := os.WriteFile(dailyPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return err
	}

	// Strip .md for wikilink
	linkName := strings.TrimSuffix(filename, ".md")
	wikilink := fmt.Sprintf("[[Notes/%s|%s]]", linkName, title)

	// Don't add duplicate links
	if strings.Contains(string(data), wikilink) {
		return nil
	}

	// Find ## Notes section and append the wikilink
	content := string(data)
	notesIdx := strings.Index(content, "## Notes")
	if notesIdx == -1 {
		content += "\n## Notes\n- " + wikilink + "\n"
	} else {
		afterNotes := notesIdx + len("## Notes")
		nextLine := strings.Index(content[afterNotes:], "\n")
		if nextLine == -1 {
			content += "\n- " + wikilink + "\n"
		} else {
			insertAt := afterNotes + nextLine + 1
			content = content[:insertAt] + "- " + wikilink + "\n" + content[insertAt:]
		}
	}

	return os.WriteFile(dailyPath, []byte(content), 0o644)
}
