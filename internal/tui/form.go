package tui

import (
	"fmt"
	"strconv"
	"strings"

	"ssh-tool/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	fieldName = iota
	fieldHost
	fieldUser
	fieldPort
	fieldKey
	fieldTags
	fieldCount
)

var fieldLabels = [fieldCount]string{
	"Name:", "Host:", "User:", "Port:", "Key:", "Tags:",
}

// formModel handles add/edit server forms.
type formModel struct {
	inputs  [fieldCount]textinput.Model
	focused int
	title   string
	editing bool // true if editing an existing server
	index   int  // index of server being edited (-1 for new)
	done    bool
	saved   bool
}

func newFormModel(title string, s *model.Server, index int) formModel {
	m := formModel{
		title:   title,
		editing: s != nil,
		index:   index,
	}

	for i := 0; i < fieldCount; i++ {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 256
		m.inputs[i] = t
	}

	m.inputs[fieldName].Placeholder = "my-server"
	m.inputs[fieldHost].Placeholder = "192.168.1.1"
	m.inputs[fieldUser].Placeholder = "root"
	m.inputs[fieldPort].Placeholder = "22"
	m.inputs[fieldKey].Placeholder = "~/.ssh/id_rsa (optional)"
	m.inputs[fieldTags].Placeholder = "web, prod (optional, comma-separated)"

	if s != nil {
		m.inputs[fieldName].SetValue(s.Name)
		m.inputs[fieldHost].SetValue(s.Host)
		m.inputs[fieldUser].SetValue(s.User)
		m.inputs[fieldPort].SetValue(strconv.Itoa(s.Port))
		m.inputs[fieldKey].SetValue(s.Key)
		m.inputs[fieldTags].SetValue(strings.Join(s.Tags, ", "))
	}

	m.inputs[m.focused].Focus()
	return m
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (formModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			return m, nil
		case "ctrl+s":
			m.done = true
			m.saved = true
			return m, nil
		case "tab", "down":
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focused == fieldCount-1 {
				// Last field: save.
				m.done = true
				m.saved = true
				return m, nil
			}
			m.focused++
			return m, m.updateFocus()
		}
	}

	// Update the focused input.
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m *formModel) updateFocus() tea.Cmd {
	var cmds []tea.Cmd
	for i := 0; i < fieldCount; i++ {
		if i == m.focused {
			cmds = append(cmds, m.inputs[i].Focus())
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m formModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	for i := 0; i < fieldCount; i++ {
		label := labelStyle.Render(fieldLabels[i])
		input := m.inputs[i].View()
		cursor := "  "
		if i == m.focused {
			cursor = focusedInputStyle.Render("> ")
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, input))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Tab/Shift+Tab: navigate | Enter: next/save | Ctrl+S: save | Esc: cancel"))
	return b.String()
}

// ToServer converts the form inputs into a Server struct.
func (m formModel) ToServer() model.Server {
	port := 22
	if p, err := strconv.Atoi(strings.TrimSpace(m.inputs[fieldPort].Value())); err == nil && p > 0 {
		port = p
	}

	var tags []string
	raw := strings.TrimSpace(m.inputs[fieldTags].Value())
	if raw != "" {
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	return model.Server{
		Name: strings.TrimSpace(m.inputs[fieldName].Value()),
		Host: strings.TrimSpace(m.inputs[fieldHost].Value()),
		User: strings.TrimSpace(m.inputs[fieldUser].Value()),
		Port: port,
		Key:  strings.TrimSpace(m.inputs[fieldKey].Value()),
		Tags: tags,
	}
}
