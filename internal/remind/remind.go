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

// lockPath returns the path for the lockfile.
func lockPath() (string, error) {
	p, err := statePath()
	if err != nil {
		return "", err
	}
	return p + ".lock", nil
}

// acquireLock acquires an advisory file lock and returns the lock file.
// The caller must call releaseLock when done.
func acquireLock() (*os.File, error) {
	lp, err := lockPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(lp), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lp, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("flock: %w", err)
	}
	return f, nil
}

// releaseLock releases and closes the lock file.
func releaseLock(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

// WithLock runs fn while holding the reminders file lock.
// Useful for external callers that need to do atomic Load+Save.
func WithLock(fn func() error) error {
	lk, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lk)
	return fn()
}

type Reminder struct {
	ID       int           `json:"id"`
	PID      int           `json:"pid"`
	Message  string        `json:"message"`
	FireAt   time.Time     `json:"fire_at"`
	Duration time.Duration `json:"duration,omitempty"` // original duration, for repeat
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
func Add(pid int, message string, fireAt time.Time, duration ...time.Duration) (int, error) {
	lk, err := acquireLock()
	if err != nil {
		return 0, err
	}
	defer releaseLock(lk)

	reminders, err := Load()
	if err != nil {
		return 0, err
	}

	id := 1
	for _, r := range reminders {
		if r.ID >= id {
			id = r.ID + 1
		}
	}

	r := Reminder{
		ID:      id,
		PID:     pid,
		Message: message,
		FireAt:  fireAt,
	}
	if len(duration) > 0 {
		r.Duration = duration[0]
	}
	reminders = append(reminders, r)

	return id, Save(reminders)
}

// Remove deletes a reminder by ID.
func Remove(id int) error {
	lk, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lk)

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
	lk, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lk)

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
	// Inline the remove logic to reuse the lock
	filtered := reminders[:0]
	for _, r := range reminders {
		if r.ID != id {
			filtered = append(filtered, r)
		}
	}
	return Save(filtered)
}

// CancelAll kills all reminder processes and clears the list.
func CancelAll() error {
	lk, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lk)

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
	lk, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lk)

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

// ParseTimeOrDuration parses either a duration ("30m", "1h30m") or a clock time
// ("3:30pm", "15:30", "3pm"). Clock times resolve to today, or tomorrow if already past.
func ParseTimeOrDuration(s string) (fireAt time.Time, duration time.Duration, err error) {
	// Try as Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(d), d, nil
	}

	// Try clock time formats
	now := time.Now()
	formats := []string{"3:04pm", "3:04PM", "3pm", "3PM", "15:04"}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err != nil {
			continue
		}
		// Build time for today
		fireAt = time.Date(now.Year(), now.Month(), now.Day(),
			t.Hour(), t.Minute(), 0, 0, now.Location())
		if fireAt.Before(now) {
			fireAt = fireAt.Add(24 * time.Hour)
		}
		return fireAt, fireAt.Sub(now), nil
	}

	return time.Time{}, 0, fmt.Errorf("invalid duration or time %q (e.g. 30m, 1h30m, 3:30pm, 15:04)", s)
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
