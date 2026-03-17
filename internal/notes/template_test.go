package notes

import (
	"strings"
	"testing"
	"time"
)

func TestRenderDailyTemplate(t *testing.T) {
	date := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	got := RenderDailyTemplate(date)

	checks := []string{
		"date: 2026-03-16",
		"# Monday, March 16, 2026",
		"## Tasks",
		"## Notes",
	}

	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("template missing %q\ngot:\n%s", want, got)
		}
	}
}
