package tui

import (
	"fmt"
	"strings"

	"ssh-tool/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

// importModel lets the user select which SSH config hosts to import.
type importModel struct {
	servers  []model.Server
	selected []bool
	cursor   int
	done     bool
	imported bool
}

func newImportModel(servers []model.Server) importModel {
	sel := make([]bool, len(servers))
	// Select all by default.
	for i := range sel {
		sel[i] = true
	}
	return importModel{
		servers:  servers,
		selected: sel,
	}
}

func (m importModel) Init() tea.Cmd {
	return nil
}

func (m importModel) Update(msg tea.Msg) (importModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.servers)-1 {
				m.cursor++
			}
		case " ":
			if len(m.selected) > 0 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "enter":
			m.done = true
			m.imported = true
			return m, nil
		}
	}
	return m, nil
}

func (m importModel) View() string {
	if len(m.servers) == 0 {
		return titleStyle.Render("No hosts found in ~/.ssh/config") + "\n\n" +
			helpStyle.Render("Press Esc to go back")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Import from ~/.ssh/config"))
	b.WriteString("\n\n")

	for i, s := range m.servers {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
		}

		check := "[ ]"
		if m.selected[i] {
			check = successStyle.Render("[x]")
		}

		desc := fmt.Sprintf("%s@%s:%d", s.User, s.Host, s.Port)
		name := s.Name
		if i == m.cursor {
			name = selectedStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, check, name, helpStyle.Render(desc)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Space: toggle | Enter: import selected | Esc: cancel"))
	return b.String()
}

// SelectedServers returns only the servers that were selected.
func (m importModel) SelectedServers() []model.Server {
	var result []model.Server
	for i, s := range m.servers {
		if m.selected[i] {
			result = append(result, s)
		}
	}
	return result
}
