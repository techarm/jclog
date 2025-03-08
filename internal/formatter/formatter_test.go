package formatter

import (
	"strings"
	"testing"
)

func TestFormatLog(t *testing.T) {
	tests := []struct {
		name        string
		fields      map[string]string
		format      string
		fieldOrder  []string
		hideMissing bool
		want        string
	}{
		{
			name: "Basic formatting",
			fields: map[string]string{
				"timestamp": "2024-03-20",
				"level":     "INFO",
				"message":   "test message",
			},
			format:      "{timestamp} [{level}] {message}",
			fieldOrder:  []string{"timestamp", "level", "message"},
			hideMissing: false,
			want:        "2024-03-20 [INFO] test message",
		},
		{
			name: "No format string",
			fields: map[string]string{
				"timestamp": "2024-03-20",
				"level":     "INFO",
				"message":   "test message",
			},
			format:      "",
			fieldOrder:  []string{"timestamp", "level", "message"},
			hideMissing: false,
			want:        "2024-03-20 INFO test message",
		},
		{
			name: "Missing fields not hidden",
			fields: map[string]string{
				"timestamp": "2024-03-20",
				"message":   "test message",
			},
			format:      "{timestamp} [{level}] {message}",
			fieldOrder:  []string{"timestamp", "level", "message"},
			hideMissing: false,
			want:        "2024-03-20 [] test message",
		},
		{
			name: "Missing fields hidden",
			fields: map[string]string{
				"timestamp": "2024-03-20",
				"message":   "test message",
			},
			format:      "{timestamp} [{level}] {message}",
			fieldOrder:  []string{"timestamp", "level", "message"},
			hideMissing: true,
			want:        "2024-03-20 test message",
		},
		{
			name: "Different log level colors",
			fields: map[string]string{
				"level": "ERROR",
			},
			format:      "{level}",
			fieldOrder:  []string{"level"},
			hideMissing: false,
			want:        "ERROR", // Note: actual output will include color codes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLog(tt.fields, tt.format, tt.fieldOrder, tt.hideMissing)
			// Remove ANSI color codes for comparison
			got = stripANSI(got)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

// stripANSI removes ANSI color codes from a string
func stripANSI(str string) string {
	// ANSI color code replacer
	r := strings.NewReplacer(
		"\x1b[0m", "",
		"\x1b[31m", "",
		"\x1b[32m", "",
		"\x1b[33m", "",
		"\x1b[37m", "",
		"\x1b[90m", "",
	)
	return r.Replace(str)
}
