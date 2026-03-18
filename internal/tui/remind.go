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
)

type remindModel struct {
	phase     remindPhase
	durations []time.Duration
	labels    []string
	cursor    int
	input     textinput.Model
	duration  time.Duration
	result    *RemindResult
	err       error
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

	return remindModel{
		phase: remindPickDuration,
		input: ti,
		durations: []time.Duration{
			5 * time.Minute,
			10 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
		},
		labels: []string{"5m", "10m", "15m", "30m", "1h"},
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
				m.duration = m.durations[m.cursor]
				m.phase = remindInputMessage
				m.input.Focus()
				return m, nil
			case "-":
				return m, func() tea.Msg { return BackMsg{} }
			}
			return m, nil

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
			case "escape":
				m.phase = remindPickDuration
				return m, nil
			}
		}
	}

	// Forward to textinput when in message phase
	if m.phase == remindInputMessage {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}
