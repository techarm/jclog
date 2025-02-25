package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/techarm/json-log-viewer/internal/formatter"
	"github.com/techarm/json-log-viewer/internal/types"
)

// ProcessLog parses JSON logs and outputs formatted results
func ProcessLog(scanner *bufio.Scanner, format string, fields []string) {
	for scanner.Scan() {
		var entry types.LogEntry
		extraFields := make(map[string]string)

		raw := make(map[string]any)
		if err := json.Unmarshal([]byte(scanner.Text()), &raw); err != nil {
			fmt.Println("Invalid JSON:", scanner.Text())
			continue
		}

		// Extract default fields
		if v, ok := raw["timestamp"].(string); ok {
			entry.Timestamp = v
		}
		if v, ok := raw["level"].(string); ok {
			entry.Level = v
		}

		// Extract message field (which is a JSON string)
		if v, ok := raw["message"].(string); ok {
			entry.Message = v

			// Try to parse `message` as JSON
			messageFields := make(map[string]string)
			var messageJSON map[string]any
			if err := json.Unmarshal([]byte(v), &messageJSON); err == nil {
				// If message is a valid JSON, extract its fields
				for key, val := range messageJSON {
					if strVal, ok := val.(string); ok {
						messageFields["message."+key] = strVal
					} else if numVal, ok := val.(float64); ok {
						messageFields["message."+key] = fmt.Sprintf("%.0f", numVal)
					}
				}
				// Merge message fields into extraFields
				for k, v := range messageFields {
					extraFields[k] = v
				}
			}
		}

		// Extract additional fields specified by --fields
		for _, field := range fields {
			if v, ok := raw[field].(string); ok {
				extraFields[field] = v
			}
		}

		fmt.Println(formatter.FormatLog(entry, format, extraFields))
	}
}
