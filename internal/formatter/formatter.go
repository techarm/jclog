package formatter

import (
	"strings"

	"github.com/fatih/color"
	"github.com/techarm/json-log-viewer/internal/types"
)

// Level color mapping
var levelColors = map[string]func(a ...interface{}) string{
	"DEBUG": color.New(color.FgCyan).SprintFunc(),
	"INFO":  color.New(color.FgBlue).SprintFunc(),
	"WARN":  color.New(color.FgYellow).SprintFunc(),
	"ERROR": color.New(color.FgRed).SprintFunc(),
}

// FormatLog applies formatting and color to log entries
func FormatLog(entry types.LogEntry, format string, extraFields map[string]string) string {
	// Replace placeholders
	result := strings.ReplaceAll(format, "{timestamp}", entry.Timestamp)
	result = strings.ReplaceAll(result, "{level}", entry.Level)
	result = strings.ReplaceAll(result, "{message}", entry.Message)

	// Replace additional fields
	for key, value := range extraFields {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Apply color to level
	if colorFunc, exists := levelColors[entry.Level]; exists {
		result = strings.ReplaceAll(result, entry.Level, colorFunc(entry.Level))
	}
	return result
}
