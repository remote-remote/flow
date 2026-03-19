package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// BackMsg signals that the user wants to go up a level.
type BackMsg struct{}

// isBackKey returns true if the key should trigger navigation back.
// Only triggers when the list is not actively filtering.
func isBackKey(msg tea.KeyPressMsg, l list.Model) bool {
	key := msg.String()
	if key == "esc" {
		return l.FilterState() == list.Unfiltered
	}
	if key == "-" {
		return l.FilterState() == list.Unfiltered
	}
	return false
}
