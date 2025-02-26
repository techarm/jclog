package formatter

import (
	"strings"

	"github.com/fatih/color"
)

// Level color mapping
var levelColors = map[string]func(a ...any) string{
	"DEBUG": color.New(color.FgHiBlack).SprintFunc(),
	"INFO":  color.New(color.FgGreen).SprintFunc(),
	"WARN":  color.New(color.FgYellow).SprintFunc(),
	"ERROR": color.New(color.FgRed).SprintFunc(),
}

// FormatLog dynamically applies formatting and color to log entries
func FormatLog(fields map[string]string, format string, fieldOrder []string) string {
	// If --fields is specified but no --format, construct a space-separated output
	if format == "" {
		values := []string{}
		for _, field := range fieldOrder {
			if val, exists := fields[field]; exists {
				values = append(values, val)
			}
		}
		result := strings.Join(values, " ")

		// Apply color based on log level
		if level, exists := fields["level"]; exists {
			if colorFunc, exists := levelColors[strings.ToUpper(level)]; exists {
				return colorFunc(result)
			}
		}
		return result
	}

	// If --format is specified, apply it dynamically
	result := format
	for _, field := range fieldOrder {
		placeholder := "{" + field + "}"
		if val, exists := fields[field]; exists {
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}

	// Apply color based on log level
	if level, exists := fields["level"]; exists {
		if colorFunc, exists := levelColors[strings.ToUpper(level)]; exists {
			return colorFunc(result)
		}
	}
	return result
}
