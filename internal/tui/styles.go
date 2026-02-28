package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (ANSI 256 for broad terminal support) — reassigned by applyTheme.
	colorPrimary   = lipgloss.Color("39")  // blue (SSH)
	colorSecondary = lipgloss.Color("245") // gray
	colorAccent    = lipgloss.Color("214") // orange (Tunnel)
	colorDanger    = lipgloss.Color("196") // red
	colorSuccess   = lipgloss.Color("40")  // green
	colorMuted     = lipgloss.Color("240") // dark gray
)

// Style vars — populated by rebuildStyles() via init() and applyTheme().
var (
	titleStyle       lipgloss.Style
	statusStyle      lipgloss.Style
	helpStyle        lipgloss.Style
	selectedStyle    lipgloss.Style
	tagStyle         lipgloss.Style
	dangerStyle      lipgloss.Style
	successStyle     lipgloss.Style
	labelStyle       lipgloss.Style
	focusedInputStyle lipgloss.Style
	blurredInputStyle lipgloss.Style

	tunnelTitleStyle lipgloss.Style
	tunnelLabelStyle lipgloss.Style

	searchBoxFocusedStyle       lipgloss.Style
	searchBoxBlurredStyle       lipgloss.Style
	tunnelSearchBoxFocusedStyle lipgloss.Style

	tableHeaderStyle       lipgloss.Style
	tableSelectedStyle     lipgloss.Style
	tunnelTableHeaderStyle lipgloss.Style
	tunnelTableSelectedStyle lipgloss.Style
)

func init() {
	rebuildStyles()
}

// applyTheme reassigns the primary/accent colors and rebuilds all styles.
func applyTheme(sshColor, tunnelColor lipgloss.Color) {
	colorPrimary = sshColor
	colorAccent = tunnelColor
	rebuildStyles()
}

func rebuildStyles() {
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		PaddingLeft(1)

	statusStyle = lipgloss.NewStyle().
		Foreground(colorMuted).
		PaddingLeft(1)

	helpStyle = lipgloss.NewStyle().
		Foreground(colorSecondary).
		PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	tagStyle = lipgloss.NewStyle().
		Foreground(colorAccent)

	dangerStyle = lipgloss.NewStyle().
		Foreground(colorDanger).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(colorSuccess)

	labelStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Width(12)

	focusedInputStyle = lipgloss.NewStyle().
		Foreground(colorPrimary)

	blurredInputStyle = lipgloss.NewStyle().
		Foreground(colorSecondary)

	tunnelTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent).
		PaddingLeft(1)

	tunnelLabelStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Width(13)

	searchBoxFocusedStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	searchBoxBlurredStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(0, 1)

	tunnelSearchBoxFocusedStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		BorderBottom(true)

	tableSelectedStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Background(lipgloss.Color("236"))

	tunnelTableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		BorderBottom(true)

	tunnelTableSelectedStyle = lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Background(lipgloss.Color("236"))
}
