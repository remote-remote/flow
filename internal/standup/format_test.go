package standup

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	data := StandupData{
		Yesterday: []Item{
			{Text: "[ENG-1] Fixed bug (Done)", URL: "https://linear.app/eng-1"},
			{Text: "PR: Review auth changes", URL: "https://github.com/pr/1"},
		},
		Today: []Item{
			{Text: "[ENG-2] Deploy fix", URL: "https://linear.app/eng-2"},
		},
	}

	got := Format(data)

	if !strings.Contains(got, "    - https://linear.app/eng-1") {
		t.Errorf("missing linear URL, got:\n%s", got)
	}
	if !strings.Contains(got, "    - https://github.com/pr/1") {
		t.Errorf("missing github URL, got:\n%s", got)
	}
	if !strings.Contains(got, "    - https://linear.app/eng-2") {
		t.Errorf("missing today URL, got:\n%s", got)
	}
	if !strings.Contains(got, "- Yesterday") {
		t.Errorf("missing Yesterday heading, got:\n%s", got)
	}
	if !strings.Contains(got, "- Today") {
		t.Errorf("missing Today heading, got:\n%s", got)
	}
}

func TestFormat_Empty(t *testing.T) {
	got := Format(StandupData{})
	if !strings.Contains(got, "- (none)") {
		t.Error("empty sections should show (none)")
	}
}

func TestFormat_NoURL(t *testing.T) {
	data := StandupData{
		Yesterday: []Item{{Text: "ENG-5: Some task from notes"}},
	}
	got := Format(data)
	if !strings.Contains(got, "    - ENG-5: Some task from notes") {
		t.Errorf("plain text item missing, got:\n%s", got)
	}
}
