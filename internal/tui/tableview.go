package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tableViewAction int

const (
	tableViewActionNone tableViewAction = iota
	tableViewActionConnect
	tableViewActionAdd
	tableViewActionEdit
	tableViewActionDelete
	tableViewActionImport
	tableViewActionToggleList   // v: switch back to list view
	tableViewActionToggleTunnel // tab: switch to tunnel table view
	tableViewActionQuit
)

type serverTableModel struct {
	tbl       table.Model
	search    textinput.Model
	allItems  []serverItem
	filtered  []serverItem
	maxHeight int // max visible rows based on terminal height
	width     int
}

func newServerTableModel(items []serverItem, width, height int) serverTableModel {
	cols := serverTableCols(width - 2) // -2 for box border left+right
	rows := serverItemsToRows(items)
	maxH := tableDataHeight(height)
	tableH := max(1, min(len(items)+2, maxH))

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableH),
	)
	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = tableSelectedStyle
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64
	ti.Width = searchInputWidth(width)
	ti.Focus() // start in search mode by default

	return serverTableModel{
		tbl:       t,
		search:    ti,
		allItems:  items,
		filtered:  items,
		maxHeight: maxH,
		width:     width,
	}
}

// Init returns the cursor-blink command for the search box.
func (m serverTableModel) Init() tea.Cmd {
	return m.search.Focus()
}

func tableDataHeight(totalHeight int) int {
	// totalHeight minus: title(1) + search box(3) + help(1) + table border/header(4)
	h := totalHeight - 9
	if h < 2 {
		h = 2
	}
	return h
}

func searchInputWidth(totalWidth int) int {
	w := totalWidth - 12
	if w < 20 {
		w = 20
	}
	return w
}

func serverTableCols(width int) []table.Column {
	nameW := 20
	hostW := 24
	userW := 12
	portW := 6
	tagsW := width - nameW - hostW - userW - portW - 14 // 14 for column separators/padding
	if tagsW < 8 {
		tagsW = 8
	}
	return []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Host", Width: hostW},
		{Title: "User", Width: userW},
		{Title: "Port", Width: portW},
		{Title: "Tags", Width: tagsW},
	}
}

func serverItemsToRows(items []serverItem) []table.Row {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.server.Name,
			item.server.Host,
			item.server.User,
			fmt.Sprintf("%d", item.server.Port),
			strings.Join(item.server.Tags, ", "),
		}
	}
	return rows
}

func (m *serverTableModel) setItems(items []serverItem) {
	m.allItems = items
	m.applyFilter(m.search.Value())
}

func (m *serverTableModel) resize(width, height int) {
	m.width = width
	m.maxHeight = tableDataHeight(height)
	m.tbl.SetHeight(max(1, min(len(m.filtered)+2, m.maxHeight)))
	m.tbl.SetColumns(serverTableCols(width - 2)) // -2 for box border
	m.search.Width = searchInputWidth(width)
}

func (m *serverTableModel) applyFilter(query string) {
	if query == "" {
		m.filtered = m.allItems
	} else {
		q := strings.ToLower(query)
		var out []serverItem
		for _, item := range m.allItems {
			hay := strings.ToLower(
				item.server.Name + " " + item.server.Host + " " +
					item.server.User + " " + strings.Join(item.server.Tags, " "),
			)
			if strings.Contains(hay, q) {
				out = append(out, item)
			}
		}
		m.filtered = out
	}
	m.tbl.SetRows(serverItemsToRows(m.filtered))
	m.tbl.SetHeight(max(1, min(len(m.filtered)+2, m.maxHeight)))
}

func (m serverTableModel) selectedItem() *serverItem {
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.filtered) {
		return nil
	}
	item := m.filtered[idx]
	return &item
}

func (m serverTableModel) Update(msg tea.Msg) (serverTableModel, tableViewAction, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "pgup", "pgdown":
			m.tbl, cmd = m.tbl.Update(msg)
			return m, tableViewActionNone, cmd
		case "enter":
			if m.selectedItem() != nil {
				return m, tableViewActionConnect, nil
			}
			return m, tableViewActionNone, nil
		case "esc":
			m.search.SetValue("")
			m.applyFilter("")
			return m, tableViewActionNone, nil
		case "ctrl+a":
			return m, tableViewActionAdd, nil
		case "ctrl+e":
			if m.selectedItem() != nil {
				return m, tableViewActionEdit, nil
			}
		case "ctrl+d":
			if m.selectedItem() != nil {
				return m, tableViewActionDelete, nil
			}
		case "ctrl+o":
			return m, tableViewActionImport, nil
		case "ctrl+v":
			return m, tableViewActionToggleList, nil
		case "tab":
			return m, tableViewActionToggleTunnel, nil
		case "ctrl+c":
			return m, tableViewActionQuit, nil
		default:
			m.search, cmd = m.search.Update(msg)
			m.applyFilter(m.search.Value())
			return m, tableViewActionNone, cmd
		}
	}

	m.tbl, cmd = m.tbl.Update(msg)
	return m, tableViewActionNone, cmd
}

func (m serverTableModel) View() string {
	count := helpStyle.Render(fmt.Sprintf("(%d/%d)", len(m.filtered), len(m.allItems)))
	// No PaddingLeft here — title must start at column 0 to align with the boxes below.
	title := lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render("SSHH") + "  " + count

	// searchBox total width = (m.width-4) content + 2 padding + 2 border = m.width
	searchBox := searchBoxFocusedStyle.Width(m.width - 4).Render(m.search.View())

	// tableBox total width = (m.width-2) content + 2 border = m.width  (matches searchBox)
	tableBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		Width(m.width - 2).
		Render(m.tbl.View())

	help := helpStyle.Render("tab: tunnels | ctrl+v: list | ↑↓: navigate | esc: clear | ctrl+a: add | ctrl+e: edit | ctrl+d: del | ctrl+o: import | enter: connect")

	return title + "\n" + searchBox + "\n" + tableBox + "\n" + help
}
