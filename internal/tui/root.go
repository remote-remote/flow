/*
Package tui is the UI layer
*/
package tui

import (
	"fmt"
	"os"

	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

/*
Menu runs a bubbletea menu and returns the selected command.
*/
func Menu(page string) string {
	p := tea.NewProgram(initMenu(page))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	m := finalModel.(model)
	return m.list.SelectedItem().(item).key
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list list.Model
}

/*
Bubbletea Stuff
*/
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "enter" || msg.String() == "space" {
			return m.handleSelection()
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

/*
Private APIs
*/
func initMenu(page string) model {
	var items []list.Item
	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "FLOW"

	switch page {
	case "":
		m.setRootItems()
	case "note":
		m.setNoteItems()
	}
	return m
}

func (m *model) setRootItems() {
	m.list.SetItems([]list.Item{
		item{title: "Notes", desc: "Work with notes", key: "note"},
		item{title: "Standup", desc: "Generate standup from yesterday's work", key: "standup"},
		item{title: "Configure", desc: "Configure Flow", key: "config"},
	})
}

func (m *model) setNoteItems() {
	m.list.SetItems([]list.Item{
		item{title: "Task note", key: "note:task", desc: "Open a note for a Linear task"},
		item{title: "Daily note", key: "note:daily", desc: "Open today's daily note"},
	})
}

func (m model) handleSelection() (tea.Model, tea.Cmd) {
	key := m.list.SelectedItem().(item).key
	if key == "note" {
		m.setNoteItems()
		return m, nil
	} else {
		return m, tea.Quit
	}
}
