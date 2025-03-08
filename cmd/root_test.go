package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/techarm/jclog/internal/config"
)

func TestParseFilterArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "Empty arguments",
			args: []string{},
			want: map[string]string{},
		},
		{
			name: "Single filter",
			args: []string{"level=INFO"},
			want: map[string]string{
				"level": "INFO",
			},
		},
		{
			name: "Multiple filters",
			args: []string{"level=INFO", "env=prod"},
			want: map[string]string{
				"level": "INFO",
				"env":   "prod",
			},
		},
		{
			name: "Invalid format",
			args: []string{"invalid"},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFilterArgs(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("parseFilterArgs() got %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseFilterArgs() got[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestRootCommand(t *testing.T) {
	// Save original args and restore them after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "jclog_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config file
	configPath := filepath.Join(tmpDir, "config.json")
	testConfig := config.DefaultConfig()
	testConfig.Profiles["test"] = config.Profile{
		Format:      "{timestamp} {message}",
		Fields:      []string{"timestamp", "message"},
		MaxDepth:    1,
		HideMissing: true,
		Filters:     []string{"level=INFO"},
		Excludes:    []string{},
	}
	if err := config.SaveConfig(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Create test log file
	logPath := filepath.Join(tmpDir, "test.log")
	logContent := `{"timestamp": "2024-03-20", "level": "INFO", "message": "test message"}
{"timestamp": "2024-03-20", "level": "DEBUG", "message": "debug message"}`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Create output capture goroutine
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "Basic command",
			args:    []string{"jclog", "--config", configPath, logPath},
			wantErr: false,
		},
		{
			name:    "With profile",
			args:    []string{"jclog", "--config", configPath, "--profile", "test", logPath},
			wantErr: false,
		},
		{
			name:    "Invalid file",
			args:    []string{"jclog", "--config", configPath, "nonexistent.log"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			cmd := NewRootCommand()
			err := cmd.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("RootCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
