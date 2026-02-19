package tui

import (
	"fmt"
	"strconv"
	"strings"

	"sshh/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Field indices for the tunnel form (navigation order).
const (
	tFieldName       = 0
	tFieldSSHHost    = 1
	tFieldSSHUser    = 2
	tFieldSSHPort    = 3
	tFieldSSHKey     = 4
	tFieldType       = 5 // virtual selector — not a textinput
	tFieldLocalPort  = 6
	tFieldRemoteHost = 7
	tFieldRemotePort = 8
	tFieldCount      = 9

	tInputCount = 8 // number of actual textinputs (all fields except type)
)

var tFieldLabels = [tFieldCount]string{
	"Name:", "SSH Host:", "SSH User:", "SSH Port:", "SSH Key:",
	"Type:", "Local Port:", "Remote Host:", "Remote Port:",
}

var tunnelTypeOptions = []model.TunnelType{
	model.TunnelLocal,
	model.TunnelRemote,
	model.TunnelDynamic,
}

// tInputIdx maps a focused field index to a textinput array index.
// Returns -1 for the virtual type selector field.
func tInputIdx(field int) int {
	if field == tFieldType {
		return -1
	}
	if field > tFieldType {
		return field - 1
	}
	return field
}

// tunnelFormModel handles add/edit tunnel forms.
type tunnelFormModel struct {
	inputs     [tInputCount]textinput.Model
	tunnelType model.TunnelType
	focused    int
	title      string
	editing    bool
	index      int
	done       bool
	saved      bool
}

func newTunnelFormModel(title string, t *model.Tunnel, index int) tunnelFormModel {
	m := tunnelFormModel{
		title:      title,
		editing:    t != nil,
		index:      index,
		tunnelType: model.TunnelLocal,
	}

	for i := 0; i < tInputCount; i++ {
		inp := textinput.New()
		inp.Prompt = ""
		inp.CharLimit = 256
		m.inputs[i] = inp
	}

	// Placeholders (input array indices 0-7 map to fields 0-4, 6-8).
	m.inputs[0].Placeholder = "my-tunnel"
	m.inputs[1].Placeholder = "server.example.com"
	m.inputs[2].Placeholder = "root"
	m.inputs[3].Placeholder = "22"
	m.inputs[4].Placeholder = "~/.ssh/id_rsa (optional)"
	m.inputs[5].Placeholder = "8080"
	m.inputs[6].Placeholder = "db.internal (not needed for dynamic)"
	m.inputs[7].Placeholder = "5432 (not needed for dynamic)"

	if t != nil {
		m.inputs[0].SetValue(t.Name)
		m.inputs[1].SetValue(t.SSHHost)
		m.inputs[2].SetValue(t.SSHUser)
		sshPort := t.SSHPort
		if sshPort == 0 {
			sshPort = 22
		}
		m.inputs[3].SetValue(strconv.Itoa(sshPort))
		m.inputs[4].SetValue(t.SSHKey)
		m.tunnelType = t.Type
		if t.LocalPort > 0 {
			m.inputs[5].SetValue(strconv.Itoa(t.LocalPort))
		}
		if t.RemoteHost != "" {
			m.inputs[6].SetValue(t.RemoteHost)
		}
		if t.RemotePort > 0 {
			m.inputs[7].SetValue(strconv.Itoa(t.RemotePort))
		}
	}

	m.inputs[0].Focus()
	return m
}

func (m tunnelFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tunnelFormModel) Update(msg tea.Msg) (tunnelFormModel, tea.Cmd) {
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
			m.focused = (m.focused + 1) % tFieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + tFieldCount) % tFieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focused == tFieldCount-1 {
				m.done = true
				m.saved = true
				return m, nil
			}
			m.focused++
			return m, m.updateFocus()
		case "left":
			if m.focused == tFieldType {
				m.cycleType(false)
				return m, nil
			}
		case "right":
			if m.focused == tFieldType {
				m.cycleType(true)
				return m, nil
			}
		}
	}

	// Forward to the focused textinput (skip if on type selector).
	idx := tInputIdx(m.focused)
	if idx >= 0 {
		var cmd tea.Cmd
		m.inputs[idx], cmd = m.inputs[idx].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *tunnelFormModel) cycleType(forward bool) {
	for i, t := range tunnelTypeOptions {
		if t == m.tunnelType {
			if forward {
				m.tunnelType = tunnelTypeOptions[(i+1)%len(tunnelTypeOptions)]
			} else {
				m.tunnelType = tunnelTypeOptions[(i-1+len(tunnelTypeOptions))%len(tunnelTypeOptions)]
			}
			return
		}
	}
	m.tunnelType = tunnelTypeOptions[0]
}

func (m *tunnelFormModel) updateFocus() tea.Cmd {
	idx := tInputIdx(m.focused)
	var cmds []tea.Cmd
	for i := 0; i < tInputCount; i++ {
		if i == idx {
			cmds = append(cmds, m.inputs[i].Focus())
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m tunnelFormModel) View() string {
	var b strings.Builder
	b.WriteString(tunnelTitleStyle.Render(m.title))
	b.WriteString("\n\n")

	for field := 0; field < tFieldCount; field++ {
		label := tunnelLabelStyle.Render(tFieldLabels[field])
		cursor := "  "
		if field == m.focused {
			cursor = focusedInputStyle.Render("> ")
		}

		if field == tFieldType {
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, m.renderTypeSelector()))
		} else {
			idx := tInputIdx(field)
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, m.inputs[idx].View()))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Tab/↑↓: navigate | ←/→: change type | Enter: next/save | Ctrl+S: save | Esc: cancel"))
	return b.String()
}

func (m tunnelFormModel) renderTypeSelector() string {
	var parts []string
	for _, t := range tunnelTypeOptions {
		if t == m.tunnelType {
			parts = append(parts, selectedStyle.Render(fmt.Sprintf("[ %s ]", t)))
		} else {
			parts = append(parts, helpStyle.Render(fmt.Sprintf("[ %s ]", t)))
		}
	}
	return strings.Join(parts, " ")
}

// ToTunnel converts the form inputs into a Tunnel struct.
func (m tunnelFormModel) ToTunnel() model.Tunnel {
	sshPort := 22
	if p, err := strconv.Atoi(strings.TrimSpace(m.inputs[3].Value())); err == nil && p > 0 {
		sshPort = p
	}
	localPort := 0
	if p, err := strconv.Atoi(strings.TrimSpace(m.inputs[5].Value())); err == nil && p > 0 {
		localPort = p
	}
	remotePort := 0
	if p, err := strconv.Atoi(strings.TrimSpace(m.inputs[7].Value())); err == nil && p > 0 {
		remotePort = p
	}

	return model.Tunnel{
		Name:       strings.TrimSpace(m.inputs[0].Value()),
		SSHHost:    strings.TrimSpace(m.inputs[1].Value()),
		SSHUser:    strings.TrimSpace(m.inputs[2].Value()),
		SSHPort:    sshPort,
		SSHKey:     strings.TrimSpace(m.inputs[4].Value()),
		Type:       m.tunnelType,
		LocalPort:  localPort,
		RemoteHost: strings.TrimSpace(m.inputs[6].Value()),
		RemotePort: remotePort,
	}
}
