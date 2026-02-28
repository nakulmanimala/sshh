package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Settings holds user preferences persisted to ~/.sshh/settings.yaml.
type Settings struct {
	SSHColor    string `yaml:"ssh_color"`
	TunnelColor string `yaml:"tunnel_color"`
}

func settingsFilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.yaml"), nil
}

// LoadSettings reads ~/.sshh/settings.yaml. Returns defaults if the file doesn't exist.
func LoadSettings() (*Settings, error) {
	p, err := settingsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{SSHColor: "39", TunnelColor: "214"}, nil
		}
		return nil, err
	}
	s := &Settings{SSHColor: "39", TunnelColor: "214"}
	if err := yaml.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Save writes the settings to ~/.sshh/settings.yaml.
func (s *Settings) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	p, err := settingsFilePath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}
