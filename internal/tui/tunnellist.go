package tui

import (
	"fmt"

	"sshh/internal/model"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// tunnelItem wraps a Tunnel for use in the bubbles list.
type tunnelItem struct {
	tunnel model.Tunnel
	index  int
}

func (t tunnelItem) Title() string       { return t.tunnel.Name }
func (t tunnelItem) FilterValue() string { return t.tunnel.Name + " " + t.tunnel.SSHHost }
func (t tunnelItem) Description() string {
	user := t.tunnel.SSHUser
	if user == "" {
		user = "~"
	}
	via := fmt.Sprintf("via %s@%s", user, t.tunnel.SSHHost)
	switch t.tunnel.Type {
	case model.TunnelLocal:
		return fmt.Sprintf("local  127.0.0.1:%d → %s:%d  %s",
			t.tunnel.LocalPort, t.tunnel.RemoteHost, t.tunnel.RemotePort, via)
	case model.TunnelRemote:
		return fmt.Sprintf("remote  %s:%d → 127.0.0.1:%d  %s",
			t.tunnel.SSHHost, t.tunnel.RemotePort, t.tunnel.LocalPort, via)
	case model.TunnelDynamic:
		return fmt.Sprintf("dynamic SOCKS  127.0.0.1:%d  %s", t.tunnel.LocalPort, via)
	default:
		return t.tunnel.SSHHost
	}
}

// buildTunnelListItems creates list items from tunnels.
func buildTunnelListItems(tunnels []model.Tunnel) []list.Item {
	items := make([]list.Item, len(tunnels))
	for i, t := range tunnels {
		items[i] = tunnelItem{tunnel: t, index: i}
	}
	return items
}

// tunnelListHelp returns the help bar text for the tunnel list view.
func tunnelListHelp() string {
	return helpStyle.Render("Tab: ssh mode | /: search | a: add | e: edit | d: delete | enter: run tunnel | q: quit")
}

// selectedTunnel returns the currently selected tunnel item, or nil if none.
func selectedTunnel(l list.Model) *tunnelItem {
	item := l.SelectedItem()
	if item == nil {
		return nil
	}
	t := item.(tunnelItem)
	return &t
}

// newTunnelList creates a configured bubbles list for tunnels.
func newTunnelList(items []list.Item, width, height int) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedStyle

	l := list.New(items, delegate, width, height)
	l.Title = "SSHH — Tunnels"
	l.Styles.Title = tunnelTitleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	return l
}

type tunnelListAction int

const (
	tunnelListActionNone tunnelListAction = iota
	tunnelListActionRun
	tunnelListActionAdd
	tunnelListActionEdit
	tunnelListActionDelete
	tunnelListActionToggleMode
	tunnelListActionQuit
)

func updateTunnelList(l *list.Model, msg tea.Msg) (tunnelListAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if l.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if selectedTunnel(*l) != nil {
				return tunnelListActionRun, nil
			}
		case "a":
			return tunnelListActionAdd, nil
		case "e":
			if selectedTunnel(*l) != nil {
				return tunnelListActionEdit, nil
			}
		case "d":
			if selectedTunnel(*l) != nil {
				return tunnelListActionDelete, nil
			}
		case "tab":
			return tunnelListActionToggleMode, nil
		case "q", "ctrl+c":
			return tunnelListActionQuit, nil
		}
	}

	var cmd tea.Cmd
	*l, cmd = l.Update(msg)
	return tunnelListActionNone, cmd
}
