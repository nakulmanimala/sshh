package awssync

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListProfiles returns the AWS CLI profile names found in ~/.aws/config and
// ~/.aws/credentials, with "default" sorted first. Falls back to ["default"]
// if neither file exists or no profiles are found.
func ListProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	if err := collectConfigProfiles(filepath.Join(home, ".aws", "config"), seen); err != nil {
		return nil, err
	}
	if err := collectCredentialsProfiles(filepath.Join(home, ".aws", "credentials"), seen); err != nil {
		return nil, err
	}

	if len(seen) == 0 {
		return []string{"default"}, nil
	}

	profiles := make([]string, 0, len(seen))
	for p := range seen {
		profiles = append(profiles, p)
	}
	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i] == "default" {
			return true
		}
		if profiles[j] == "default" {
			return false
		}
		return profiles[i] < profiles[j]
	})
	return profiles, nil
}

// collectConfigProfiles reads section headers from ~/.aws/config, where
// sections look like "[default]" or "[profile name]".
func collectConfigProfiles(path string, seen map[string]bool) error {
	return scanSections(path, func(section string) {
		if section == "default" {
			seen["default"] = true
			return
		}
		if name, ok := strings.CutPrefix(section, "profile "); ok {
			name = strings.TrimSpace(name)
			if name != "" {
				seen[name] = true
			}
		}
	})
}

// collectCredentialsProfiles reads section headers from ~/.aws/credentials,
// where sections are the bare profile name (including "default").
func collectCredentialsProfiles(path string, seen map[string]bool) error {
	return scanSections(path, func(section string) {
		if section != "" {
			seen[section] = true
		}
	})
}

// scanSections calls fn with the contents of each "[...]" section header
// found in the file at path. Missing files are treated as having no sections.
func scanSections(path string, fn func(section string)) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
			continue
		}
		fn(strings.TrimSpace(line[1 : len(line)-1]))
	}
	return scanner.Err()
}
