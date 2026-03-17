package tui

import (
	"fmt"
	"os"

	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/linear"
)

type projectItem struct {
	project linear.Project
}

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return "" }
func (i projectItem) FilterValue() string { return i.project.FilterValue() }

type projectPickerModel struct {
	list     list.Model
	selected *linear.Project
}

// PickProject shows a fuzzy-filterable list of projects and returns the selected one.
func PickProject(projects []linear.Project) *linear.Project {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{project: p}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Project"

	p := tea.NewProgram(projectPickerModel{list: l})
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	m := finalModel.(projectPickerModel)
	return m.selected
}

func (m projectPickerModel) Init() tea.Cmd {
	return nil
}

func (m projectPickerModel) View() tea.View {
	v := tea.NewView(m.list.View())
	v.AltScreen = true
	return v
}

func (m projectPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			if sel := m.list.SelectedItem(); sel != nil {
				proj := sel.(projectItem).project
				m.selected = &proj
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
