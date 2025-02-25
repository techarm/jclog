package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/techarm/jclog/internal/formatter"
	"github.com/techarm/jclog/internal/types"
)

// FieldAliases defines field name aliases for better flexibility
var FieldAliases = map[string][]string{
	"timestamp": {"timestamp", "time"},
	"level":     {"level", "lvl"},
	"message":   {"message", "msg"},
}

// ProcessLog parses JSON logs and outputs formatted results
func ProcessLog(scanner *bufio.Scanner, format string, fields []string, maxDepth int64) {
	for scanner.Scan() {
		var entry types.LogEntry
		extraFields := make(map[string]string)

		raw := make(map[string]any)
		if err := json.Unmarshal([]byte(scanner.Text()), &raw); err != nil {
			// fmt.Println("Invalid JSON:", scanner.Text())
			fmt.Println(scanner.Text())
			continue
		}

		// Extract log level (case insensitive)
		entry.Timestamp = getFieldValue(raw, "timestamp")
		entry.Level = getFieldValue(raw, "level")
		entry.Level = strings.ToUpper(entry.Level) // Normalize log level

		// Extract message field (either "message" or "msg")
		entry.Message = getFieldValue(raw, "message")
		if entry.Message != "" {
			messageFields := make(map[string]string)
			flattenJSONString(entry.Message, "message", messageFields, maxDepth, 1)
			for k, v := range messageFields {
				extraFields[k] = v
			}
		}

		// Extract additional fields specified by --fields
		for _, field := range fields {
			extraFields[field] = getFieldValue(raw, field)
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
		case map[string]any:
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

// getFieldValue retrieves the first available value from field aliases
func getFieldValue(data map[string]any, field string) string {
	aliases, exists := FieldAliases[field]
	if !exists {
		aliases = []string{field} // No alias, just use the field name
	}

	for _, alias := range aliases {
		if v, ok := data[alias].(string); ok {
			return v
		}
		if v, ok := data[alias].(float64); ok {
			return fmt.Sprintf("%.0f", v)
		}
	}
	return ""
}
