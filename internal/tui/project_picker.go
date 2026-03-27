package tui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/linear"
)

type projectPickPhase int

const (
	projectPickLoading projectPickPhase = iota
	projectPickPicking
)

type projectPickerModel struct {
	phase    projectPickPhase
	spinner  spinner.Model
	list     list.Model
	selected string
	err      error
	width    int
	height   int
}

// ProjectPicker runs a standalone TUI to pick a Linear project.
func ProjectPicker() MenuResult {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	m := projectPickerModel{
		phase:   projectPickLoading,
		spinner: s,
	}

	p := tea.NewProgram(standaloneModel{inner: m})
	finalModel, err := p.Run()
	if err != nil {
		return MenuResult{Err: err}
	}

	fm := finalModel.(standaloneModel).inner.(projectPickerModel)
	if fm.err != nil {
		return MenuResult{Err: fm.err}
	}
	return MenuResult{ProjectName: fm.selected}
}

func (m projectPickerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			projects, err := linear.Projects()
			return projectsLoadedMsg{projects: projects, err: err}
		},
	)
}

func (m projectPickerModel) View() tea.View {
	var s string
	if m.err != nil {
		s = fmt.Sprintf("\n  Error: %s\n\n  Press any key to exit.\n", m.err)
	} else {
		switch m.phase {
		case projectPickLoading:
			s = fmt.Sprintf("\n  %s Loading projects...\n", m.spinner.View())
		case projectPickPicking:
			s = m.list.View()
		}
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m projectPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.phase == projectPickPicking {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.err != nil {
			return m, tea.Quit
		}
		if m.phase == projectPickPicking {
			if isBackKey(msg, m.list) {
				return m, tea.Quit
			}
			if (msg.String() == "enter" || msg.String() == "space") && m.list.FilterState() != list.Filtering {
				if sel := m.list.SelectedItem(); sel != nil {
					m.selected = sel.(projectItem).project.Name
					return m, tea.Quit
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case projectsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.projects) == 0 {
			m.err = fmt.Errorf("no projects found")
			return m, nil
		}
		items := make([]list.Item, len(msg.projects))
		for i, p := range msg.projects {
			items[i] = projectItem{project: p}
		}
		m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.list.Title = "Select Project"
		m.phase = projectPickPicking
		return m, nil

	default:
		if m.phase == projectPickPicking {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}
