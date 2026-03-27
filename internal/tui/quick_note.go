package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type quickNoteModel struct {
	input  textinput.Model
	title  string
	done   bool
	err    error
	width  int
	height int
}

// QuickNotePrompt runs a standalone TUI to prompt for a quick note title.
func QuickNotePrompt() MenuResult {
	m := newQuickNoteModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return MenuResult{Err: err}
	}
	fm := finalModel.(quickNoteModel)
	if fm.done {
		return MenuResult{
			Action:         "note:quick:done",
			QuickNoteTitle: fm.title,
		}
	}
	return MenuResult{}
}

func newQuickNoteModel() quickNoteModel {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "meeting notes, idea, etc."
	ti.CharLimit = 120
	ti.SetWidth(60)

	// Make placeholder visible on dark backgrounds
	styles := ti.Styles()
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	ti.SetStyles(styles)

	ti.Focus()

	return quickNoteModel{input: ti}
}

func (m quickNoteModel) Init() tea.Cmd {
	return nil
}

func (m quickNoteModel) View() tea.View {
	s := "\n  Quick Note\n\n"
	s += fmt.Sprintf("  Title: %s\n\n", m.input.View())
	s += "  enter to create, esc to go back\n"

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m quickNoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch msg.String() {
		case "enter":
			m.title = m.input.Value()
			m.done = true
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
