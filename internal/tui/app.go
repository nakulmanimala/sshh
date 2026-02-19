package tui

import (
	"fmt"

	"sshh/internal/config"
	"sshh/internal/history"
	"sshh/internal/model"
	"sshh/internal/sshconfig"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewList view = iota
	viewForm
	viewConfirm
	viewImport
	viewTunnelList
	viewTunnelForm
	viewTunnelConfirm
)

// Model is the root Bubble Tea model.
type Model struct {
	cfg  *config.Config
	hist *history.History

	// SSH mode state.
	serverList  list.Model
	listInited  bool
	form        formModel
	confirm     confirmModel
	imprt       importModel
	deleteIndex int

	// Tunnel mode state.
	tunnelList       list.Model
	tunnelListInited bool
	tunnelForm       tunnelFormModel
	tunnelConfirm    confirmModel
	tunnelDeleteIndex int

	activeView view
	width      int
	height     int

	// Set when user selects an action that requires leaving the TUI.
	ConnectTo *model.Server
	RunTunnel *model.Tunnel

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
		m.refreshTunnelList()
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
	case viewTunnelList:
		return m.updateTunnelListView(msg)
	case viewTunnelForm:
		return m.updateTunnelFormView(msg)
	case viewTunnelConfirm:
		return m.updateTunnelConfirmView(msg)
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
	case viewTunnelList:
		return m.renderTunnelListView()
	case viewTunnelForm:
		return m.tunnelForm.View() + "\n"
	case viewTunnelConfirm:
		return m.tunnelConfirm.View() + "\n"
	default:
		return m.renderListView()
	}
}

// --- SSH server list ---

func (m *Model) refreshList() {
	sorted := m.hist.SortByRecent(m.cfg.Servers)

	originalIndices := make([]int, len(sorted))
	for i, s := range sorted {
		idx, _ := m.cfg.FindByName(s.Name)
		originalIndices[i] = idx
	}

	items := buildListItems(sorted, originalIndices)

	w, h := m.dims()
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
		return titleStyle.Render("SSHH") + "\n\n" + helpStyle.Render("Loading...")
	}
	return m.serverList.View() + "\n" + listHelp()
}

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
			var newServers []model.Server
			for _, s := range servers {
				if idx, _ := m.cfg.FindByName(s.Name); idx == -1 {
					newServers = append(newServers, s)
				}
			}
			m.imprt = newImportModel(newServers)
		}
		m.activeView = viewImport
	case listActionToggleMode:
		m.activeView = viewTunnelList
		m.refreshTunnelList()
	case listActionQuit:
		return m, tea.Quit
	}

	return m, cmd
}

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

// --- Tunnel list ---

func (m *Model) refreshTunnelList() {
	items := buildTunnelListItems(m.cfg.Tunnels)
	w, h := m.dims()
	if !m.tunnelListInited {
		m.tunnelList = newTunnelList(items, w, h)
		m.tunnelListInited = true
	} else {
		m.tunnelList.SetItems(items)
		m.tunnelList.SetSize(w, h)
	}
}

func (m Model) renderTunnelListView() string {
	if !m.tunnelListInited {
		return tunnelTitleStyle.Render("SSHH — Tunnels") + "\n\n" + helpStyle.Render("Loading...")
	}
	return m.tunnelList.View() + "\n" + tunnelListHelp()
}

func (m Model) updateTunnelListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	action, cmd := updateTunnelList(&m.tunnelList, msg)

	switch action {
	case tunnelListActionRun:
		t := selectedTunnel(m.tunnelList)
		if t != nil {
			tun := t.tunnel
			m.RunTunnel = &tun
			return m, tea.Quit
		}
	case tunnelListActionAdd:
		m.tunnelForm = newTunnelFormModel("Add Tunnel", nil, -1)
		m.activeView = viewTunnelForm
		return m, m.tunnelForm.Init()
	case tunnelListActionEdit:
		t := selectedTunnel(m.tunnelList)
		if t != nil {
			m.tunnelForm = newTunnelFormModel("Edit Tunnel", &t.tunnel, t.index)
			m.activeView = viewTunnelForm
			return m, m.tunnelForm.Init()
		}
	case tunnelListActionDelete:
		t := selectedTunnel(m.tunnelList)
		if t != nil {
			m.tunnelDeleteIndex = t.index
			m.tunnelConfirm = newConfirmModel(fmt.Sprintf("Delete tunnel %q?", t.tunnel.Name))
			m.activeView = viewTunnelConfirm
		}
	case tunnelListActionToggleMode:
		m.activeView = viewList
		m.refreshList()
	case tunnelListActionQuit:
		return m, tea.Quit
	}

	return m, cmd
}

func (m Model) updateTunnelFormView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tunnelForm, cmd = m.tunnelForm.Update(msg)

	if m.tunnelForm.done {
		if m.tunnelForm.saved {
			t := m.tunnelForm.ToTunnel()
			if t.Name != "" && t.SSHHost != "" {
				if m.tunnelForm.editing {
					if err := m.cfg.UpdateTunnel(m.tunnelForm.index, t); err != nil {
						m.err = err
					}
				} else {
					if err := m.cfg.AddTunnel(t); err != nil {
						m.err = err
					}
				}
			}
		}
		m.activeView = viewTunnelList
		m.refreshTunnelList()
	}

	return m, cmd
}

func (m Model) updateTunnelConfirmView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tunnelConfirm, cmd = m.tunnelConfirm.Update(msg)

	if m.tunnelConfirm.done {
		if m.tunnelConfirm.confirmed {
			if err := m.cfg.DeleteTunnel(m.tunnelDeleteIndex); err != nil {
				m.err = err
			}
		}
		m.activeView = viewTunnelList
		m.refreshTunnelList()
	}

	return m, cmd
}

// dims returns the usable width and height for list views.
func (m Model) dims() (int, int) {
	w := m.width
	h := m.height - 2
	if w == 0 {
		w = 80
	}
	if h < 10 {
		h = 20
	}
	return w, h
}
