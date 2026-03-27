package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/remote-remote/flow/internal/remind"
)

type remindPhase int

const (
	remindPickDuration remindPhase = iota
	remindInputCustom
	remindInputMessage
)

type durationItem struct {
	label    string
	duration time.Duration
}

func (i durationItem) Title() string       { return i.label }
func (i durationItem) Description() string { return "" }
func (i durationItem) FilterValue() string { return i.label }

type remindModel struct {
	phase     remindPhase
	list      list.Model
	input     textinput.Model
	custom    textinput.Model
	duration  time.Duration
	result    *RemindResult
	err       error
	customErr string
	width     int
	height    int
}

type RemindResult struct {
	Duration time.Duration
	Message  string
}

func newRemindModel() remindModel {
	placeholderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))

	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "check deployment, standup, etc."
	ti.CharLimit = 100
	ti.SetWidth(60)
	tiStyles := ti.Styles()
	tiStyles.Focused.Placeholder = placeholderStyle
	tiStyles.Blurred.Placeholder = placeholderStyle
	ti.SetStyles(tiStyles)
	ti.Blur()

	ci := textinput.New()
	ci.Prompt = ""
	ci.Placeholder = "e.g. 30m, 1h30m, 3:30pm, 15:04"
	ci.CharLimit = 20
	ci.SetWidth(40)
	ciStyles := ci.Styles()
	ciStyles.Focused.Placeholder = placeholderStyle
	ciStyles.Blurred.Placeholder = placeholderStyle
	ci.SetStyles(ciStyles)
	ci.Blur()

	items := []list.Item{
		durationItem{label: "5 minutes", duration: 5 * time.Minute},
		durationItem{label: "10 minutes", duration: 10 * time.Minute},
		durationItem{label: "15 minutes", duration: 15 * time.Minute},
		durationItem{label: "30 minutes", duration: 30 * time.Minute},
		durationItem{label: "1 hour", duration: 1 * time.Hour},
		durationItem{label: "Custom...", duration: 0},
	}

	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	l := list.New(items, d, 0, 0)
	l.Title = "Set Reminder"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return remindModel{
		phase:  remindPickDuration,
		list:   l,
		input:  ti,
		custom: ci,
	}
}

func (m *remindModel) setSize(w, h int) {
	m.width = w
	m.height = h
	dh, dv := docStyle.GetFrameSize()
	m.list.SetSize(w-dh, h-dv)
}

func (m remindModel) Init() tea.Cmd {
	return nil
}

func (m remindModel) View() tea.View {
	var s string

	switch m.phase {
	case remindPickDuration:
		s = docStyle.Render(m.list.View())

	case remindInputCustom:
		s = "\n  Set Reminder — Custom Time\n\n"
		s += fmt.Sprintf("  %s\n\n", m.custom.View())
		if m.customErr != "" {
			s += fmt.Sprintf("  %s\n\n", m.customErr)
		}
		s += "  enter to confirm, esc to go back\n"

	case remindInputMessage:
		s = fmt.Sprintf("\n  Remind in %s\n\n", remind.FormatDuration(m.duration))
		s += fmt.Sprintf("  %s\n\n", m.input.View())
		s += "  enter to confirm, esc to go back\n"
	}

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m remindModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.phase {
		case remindPickDuration:
			if isBackKey(msg, m.list) {
				return m, func() tea.Msg { return BackMsg{} }
			}
			if msg.String() == "enter" || msg.String() == "space" {
				sel := m.list.SelectedItem()
				if sel == nil {
					return m, nil
				}
				di := sel.(durationItem)
				if di.duration == 0 {
					m.phase = remindInputCustom
					m.custom.Focus()
					return m, nil
				}
				m.duration = di.duration
				m.phase = remindInputMessage
				m.input.Focus()
				return m, nil
			}

		case remindInputCustom:
			switch msg.String() {
			case "enter":
				val := m.custom.Value()
				if val == "" {
					return m, nil
				}
				_, dur, err := remind.ParseTimeOrDuration(val)
				if err != nil {
					m.customErr = "invalid — use e.g. 30m, 1h30m, 3:30pm"
					return m, nil
				}
				m.duration = dur
				m.customErr = ""
				m.phase = remindInputMessage
				m.input.Focus()
				return m, nil
			case "esc":
				m.phase = remindPickDuration
				m.customErr = ""
				return m, nil
			}

		case remindInputMessage:
			switch msg.String() {
			case "enter":
				message := m.input.Value()
				if message == "" {
					message = "reminder"
				}
				m.result = &RemindResult{
					Duration: m.duration,
					Message:  message,
				}
				return m, nil
			case "esc":
				m.phase = remindPickDuration
				return m, nil
			}
		}
	}

	// Forward to active sub-component
	switch m.phase {
	case remindPickDuration:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case remindInputMessage:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	case remindInputCustom:
		var cmd tea.Cmd
		m.custom, cmd = m.custom.Update(msg)
		return m, cmd
	}

	return m, nil
}
