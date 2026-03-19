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
	Action string

	// WorkResult is set when the work flow completed inline.
	WorkResult *WorkResult

	// Issue is the resolved issue from an inline task/work flow.
	Issue any

	// RemindResult is set when a reminder was created inline.
	RemindResult *RemindResult

	// QuickNoteTitle is set when the quick note flow completed inline.
	QuickNoteTitle string

	// Err is set when a delegated sub-model encountered an error.
	Err error
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
}

// Menu runs the root menu. Actions that have their own TUI (work, task note)
// run inline — no screen flash. Other actions return an Action string for the cmd layer.
func Menu(page string) MenuResult {
	m := rootModel{
		phase: rootMenu,
		page:  page,
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
			item{title: "Remind", desc: "Set or manage reminders", key: "remind"},
			item{title: "Standup", desc: "Generate standup from yesterday's work", key: "standup"},
			item{title: "Configure", desc: "Configure Flow", key: "config"},
		})
	case "note":
		m.list.SetItems([]list.Item{
			item{title: "Task note", key: "note:task", desc: "Open a note for a Linear task"},
			item{title: "Daily note", key: "note:daily", desc: "Open today's daily note"},
			item{title: "Quick note", key: "note:quick", desc: "Create a quick titled note"},
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
		if _, ok := msg.(BackMsg); ok {
			return m.returnToMenu()
		}

		updated, cmd := m.delegate.Update(msg)
		m.delegate = updated

		if m.isDelegateComplete() {
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
		if isBackKey(msg, m.list) {
			if m.page != "" {
				m.page = ""
				m.initList()
				h, v := docStyle.GetFrameSize()
				m.list.SetSize(m.width-h, m.height-v)
				return m, nil
			}
			return m, tea.Quit
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
			item{title: "Quick note", key: "note:quick", desc: "Create a quick titled note"},
		})
		return m, nil

	case "work":
		return m.delegateToWork()

	case "note:task":
		return m.delegateToTaskPicker()

	case "note:quick":
		return m.delegateToQuickNote()

	case "remind":
		return m.delegateToRemind()

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

func (m rootModel) delegateToQuickNote() (tea.Model, tea.Cmd) {
	sub := newQuickNoteModel()
	sub.width = m.width
	sub.height = m.height
	m.phase = rootDelegated
	m.delegate = sub
	return m, sub.Init()
}

func (m rootModel) delegateToRemind() (tea.Model, tea.Cmd) {
	sub := newRemindModel()
	sub.width = m.width
	sub.height = m.height
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

func (m *rootModel) isDelegateComplete() bool {
	switch sub := m.delegate.(type) {
	case workModel:
		return sub.selected != nil || sub.err != nil
	case taskPickerModel:
		return sub.selected != nil || sub.err != nil
	case remindModel:
		return sub.result != nil || sub.err != nil
	case quickNoteModel:
		return sub.done || sub.err != nil
	}
	return false
}

func (m rootModel) collectDelegateResult() (tea.Model, tea.Cmd) {
	switch sub := m.delegate.(type) {
	case workModel:
		if sub.err != nil {
			m.result = MenuResult{Err: sub.err}
		} else if sub.selected != nil {
			m.result = MenuResult{
				Action:     "work:done",
				WorkResult: &WorkResult{Issue: sub.selected, Dirty: sub.dirty},
			}
		}
	case taskPickerModel:
		if sub.err != nil {
			m.result = MenuResult{Err: sub.err}
		} else if sub.selected != nil {
			m.result = MenuResult{
				Action: "note:task:done",
				Issue:  sub.selected,
			}
		}
	case remindModel:
		if sub.err != nil {
			m.result = MenuResult{Err: sub.err}
		} else if sub.result != nil {
			m.result = MenuResult{
				Action:       "remind:done",
				RemindResult: sub.result,
			}
		}
	case quickNoteModel:
		if sub.err != nil {
			m.result = MenuResult{Err: sub.err}
		} else if sub.done {
			m.result = MenuResult{
				Action:         "note:quick:done",
				QuickNoteTitle: sub.title,
			}
		}
	}
	return m, tea.Quit
}
