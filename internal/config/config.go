package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ayushsarode/orb/internal/transport"
)

const ConfigFile = ".orb/config"

// Config represents the repository configuration
type Config struct {
	Values  map[string]string
	Remotes []transport.RemoteConfig
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Values:  make(map[string]string),
		Remotes: []transport.RemoteConfig{},
	}

	file, err := os.Open(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var section string
	var inRemoteSection bool
	var currentRemote transport.RemoteConfig

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse section headers: [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// If we were in a remote section, add the remote to the list
			if inRemoteSection && currentRemote.Name != "" {
				cfg.Remotes = append(cfg.Remotes, currentRemote)
				currentRemote = transport.RemoteConfig{}
			}

			section = line[1 : len(line)-1]

			// Check if this is a remote section
			if strings.HasPrefix(section, "remote ") {
				inRemoteSection = true
				currentRemote.Name = strings.TrimPrefix(section, "remote ")
			} else {
				inRemoteSection = false
			}
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Handle remote section values
		if inRemoteSection {
			if key == "url" {
				currentRemote.URL = value
			}
			// Store as remote.<name>.<key> = value for backward compatibility
			cfg.Values[fmt.Sprintf("remote.%s.%s", currentRemote.Name, key)] = value
		} else {
			// Store as section.key = value for regular sections
			if section != "" {
				key = section + "." + key
			}
			cfg.Values[key] = value
		}
	}

	// Add the last remote if we were in a remote section
	if inRemoteSection && currentRemote.Name != "" {
		cfg.Remotes = append(cfg.Remotes, currentRemote)
	}

	return cfg, nil
}

// Get returns a configuration value
func (cfg *Config) Get(key string) string {
	return cfg.Values[key]
}

// GetString returns a configuration value or a default if not found
func (cfg *Config) GetString(key, defaultValue string) string {
	if value, ok := cfg.Values[key]; ok {
		return value
	}
	return defaultValue
}

// Set sets a configuration value
func (cfg *Config) Set(key, value string) {
	cfg.Values[key] = value
}

// Save writes the configuration to the config file
// SaveConfig writes a configuration to the config file
func SaveConfig(cfg *Config) error {
	// Open or create config file
	file, err := os.OpenFile(ConfigFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer file.Close()

	// Group configuration by section
	sections := make(map[string]map[string]string)

	for fullKey, value := range cfg.Values {
		// Skip remote configurations as we'll handle them separately
		if strings.HasPrefix(fullKey, "remote.") && strings.Count(fullKey, ".") >= 2 {
			continue
		}

		parts := strings.SplitN(fullKey, ".", 2)

		if len(parts) != 2 {
			// Keys without a section go into the empty section
			if sections[""] == nil {
				sections[""] = make(map[string]string)
			}
			sections[""][fullKey] = value
			continue
		}

		section, key := parts[0], parts[1]
		if sections[section] == nil {
			sections[section] = make(map[string]string)
		}
		sections[section][key] = value
	}

	// Write each regular section
	for section, values := range sections {
		if section != "" {
			fmt.Fprintf(file, "[%s]\n", section)
		}

		for key, value := range values {
			fmt.Fprintf(file, "\t%s = %s\n", key, value)
		}

		fmt.Fprintln(file)
	}

	// Write remote sections
	for _, remote := range cfg.Remotes {
		fmt.Fprintf(file, "[remote %s]\n", remote.Name)
		fmt.Fprintf(file, "\turl = %s\n", remote.URL)
		fmt.Fprintln(file)
	}

	return nil
}

// Save saves the configuration to the config file
func (cfg *Config) Save() error {
	return SaveConfig(cfg)
}
