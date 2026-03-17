/*
Package tui is the UI layer
*/
package tui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

// MenuResult is returned from the root menu to tell the cmd layer what to do.
type MenuResult struct {
	// Action is the selected command key (e.g. "note:daily", "standup", "config").
	// Empty if the user selected a flow handled inline (work, note:task).
	Action string

	// WorkResult is set when the work flow completed inline.
	WorkResult *WorkResult

	// TaskIssue is set when the task note flow completed inline.
	TaskIssue interface{ GetIdentifier() string }

	// Issue is the resolved issue from an inline task/work flow.
	Issue interface{}
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type rootPhase int

const (
	rootMenu rootPhase = iota
	rootDelegated
)

type rootModel struct {
	phase    rootPhase
	list     list.Model
	delegate tea.Model // sub-model for work/task flows
	result   MenuResult
	width    int
	height   int
	page     string
	startFn  func(string) IssueStartedMsg
}

// Menu runs the root menu. Actions that have their own TUI (work, task note)
// run inline — no screen flash. Other actions return an Action string for the cmd layer.
func Menu(page string, startFn func(string) IssueStartedMsg) MenuResult {
	m := rootModel{
		phase:   rootMenu,
		page:    page,
		startFn: startFn,
	}
	m.initList()

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return MenuResult{}
	}
	return finalModel.(rootModel).result
}

func (m *rootModel) initList() {
	m.list = list.New(nil, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = "FLOW"

	switch m.page {
	case "":
		m.list.SetItems([]list.Item{
			item{title: "Work", desc: "Pick a task to work on", key: "work"},
			item{title: "Notes", desc: "Work with notes", key: "note"},
			item{title: "Standup", desc: "Generate standup from yesterday's work", key: "standup"},
			item{title: "Configure", desc: "Configure Flow", key: "config"},
		})
	case "note":
		m.list.SetItems([]list.Item{
			item{title: "Task note", key: "note:task", desc: "Open a note for a Linear task"},
			item{title: "Daily note", key: "note:daily", desc: "Open today's daily note"},
		})
	}
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) View() tea.View {
	if m.phase == rootDelegated && m.delegate != nil {
		return m.delegate.View()
	}
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If delegated to a sub-model, forward everything there
	if m.phase == rootDelegated && m.delegate != nil {
		// Check for BackMsg before forwarding
		if _, ok := msg.(BackMsg); ok {
			return m.returnToMenu()
		}

		updated, cmd := m.delegate.Update(msg)
		m.delegate = updated

		// Check if sub-model is done (quit command)
		if isQuitCmd(cmd) {
			return m.collectDelegateResult()
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if isBackKey(msg, m.list) && m.page != "" {
			m.page = ""
			m.list.SetItems([]list.Item{
				item{title: "Work", desc: "Pick a task to work on", key: "work"},
				item{title: "Notes", desc: "Work with notes", key: "note"},
				item{title: "Standup", desc: "Generate standup from yesterday's work", key: "standup"},
				item{title: "Configure", desc: "Configure Flow", key: "config"},
			})
			return m, nil
		}
		if msg.String() == "enter" || msg.String() == "space" {
			return m.handleSelection()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m rootModel) handleSelection() (tea.Model, tea.Cmd) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return m, nil
	}
	key := sel.(item).key

	switch key {
	case "note":
		m.page = "note"
		m.list.SetItems([]list.Item{
			item{title: "Task note", key: "note:task", desc: "Open a note for a Linear task"},
			item{title: "Daily note", key: "note:daily", desc: "Open today's daily note"},
		})
		return m, nil

	case "work":
		return m.delegateToWork()

	case "note:task":
		return m.delegateToTaskPicker()

	default:
		// Non-TUI actions: exit and let cmd layer handle
		m.result = MenuResult{Action: key}
		return m, tea.Quit
	}
}

func (m rootModel) delegateToWork() (tea.Model, tea.Cmd) {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	sub := workModel{
		phase:   workLoadingProjects,
		spinner: s,
		startFn: m.startFn,
		width:   m.width,
		height:  m.height,
	}
	m.phase = rootDelegated
	m.delegate = sub
	return m, sub.Init()
}

func (m rootModel) delegateToTaskPicker() (tea.Model, tea.Cmd) {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	sub := taskPickerModel{
		phase:   taskLoading,
		spinner: s,
		width:   m.width,
		height:  m.height,
	}
	m.phase = rootDelegated
	m.delegate = sub
	return m, sub.Init()
}

func (m rootModel) returnToMenu() (tea.Model, tea.Cmd) {
	m.phase = rootMenu
	m.delegate = nil
	m.initList()
	h, v := docStyle.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)
	return m, nil
}

func (m rootModel) collectDelegateResult() (tea.Model, tea.Cmd) {
	switch sub := m.delegate.(type) {
	case workModel:
		if sub.selected != nil {
			m.result = MenuResult{
				Action:     "work:done",
				WorkResult: &WorkResult{Issue: sub.selected, Dirty: sub.dirty},
			}
		}
	case taskPickerModel:
		if sub.selected != nil {
			m.result = MenuResult{
				Action: "note:task:done",
				Issue:  sub.selected,
			}
		}
	}
	return m, tea.Quit
}

// isQuitCmd checks if a tea.Cmd is tea.Quit.
// We detect quit by running the cmd and checking if it produces a tea.QuitMsg.
func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	return ok
}
