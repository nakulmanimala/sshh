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
	tableViewActionPickColor    // ctrl+t: open color picker
	tableViewActionQuit
)

type serverTableModel struct {
	tbl          table.Model
	search       textinput.Model
	allItems     []serverItem
	filtered     []serverItem
	maxHeight    int // max visible rows based on terminal height
	width        int
	tableRowWidth int // exact rendered width of one table row (no box border)
}

func newServerTableModel(items []serverItem, width, height int) serverTableModel {
	cols := serverTableCols(width, items)
	rows := serverItemsToRows(items)
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
	s.Header = tableHeaderStyle
	s.Selected = tableSelectedStyle
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64
	ti.Width = searchInputWidth(trw)
	ti.Focus() // start in search mode by default

	return serverTableModel{
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
func (m serverTableModel) Init() tea.Cmd {
	return m.search.Focus()
}

func tableDataHeight(totalHeight int) int {
	// totalHeight minus: title(1) + search box(3) + table border/header(4)
	h := totalHeight - 8
	if h < 2 {
		h = 2
	}
	return h
}

// renderShortcutsPanel renders a vertical key → description shortcuts panel.
func renderShortcutsPanel(shortcuts [][2]string, accentColor lipgloss.Color) string {
	keyStyle := lipgloss.NewStyle().Foreground(accentColor)
	descStyle := lipgloss.NewStyle().Foreground(colorMuted)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorSecondary)
	dividerStyle := lipgloss.NewStyle().Foreground(colorMuted)

	lines := make([]string, 0, len(shortcuts)+2)
	lines = append(lines, headerStyle.Render("Shortcuts"))
	lines = append(lines, dividerStyle.Render("─────────"))
	for _, s := range shortcuts {
		key := keyStyle.Render(fmt.Sprintf("%-8s", s[0]))
		desc := descStyle.Render(s[1])
		lines = append(lines, key+" "+desc)
	}
	return strings.Join(lines, "\n")
}

func searchInputWidth(rowWidth int) int {
	// searchBoxFocusedStyle uses Width(rowWidth); padding(0,1) gives wrapAt = rowWidth-2,
	// so the textinput should be rowWidth-2 wide (content area inside padding).
	w := rowWidth - 2
	if w < 20 {
		w = 20
	}
	return w
}

// tableRowWidth returns the exact visual width of a rendered table row:
// each cell is padded to its Width and surrounded by 1 space on each side.
func tableRowWidth(cols []table.Column) int {
	total := 0
	for _, c := range cols {
		total += c.Width
	}
	return total + 2*len(cols) // 1 space left + 1 space right per cell
}

func serverTableCols(termWidth int, items []serverItem) []table.Column {
	// Default minimums give the table a reasonable size even with no data.
	nameW := 15
	hostW := 20
	userW := 8
	portW := 5
	tagsW := 10
	for _, item := range items {
		if n := len(item.server.Name); n > nameW {
			nameW = n
		}
		if n := len(item.server.Host); n > hostW {
			hostW = n
		}
		if n := len(item.server.User); n > userW {
			userW = n
		}
		if n := len(fmt.Sprintf("%d", item.server.Port)); n > portW {
			portW = n
		}
		if n := len(strings.Join(item.server.Tags, ", ")); n > tagsW {
			tagsW = n
		}
	}
	// +1 breathing room on each column.
	nameW++
	hostW++
	userW++
	portW++
	tagsW++

	// Cap last column so the total row never exceeds the terminal width.
	// Row width = sum(colWidth) + 2*numCols; box border adds 2 more.
	maxTagsW := termWidth - 2 - nameW - hostW - userW - portW - 2*5
	if maxTagsW < 8 {
		maxTagsW = 8
	}
	if tagsW > maxTagsW {
		tagsW = maxTagsW
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
	cols := serverTableCols(width, m.allItems)
	m.tbl.SetColumns(cols)
	trw := tableRowWidth(cols)
	m.tbl.SetWidth(trw)
	m.tableRowWidth = trw
	m.search.Width = searchInputWidth(trw)
}

func (m *serverTableModel) applyFilter(query string) {
	if normalizeForSearch(query) == "" {
		m.filtered = m.allItems
	} else {
		var out []serverItem
		for _, item := range m.allItems {
			hay := item.server.Name + " " + item.server.Host + " " +
				item.server.User + " " + strings.Join(item.server.Tags, " ")
			if matchesNormalized(hay, query) {
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
		case "ctrl+t":
			return m, tableViewActionPickColor, nil
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

	trw := m.tableRowWidth
	// tableBox: border adds 2, so content width = trw, outer width = trw+2
	tableBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		Width(trw).
		Render(m.tbl.View())

	// searchBox: lipgloss Width includes padding; border is added on top.
	// Width(trw) → padded area = trw, border adds 2 → outer = trw+2 (same as tableBox).
	searchBox := searchBoxFocusedStyle.Width(trw).Render(m.search.View())

	shortcuts := [][2]string{
		{"tab", "tunnels"},
		{"ctrl+v", "list"},
		{"↑↓", "navigate"},
		{"esc", "clear"},
		{"ctrl+a", "add"},
		{"ctrl+e", "edit"},
		{"ctrl+d", "delete"},
		{"ctrl+o", "import"},
		{"ctrl+t", "theme"},
		{"enter", "connect"},
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
		Render(renderShortcutsPanel(shortcuts, colorPrimary))

	panel := lipgloss.NewStyle().PaddingLeft(2).Render(shortcutsBox)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, panel)
	return title + "\n" + mainArea
}
