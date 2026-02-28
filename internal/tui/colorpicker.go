package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type colorPickerTarget int

const (
	colorPickerTargetSSH colorPickerTarget = iota
	colorPickerTargetTunnel
)

type colorPickerAction int

const (
	colorPickerActionNone colorPickerAction = iota
	colorPickerActionSelect
	colorPickerActionCancel
)

// palette is a curated set of vibrant ANSI 256 colors.
var palette = []lipgloss.Color{
	// Blues
	"27", "33", "39", "45", "69", "75",
	// Greens
	"34", "40", "46", "82", "118", "120",
	// Cyans
	"51", "87", "123", "159", "14", "50",
	// Purples
	"99", "105", "135", "141", "147", "57",
	// Pinks / Magentas
	"201", "205", "213", "219", "207", "171",
	// Reds / Oranges / Yellows
	"196", "203", "202", "208", "214", "220",
}

const pickerCols = 6

type colorPickerModel struct {
	target colorPickerTarget
	cursor int
	width  int
	height int
}

func newColorPickerModel(target colorPickerTarget, width, height int) colorPickerModel {
	// Pre-select the currently active color.
	current := colorPrimary
	if target == colorPickerTargetTunnel {
		current = colorAccent
	}
	cursor := 0
	for i, c := range palette {
		if c == current {
			cursor = i
			break
		}
	}
	return colorPickerModel{
		target: target,
		cursor: cursor,
		width:  width,
		height: height,
	}
}

func (m colorPickerModel) selectedColor() lipgloss.Color {
	return palette[m.cursor]
}

func (m colorPickerModel) Update(msg tea.Msg) (colorPickerModel, colorPickerAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(palette)-1 {
				m.cursor++
			}
		case "up":
			if m.cursor >= pickerCols {
				m.cursor -= pickerCols
			}
		case "down":
			if m.cursor+pickerCols < len(palette) {
				m.cursor += pickerCols
			}
		case "enter":
			return m, colorPickerActionSelect, nil
		case "esc", "ctrl+c":
			return m, colorPickerActionCancel, nil
		}
	}
	return m, colorPickerActionNone, nil
}

func (m colorPickerModel) View() string {
	accentColor := colorPrimary
	titleText := "Pick SSH Color"
	if m.target == colorPickerTargetTunnel {
		accentColor = colorAccent
		titleText = "Pick Tunnel Color"
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	dividerStyle := lipgloss.NewStyle().Foreground(colorMuted)
	helpTextStyle := lipgloss.NewStyle().Foreground(colorSecondary)

	// Build the swatch grid.
	rows := []string{}
	for row := 0; row*pickerCols < len(palette); row++ {
		var cells []string
		for col := 0; col < pickerCols; col++ {
			idx := row*pickerCols + col
			if idx >= len(palette) {
				break
			}
			c := palette[idx]
			swatch := lipgloss.NewStyle().
				Background(c).
				Padding(0, 2).
				Render("  ")

			if idx == m.cursor {
				swatch = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("255")).
					Render(swatch)
			} else {
				// Add spacing to match the border width of the selected swatch.
				swatch = lipgloss.NewStyle().
					Padding(1, 1).
					Render(swatch)
			}
			cells = append(cells, swatch)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center, cells...))
	}

	// Preview of selected color.
	sel := palette[m.cursor]
	previewSwatch := lipgloss.NewStyle().Background(sel).Padding(0, 2).Render("  ")
	previewLabel := lipgloss.NewStyle().Foreground(sel).Bold(true).Render(fmt.Sprintf("ANSI %s", string(sel)))
	preview := "  " + previewSwatch + "  " + previewLabel

	body := strings.Join([]string{
		titleStyle.Render(titleText),
		dividerStyle.Render("─────────────────────"),
		"",
		strings.Join(rows, "\n"),
		"",
		preview,
		"",
		helpTextStyle.Render("←→↑↓: navigate   enter: select   esc: cancel"),
	}, "\n")

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Render(body)

	// Center in terminal.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
