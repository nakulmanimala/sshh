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
	tbl            table.Model
	search         textinput.Model
	allItems       []tunnelItem
	filtered       []tunnelItem
	maxHeight      int // max visible rows based on terminal height
	width          int
	effectiveWidth int // actual rendered width based on content
}

func newTunnelTableModel(items []tunnelItem, width, height int) tunnelTableModel {
	cols := tunnelTableCols(width-2, items) // -2 for box border left+right
	rows := tunnelItemsToRows(items)
	maxH := tableDataHeight(height)
	tableH := max(1, min(len(items)+2, maxH))
	ew := tableEffectiveWidth(cols)

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
	ti.Width = searchInputWidth(ew)
	ti.Focus() // start in search mode by default

	return tunnelTableModel{
		tbl:            t,
		search:         ti,
		allItems:       items,
		filtered:       items,
		maxHeight:      maxH,
		width:          width,
		effectiveWidth: ew,
	}
}

// Init returns the cursor-blink command for the search box.
func (m tunnelTableModel) Init() tea.Cmd {
	return m.search.Focus()
}

func tunnelTableCols(width int, items []tunnelItem) []table.Column {
	// Start each column at its header length, then grow to fit content.
	nameW := len("Name")
	typeW := len("Type")
	viaW := len("Via")
	localW := len("Local")
	remoteW := len("Remote")
	for _, item := range items {
		if n := len(item.tunnel.Name); n > nameW {
			nameW = n
		}
		if n := len(string(item.tunnel.Type)); n > typeW {
			typeW = n
		}
		if n := len(tunnelViaStr(item)); n > viaW {
			viaW = n
		}
		if n := len(fmt.Sprintf(":%d", item.tunnel.LocalPort)); n > localW {
			localW = n
		}
		if n := len(tunnelRemoteStr(item)); n > remoteW {
			remoteW = n
		}
	}
	// +1 breathing room on each column.
	nameW++
	typeW++
	viaW++
	localW++
	remoteW++
	const overhead = 14 // cell padding + separators for 5 columns
	// Cap last column so the table never exceeds the terminal width.
	maxRemoteW := width - nameW - typeW - viaW - localW - overhead
	if maxRemoteW < 10 {
		maxRemoteW = 10
	}
	if remoteW > maxRemoteW {
		remoteW = maxRemoteW
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
	m.tbl.SetHeight(max(1, min(len(m.filtered)+2, m.maxHeight)))
	cols := tunnelTableCols(width-2, m.allItems) // -2 for box border
	m.tbl.SetColumns(cols)
	m.effectiveWidth = tableEffectiveWidth(cols)
	m.search.Width = searchInputWidth(m.effectiveWidth)
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
	m.tbl.SetHeight(max(1, min(len(m.filtered)+2, m.maxHeight)))
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
		switch msg.String() {
		case "up", "down", "pgup", "pgdown":
			m.tbl, cmd = m.tbl.Update(msg)
			return m, tunnelTableActionNone, cmd
		case "enter":
			if m.selectedItem() != nil {
				return m, tunnelTableActionRun, nil
			}
			return m, tunnelTableActionNone, nil
		case "esc":
			m.search.SetValue("")
			m.applyFilter("")
			return m, tunnelTableActionNone, nil
		case "ctrl+a":
			return m, tunnelTableActionAdd, nil
		case "ctrl+e":
			if m.selectedItem() != nil {
				return m, tunnelTableActionEdit, nil
			}
		case "ctrl+d":
			if m.selectedItem() != nil {
				return m, tunnelTableActionDelete, nil
			}
		case "ctrl+v":
			return m, tunnelTableActionToggleList, nil
		case "tab":
			return m, tunnelTableActionToggleSSH, nil
		case "ctrl+c":
			return m, tunnelTableActionQuit, nil
		default:
			m.search, cmd = m.search.Update(msg)
			m.applyFilter(m.search.Value())
			return m, tunnelTableActionNone, cmd
		}
	}

	m.tbl, cmd = m.tbl.Update(msg)
	return m, tunnelTableActionNone, cmd
}

func (m tunnelTableModel) View() string {
	count := helpStyle.Render(fmt.Sprintf("(%d/%d)", len(m.filtered), len(m.allItems)))
	// No PaddingLeft — align with boxes below.
	title := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("SSHH — Tunnels") + "  " + count

	ew := m.effectiveWidth
	// searchBox: content width = ew-4 (2 border + 2 padding), total = ew
	searchBox := tunnelSearchBoxFocusedStyle.Width(ew - 4).Render(m.search.View())
	// tableBox: content width = ew-2 (2 border), total = ew
	tableBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorAccent).
		Width(ew - 2).
		Render(m.tbl.View())

	help := helpStyle.Render("tab: ssh | ctrl+v: list | ↑↓: navigate | esc: clear | ctrl+a: add | ctrl+e: edit | ctrl+d: del | enter: run")

	return title + "\n" + searchBox + "\n" + tableBox + "\n" + help
}
