package config

import (
	"os"
	"path/filepath"

	"sshh/internal/model"

	"gopkg.in/yaml.v3"
)

// Config holds the list of saved servers and tunnels.
type Config struct {
	Servers []model.Server `yaml:"servers"`
	Tunnels []model.Tunnel `yaml:"tunnels"`
}

// Dir returns the config directory path (~/.sshh/).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sshh"), nil
}

// filePath returns the full path to config.yaml.
func filePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config from disk. Returns an empty config if the file doesn't exist.
func Load() (*Config, error) {
	p, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func (c *Config) Save() error {
	dir, err := Dir()
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

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// AddServer appends a server and saves.
func (c *Config) AddServer(s model.Server) error {
	c.Servers = append(c.Servers, s)
	return c.Save()
}

// UpdateServer replaces the server at index i and saves.
func (c *Config) UpdateServer(i int, s model.Server) error {
	if i < 0 || i >= len(c.Servers) {
		return nil
	}
	c.Servers[i] = s
	return c.Save()
}

// DeleteServer removes the server at index i and saves.
func (c *Config) DeleteServer(i int) error {
	if i < 0 || i >= len(c.Servers) {
		return nil
	}
	c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
	return c.Save()
}

// FindByName returns the index and server with the given name, or -1 if not found.
func (c *Config) FindByName(name string) (int, *model.Server) {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return i, &c.Servers[i]
		}
	}
	return -1, nil
}

// AddTunnel appends a tunnel and saves.
func (c *Config) AddTunnel(t model.Tunnel) error {
	c.Tunnels = append(c.Tunnels, t)
	return c.Save()
}

// UpdateTunnel replaces the tunnel at index i and saves.
func (c *Config) UpdateTunnel(i int, t model.Tunnel) error {
	if i < 0 || i >= len(c.Tunnels) {
		return nil
	}
	c.Tunnels[i] = t
	return c.Save()
}

// DeleteTunnel removes the tunnel at index i and saves.
func (c *Config) DeleteTunnel(i int) error {
	if i < 0 || i >= len(c.Tunnels) {
		return nil
	}
	c.Tunnels = append(c.Tunnels[:i], c.Tunnels[i+1:]...)
	return c.Save()
}

// FindTunnelByName returns the index and tunnel with the given name, or -1 if not found.
func (c *Config) FindTunnelByName(name string) (int, *model.Tunnel) {
	for i := range c.Tunnels {
		if c.Tunnels[i].Name == name {
			return i, &c.Tunnels[i]
		}
	}
	return -1, nil
}
