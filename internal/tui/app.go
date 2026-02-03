package tui

import (
	"fmt"

	"ssh-tool/internal/config"
	"ssh-tool/internal/history"
	"ssh-tool/internal/model"
	"ssh-tool/internal/sshconfig"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewList view = iota
	viewForm
	viewConfirm
	viewImport
)

// Model is the root Bubble Tea model.
type Model struct {
	cfg  *config.Config
	hist *history.History

	serverList  list.Model
	listInited  bool
	form        formModel
	confirm     confirmModel
	imprt       importModel
	deleteIndex int // config index of server being deleted

	activeView view
	width      int
	height     int

	// Set when the user selects a server to connect to.
	ConnectTo *model.Server

	err error
}

// NewModel creates the initial app model.
func NewModel(cfg *config.Config, hist *history.History) Model {
	return Model{
		cfg:        cfg,
		hist:       hist,
		activeView: viewList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshList()
		return m, nil
	}

	switch m.activeView {
	case viewList:
		return m.updateListView(msg)
	case viewForm:
		return m.updateFormView(msg)
	case viewConfirm:
		return m.updateConfirmView(msg)
	case viewImport:
		return m.updateImportView(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return dangerStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	switch m.activeView {
	case viewForm:
		return m.form.View() + "\n"
	case viewConfirm:
		return m.confirm.View() + "\n"
	case viewImport:
		return m.imprt.View() + "\n"
	default:
		return m.renderListView()
	}
}

// refreshList rebuilds the list items from config, sorted by recent usage.
func (m *Model) refreshList() {
	sorted := m.hist.SortByRecent(m.cfg.Servers)

	originalIndices := make([]int, len(sorted))
	for i, s := range sorted {
		idx, _ := m.cfg.FindByName(s.Name)
		originalIndices[i] = idx
	}

	items := buildListItems(sorted, originalIndices)

	w := m.width
	h := m.height - 2
	if w == 0 {
		w = 80
	}
	if h < 10 {
		h = 20
	}

	if !m.listInited {
		m.serverList = newServerList(items, w, h)
		m.listInited = true
	} else {
		m.serverList.SetItems(items)
		m.serverList.SetSize(w, h)
	}
}

func (m Model) renderListView() string {
	if !m.listInited {
		return titleStyle.Render("SSH-TOOL") + "\n\n" + helpStyle.Render("Loading...")
	}
	return m.serverList.View() + "\n" + listHelp()
}

// --- List view updates ---

func (m Model) updateListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	action, cmd := updateList(&m.serverList, msg)

	switch action {
	case listActionConnect:
		s := selectedServer(m.serverList)
		if s != nil {
			srv := s.server
			m.ConnectTo = &srv
			return m, tea.Quit
		}
	case listActionAdd:
		m.form = newFormModel("Add Server", nil, -1)
		m.activeView = viewForm
		return m, m.form.Init()
	case listActionEdit:
		s := selectedServer(m.serverList)
		if s != nil {
			m.form = newFormModel("Edit Server", &s.server, s.index)
			m.activeView = viewForm
			return m, m.form.Init()
		}
	case listActionDelete:
		s := selectedServer(m.serverList)
		if s != nil {
			m.deleteIndex = s.index
			m.confirm = newConfirmModel(fmt.Sprintf("Delete server %q?", s.server.Name))
			m.activeView = viewConfirm
		}
	case listActionImport:
		servers, err := sshconfig.Parse()
		if err != nil {
			m.imprt = newImportModel(nil)
		} else {
			// Filter out servers that already exist.
			var newServers []model.Server
			for _, s := range servers {
				if idx, _ := m.cfg.FindByName(s.Name); idx == -1 {
					newServers = append(newServers, s)
				}
			}
			m.imprt = newImportModel(newServers)
		}
		m.activeView = viewImport
	case listActionQuit:
		return m, tea.Quit
	}

	return m, cmd
}

// --- Form view updates ---

func (m Model) updateFormView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.done {
		if m.form.saved {
			srv := m.form.ToServer()
			if srv.Name != "" && srv.Host != "" {
				if m.form.editing {
					if err := m.cfg.UpdateServer(m.form.index, srv); err != nil {
						m.err = err
					}
				} else {
					if err := m.cfg.AddServer(srv); err != nil {
						m.err = err
					}
				}
			}
		}
		m.activeView = viewList
		m.refreshList()
	}

	return m, cmd
}

// --- Confirm view updates ---

func (m Model) updateConfirmView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)

	if m.confirm.done {
		if m.confirm.confirmed {
			if err := m.cfg.DeleteServer(m.deleteIndex); err != nil {
				m.err = err
			}
		}
		m.activeView = viewList
		m.refreshList()
	}

	return m, cmd
}

// --- Import view updates ---

func (m Model) updateImportView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.imprt, cmd = m.imprt.Update(msg)

	if m.imprt.done {
		if m.imprt.imported {
			for _, s := range m.imprt.SelectedServers() {
				if err := m.cfg.AddServer(s); err != nil {
					m.err = err
					break
				}
			}
		}
		m.activeView = viewList
		m.refreshList()
	}

	return m, cmd
}
