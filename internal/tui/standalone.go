package tui

import tea "charm.land/bubbletea/v2"

// standaloneModel wraps a sub-model and converts completion (selected/err set) into tea.Quit.
// Used when running work/task flows outside the root menu.
type standaloneModel struct {
	inner tea.Model
}

func (m standaloneModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m standaloneModel) View() tea.View {
	return m.inner.View()
}

func (m standaloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.inner.Update(msg)
	m.inner = updated

	// Check if the inner model is done
	switch sub := m.inner.(type) {
	case workModel:
		if sub.selected != nil || sub.err != nil {
			return m, tea.Quit
		}
	case taskPickerModel:
		if sub.selected != nil || sub.err != nil {
			return m, tea.Quit
		}
	}

	return m, cmd
}
