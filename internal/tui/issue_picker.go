package tui

import (
	"fmt"
	"os"

	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/linear"
)

type issueItem struct {
	issue linear.Issue
}

func (i issueItem) Title() string       { return i.issue.Identifier + ": " + i.issue.Title }
func (i issueItem) Description() string { return i.issue.State.Name }
func (i issueItem) FilterValue() string { return i.issue.FilterValue() }

type issuePickerModel struct {
	list     list.Model
	selected *linear.Issue
	quitted  bool
}

// PickIssue shows a fuzzy-filterable list of issues and returns the selected one, or nil if cancelled.
func PickIssue(issues []linear.Issue) *linear.Issue {
	items := make([]list.Item, len(issues))
	for i, iss := range issues {
		items[i] = issueItem{issue: iss}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Issue"

	p := tea.NewProgram(issuePickerModel{list: l})
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	m := finalModel.(issuePickerModel)
	return m.selected
}

func (m issuePickerModel) Init() tea.Cmd {
	return nil
}

func (m issuePickerModel) View() tea.View {
	if m.quitted && m.selected != nil {
		v := tea.NewView(fmt.Sprintf("Selected: %s: %s\n", m.selected.Identifier, m.selected.Title))
		return v
	}
	v := tea.NewView(m.list.View())
	v.AltScreen = true
	return v
}

func (m issuePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitted = true
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			if sel := m.list.SelectedItem(); sel != nil {
				issue := sel.(issueItem).issue
				m.selected = &issue
				m.quitted = true
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
