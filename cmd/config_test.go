package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigCommands(t *testing.T) {
	// Save original args and restore them after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test environment
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	ctx := context.Background()
	rootCmd := NewRootCommand()

	// Test initialize configuration
	t.Run("Init Config", func(t *testing.T) {
		args := []string{"jclog", "config", "init"}
		if err := rootCmd.Run(ctx, args); err != nil {
			t.Errorf("Config init failed: %v", err)
		}

		configPath := filepath.Join(tmpDir, ".jclog.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
	})

	// Test add profile
	t.Run("Add Profile", func(t *testing.T) {
		args := []string{"jclog", "config", "add-profile",
			"--name", "test",
			"--format", "{timestamp} {message}",
			"--fields", "timestamp,message",
			"--max-depth", "1",
			"--hide-missing",
			"--filter", "level=INFO",
		}
		if err := rootCmd.Run(ctx, args); err != nil {
			t.Errorf("Add profile failed: %v", err)
		}
	})

	// Test show configuration
	t.Run("Show Config", func(t *testing.T) {
		args := []string{"jclog", "config", "show"}
		if err := rootCmd.Run(ctx, args); err != nil {
			t.Errorf("Show config failed: %v", err)
		}
	})

	// Test set active profile
	t.Run("Set Active Profile", func(t *testing.T) {
		args := []string{"jclog", "config", "set-active", "--name", "test"}
		if err := rootCmd.Run(ctx, args); err != nil {
			t.Errorf("Set active profile failed: %v", err)
		}
	})

	// Test remove profile
	t.Run("Remove Profile", func(t *testing.T) {
		args := []string{"jclog", "config", "remove-profile", "--name", "test"}
		if err := rootCmd.Run(ctx, args); err != nil {
			t.Errorf("Remove profile failed: %v", err)
		}
	})

	// Test error cases
	t.Run("Error Cases", func(t *testing.T) {
		// Try to remove default profile
		args := []string{"jclog", "config", "remove-profile", "--name", "default"}
		if err := rootCmd.Run(ctx, args); err == nil {
			t.Error("Expected error when removing default profile")
		}

		// Try to set non-existent profile as active
		args = []string{"jclog", "config", "set-active", "--name", "nonexistent"}
		if err := rootCmd.Run(ctx, args); err == nil {
			t.Error("Expected error when setting nonexistent profile as active")
		}
	})
}
