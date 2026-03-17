package standup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/github"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
)

type Item struct {
	Text   string
	URL    string
	Source string // "linear", "github", "notes"
}

type StandupData struct {
	Yesterday []Item
	Today     []Item
}

// Aggregate collects standup data from Linear, GitHub, and notes.
func Aggregate(cfg *config.Config, date time.Time) StandupData {
	yesterday := date.AddDate(0, 0, -1)
	// Skip weekends: if today is Monday, look back to Friday
	if date.Weekday() == time.Monday {
		yesterday = date.AddDate(0, 0, -3)
	}

	days := int(date.Sub(yesterday).Hours()/24) + 1
	since := fmt.Sprintf("%dd", days)

	var data StandupData

	// Linear: issues changed since yesterday
	if issues, err := linear.IssuesChangedSince(since); err == nil {
		for _, iss := range issues {
			data.Yesterday = append(data.Yesterday, Item{
				Text:   fmt.Sprintf("[%s] %s (%s)", iss.Identifier, iss.Title, iss.State.Name),
				URL:    iss.URL,
				Source: "linear",
			})
		}
	}

	// Linear: active/todo issues for today
	if issues, err := linear.AssignedIssues(); err == nil {
		for _, iss := range issues {
			state := iss.State.Name
			if state == "Done" || state == "Canceled" || state == "Cancelled" || state == "Duplicate" {
				continue
			}
			data.Today = append(data.Today, Item{
				Text:   fmt.Sprintf("[%s] %s", iss.Identifier, iss.Title),
				URL:    iss.URL,
				Source: "linear",
			})
		}
	}

	// GitHub: PRs
	if prs, err := github.PRsOpenedOrMerged(yesterday); err == nil {
		for _, pr := range prs {
			data.Yesterday = append(data.Yesterday, Item{
				Text:   fmt.Sprintf("PR: %s", pr.Title),
				URL:    pr.URL,
				Source: "github",
			})
		}
	}

	// Notes: task wikilinks from yesterday's daily note
	if links := taskLinksFromDaily(cfg, yesterday); len(links) > 0 {
		for _, link := range links {
			data.Yesterday = append(data.Yesterday, Item{
				Text:   link,
				Source: "notes",
			})
		}
	}

	// Notes: recently modified task notes
	if items := recentTaskNotes(cfg, yesterday); len(items) > 0 {
		for _, item := range items {
			dup := false
			for _, existing := range data.Yesterday {
				if strings.Contains(existing.Text, item) {
					dup = true
					break
				}
			}
			if !dup {
				data.Yesterday = append(data.Yesterday, Item{
					Text:   item,
					Source: "notes",
				})
			}
		}
	}

	return data
}

func taskLinksFromDaily(cfg *config.Config, date time.Time) []string {
	dailyPath, err := config.DailyNotePath(cfg.VaultPath, date)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(dailyPath)
	if err != nil {
		return nil
	}

	var links []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [[Tasks/") {
			if idx := strings.Index(line, "|"); idx != -1 {
				end := strings.Index(line[idx:], "]]")
				if end != -1 {
					links = append(links, line[idx+1:idx+end])
				}
			}
		}
	}
	return links
}

func recentTaskNotes(cfg *config.Config, since time.Time) []string {
	tasksDir := filepath.Join(cfg.VaultPath, "Tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}

	var items []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(since) {
			id := strings.TrimSuffix(entry.Name(), ".md")
			title := readTaskTitle(filepath.Join(tasksDir, entry.Name()))
			if title != "" {
				items = append(items, fmt.Sprintf("%s: %s", id, title))
			} else {
				items = append(items, id)
			}
		}
	}
	return items
}

func readTaskTitle(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			title := strings.TrimPrefix(line, "title:")
			title = strings.TrimSpace(title)
			title = strings.Trim(title, `"`)
			return title
		}
	}
	return ""
}

func TaskNotePath(vaultPath, identifier string) string {
	return notes.TaskNotePath(vaultPath, identifier)
}
