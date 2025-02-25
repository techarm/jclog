package formatter

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/techarm/json-log-viewer/internal/types"
)

// Level color mapping
var levelColors = map[string]func(a ...any) string{
	"DEBUG": color.New(color.FgHiBlack).SprintFunc(),
	"INFO":  color.New(color.FgGreen).SprintFunc(),
	"WARN":  color.New(color.FgYellow).SprintFunc(),
	"ERROR": color.New(color.FgRed).SprintFunc(),
}

// padLevel ensures the log level is always 5 characters long
func padLevel(level string) string {
	return fmt.Sprintf("%-5s", level) // Left-aligned padding with spaces
}

// FormatLog applies formatting and color to log entries
func FormatLog(entry types.LogEntry, format string, extraFields map[string]string) string {
	// Pad the log level
	paddedLevel := padLevel(entry.Level)

	// Replace placeholders
	result := strings.ReplaceAll(format, "{timestamp}", entry.Timestamp)
	result = strings.ReplaceAll(result, "{level}", paddedLevel)
	result = strings.ReplaceAll(result, "{message}", entry.Message)

	// Replace additional fields, including extracted message fields
	for key, value := range extraFields {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Apply color to the entire log line
	if colorFunc, exists := levelColors[entry.Level]; exists {
		return colorFunc(result) // Apply color to the whole line
	}
	return result // Return uncolored line if no level match
}
