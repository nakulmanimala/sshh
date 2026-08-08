package tui

import (
	"fmt"

	"sshh/internal/awssync"
	"sshh/internal/config"
	"sshh/internal/history"
	"sshh/internal/model"
	"sshh/internal/sshconfig"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewList view = iota
	viewForm
	viewConfirm
	viewImport
	viewAWSUsername
	viewAWSProfile
	viewAWSSync
	viewTunnelList
	viewTunnelForm
	viewTunnelConfirm
	viewTableList
	viewTableTunnelList
	viewColorPicker
)

// Model is the root Bubble Tea model.
type Model struct {
	cfg       *config.Config
	tunnelCfg *config.TunnelConfig
	hist      *history.History

	// SSH mode state.
	serverList  list.Model
	listInited  bool
	form        formModel
	confirm     confirmModel
	imprt       importModel
	awsUsername awsUsernameModel
	awsProfile  awsProfileModel
	awsSync     awsSyncModel
	deleteIndex int

	// SSH table view state.
	serverTable       serverTableModel
	serverTableInited bool

	// Tunnel mode state.
	tunnelList        list.Model
	tunnelListInited  bool
	tunnelForm        tunnelFormModel
	tunnelConfirm     confirmModel
	tunnelDeleteIndex int

	// Tunnel table view state.
	tunnelTable       tunnelTableModel
	tunnelTableInited bool

	// Color picker state.
	colorPicker colorPickerModel
	settings    *config.Settings

	// returnToView records where to go back after form/confirm sub-views.
	returnToView view

	activeView view
	width      int
	height     int

	// Set when user selects an action that requires leaving the TUI.
	ConnectTo *model.Server
	RunTunnel *model.Tunnel

	err error
}

// NewModel creates the initial app model.
func NewModel(cfg *config.Config, tunnelCfg *config.TunnelConfig, hist *history.History, settings *config.Settings) Model {
	applyTheme(lipgloss.Color(settings.SSHColor), lipgloss.Color(settings.TunnelColor))
	return Model{
		cfg:        cfg,
		tunnelCfg:  tunnelCfg,
		hist:       hist,
		settings:   settings,
		activeView: viewTableList,
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
		switch m.activeView {
		case viewTableList:
			return m, m.serverTable.Init()
		case viewTableTunnelList:
			return m, m.tunnelTable.Init()
		case viewConfirm:
			m.confirm.resize(m.width, m.height)
		case viewTunnelConfirm:
			m.tunnelConfirm.resize(m.width, m.height)
		}
		return m, nil
	case awsInstancesMsg:
		m.awsSync.setResult(msg.servers, m.cfg.Servers, m.settings.AWSDefaultUser, msg.err)
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
	case viewAWSUsername:
		return m.updateAWSUsernameView(msg)
	case viewAWSProfile:
		return m.updateAWSProfileView(msg)
	case viewAWSSync:
		return m.updateAWSSyncView(msg)
	case viewTunnelList:
		return m.updateTunnelListView(msg)
	case viewTunnelForm:
		return m.updateTunnelFormView(msg)
	case viewTunnelConfirm:
		return m.updateTunnelConfirmView(msg)
	case viewTableList:
		return m.updateServerTableView(msg)
	case viewTableTunnelList:
		return m.updateTunnelTableView(msg)
	case viewColorPicker:
		return m.updateColorPickerView(msg)
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
	case viewAWSUsername:
		return m.awsUsername.View() + "\n"
	case viewAWSProfile:
		return m.awsProfile.View() + "\n"
	case viewAWSSync:
		return m.awsSync.View() + "\n"
	case viewTunnelList:
		return m.renderTunnelListView()
	case viewTunnelForm:
		return m.tunnelForm.View() + "\n"
	case viewTunnelConfirm:
		return m.tunnelConfirm.View() + "\n"
	case viewTableList:
		if !m.serverTableInited {
			return titleStyle.Render("SSHH") + "\n\n" + helpStyle.Render("Loading...")
		}
		return m.serverTable.View() + "\n"
	case viewTableTunnelList:
		if !m.tunnelTableInited {
			return tunnelTitleStyle.Render("SSHH — Tunnels") + "\n\n" + helpStyle.Render("Loading...")
		}
		return m.tunnelTable.View() + "\n"
	case viewColorPicker:
		return m.colorPicker.View()
	default:
		return m.renderListView()
	}
}

// --- SSH server list ---

func (m *Model) refreshList() {
	sorted := m.hist.SortByRecent(m.cfg.Servers)

	// Build a name→index map once (O(n)) instead of calling FindByName
	// in a loop (O(n²)).
	nameToIdx := make(map[string]int, len(m.cfg.Servers))
	for i, s := range m.cfg.Servers {
		nameToIdx[s.Name] = i
	}

	originalIndices := make([]int, len(sorted))
	for i, s := range sorted {
		originalIndices[i] = nameToIdx[s.Name]
	}

	listItems := buildListItems(sorted, originalIndices)
	srvItems := buildServerItems(sorted, originalIndices)

	w, h := m.dims()
	if !m.listInited {
		m.serverList = newServerList(listItems, w, h)
		m.listInited = true
	} else {
		m.serverList.SetItems(listItems)
		m.serverList.SetSize(w, h)
	}

	if !m.serverTableInited {
		m.serverTable = newServerTableModel(srvItems, w, h)
		m.serverTableInited = true
	} else {
		m.serverTable.setItems(srvItems)
		m.serverTable.resize(w, h)
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
		m.returnToView = viewList
		m.form = newFormModel("Add Server", nil, -1)
		m.activeView = viewForm
		return m, m.form.Init()
	case listActionEdit:
		s := selectedServer(m.serverList)
		if s != nil {
			m.returnToView = viewList
			m.form = newFormModel("Edit Server", &s.server, s.index)
			m.activeView = viewForm
			return m, m.form.Init()
		}
	case listActionDelete:
		s := selectedServer(m.serverList)
		if s != nil {
			m.returnToView = viewList
			m.deleteIndex = s.index
			m.confirm = newConfirmModel(fmt.Sprintf("Delete server %q?", s.server.Name))
			m.confirm.resize(m.width, m.height)
			m.activeView = viewConfirm
		}
	case listActionImport:
		m.returnToView = viewList
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
	case listActionAWSSync:
		return m.startAWSSync(viewList)
	case listActionToggleMode:
		m.activeView = viewTunnelList
		m.refreshTunnelList()
	case listActionToggleTable:
		m.activeView = viewTableList
		m.refreshList()
		return m, m.serverTable.Init()
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
		m.activeView = m.returnToView
		m.refreshList()
		if m.returnToView == viewTableList {
			return m, m.serverTable.Init()
		}
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
		m.activeView = m.returnToView
		m.refreshList()
		if m.returnToView == viewTableList {
			return m, m.serverTable.Init()
		}
	}

	return m, cmd
}

func (m Model) updateImportView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.imprt, cmd = m.imprt.Update(msg)

	if m.imprt.done {
		if m.imprt.imported {
			// AddServers appends all at once and saves a single time,
			// avoiding N disk writes for N imported servers.
			if selected := m.imprt.SelectedServers(); len(selected) > 0 {
				if err := m.cfg.AddServers(selected); err != nil {
					m.err = err
				}
			}
		}
		m.activeView = m.returnToView
		m.refreshList()
		if m.returnToView == viewTableList {
			return m, m.serverTable.Init()
		}
	}

	return m, cmd
}

// --- AWS sync ---

// startAWSSync begins the AWS sync flow from the given view, prompting for a
// default username first if one isn't configured yet.
func (m Model) startAWSSync(from view) (tea.Model, tea.Cmd) {
	m.returnToView = from
	if m.settings.AWSDefaultUser == "" {
		m.awsUsername = newAWSUsernameModel("")
		m.activeView = viewAWSUsername
		return m, m.awsUsername.Init()
	}
	return m.enterAWSProfilePicker()
}

// enterAWSProfilePicker lists AWS CLI profiles and shows the picker view.
func (m Model) enterAWSProfilePicker() (tea.Model, tea.Cmd) {
	profiles, err := awssync.ListProfiles()
	if err != nil || len(profiles) == 0 {
		profiles = []string{"default"}
	}
	m.awsProfile = newAWSProfileModel(profiles)
	m.activeView = viewAWSProfile
	return m, nil
}

func (m Model) updateAWSUsernameView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.awsUsername, cmd = m.awsUsername.Update(msg)

	if m.awsUsername.done {
		if m.awsUsername.saved {
			m.settings.AWSDefaultUser = m.awsUsername.Username()
			if err := m.settings.Save(); err != nil {
				m.err = err
				return m, nil
			}
			return m.enterAWSProfilePicker()
		}
		m.activeView = m.returnToView
		return m, nil
	}

	return m, cmd
}

func (m Model) updateAWSProfileView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.awsProfile, cmd = m.awsProfile.Update(msg)

	if m.awsProfile.done {
		if m.awsProfile.picked {
			profile := m.awsProfile.SelectedProfile()
			m.awsSync = newAWSSyncModel()
			m.activeView = viewAWSSync
			return m, fetchAWSInstancesCmd(profile)
		}
		m.activeView = m.returnToView
		return m, nil
	}

	return m, cmd
}

func (m Model) updateAWSSyncView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.awsSync, cmd = m.awsSync.Update(msg)

	if m.awsSync.done {
		if m.awsSync.applied {
			updates := m.awsSync.SelectedUpdates()
			adds := m.awsSync.SelectedAdds()
			removes := m.awsSync.SelectedRemoves()
			if len(updates) > 0 || len(adds) > 0 || len(removes) > 0 {
				if err := m.cfg.ApplyAWSSync(updates, adds, removes); err != nil {
					m.err = err
				}
			}
		}
		m.activeView = m.returnToView
		m.refreshList()
		if m.returnToView == viewTableList {
			return m, m.serverTable.Init()
		}
	}

	return m, cmd
}

// --- Tunnel list ---

func (m *Model) refreshTunnelList() {
	tnlItems := buildTunnelItems(m.tunnelCfg.Tunnels)
	listItems := buildTunnelListItems(m.tunnelCfg.Tunnels)

	w, h := m.dims()
	if !m.tunnelListInited {
		m.tunnelList = newTunnelList(listItems, w, h)
		m.tunnelListInited = true
	} else {
		m.tunnelList.SetItems(listItems)
		m.tunnelList.SetSize(w, h)
	}

	if !m.tunnelTableInited {
		m.tunnelTable = newTunnelTableModel(tnlItems, w, h)
		m.tunnelTableInited = true
	} else {
		m.tunnelTable.setItems(tnlItems)
		m.tunnelTable.resize(w, h)
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
		m.returnToView = viewTunnelList
		m.tunnelForm = newTunnelFormModel("Add Tunnel", nil, -1)
		m.activeView = viewTunnelForm
		return m, m.tunnelForm.Init()
	case tunnelListActionEdit:
		t := selectedTunnel(m.tunnelList)
		if t != nil {
			m.returnToView = viewTunnelList
			m.tunnelForm = newTunnelFormModel("Edit Tunnel", &t.tunnel, t.index)
			m.activeView = viewTunnelForm
			return m, m.tunnelForm.Init()
		}
	case tunnelListActionDelete:
		t := selectedTunnel(m.tunnelList)
		if t != nil {
			m.returnToView = viewTunnelList
			m.tunnelDeleteIndex = t.index
			m.tunnelConfirm = newConfirmModel(fmt.Sprintf("Delete tunnel %q?", t.tunnel.Name))
			m.tunnelConfirm.resize(m.width, m.height)
			m.activeView = viewTunnelConfirm
		}
	case tunnelListActionToggleMode:
		m.activeView = viewList
		m.refreshList()
	case tunnelListActionToggleTable:
		m.activeView = viewTableTunnelList
		m.refreshTunnelList()
		return m, m.tunnelTable.Init()
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
					if err := m.tunnelCfg.UpdateTunnel(m.tunnelForm.index, t); err != nil {
						m.err = err
					}
				} else {
					if err := m.tunnelCfg.AddTunnel(t); err != nil {
						m.err = err
					}
				}
			}
		}
		m.activeView = m.returnToView
		m.refreshTunnelList()
		if m.returnToView == viewTableTunnelList {
			return m, m.tunnelTable.Init()
		}
	}

	return m, cmd
}

func (m Model) updateTunnelConfirmView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tunnelConfirm, cmd = m.tunnelConfirm.Update(msg)

	if m.tunnelConfirm.done {
		if m.tunnelConfirm.confirmed {
			if err := m.tunnelCfg.DeleteTunnel(m.tunnelDeleteIndex); err != nil {
				m.err = err
			}
		}
		m.activeView = m.returnToView
		m.refreshTunnelList()
		if m.returnToView == viewTableTunnelList {
			return m, m.tunnelTable.Init()
		}
	}

	return m, cmd
}

// --- Server table view ---

func (m Model) updateServerTableView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action tableViewAction
	var cmd tea.Cmd
	m.serverTable, action, cmd = m.serverTable.Update(msg)

	switch action {
	case tableViewActionConnect:
		s := m.serverTable.selectedItem()
		if s != nil {
			srv := s.server
			m.ConnectTo = &srv
			return m, tea.Quit
		}
	case tableViewActionAdd:
		m.returnToView = viewTableList
		m.form = newFormModel("Add Server", nil, -1)
		m.activeView = viewForm
		return m, m.form.Init()
	case tableViewActionEdit:
		s := m.serverTable.selectedItem()
		if s != nil {
			m.returnToView = viewTableList
			m.form = newFormModel("Edit Server", &s.server, s.index)
			m.activeView = viewForm
			return m, m.form.Init()
		}
	case tableViewActionDelete:
		s := m.serverTable.selectedItem()
		if s != nil {
			m.returnToView = viewTableList
			m.deleteIndex = s.index
			m.confirm = newConfirmModel(fmt.Sprintf("Delete server %q?", s.server.Name))
			m.confirm.resize(m.width, m.height)
			m.activeView = viewConfirm
		}
	case tableViewActionImport:
		m.returnToView = viewTableList
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
	case tableViewActionAWSSync:
		return m.startAWSSync(viewTableList)
	case tableViewActionToggleList:
		m.activeView = viewList
		m.refreshList()
	case tableViewActionToggleTunnel:
		m.activeView = viewTableTunnelList
		m.refreshTunnelList()
		return m, m.tunnelTable.Init()
	case tableViewActionPickColor:
		m.returnToView = viewTableList
		m.colorPicker = newColorPickerModel(colorPickerTargetSSH, m.width, m.height)
		m.activeView = viewColorPicker
	case tableViewActionQuit:
		return m, tea.Quit
	}

	return m, cmd
}

// --- Tunnel table view ---

func (m Model) updateTunnelTableView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action tunnelTableAction
	var cmd tea.Cmd
	m.tunnelTable, action, cmd = m.tunnelTable.Update(msg)

	switch action {
	case tunnelTableActionRun:
		t := m.tunnelTable.selectedItem()
		if t != nil {
			tun := t.tunnel
			m.RunTunnel = &tun
			return m, tea.Quit
		}
	case tunnelTableActionAdd:
		m.returnToView = viewTableTunnelList
		m.tunnelForm = newTunnelFormModel("Add Tunnel", nil, -1)
		m.activeView = viewTunnelForm
		return m, m.tunnelForm.Init()
	case tunnelTableActionEdit:
		t := m.tunnelTable.selectedItem()
		if t != nil {
			m.returnToView = viewTableTunnelList
			m.tunnelForm = newTunnelFormModel("Edit Tunnel", &t.tunnel, t.index)
			m.activeView = viewTunnelForm
			return m, m.tunnelForm.Init()
		}
	case tunnelTableActionDelete:
		t := m.tunnelTable.selectedItem()
		if t != nil {
			m.returnToView = viewTableTunnelList
			m.tunnelDeleteIndex = t.index
			m.tunnelConfirm = newConfirmModel(fmt.Sprintf("Delete tunnel %q?", t.tunnel.Name))
			m.tunnelConfirm.resize(m.width, m.height)
			m.activeView = viewTunnelConfirm
		}
	case tunnelTableActionToggleList:
		m.activeView = viewTunnelList
		m.refreshTunnelList()
	case tunnelTableActionToggleSSH:
		m.activeView = viewTableList
		m.refreshList()
		return m, m.serverTable.Init()
	case tunnelTableActionPickColor:
		m.returnToView = viewTableTunnelList
		m.colorPicker = newColorPickerModel(colorPickerTargetTunnel, m.width, m.height)
		m.activeView = viewColorPicker
	case tunnelTableActionQuit:
		return m, tea.Quit
	}

	return m, cmd
}

// --- Color picker ---

func (m Model) updateColorPickerView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var action colorPickerAction
	m.colorPicker, action, _ = m.colorPicker.Update(msg)

	switch action {
	case colorPickerActionSelect:
		newColor := m.colorPicker.selectedColor()
		if m.colorPicker.target == colorPickerTargetSSH {
			applyTheme(newColor, colorAccent)
			m.settings.SSHColor = string(newColor)
		} else {
			applyTheme(colorPrimary, newColor)
			m.settings.TunnelColor = string(newColor)
		}
		_ = m.settings.Save()
		// Force table reinit so new styles are applied.
		m.serverTableInited = false
		m.tunnelTableInited = false
		m.activeView = m.returnToView
		m.refreshList()
		m.refreshTunnelList()
		if m.returnToView == viewTableList {
			return m, m.serverTable.Init()
		}
		return m, m.tunnelTable.Init()
	case colorPickerActionCancel:
		m.activeView = m.returnToView
	}

	return m, nil
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
