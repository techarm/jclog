package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
}

// Profile represents a single configuration profile
type Profile struct {
	Format           string            `json:"format"`
	Fields           []string          `json:"fields"`
	MaxDepth         int               `json:"max_depth"`
	HideMissing      bool              `json:"hide_missing"`
	Filters          []string          `json:"filters"`
	Excludes         []string          `json:"excludes"`
	LevelMappings    map[string]string `json:"level_mappings"`
	AutoConvertLevel bool              `json:"auto_convert_level"`
	TimeFormat       string            `json:"time_format"`
}

// DefaultConfig creates a new configuration with default values
func DefaultConfig() *Config {
	return &Config{
		ActiveProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				Format:           "{time} [{level}] {msg} ({name})",
				MaxDepth:         2,
				HideMissing:      false,
				Filters:          []string{},
				Excludes:         []string{},
				AutoConvertLevel: false,
				TimeFormat:       "2006/01/02 15:04:05.000",
				LevelMappings: map[string]string{
					"10": "TRACE",
					"20": "DEBUG",
					"30": "INFO",
					"40": "WARN",
					"50": "ERROR",
					"60": "FATAL",
				},
			},
		},
	}
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Ensure default profile exists
	if _, exists := config.Profiles["default"]; !exists {
		config.Profiles["default"] = DefaultConfig().Profiles["default"]
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default path for the config file
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".jclog.json"
	}
	return filepath.Join(homeDir, ".jclog.json")
}

// GetActiveProfile returns the active profile from the configuration
func (c *Config) GetActiveProfile() Profile {
	if profile, exists := c.Profiles[c.ActiveProfile]; exists {
		return profile
	}
	return c.Profiles["default"]
}
