package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// awsUsernameModel prompts for the default SSH username applied to servers
// newly discovered via AWS sync.
type awsUsernameModel struct {
	input textinput.Model
	done  bool
	saved bool
}

func newAWSUsernameModel(current string) awsUsernameModel {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 256
	ti.Width = 40
	ti.Placeholder = "ec2-user"
	ti.SetValue(current)
	ti.Focus()
	return awsUsernameModel{input: ti}
}

func (m awsUsernameModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m awsUsernameModel) Update(msg tea.Msg) (awsUsernameModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			return m, nil
		case "enter":
			if strings.TrimSpace(m.input.Value()) != "" {
				m.done = true
				m.saved = true
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m awsUsernameModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("AWS Sync — Default Username"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("SSH username applied to newly discovered EC2 instances."))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("User:") + " " + m.input.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter: save | Esc: cancel"))
	return b.String()
}

// Username returns the trimmed entered username.
func (m awsUsernameModel) Username() string {
	return strings.TrimSpace(m.input.Value())
}
