package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ActiveProfile != "default" {
		t.Errorf("Expected active profile to be 'default', got '%s'", cfg.ActiveProfile)
	}

	if _, exists := cfg.Profiles["default"]; !exists {
		t.Error("Expected default profile to exist")
	}

	defaultProfile := cfg.Profiles["default"]
	if defaultProfile.MaxDepth != 2 {
		t.Errorf("Expected default max depth to be 2, got %d", defaultProfile.MaxDepth)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test loading non-existent config file
	configPath := filepath.Join(tmpDir, "nonexistent.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error when loading nonexistent config, got %v", err)
	}
	if cfg.ActiveProfile != "default" {
		t.Errorf("Expected default profile when loading nonexistent config, got %s", cfg.ActiveProfile)
	}

	// Create test config file
	testConfig := &Config{
		ActiveProfile: "test",
		Profiles: map[string]Profile{
			"test": {
				Format:      "{timestamp} {message}",
				Fields:      []string{"timestamp", "message"},
				MaxDepth:    1,
				HideMissing: true,
				Filters:     []string{"level=INFO"},
				Excludes:    []string{"level=DEBUG"},
			},
		},
	}

	configPath = filepath.Join(tmpDir, "config.json")
	if err := SaveConfig(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Test loading config file
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.ActiveProfile != "test" {
		t.Errorf("Expected active profile to be 'test', got '%s'", loadedConfig.ActiveProfile)
	}

	testProfile := loadedConfig.Profiles["test"]
	if testProfile.Format != "{timestamp} {message}" {
		t.Errorf("Expected format '{timestamp} {message}', got '%s'", testProfile.Format)
	}
}

func TestGetActiveProfile(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "test",
		Profiles: map[string]Profile{
			"default": {
				Format: "default_format",
			},
			"test": {
				Format: "test_format",
			},
		},
	}

	// Test getting existing profile
	profile := cfg.GetActiveProfile()
	if profile.Format != "test_format" {
		t.Errorf("Expected format 'test_format', got '%s'", profile.Format)
	}

	// Test getting non-existent profile (should return default)
	cfg.ActiveProfile = "nonexistent"
	profile = cfg.GetActiveProfile()
	if profile.Format != "default_format" {
		t.Errorf("Expected format 'default_format', got '%s'", profile.Format)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test saving config to non-existent directory
	configPath := filepath.Join(tmpDir, "subdir", "config.json")
	cfg := DefaultConfig()
	if err := SaveConfig(cfg, configPath); err != nil {
		t.Errorf("Failed to save config to new directory: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Test saving and loading config
	cfg.ActiveProfile = "test"
	cfg.Profiles["test"] = Profile{
		Format: "test_format",
	}

	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.ActiveProfile != "test" {
		t.Errorf("Expected active profile 'test', got '%s'", loadedCfg.ActiveProfile)
	}

	if loadedCfg.Profiles["test"].Format != "test_format" {
		t.Errorf("Expected format 'test_format', got '%s'", loadedCfg.Profiles["test"].Format)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()
	if path == "" {
		t.Error("Expected non-empty default config path")
	}

	// Test case when HOME directory is not set
	oldHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", oldHome)

	path = GetDefaultConfigPath()
	if path != ".jclog.json" {
		t.Errorf("Expected '.jclog.json' when HOME is not set, got '%s'", path)
	}
}
