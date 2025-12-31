package aws

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GetProfiles returns a list of all AWS profiles found in ~/.aws/config and ~/.aws/credentials
func GetProfiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get home directory: %w", err)
	}

	profiles := make(map[string]struct{})

	configPath := filepath.Join(home, ".aws", "config")
	if _, err := os.Stat(configPath); err == nil {
		if err := parseProfiles(configPath, profiles, true); err != nil {
			return nil, err
		}
	}

	credsPath := filepath.Join(home, ".aws", "credentials")
	if _, err := os.Stat(credsPath); err == nil {
		if err := parseProfiles(credsPath, profiles, false); err != nil {
			return nil, err
		}
	}

	result := make([]string, 0, len(profiles))
	for p := range profiles {
		result = append(result, p)
	}
	sort.Strings(result)

	if len(result) == 0 {
		result = append(result, "default")
	}

	return result, nil
}

func parseProfiles(path string, profiles map[string]struct{}, isConfig bool) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.Trim(line, "[]")
			if isConfig && strings.HasPrefix(name, "profile ") {
				name = strings.TrimPrefix(name, "profile ")
			}
			if name != "" {
				profiles[name] = struct{}{}
			}
		}
	}

	return scanner.Err()
}
