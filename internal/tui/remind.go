package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/remind"
)

type remindPhase int

const (
	remindPickDuration remindPhase = iota
	remindInputMessage
	remindInputCustom
)

type remindModel struct {
	phase     remindPhase
	durations []time.Duration
	labels    []string
	cursor    int
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
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Blur()

	ci := textinput.New()
	ci.Prompt = ""
	ci.Placeholder = "e.g. 30m, 1h30m, 3:30pm, 15:04"
	ci.CharLimit = 20
	ci.Blur()

	return remindModel{
		phase:  remindPickDuration,
		input:  ti,
		custom: ci,
		durations: []time.Duration{
			5 * time.Minute,
			10 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			0, // sentinel for "Custom"
		},
		labels: []string{"5m", "10m", "15m", "30m", "1h", "Custom..."},
	}
}

func (m remindModel) Init() tea.Cmd {
	return nil
}

func (m remindModel) View() tea.View {
	var s string

	switch m.phase {
	case remindPickDuration:
		s = "\n  Set Reminder\n\n"
		for i, label := range m.labels {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			s += fmt.Sprintf("  %s%s\n", cursor, label)
		}
		s += "\n  j/k to move, enter to select, - to go back\n"

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
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.phase {
		case remindPickDuration:
			switch msg.String() {
			case "j", "down":
				if m.cursor < len(m.durations)-1 {
					m.cursor++
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "enter":
				if m.durations[m.cursor] == 0 {
					// Custom option
					m.phase = remindInputCustom
					m.custom.Focus()
					return m, nil
				}
				m.duration = m.durations[m.cursor]
				m.phase = remindInputMessage
				m.input.Focus()
				return m, nil
			case "-":
				return m, func() tea.Msg { return BackMsg{} }
			}
			return m, nil

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

	// Forward to active text input
	switch m.phase {
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
