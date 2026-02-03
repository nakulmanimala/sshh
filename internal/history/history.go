package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"ssh-tool/internal/config"
	"ssh-tool/internal/model"
)

// History tracks when each server was last connected to.
type History struct {
	Entries map[string]time.Time `json:"entries"`
}

// filePath returns the full path to history.json.
func filePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

// Load reads history from disk. Returns empty history if the file doesn't exist.
func Load() (*History, error) {
	p, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{Entries: make(map[string]time.Time)}, nil
		}
		return nil, err
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}
	if h.Entries == nil {
		h.Entries = make(map[string]time.Time)
	}
	return &h, nil
}

// Save writes history to disk.
func (h *History) Save() error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	p, err := filePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// Record marks a server as just used and saves.
func (h *History) Record(serverName string) error {
	h.Entries[serverName] = time.Now()
	return h.Save()
}

// SortByRecent returns a copy of servers sorted by most recently used first.
// Servers with no history are placed at the end in their original order.
func (h *History) SortByRecent(servers []model.Server) []model.Server {
	sorted := make([]model.Server, len(servers))
	copy(sorted, servers)

	sort.SliceStable(sorted, func(i, j int) bool {
		ti, oki := h.Entries[sorted[i].Name]
		tj, okj := h.Entries[sorted[j].Name]
		if oki && okj {
			return ti.After(tj)
		}
		if oki {
			return true
		}
		return false
	})
	return sorted
}
