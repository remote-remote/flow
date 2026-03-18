// Package remind manages timer-based reminders that fire via tmux popups.
package remind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Reminder struct {
	ID      int       `json:"id"`
	PID     int       `json:"pid"`
	Message string    `json:"message"`
	FireAt  time.Time `json:"fire_at"`
}

func statePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "flow", "reminders.json"), nil
}

func Load() ([]Reminder, error) {
	p, err := statePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var reminders []Reminder
	if err := json.Unmarshal(data, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func Save(reminders []Reminder) error {
	p, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reminders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// Add persists a new reminder and returns its ID.
func Add(pid int, message string, fireAt time.Time) (int, error) {
	reminders, _ := Load()

	id := 1
	for _, r := range reminders {
		if r.ID >= id {
			id = r.ID + 1
		}
	}

	reminders = append(reminders, Reminder{
		ID:      id,
		PID:     pid,
		Message: message,
		FireAt:  fireAt,
	})

	return id, Save(reminders)
}

// Remove deletes a reminder by ID.
func Remove(id int) error {
	reminders, err := Load()
	if err != nil {
		return err
	}
	filtered := reminders[:0]
	for _, r := range reminders {
		if r.ID != id {
			filtered = append(filtered, r)
		}
	}
	return Save(filtered)
}

// Cancel kills the process for a reminder and removes it.
func Cancel(id int) error {
	reminders, err := Load()
	if err != nil {
		return err
	}
	for _, r := range reminders {
		if r.ID == id {
			// Try to kill the process
			if proc, err := os.FindProcess(r.PID); err == nil {
				proc.Signal(syscall.SIGTERM)
			}
			break
		}
	}
	return Remove(id)
}

// CancelAll kills all reminder processes and clears the list.
func CancelAll() error {
	reminders, err := Load()
	if err != nil {
		return err
	}
	for _, r := range reminders {
		if proc, err := os.FindProcess(r.PID); err == nil {
			proc.Signal(syscall.SIGTERM)
		}
	}
	return Save(nil)
}

// Prune removes reminders whose processes are no longer running.
func Prune() error {
	reminders, err := Load()
	if err != nil {
		return err
	}
	alive := reminders[:0]
	for _, r := range reminders {
		if isProcessAlive(r.PID) {
			alive = append(alive, r)
		}
	}
	return Save(alive)
}

// Active returns reminders that are still pending (process alive).
func Active() ([]Reminder, error) {
	if err := Prune(); err != nil {
		return nil, err
	}
	return Load()
}

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without actually sending a signal
	return proc.Signal(syscall.Signal(0)) == nil
}

// FormatDuration formats a duration for display (e.g. "25m", "1h30m").
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return strconv.Itoa(int(d.Seconds())) + "s"
	}
	if d < time.Hour {
		return strconv.Itoa(int(d.Minutes())) + "m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
