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
	tunnelTableActionPickColor  // ctrl+t: open color picker
	tunnelTableActionQuit
)

type tunnelTableModel struct {
	tbl           table.Model
	search        textinput.Model
	allItems      []tunnelItem
	filtered      []tunnelItem
	maxHeight     int // max visible rows based on terminal height
	width         int
	tableRowWidth int // exact rendered width of one table row (no box border)
}

func newTunnelTableModel(items []tunnelItem, width, height int) tunnelTableModel {
	cols := tunnelTableCols(width, items)
	rows := tunnelItemsToRows(items)
	maxH := tableDataHeight(height)
	tableH := max(1, min(len(items)+2, maxH))
	trw := tableRowWidth(cols)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableH),
		table.WithWidth(trw),
	)
	s := table.DefaultStyles()
	s.Header = tunnelTableHeaderStyle
	s.Selected = tunnelTableSelectedStyle
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64
	ti.Width = searchInputWidth(trw)
	ti.Focus() // start in search mode by default

	return tunnelTableModel{
		tbl:           t,
		search:        ti,
		allItems:      items,
		filtered:      items,
		maxHeight:     maxH,
		width:         width,
		tableRowWidth: trw,
	}
}

// Init returns the cursor-blink command for the search box.
func (m tunnelTableModel) Init() tea.Cmd {
	return m.search.Focus()
}

func tunnelTableCols(termWidth int, items []tunnelItem) []table.Column {
	// Default minimums give the table a reasonable size even with no data.
	nameW := 15
	typeW := 7  // "dynamic"
	viaW := 20
	localW := 6 // ":65535"
	remoteW := 20
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

	// Cap last column so the total row never exceeds the terminal width.
	// Row width = sum(colWidth) + 2*numCols; box border adds 2 more.
	maxRemoteW := termWidth - 2 - nameW - typeW - viaW - localW - 2*5
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
	cols := tunnelTableCols(width, m.allItems)
	m.tbl.SetColumns(cols)
	trw := tableRowWidth(cols)
	m.tbl.SetWidth(trw)
	m.tableRowWidth = trw
	m.search.Width = searchInputWidth(trw)
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
		case "ctrl+t":
			return m, tunnelTableActionPickColor, nil
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

	trw := m.tableRowWidth
	// tableBox: border adds 2, so content width = trw, outer width = trw+2
	tableBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorAccent).
		Width(trw).
		Render(m.tbl.View())

	// searchBox: lipgloss Width includes padding; border is added on top.
	// Width(trw) → padded area = trw, border adds 2 → outer = trw+2 (same as tableBox).
	searchBox := tunnelSearchBoxFocusedStyle.Width(trw).Render(m.search.View())

	shortcuts := [][2]string{
		{"tab", "ssh"},
		{"ctrl+v", "list"},
		{"↑↓", "navigate"},
		{"esc", "clear"},
		{"ctrl+a", "add"},
		{"ctrl+e", "edit"},
		{"ctrl+d", "delete"},
		{"ctrl+t", "theme"},
		{"enter", "run"},
	}

	leftBlock := searchBox + "\n" + tableBox
	leftHeight := strings.Count(leftBlock, "\n") + 1
	innerHeight := leftHeight - 2 // subtract border top and bottom
	if innerHeight < 1 {
		innerHeight = 1
	}

	shortcutsBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		Height(innerHeight).
		Padding(0, 1).
		Render(renderShortcutsPanel(shortcuts, colorAccent))

	panel := lipgloss.NewStyle().PaddingLeft(2).Render(shortcutsBox)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, panel)
	return title + "\n" + mainArea
}
