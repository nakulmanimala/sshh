package config

import (
	"os"
	"path/filepath"

	"sshh/internal/model"

	"gopkg.in/yaml.v3"
)

// TunnelConfig holds the list of saved tunnel templates.
type TunnelConfig struct {
	Tunnels []model.Tunnel `yaml:"tunnels"`
}

// tunnelFilePath returns the full path to tunnels.yaml.
func tunnelFilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tunnels.yaml"), nil
}

// LoadTunnels reads tunnels from ~/.sshh/tunnels.yaml.
// Returns an empty config if the file doesn't exist.
func LoadTunnels() (*TunnelConfig, error) {
	p, err := tunnelFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &TunnelConfig{}, nil
		}
		return nil, err
	}

	var tc TunnelConfig
	if err := yaml.Unmarshal(data, &tc); err != nil {
		return nil, err
	}
	return &tc, nil
}

// Save writes the tunnel config to ~/.sshh/tunnels.yaml.
func (tc *TunnelConfig) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	p, err := tunnelFilePath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// AddTunnel appends a tunnel and saves.
func (tc *TunnelConfig) AddTunnel(t model.Tunnel) error {
	tc.Tunnels = append(tc.Tunnels, t)
	return tc.Save()
}

// UpdateTunnel replaces the tunnel at index i and saves.
func (tc *TunnelConfig) UpdateTunnel(i int, t model.Tunnel) error {
	if i < 0 || i >= len(tc.Tunnels) {
		return nil
	}
	tc.Tunnels[i] = t
	return tc.Save()
}

// DeleteTunnel removes the tunnel at index i and saves.
func (tc *TunnelConfig) DeleteTunnel(i int) error {
	if i < 0 || i >= len(tc.Tunnels) {
		return nil
	}
	tc.Tunnels = append(tc.Tunnels[:i], tc.Tunnels[i+1:]...)
	return tc.Save()
}

// FindTunnelByName returns the index and tunnel with the given name, or -1 if not found.
func (tc *TunnelConfig) FindTunnelByName(name string) (int, *model.Tunnel) {
	for i := range tc.Tunnels {
		if tc.Tunnels[i].Name == name {
			return i, &tc.Tunnels[i]
		}
	}
	return -1, nil
}
