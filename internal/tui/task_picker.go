package tui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/linear"
)

type taskPhase int

const (
	taskLoading taskPhase = iota
	taskPicking
	taskFetchingDetails
)

type assignedLoadedMsg struct {
	issues []linear.Issue
	err    error
}

type issueDetailMsg struct {
	issue *linear.Issue
	err   error
}

type taskPickerModel struct {
	phase      taskPhase
	spinner    spinner.Model
	list       list.Model
	selected   *linear.Issue
	identifier string // set when fetching a known identifier directly
	err        error
	width      int
	height     int
}

// RunTaskPicker shows a spinner while fetching assigned issues, lets user pick,
// then fetches full details (with URL) before returning. TUI stays up the whole time.
func RunTaskPicker() (*linear.Issue, error) {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	inner := taskPickerModel{
		phase:   taskLoading,
		spinner: s,
	}

	p := tea.NewProgram(standaloneModel{inner: inner})
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(standaloneModel).inner.(taskPickerModel)
	if fm.err != nil {
		return nil, fm.err
	}
	return fm.selected, nil
}

// RunTaskPickerForIdentifier shows a spinner while fetching a specific issue.
// Keeps TUI up during the fetch so there's no flash.
func RunTaskPickerForIdentifier(identifier string) (*linear.Issue, error) {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	inner := taskPickerModel{
		phase:      taskFetchingDetails,
		spinner:    s,
		identifier: identifier,
	}

	p := tea.NewProgram(standaloneModel{inner: inner})
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(standaloneModel).inner.(taskPickerModel)
	if fm.err != nil {
		return nil, fm.err
	}
	return fm.selected, nil
}

func fetchIssueDetail(identifier string) tea.Cmd {
	return func() tea.Msg {
		issue, err := linear.IssueByIdentifier(identifier)
		return issueDetailMsg{issue: issue, err: err}
	}
}

func (m taskPickerModel) Init() tea.Cmd {
	if m.identifier != "" {
		return tea.Batch(
			m.spinner.Tick,
			fetchIssueDetail(m.identifier),
		)
	}
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			issues, err := linear.RecentIssues()
			return assignedLoadedMsg{issues: issues, err: err}
		},
	)
}

func (m taskPickerModel) View() tea.View {
	var s string
	if m.err != nil {
		s = fmt.Sprintf("\n  Error: %s\n\n  Press any key to go back.\n", m.err)
	} else {
		switch m.phase {
		case taskLoading:
			s = fmt.Sprintf("\n  %s Loading issues...\n", m.spinner.View())
		case taskPicking:
			s = m.list.View()
		case taskFetchingDetails:
			s = fmt.Sprintf("\n  %s Loading issue details...\n", m.spinner.View())
		}
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m taskPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.phase == taskPicking {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.phase == taskPicking {
			if isBackKey(msg, m.list) {
				return m, func() tea.Msg { return BackMsg{} }
			}
			if msg.String() == "enter" && m.list.FilterState() != list.Filtering {
				if sel := m.list.SelectedItem(); sel != nil {
					issue := sel.(issueItem).issue
					m.phase = taskFetchingDetails
					return m, tea.Batch(
						m.spinner.Tick,
						fetchIssueDetail(issue.Identifier),
					)
				}
				return m, nil
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

	case assignedLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.issues) == 0 {
			m.err = fmt.Errorf("no assigned issues found")
			return m, nil
		}
		items := make([]list.Item, len(msg.issues))
		for i, iss := range msg.issues {
			items[i] = issueItem{issue: iss}
		}
		m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.list.Title = "Select Issue"
		m.phase = taskPicking
		return m, nil

	case issueDetailMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.selected = msg.issue
		return m, nil

	default:
		// Forward unhandled messages (e.g. FilterMatchesMsg) to the list
		if m.phase == taskPicking {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}
