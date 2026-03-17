package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/remote-remote/flow/internal/config"
)

type configState int

const (
	stateVaultInput configState = iota
	stateError
	stateDone
)

type configModel struct {
	state    configState
	input    string
	cursor   int
	err      string
	finished bool
}

func ConfigWizard() error {
	p := tea.NewProgram(configModel{state: stateVaultInput})
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	m := finalModel.(configModel)
	if !m.finished {
		return fmt.Errorf("configuration cancelled")
	}
	return nil
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) View() tea.View {
	var s string

	switch m.state {
	case stateVaultInput:
		s = "Obsidian vault path: " + m.input + "█\n\n(enter to confirm, ctrl+c to cancel)"
	case stateError:
		s = fmt.Sprintf("Error: %s\n\nObsidian vault path: %s█", m.err, m.input)
	case stateDone:
		s = fmt.Sprintf("Config saved! Vault: %s\n", m.input)
	}

	v := tea.NewView("\n  " + s + "\n")
	return v
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.validateAndSave()
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m configModel) validateAndSave() (tea.Model, tea.Cmd) {
	path := m.input

	// Expand ~
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	// Check .obsidian/ exists
	obsidianDir := filepath.Join(path, ".obsidian")
	if _, err := os.Stat(obsidianDir); os.IsNotExist(err) {
		m.state = stateError
		m.err = fmt.Sprintf(".obsidian/ not found in %s — is this an Obsidian vault?", path)
		return m, nil
	}

	cfg := &config.Config{VaultPath: path}
	if err := config.Save(cfg); err != nil {
		m.state = stateError
		m.err = err.Error()
		return m, nil
	}

	m.input = path
	m.state = stateDone
	m.finished = true
	return m, tea.Quit
}
