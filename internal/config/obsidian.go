package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// momentToGo converts Moment.js format tokens to Go time layout tokens.
var momentToGo = strings.NewReplacer(
	"YYYY", "2006",
	"YY", "06",
	"MM", "01",
	"DD", "02",
	"ddd", "Mon",
	"dddd", "Monday",
)

type dailyNotesJSON struct {
	Folder string `json:"folder"`
	Format string `json:"format"`
}

type periodicNotesJSON struct {
	Daily struct {
		Folder string `json:"folder"`
		Format string `json:"format"`
	} `json:"daily"`
}

// DailyNotePath resolves the full filesystem path for a daily note on the given date.
// It checks Obsidian's daily-notes.json, then periodic-notes plugin config, then falls back to defaults.
func DailyNotePath(vaultPath string, date time.Time) (string, error) {
	folder, format := dailyNoteSettings(vaultPath)

	filename := date.Format(momentToGo.Replace(format)) + ".md"

	if folder != "" {
		return filepath.Join(vaultPath, folder, filename), nil
	}
	return filepath.Join(vaultPath, filename), nil
}

func dailyNoteSettings(vaultPath string) (folder, format string) {
	// Try core daily-notes.json
	if f, fmt, ok := readDailyNotesConfig(vaultPath); ok {
		return f, fmt
	}

	// Try periodic-notes plugin
	if f, fmt, ok := readPeriodicNotesConfig(vaultPath); ok {
		return f, fmt
	}

	// Fallback
	return "", "YYYY-MM-DD"
}

func readDailyNotesConfig(vaultPath string) (folder, format string, ok bool) {
	data, err := os.ReadFile(filepath.Join(vaultPath, ".obsidian", "daily-notes.json"))
	if err != nil {
		return "", "", false
	}
	var cfg dailyNotesJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", false
	}
	f := cfg.Format
	if f == "" {
		f = "YYYY-MM-DD"
	}
	return cfg.Folder, f, true
}

func readPeriodicNotesConfig(vaultPath string) (folder, format string, ok bool) {
	data, err := os.ReadFile(filepath.Join(vaultPath, ".obsidian", "plugins", "periodic-notes", "data.json"))
	if err != nil {
		return "", "", false
	}
	var cfg periodicNotesJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", false
	}
	f := cfg.Daily.Format
	if f == "" {
		f = "YYYY-MM-DD"
	}
	return cfg.Daily.Folder, f, true
}
