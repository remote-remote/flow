package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/remote-remote/flow/internal/remind"
)

// ReminderPopupResult describes how the user responded to a fired reminder popup.
type ReminderPopupResult struct {
	Action string        // "dismiss", "snooze", "repeat", or "" if nothing chosen
	Snooze time.Duration // populated when Action == "snooze"
}

type popupPhase int

const (
	popupPickAction popupPhase = iota
	popupCustomSnooze
)

type reminderPopupModel struct {
	phase   popupPhase
	message string
	origDur time.Duration
	input   textinput.Model
	err     string
	result  *ReminderPopupResult
	width   int
	height  int
}

// RunReminderPopup shows the fired-reminder popup TUI and returns the user's choice.
// Intended to be invoked inside a tmux popup, but works in any terminal.
func RunReminderPopup(message string, origDur time.Duration) (ReminderPopupResult, error) {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))

	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "30m, 1h, 2h30m"
	ti.CharLimit = 20
	ti.SetWidth(18)
	styles := ti.Styles()
	styles.Focused.Placeholder = dim
	styles.Blurred.Placeholder = dim
	ti.SetStyles(styles)
	ti.Blur()

	m := reminderPopupModel{
		phase:   popupPickAction,
		message: message,
		origDur: origDur,
		input:   ti,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return ReminderPopupResult{}, err
	}
	fm := finalModel.(reminderPopupModel)
	if fm.result != nil {
		return *fm.result, nil
	}
	return ReminderPopupResult{}, nil
}

func (m reminderPopupModel) Init() tea.Cmd { return nil }

var (
	popupAccent     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	popupTitle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
	popupKey        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	popupLabel      = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	popupDim        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	popupErr        = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	popupSection    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	popupCardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(0, 2)
)

func (m reminderPopupModel) View() tea.View {
	w, h := m.width, m.height
	if w == 0 {
		w = 50
	}
	if h == 0 {
		h = 16
	}

	cardW := min(70, max(28, w-4))
	innerW := cardW - 6 // border (2) + padding (4)

	header := popupAccent.Render("⏰  ") + popupTitle.Render(truncate(m.message, innerW-4))

	var body string
	switch m.phase {
	case popupPickAction:
		actions := []string{
			actionLine("d", "dismiss"),
			actionLine("1", "snooze 5m"),
			actionLine("2", "snooze 10m"),
			actionLine("3", "snooze 15m"),
			actionLine("s", "snooze custom…"),
		}
		if m.origDur > 0 {
			actions = append(actions,
				actionLine("r", "repeat ("+remind.FormatDuration(m.origDur)+")"))
		}
		body = lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			popupSection.Render("ACTIONS"),
			lipgloss.JoinVertical(lipgloss.Left, actions...),
		)

	case popupCustomSnooze:
		field := popupLabel.Render("snooze for  ") + m.input.View()
		lines := []string{
			header,
			"",
			popupSection.Render("CUSTOM SNOOZE"),
			field,
		}
		if m.err != "" {
			lines = append(lines, popupErr.Render(m.err))
		}
		lines = append(lines, "", popupDim.Render("enter confirm · esc back"))
		body = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	card := popupCardBorder.Width(cardW).Render(body)
	rendered := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, card)

	v := tea.NewView(rendered)
	v.AltScreen = true
	return v
}

func actionLine(k, label string) string {
	return fmt.Sprintf("%s  %s", popupKey.Render(k), popupLabel.Render(label))
}

func truncate(s string, n int) string {
	if n <= 1 {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func (m reminderPopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		iw := min(24, max(10, m.width-18))
		m.input.SetWidth(iw)
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.phase {
		case popupPickAction:
			switch msg.String() {
			case "d", "q", "esc", "enter":
				m.result = &ReminderPopupResult{Action: "dismiss"}
				return m, tea.Quit
			case "1":
				m.result = &ReminderPopupResult{Action: "snooze", Snooze: 5 * time.Minute}
				return m, tea.Quit
			case "2":
				m.result = &ReminderPopupResult{Action: "snooze", Snooze: 10 * time.Minute}
				return m, tea.Quit
			case "3":
				m.result = &ReminderPopupResult{Action: "snooze", Snooze: 15 * time.Minute}
				return m, tea.Quit
			case "s":
				m.phase = popupCustomSnooze
				m.input.Focus()
				return m, nil
			case "r":
				if m.origDur > 0 {
					m.result = &ReminderPopupResult{Action: "repeat"}
					return m, tea.Quit
				}
			}
			return m, nil

		case popupCustomSnooze:
			switch msg.String() {
			case "enter":
				val := m.input.Value()
				if val == "" {
					return m, nil
				}
				_, dur, err := remind.ParseTimeOrDuration(val)
				if err != nil {
					m.err = "invalid — try 30m, 1h30m"
					return m, nil
				}
				m.result = &ReminderPopupResult{Action: "snooze", Snooze: dur}
				return m, tea.Quit
			case "esc":
				m.phase = popupPickAction
				m.input.Blur()
				m.input.SetValue("")
				m.err = ""
				return m, nil
			}
		}
	}

	if m.phase == popupCustomSnooze {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

