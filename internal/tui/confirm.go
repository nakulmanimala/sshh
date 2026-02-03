package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// confirmModel is a simple y/n confirmation dialog.
type confirmModel struct {
	prompt    string
	confirmed bool
	done      bool
}

func newConfirmModel(prompt string) confirmModel {
	return confirmModel{prompt: prompt}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			m.done = true
		case "n", "N", "esc":
			m.confirmed = false
			m.done = true
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	return dangerStyle.Render(m.prompt) + " " + helpStyle.Render("[y/n]")
}
