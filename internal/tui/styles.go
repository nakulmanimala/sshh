package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (ANSI 256 for broad terminal support).
	colorPrimary   = lipgloss.Color("39")  // blue
	colorSecondary = lipgloss.Color("245") // gray
	colorAccent    = lipgloss.Color("214") // orange
	colorDanger    = lipgloss.Color("196") // red
	colorSuccess   = lipgloss.Color("40")  // green
	colorMuted     = lipgloss.Color("240") // dark gray

	// Title style.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			PaddingLeft(1)

	// Status bar at bottom.
	statusStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			PaddingLeft(1)

	// Help text.
	helpStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			PaddingLeft(1)

	// Selected item highlight.
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Tag style.
	tagStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Error / danger text.
	dangerStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	// Success text.
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Form label.
	labelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Width(12)

	// Focused input.
	focusedInputStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	// Blurred input.
	blurredInputStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)
)
