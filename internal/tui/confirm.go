package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// confirmModel is a centered modal confirmation dialog with selectable buttons.
type confirmModel struct {
	prompt    string
	confirmed bool
	done      bool
	selected  int // 0 = Confirm, 1 = Cancel
	width     int
	height    int
}

func newConfirmModel(prompt string) confirmModel {
	return confirmModel{
		prompt:   prompt,
		selected: 1, // Cancel pre-selected (safer default)
	}
}

func (m *confirmModel) resize(w, h int) {
	m.width = w
	m.height = h
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.selected = 0
		case "right", "l":
			m.selected = 1
		case "enter":
			m.confirmed = (m.selected == 0)
			m.done = true
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
	// Button styles.
	confirmFocused := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(colorDanger).
		Padding(0, 2)

	confirmMuted := lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 2)

	cancelFocused := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(colorMuted).
		Padding(0, 2)

	cancelMuted := lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 2)

	var confirmBtn, cancelBtn string
	if m.selected == 0 {
		confirmBtn = confirmFocused.Render("Confirm")
		cancelBtn = cancelMuted.Render("Cancel")
	} else {
		confirmBtn = confirmMuted.Render("Confirm")
		cancelBtn = cancelFocused.Render("Cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirmBtn, "   ", cancelBtn)

	promptLine := dangerStyle.Render(m.prompt)
	hintLine := helpStyle.Render("←/→ move  enter select  y confirm  n/esc cancel")

	innerWidth := lipgloss.Width(buttons)
	if pw := lipgloss.Width(promptLine); pw > innerWidth {
		innerWidth = pw
	}

	padded := func(s string) string {
		w := lipgloss.Width(s)
		if w < innerWidth {
			s = s + strings.Repeat(" ", innerWidth-w)
		}
		return s
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		padded(promptLine),
		"",
		padded(buttons),
		"",
		padded(hintLine),
	)

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorDanger).
		Padding(1, 3).
		Render(content)

	if m.width == 0 || m.height == 0 {
		return box
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
