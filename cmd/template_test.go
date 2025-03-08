package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestTemplateCommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "List templates",
			args:    []string{"jclog", "template", "list"},
			wantErr: false,
			contains: []string{
				"basic",
				"detailed",
				"compact",
				"debug",
				"json",
				"metrics",
			},
		},
		{
			name:     "Show template details - basic",
			args:     []string{"jclog", "template", "show", "basic"},
			wantErr:  false,
			contains: []string{"{timestamp} [{level}] {message}"},
		},
		{
			name:    "Show template details - unknown template",
			args:    []string{"jclog", "template", "show", "unknown"},
			wantErr: true,
		},
		{
			name:    "Show template details - missing name",
			args:    []string{"jclog", "template", "show"},
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
				t.Errorf("Template command error = %v, wantErr %v", err, tt.wantErr)
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
