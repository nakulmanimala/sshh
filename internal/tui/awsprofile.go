package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// awsProfileModel lets the user pick which AWS CLI profile to sync from.
type awsProfileModel struct {
	profiles []string
	cursor   int
	done     bool
	picked   bool
}

func newAWSProfileModel(profiles []string) awsProfileModel {
	return awsProfileModel{profiles: profiles}
}

func (m awsProfileModel) Init() tea.Cmd {
	return nil
}

func (m awsProfileModel) Update(msg tea.Msg) (awsProfileModel, tea.Cmd) {
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
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.profiles) > 0 {
				m.done = true
				m.picked = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m awsProfileModel) View() string {
	if len(m.profiles) == 0 {
		return titleStyle.Render("AWS Sync — Select Profile") + "\n\n" +
			helpStyle.Render("No AWS profiles found in ~/.aws/config or ~/.aws/credentials.") + "\n\n" +
			helpStyle.Render("Press Esc to go back")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("AWS Sync — Select Profile"))
	b.WriteString("\n\n")

	for i, p := range m.profiles {
		cursor := "  "
		name := p
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			name = selectedStyle.Render(name)
		}
		b.WriteString(cursor + name + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: move | Enter: select | Esc: cancel"))
	return b.String()
}

// SelectedProfile returns the profile the user picked.
func (m awsProfileModel) SelectedProfile() string {
	if m.cursor < 0 || m.cursor >= len(m.profiles) {
		return ""
	}
	return m.profiles[m.cursor]
}
