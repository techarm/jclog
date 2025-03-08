package logparser

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestProcessLog(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		format        string
		maxDepth      int
		filters       map[string]string
		excludes      map[string]string
		levelMappings map[string]string
		wantOutput    bool
	}{
		{
			name:          "Basic JSON log",
			input:         `{"timestamp": "2024-03-20", "level": "INFO", "message": "test message"}`,
			format:        "{timestamp} [{level}] {message}",
			maxDepth:      2,
			filters:       map[string]string{},
			excludes:      map[string]string{},
			levelMappings: nil,
			wantOutput:    true,
		},
		{
			name:          "Invalid JSON",
			input:         "invalid json",
			format:        "{timestamp} [{level}] {message}",
			maxDepth:      2,
			filters:       map[string]string{},
			excludes:      map[string]string{},
			levelMappings: nil,
			wantOutput:    true, // will output error message
		},
		{
			name:          "Log with filter",
			input:         `{"timestamp": "2024-03-20", "level": "INFO", "message": "test message"}`,
			format:        "{timestamp} [{level}] {message}",
			maxDepth:      2,
			filters:       map[string]string{"level": "INFO"},
			excludes:      map[string]string{},
			levelMappings: nil,
			wantOutput:    true,
		},
		{
			name:          "Log with exclude",
			input:         `{"timestamp": "2024-03-20", "level": "DEBUG", "message": "test message"}`,
			format:        "{timestamp} [{level}] {message}",
			maxDepth:      2,
			filters:       map[string]string{},
			excludes:      map[string]string{"level": "DEBUG"},
			levelMappings: nil,
			wantOutput:    false,
		},
		{
			name:          "Nested message",
			input:         `{"timestamp": "2024-03-20", "level": "INFO", "message": "{\"nested\": \"value\"}"}`,
			format:        "{timestamp} [{level}] {message.nested}",
			maxDepth:      2,
			filters:       map[string]string{},
			excludes:      map[string]string{},
			levelMappings: nil,
			wantOutput:    true,
		},
		{
			name:     "Bunyan log with level mapping",
			input:    `{"time": "2024-03-20", "level": 30, "msg": "test message"}`,
			format:   "{time} [{level}] {msg}",
			maxDepth: 2,
			filters:  map[string]string{},
			excludes: map[string]string{},
			levelMappings: map[string]string{
				"30": "INFO",
				"40": "WARN",
				"50": "ERROR",
			},
			wantOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create output capture goroutine
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			// Process log
			reader := strings.NewReader(tt.input)
			scanner := bufio.NewScanner(reader)
			ProcessLog(scanner, tt.format, tt.maxDepth, false, tt.filters, tt.excludes, tt.levelMappings)

			// Close write end of pipe and read output
			w.Close()
			out := <-outC

			// Cleanup and restore stdout
			os.Stdout = old

			if tt.wantOutput && out == "" {
				t.Errorf("Expected output but got none")
			} else if !tt.wantOutput && out != "" {
				t.Errorf("Expected no output but got: %s", out)
			}
		})
	}
}

func TestFlattenJSONString(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		prefix      string
		maxDepth    int
		wantResults map[string]string
	}{
		{
			name:     "Simple string",
			jsonStr:  "simple string",
			prefix:   "test",
			maxDepth: 1,
			wantResults: map[string]string{
				"test": "simple string",
			},
		},
		{
			name:     "Simple JSON",
			jsonStr:  `{"key": "value"}`,
			prefix:   "test",
			maxDepth: 1,
			wantResults: map[string]string{
				"test.key": "value",
			},
		},
		{
			name:     "Nested JSON",
			jsonStr:  `{"outer": {"inner": "value"}}`,
			prefix:   "test",
			maxDepth: 2,
			wantResults: map[string]string{
				"test.outer.inner": "value",
			},
		},
		{
			name:     "Exceeds max depth",
			jsonStr:  `{"l1": {"l2": {"l3": "value"}}}`,
			prefix:   "test",
			maxDepth: 1,
			wantResults: map[string]string{
				"test.l1": `{"l2":{"l3":"value"}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]string)
			flattenJSONString(tt.jsonStr, tt.prefix, result, tt.maxDepth, 1)

			if len(result) != len(tt.wantResults) {
				t.Errorf("Expected %d results, got %d", len(tt.wantResults), len(result))
			}

			for k, v := range tt.wantResults {
				if got, ok := result[k]; !ok || got != v {
					t.Errorf("Expected %s=%s, got %s", k, v, got)
				}
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		field     string
		wantValue string
	}{
		{
			name: "Direct field",
			data: map[string]any{
				"test": "value",
			},
			field:     "test",
			wantValue: "value",
		},
		{
			name: "Alias field",
			data: map[string]any{
				"msg": "message value",
			},
			field:     "message",
			wantValue: "message value",
		},
		{
			name: "Numeric field",
			data: map[string]any{
				"number": float64(123),
			},
			field:     "number",
			wantValue: "123",
		},
		{
			name: "Non-existent field",
			data: map[string]any{
				"exists": "value",
			},
			field:     "nonexistent",
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFieldValue(tt.data, tt.field)
			if got != tt.wantValue {
				t.Errorf("getFieldValue() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestMatchFilters(t *testing.T) {
	tests := []struct {
		name    string
		fields  map[string]string
		filters map[string]string
		want    bool
	}{
		{
			name: "Exact match",
			fields: map[string]string{
				"level": "INFO",
				"env":   "prod",
			},
			filters: map[string]string{
				"level": "INFO",
			},
			want: true,
		},
		{
			name: "No match",
			fields: map[string]string{
				"level": "DEBUG",
			},
			filters: map[string]string{
				"level": "INFO",
			},
			want: false,
		},
		{
			name: "Field does not exist",
			fields: map[string]string{
				"level": "INFO",
			},
			filters: map[string]string{
				"nonexistent": "value",
			},
			want: false,
		},
		{
			name: "Empty filter",
			fields: map[string]string{
				"level": "INFO",
			},
			filters: map[string]string{},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchFilters(tt.fields, tt.filters); got != tt.want {
				t.Errorf("matchFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}
