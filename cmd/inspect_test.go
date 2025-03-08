package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInspectCommand(t *testing.T) {
	// Create a temporary test file
	testLog := `{"timestamp":"2024-03-20T10:00:00Z","level":"info","message":"test message","service":"test-service"}`
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testLog); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "Inspect log file",
			args:    []string{"jclog", "inspect", tmpFile.Name()},
			wantErr: false,
			contains: []string{
				"timestamp",
				"level",
				"message",
				"service",
				"Type: string",
			},
		},
		{
			name:    "Missing file path",
			args:    []string{"jclog", "inspect"},
			wantErr: true,
		},
		{
			name:    "Non-existent file",
			args:    []string{"jclog", "inspect", "non-existent.log"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			cmd := NewRootCommand()
			err := cmd.Run(context.Background(), tt.args)

			// Copy output in background
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			// Close write end of pipe
			w.Close()

			// Restore stdout
			os.Stdout = oldStdout

			// Read output
			out := <-outC

			if (err != nil) != tt.wantErr {
				t.Errorf("Inspect command error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for _, want := range tt.contains {
					if !strings.Contains(out, want) {
						t.Errorf("Output should contain %q but got:\n%s", want, out)
					}
				}
			}
		})
	}
}
