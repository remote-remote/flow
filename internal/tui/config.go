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
	err      string
	finished bool
}

func ConfigWizard() error {
	m := configModel{state: stateVaultInput}

	// Pre-populate from existing config
	if existing, err := config.Load(); err == nil {
		m.input = existing.VaultPath
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	fm := finalModel.(configModel)
	if !fm.finished {
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
		s = fmt.Sprintf("Error: %s\n\n> %s█", m.err, m.input)
	case stateDone:
		s = fmt.Sprintf("Config saved! Vault: %s\n", m.input)
	}

	v := tea.NewView("\n  " + s + "\n")
	return v
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.PasteMsg:
		m.input += msg.Content
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
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

func (m configModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateVaultInput:
		return m.validateAndSave()
	case stateError:
		m.input = ""
		m.state = stateVaultInput
		return m, nil
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
