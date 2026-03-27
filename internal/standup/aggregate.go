package standup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/remote-remote/flow/internal/config"
	"github.com/remote-remote/flow/internal/github"
	"github.com/remote-remote/flow/internal/linear"
	"github.com/remote-remote/flow/internal/notes"
)

var identifierRe = regexp.MustCompile(`^([A-Z]+-\d+)`)

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

	var data StandupData
	seenYesterday := make(map[string]bool) // track identifiers to dedup

	// Linear: issues worked on since yesterday (in progress + recently completed)
	if issues, err := linear.IssuesWorkedSince(yesterday); err == nil {
		for _, iss := range issues {
			seenYesterday[iss.Identifier] = true
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
			cat := iss.State.Type
			if cat == "completed" || cat == "canceled" {
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

	// Notes: task wikilinks from yesterday's daily note (skip if already seen)
	if links := taskLinksFromDaily(cfg, yesterday); len(links) > 0 {
		for _, link := range links {
			id := identifierRe.FindString(link)
			if id != "" && seenYesterday[id] {
				continue
			}
			if id != "" {
				seenYesterday[id] = true
			}
			data.Yesterday = append(data.Yesterday, Item{
				Text:   link,
				URL:    resolveURLFromText(link),
				Source: "notes",
			})
		}
	}

	// Notes: recently modified task notes (skip if already seen)
	if items := recentTaskNotes(cfg, yesterday); len(items) > 0 {
		for _, item := range items {
			id := identifierRe.FindString(item)
			if id != "" && seenYesterday[id] {
				continue
			}
			if id != "" {
				seenYesterday[id] = true
			}
			data.Yesterday = append(data.Yesterday, Item{
				Text:   item,
				URL:    resolveURLFromText(item),
				Source: "notes",
			})
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
	return notes.TaskNotePathByID(vaultPath, identifier)
}

// resolveURLFromText extracts an identifier from text like "ENG-42: Fix thing" and resolves its URL.
func resolveURLFromText(text string) string {
	if m := identifierRe.FindString(text); m != "" {
		return resolveURLByIdentifier(m)
	}
	return ""
}

func resolveURLByIdentifier(identifier string) string {
	if full, err := linear.IssueByIdentifier(identifier); err == nil && full.URL != "" {
		return full.URL
	}
	return ""
}
