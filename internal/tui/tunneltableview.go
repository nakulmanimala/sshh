package tui

import (
	"fmt"
	"strings"

	"sshh/internal/model"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tunnelTableAction int

const (
	tunnelTableActionNone tunnelTableAction = iota
	tunnelTableActionRun
	tunnelTableActionAdd
	tunnelTableActionEdit
	tunnelTableActionDelete
	tunnelTableActionToggleList // v: switch back to tunnel list view
	tunnelTableActionToggleSSH  // tab: switch to SSH table view
	tunnelTableActionQuit
)

type tunnelTableModel struct {
	tbl       table.Model
	search    textinput.Model
	allItems  []tunnelItem
	filtered  []tunnelItem
	searching bool
	maxHeight int // max visible rows based on terminal height
	width     int
}

func newTunnelTableModel(items []tunnelItem, width, height int) tunnelTableModel {
	cols := tunnelTableCols(width - 2) // -2 for box border left+right
	rows := tunnelItemsToRows(items)
	maxH := tableDataHeight(height)
	tableH := max(1, min(len(items), maxH))

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableH),
	)
	s := table.DefaultStyles()
	s.Header = tunnelTableHeaderStyle
	s.Selected = tunnelTableSelectedStyle
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64
	ti.Width = searchInputWidth(width)
	ti.Focus() // start in search mode by default

	return tunnelTableModel{
		tbl:       t,
		search:    ti,
		allItems:  items,
		filtered:  items,
		searching: true,
		maxHeight: maxH,
		width:     width,
	}
}

// Init returns the cursor-blink command for the search box.
func (m tunnelTableModel) Init() tea.Cmd {
	if m.searching {
		return m.search.Focus()
	}
	return nil
}

func tunnelTableCols(width int) []table.Column {
	nameW := 18
	typeW := 9
	viaW := 24
	localW := 10
	remoteW := width - nameW - typeW - viaW - localW - 14
	if remoteW < 10 {
		remoteW = 10
	}
	return []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Type", Width: typeW},
		{Title: "Via", Width: viaW},
		{Title: "Local", Width: localW},
		{Title: "Remote", Width: remoteW},
	}
}

func tunnelViaStr(t tunnelItem) string {
	host := t.tunnel.SSHHost
	if t.tunnel.SSHUser != "" {
		host = t.tunnel.SSHUser + "@" + host
	}
	if t.tunnel.SSHPort > 0 && t.tunnel.SSHPort != 22 {
		host = fmt.Sprintf("%s:%d", host, t.tunnel.SSHPort)
	}
	return host
}

func tunnelRemoteStr(t tunnelItem) string {
	switch t.tunnel.Type {
	case model.TunnelLocal:
		return fmt.Sprintf("%s:%d", t.tunnel.RemoteHost, t.tunnel.RemotePort)
	case model.TunnelRemote:
		return fmt.Sprintf("→ :%d", t.tunnel.LocalPort)
	case model.TunnelDynamic:
		return "SOCKS5"
	default:
		return ""
	}
}

func tunnelItemsToRows(items []tunnelItem) []table.Row {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.tunnel.Name,
			string(item.tunnel.Type),
			tunnelViaStr(item),
			fmt.Sprintf(":%d", item.tunnel.LocalPort),
			tunnelRemoteStr(item),
		}
	}
	return rows
}

func (m *tunnelTableModel) setItems(items []tunnelItem) {
	m.allItems = items
	m.applyFilter(m.search.Value())
}

func (m *tunnelTableModel) resize(width, height int) {
	m.width = width
	m.maxHeight = tableDataHeight(height)
	m.tbl.SetHeight(max(1, min(len(m.filtered), m.maxHeight)))
	m.tbl.SetColumns(tunnelTableCols(width - 2)) // -2 for box border
	m.search.Width = searchInputWidth(width)
}

func (m *tunnelTableModel) applyFilter(query string) {
	if query == "" {
		m.filtered = m.allItems
	} else {
		q := strings.ToLower(query)
		var out []tunnelItem
		for _, item := range m.allItems {
			hay := strings.ToLower(item.tunnel.Name + " " + item.tunnel.SSHHost + " " + string(item.tunnel.Type))
			if strings.Contains(hay, q) {
				out = append(out, item)
			}
		}
		m.filtered = out
	}
	m.tbl.SetRows(tunnelItemsToRows(m.filtered))
	m.tbl.SetHeight(max(1, min(len(m.filtered), m.maxHeight)))
}

func (m tunnelTableModel) selectedItem() *tunnelItem {
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.filtered) {
		return nil
	}
	item := m.filtered[idx]
	return &item
}

func (m tunnelTableModel) Update(msg tea.Msg) (tunnelTableModel, tunnelTableAction, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "up", "down", "pgup", "pgdown":
				// Navigate rows while keeping search active.
				m.tbl, cmd = m.tbl.Update(msg)
				return m, tunnelTableActionNone, cmd
			case "enter":
				if m.selectedItem() != nil {
					return m, tunnelTableActionRun, nil
				}
				return m, tunnelTableActionNone, nil
			case "esc":
				// Blur the search box but keep the filter active so the
				// user can act on the filtered selection (e/d/enter).
				m.searching = false
				m.search.Blur()
				return m, tunnelTableActionNone, nil
			case "tab":
				return m, tunnelTableActionToggleSSH, nil
			case "ctrl+c":
				return m, tunnelTableActionQuit, nil
			default:
				// Route all other keys to search.
				m.search, cmd = m.search.Update(msg)
				m.applyFilter(m.search.Value())
				return m, tunnelTableActionNone, cmd
			}
		}

		// Nav-only mode (search blurred).
		switch msg.String() {
		case "/", "esc":
			// esc in nav mode = clear filter and re-enter search.
			if msg.String() == "esc" {
				m.search.SetValue("")
				m.applyFilter("")
			}
			m.searching = true
			cmd = m.search.Focus()
			return m, tunnelTableActionNone, cmd
		case "enter":
			if m.selectedItem() != nil {
				return m, tunnelTableActionRun, nil
			}
		case "a":
			return m, tunnelTableActionAdd, nil
		case "e":
			if m.selectedItem() != nil {
				return m, tunnelTableActionEdit, nil
			}
		case "d":
			if m.selectedItem() != nil {
				return m, tunnelTableActionDelete, nil
			}
		case "v":
			return m, tunnelTableActionToggleList, nil
		case "tab":
			return m, tunnelTableActionToggleSSH, nil
		case "q", "ctrl+c":
			return m, tunnelTableActionQuit, nil
		}
	}

	m.tbl, cmd = m.tbl.Update(msg)
	return m, tunnelTableActionNone, cmd
}

func (m tunnelTableModel) View() string {
	count := helpStyle.Render(fmt.Sprintf("(%d/%d)", len(m.filtered), len(m.allItems)))
	title := tunnelTitleStyle.Render("SSHH — Tunnels") + "  " + count

	searchStyle := searchBoxBlurredStyle
	if m.searching {
		searchStyle = tunnelSearchBoxFocusedStyle
	}
	searchBox := searchStyle.Width(m.width - 4).Render(m.search.View())

	tableBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorAccent).
		Render(m.tbl.View())

	help := helpStyle.Render("Tab: ssh | v: list view | ↑↓: navigate | esc: clear | a: add | e: edit | d: del | enter: run | q: quit")

	return title + "\n" + searchBox + "\n" + tableBox + "\n" + help
}
