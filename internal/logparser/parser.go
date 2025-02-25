package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/techarm/json-log-viewer/internal/formatter"
	"github.com/techarm/json-log-viewer/internal/types"
)

// ProcessLog parses JSON logs and outputs formatted results
func ProcessLog(scanner *bufio.Scanner, format string, fields []string, maxDepth int64) {
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

		// Extract and recursively decode the "message" field
		if v, ok := raw["message"].(string); ok {
			entry.Message = v
			messageFields := make(map[string]string)
			flattenJSONString(v, "message", messageFields, maxDepth, 1)
			for k, v := range messageFields {
				extraFields[k] = v
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

// flattenJSONString tries to decode nested JSON strings recursively
func flattenJSONString(jsonStr string, prefix string, result map[string]string, maxDepth int64, currentDepth int64) {
	if currentDepth > maxDepth {
		return // Stop if max depth is reached
	}

	// Try to parse the string as JSON
	parsedJSON, err := tryParseJSON(jsonStr)
	if err != nil {
		result[prefix] = jsonStr // Store as a normal string if it's not JSON
		return
	}

	// Recursively flatten the JSON object
	for key, val := range parsedJSON {
		newKey := fmt.Sprintf("%s.%s", prefix, key)
		switch v := val.(type) {
		case string:
			flattenJSONString(v, newKey, result, maxDepth, currentDepth+1) // Recursively parse if it's an escaped JSON string
		case float64:
			result[newKey] = fmt.Sprintf("%.0f", v)
		case map[string]interface{}:
			nestedJSON, err := json.Marshal(v)
			if err == nil {
				flattenJSONString(string(nestedJSON), newKey, result, maxDepth, currentDepth+1)
			}
		}
	}
}

// tryParseJSON attempts to parse a JSON string and returns a map if successful
func tryParseJSON(jsonStr string) (map[string]any, error) {
	var parsed map[string]any
	err := json.Unmarshal([]byte(jsonStr), &parsed)
	return parsed, err
}
