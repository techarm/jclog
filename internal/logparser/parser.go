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
		if v, ok := raw["message"].(string); ok {
			entry.Message = v
		}

		// Extract additional fields
		for _, field := range fields {
			if v, ok := raw[field].(string); ok {
				extraFields[field] = v
			}
		}

		fmt.Println(formatter.FormatLog(entry, format, extraFields))
	}
}
