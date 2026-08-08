package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sshh/internal/awssync"
	"sshh/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

// awsInstancesMsg carries the result of an async AWS EC2 fetch.
type awsInstancesMsg struct {
	servers []model.Server
	err     error
}

// fetchAWSInstancesCmd fetches sshh_accessible EC2 instances for profile.
func fetchAWSInstancesCmd(profile string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		servers, err := awssync.FetchAccessibleInstances(ctx, profile)
		return awsInstancesMsg{servers: servers, err: err}
	}
}

type awsSyncState int

const (
	awsSyncLoading awsSyncState = iota
	awsSyncError
	awsSyncEmpty
	awsSyncReview
)

// awsSyncModel shows the AWS EC2 fetch progress and the review/confirm
// checklist of pending inventory changes.
type awsSyncModel struct {
	state    awsSyncState
	items    []awssync.SyncItem
	selected []bool
	cursor   int
	errMsg   string
	done     bool
	applied  bool
}

func newAWSSyncModel() awsSyncModel {
	return awsSyncModel{state: awsSyncLoading}
}

func (m awsSyncModel) Init() tea.Cmd {
	return nil
}

// setResult transitions out of the loading state once the fetch completes.
func (m *awsSyncModel) setResult(servers []model.Server, existing []model.Server, defaultUser string, err error) {
	if err != nil {
		m.state = awsSyncError
		m.errMsg = err.Error()
		return
	}

	items := awssync.Diff(existing, servers, defaultUser)
	if len(items) == 0 {
		m.state = awsSyncEmpty
		return
	}

	m.state = awsSyncReview
	m.items = items
	m.selected = make([]bool, len(items))
	for i, item := range items {
		// Adds/updates are pre-selected like the SSH-config import view.
		// Stale removals default to unselected — deletion is destructive
		// and warrants an explicit opt-in each time.
		m.selected[i] = item.Action != awssync.SyncStale
	}
}

func (m awsSyncModel) Update(msg tea.Msg) (awsSyncModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if keyMsg.String() == "esc" {
		m.done = true
		return m, nil
	}

	if m.state != awsSyncReview {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case " ":
		if len(m.selected) > 0 {
			m.selected[m.cursor] = !m.selected[m.cursor]
		}
	case "enter":
		m.done = true
		m.applied = true
	}
	return m, nil
}

func (m awsSyncModel) View() string {
	switch m.state {
	case awsSyncLoading:
		return titleStyle.Render("AWS Sync") + "\n\n" + helpStyle.Render("Fetching EC2 instances...")
	case awsSyncError:
		return titleStyle.Render("AWS Sync") + "\n\n" +
			dangerStyle.Render("Error: "+m.errMsg) + "\n\n" +
			helpStyle.Render("Press Esc to go back")
	case awsSyncEmpty:
		return titleStyle.Render("AWS Sync") + "\n\n" +
			helpStyle.Render("No changes found — inventory is already up to date.") + "\n\n" +
			helpStyle.Render("Press Esc to go back")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("AWS Sync — Review Changes"))
	b.WriteString("\n\n")

	selectedCount := 0
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
		}

		check := "[ ]"
		if m.selected[i] {
			check = successStyle.Render("[x]")
			selectedCount++
		}

		name := item.Server.Name
		if i == m.cursor {
			name = selectedStyle.Render(name)
		}

		var desc string
		switch item.Action {
		case awssync.SyncNew:
			desc = successStyle.Render("[NEW]") + " " +
				fmt.Sprintf("%s@%s:%d", item.Server.User, item.Server.Host, item.Server.Port)
		case awssync.SyncUpdateIP:
			desc = tagStyle.Render("[IP]") + " " +
				fmt.Sprintf("%s -> %s", item.OldHost, item.Server.Host)
		case awssync.SyncStale:
			desc = dangerStyle.Render("[STALE]") + " " +
				fmt.Sprintf("not found in AWS (was %s) — remove?", item.Server.Host)
		}

		b.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, check, name, desc))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("%d/%d selected", selectedCount, len(m.items))))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Space: toggle | Enter: apply selected | Esc: cancel"))
	return b.String()
}

// SelectedUpdates returns the config-index -> updated-server map for selected IP changes.
func (m awsSyncModel) SelectedUpdates() map[int]model.Server {
	updates := make(map[int]model.Server)
	for i, item := range m.items {
		if item.Action != awssync.SyncUpdateIP || !m.selected[i] {
			continue
		}
		updates[item.ExistingIndex] = item.Server
	}
	return updates
}

// SelectedAdds returns the new servers to append for selected new instances.
func (m awsSyncModel) SelectedAdds() []model.Server {
	var adds []model.Server
	for i, item := range m.items {
		if item.Action != awssync.SyncNew || !m.selected[i] {
			continue
		}
		adds = append(adds, item.Server)
	}
	return adds
}

// SelectedRemoves returns the config indices to remove for selected stale entries.
func (m awsSyncModel) SelectedRemoves() []int {
	var removes []int
	for i, item := range m.items {
		if item.Action != awssync.SyncStale || !m.selected[i] {
			continue
		}
		removes = append(removes, item.ExistingIndex)
	}
	return removes
}
