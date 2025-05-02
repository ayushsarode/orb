package config

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

const ConfigFile = ".orb/config"

// Config represents the repository configuration
type Config struct {
    Values map[string]string
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
    cfg := &Config{
        Values: make(map[string]string),
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
    
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        
        // Parse section headers: [section]
        if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
            section = line[1 : len(line)-1]
            continue
        }
        
        // Parse key-value pairs
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        
        // Store as section.key = value
        if section != "" {
            key = section + "." + key
        }
        
        cfg.Values[key] = value
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
func (cfg *Config) Save() error {
    // Group configuration by section
    sections := make(map[string]map[string]string)
    
    for fullKey, value := range cfg.Values {
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
    
    // Write to file
    file, err := os.Create(ConfigFile)
    if err != nil {
        return fmt.Errorf("creating config file: %w", err)
    }
    defer file.Close()
    
    // Write each section
    for section, values := range sections {
        if section != "" {
            fmt.Fprintf(file, "[%s]\n", section)
        }
        
        for key, value := range values {
            fmt.Fprintf(file, "\t%s = %s\n", key, value)
        }
        
        fmt.Fprintln(file)
    }
    
    return nil
}