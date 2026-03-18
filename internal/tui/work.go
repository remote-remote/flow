package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/linear"
)

func gitDirty() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(out)) != ""
}

type workPhase int

const (
	workLoadingProjects workPhase = iota
	workPickProject
	workLoadingIssues
	workPickIssue
	workStartingIssue
)

type projectsLoadedMsg struct {
	projects []linear.Project
	err      error
}

type issuesLoadedMsg struct {
	issues []linear.Issue
	err    error
}

// IssueStartedMsg is the result of starting an issue (exported for cmd layer).
type IssueStartedMsg struct {
	issue *linear.Issue
	dirty bool
	err   error
}

type workModel struct {
	phase        workPhase
	spinner      spinner.Model
	list         list.Model
	projectItems []list.Item // saved for back navigation
	project      *linear.Project
	selected     *linear.Issue
	dirty        bool
	err          error
	width        int
	height       int
}

type WorkResult struct {
	Issue *linear.Issue
	Dirty bool
}

// RunWorkFlow shows a single TUI: load projects → pick → load issues → pick → start issue.
// Stays in alt-screen the entire time.
func RunWorkFlow() (*WorkResult, error) {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	inner := workModel{
		phase:   workLoadingProjects,
		spinner: s,
	}

	p := tea.NewProgram(standaloneModel{inner: inner})
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(standaloneModel).inner.(workModel)
	if fm.err != nil {
		return nil, fm.err
	}
	if fm.selected == nil {
		return nil, nil
	}
	return &WorkResult{Issue: fm.selected, Dirty: fm.dirty}, nil
}

// StartIssueResult creates an IssueStartedMsg — exported for use by cmd layer.
func StartIssueResult(identifier string, dirty bool) IssueStartedMsg {
	var startErr error
	if dirty {
		startErr = linear.StartIssue(identifier)
	} else {
		startErr = linear.StartIssueWithCheckout(identifier)
	}
	issue, err := linear.IssueByIdentifier(identifier)
	if err != nil && startErr == nil {
		startErr = err
	}
	return IssueStartedMsg{issue: issue, dirty: dirty, err: startErr}
}

func (m IssueStartedMsg) Issue() *linear.Issue { return m.issue }
func (m IssueStartedMsg) Dirty() bool          { return m.dirty }
func (m IssueStartedMsg) Err() error           { return m.err }

func (m workModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			projects, err := linear.Projects()
			return projectsLoadedMsg{projects: projects, err: err}
		},
	)
}

func (m workModel) View() tea.View {
	var s string
	switch m.phase {
	case workLoadingProjects:
		s = fmt.Sprintf("\n  %s Loading projects...\n", m.spinner.View())
	case workPickProject:
		s = m.list.View()
	case workLoadingIssues:
		s = fmt.Sprintf("\n  %s Loading issues for %s...\n", m.spinner.View(), m.project.Name)
	case workPickIssue:
		s = m.list.View()
	case workStartingIssue:
		s = fmt.Sprintf("\n  %s Starting issue...\n", m.spinner.View())
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m workModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.phase == workPickProject || m.phase == workPickIssue {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.phase == workPickProject || m.phase == workPickIssue {
			if isBackKey(msg, m.list) {
				return m.handleBack()
			}
			if msg.String() == "enter" {
				return m.handleSelection()
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
		m.projectItems = items
		m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.list.Title = "Select Project"
		m.phase = workPickProject
		return m, nil

	case issuesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.issues) == 0 {
			m.err = fmt.Errorf("no issues in this project")
			return m, nil
		}
		items := make([]list.Item, len(msg.issues))
		for i, iss := range msg.issues {
			items[i] = issueItem{issue: iss}
		}
		m.list = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.list.Title = fmt.Sprintf("Select Issue — %s", m.project.Name)
		m.phase = workPickIssue
		return m, nil

	case IssueStartedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.selected = msg.issue
		m.dirty = msg.dirty
		return m, nil
	}

	return m, nil
}

func (m workModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.phase {
	case workPickIssue:
		// Go back to project picker
		m.list = list.New(m.projectItems, list.NewDefaultDelegate(), m.width, m.height)
		m.list.Title = "Select Project"
		m.project = nil
		m.phase = workPickProject
		return m, nil
	case workPickProject:
		// Go back to root menu
		return m, func() tea.Msg { return BackMsg{} }
	}
	return m, nil
}

func (m workModel) handleSelection() (tea.Model, tea.Cmd) {
	sel := m.list.SelectedItem()
	if sel == nil {
		return m, nil
	}

	switch m.phase {
	case workPickProject:
		proj := sel.(projectItem).project
		m.project = &proj
		m.phase = workLoadingIssues
		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				issues, err := linear.ProjectIssues(proj.Name)
				return issuesLoadedMsg{issues: issues, err: err}
			},
		)
	case workPickIssue:
		issue := sel.(issueItem).issue
		m.phase = workStartingIssue
		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				return StartIssueResult(issue.Identifier, gitDirty())
			},
		)
	}

	return m, nil
}
