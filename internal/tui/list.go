package tui

import (
	"fmt"
	"strings"

	"sshh/internal/model"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// serverItem wraps a Server for use in the bubbles list.
type serverItem struct {
	server model.Server
	index  int // index in config.Servers
}

func (s serverItem) Title() string       { return s.server.Name }
func (s serverItem) FilterValue() string { return s.server.Name + " " + strings.Join(s.server.Tags, " ") }
func (s serverItem) Description() string {
	desc := fmt.Sprintf("%s@%s:%d", s.server.User, s.server.Host, s.server.Port)
	if len(s.server.Tags) > 0 {
		desc += "  " + tagStyle.Render("["+strings.Join(s.server.Tags, ", ")+"]")
	}
	return desc
}

// buildListItems creates list items from servers, preserving original config indices.
func buildListItems(servers []model.Server, originalIndices []int) []list.Item {
	items := make([]list.Item, len(servers))
	for i, s := range servers {
		idx := i
		if originalIndices != nil && i < len(originalIndices) {
			idx = originalIndices[i]
		}
		items[i] = serverItem{server: s, index: idx}
	}
	return items
}

// listHelp returns the help bar text for the server list view.
func listHelp() string {
	return helpStyle.Render("Tab: tunnel mode | /: search | a: add | e: edit | d: delete | i: import | enter: connect | q: quit")
}

// selectedServer returns the currently selected server item, or nil if none.
func selectedServer(l list.Model) *serverItem {
	item := l.SelectedItem()
	if item == nil {
		return nil
	}
	s := item.(serverItem)
	return &s
}

// newServerList creates a configured bubbles list for servers.
func newServerList(items []list.Item, width, height int) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedStyle

	l := list.New(items, delegate, width, height)
	l.Title = "SSHH"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We use custom help.
	l.DisableQuitKeybindings()
	return l
}

// listUpdate handles key events on the list view and returns actions.
type listAction int

const (
	listActionNone listAction = iota
	listActionConnect
	listActionAdd
	listActionEdit
	listActionDelete
	listActionImport
	listActionToggleMode
	listActionQuit
)

func updateList(l *list.Model, msg tea.Msg) (listAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept keys while filtering.
		if l.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if selectedServer(*l) != nil {
				return listActionConnect, nil
			}
		case "a":
			return listActionAdd, nil
		case "e":
			if selectedServer(*l) != nil {
				return listActionEdit, nil
			}
		case "d":
			if selectedServer(*l) != nil {
				return listActionDelete, nil
			}
		case "i":
			return listActionImport, nil
		case "tab":
			return listActionToggleMode, nil
		case "q", "ctrl+c":
			return listActionQuit, nil
		}
	}

	var cmd tea.Cmd
	*l, cmd = l.Update(msg)
	return listActionNone, cmd
}
