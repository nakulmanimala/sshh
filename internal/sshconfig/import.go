package sshconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sshh/internal/model"
)

// Parse reads ~/.ssh/config and returns discovered server entries.
// It skips wildcard hosts (Host *).
func Parse() ([]model.Server, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return ParseFile(filepath.Join(home, ".ssh", "config"))
}

// ParseFile reads the given ssh config file and returns discovered servers.
func ParseFile(path string) ([]model.Server, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var servers []model.Server
	var current *model.Server

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first whitespace or '='.
		key, val := splitDirective(line)
		key = strings.ToLower(key)

		switch key {
		case "host":
			// Finish previous block.
			if current != nil {
				servers = append(servers, finalize(*current))
			}
			// Skip wildcard patterns.
			if strings.Contains(val, "*") || strings.Contains(val, "?") {
				current = nil
				continue
			}
			current = &model.Server{
				Name: val,
				Port: 22,
				User: "root",
			}
		case "hostname":
			if current != nil {
				current.Host = val
			}
		case "user":
			if current != nil {
				current.User = val
			}
		case "port":
			if current != nil {
				if p, err := strconv.Atoi(val); err == nil {
					current.Port = p
				}
			}
		case "identityfile":
			if current != nil {
				current.Key = expandTilde(val)
			}
		}
	}

	// Don't forget the last block.
	if current != nil {
		servers = append(servers, finalize(*current))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return servers, nil
}

// finalize ensures required fields have defaults.
func finalize(s model.Server) model.Server {
	if s.Host == "" {
		s.Host = s.Name
	}
	return s
}

// splitDirective splits "Key value" or "Key=value" into (key, value).
func splitDirective(line string) (string, string) {
	// Try '=' first.
	if idx := strings.Index(line, "="); idx != -1 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
	}
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 {
		return parts[0], strings.TrimSpace(parts[1])
	}
	return parts[0], ""
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
