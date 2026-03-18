package remind

import (
	"testing"
	"time"
)

func TestAddAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := Add(99999, "test reminder", time.Now().Add(5*time.Minute))
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if id != 1 {
		t.Errorf("id = %d, want 1", id)
	}

	reminders, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reminders) != 1 {
		t.Fatalf("got %d reminders, want 1", len(reminders))
	}
	if reminders[0].Message != "test reminder" {
		t.Errorf("message = %q, want %q", reminders[0].Message, "test reminder")
	}
}

func TestRemove(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	Add(99999, "first", time.Now().Add(5*time.Minute))
	Add(99999, "second", time.Now().Add(10*time.Minute))

	if err := Remove(1); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	reminders, _ := Load()
	if len(reminders) != 1 {
		t.Fatalf("got %d reminders, want 1", len(reminders))
	}
	if reminders[0].Message != "second" {
		t.Errorf("remaining = %q, want %q", reminders[0].Message, "second")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{90 * time.Minute, "1h30m"},
		{2 * time.Hour, "2h"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
