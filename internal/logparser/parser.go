package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/techarm/jclog/internal/formatter"
)

// FieldAliases defines field name aliases for better flexibility
var FieldAliases = map[string][]string{
	"timestamp": {"timestamp", "time"},
	"level":     {"level", "lvl"},
	"message":   {"message", "msg"},
}

// Default fields when --fields is not specified
var defaultFields = []string{"timestamp", "level", "message"}

// ProcessLog parses JSON logs and outputs formatted results
func ProcessLog(scanner *bufio.Scanner, format string, fields []string, maxDepth int, hideMissing bool, filters map[string]string, excludes map[string]string) {
	// If no --fields is provided, use default fields
	if len(fields) == 0 {
		fields = defaultFields
	}

	for scanner.Scan() {
		// Parse the log line as JSON
		raw := make(map[string]any)
		if err := json.Unmarshal([]byte(scanner.Text()), &raw); err != nil {
			fmt.Println("Invalid JSON:", scanner.Text())
			continue
		}

		// Extract only the fields that are specified in --fields
		extractedFields := make(map[string]string)
		for _, field := range fields {
			extractedFields[field] = getFieldValue(raw, field)
		}

		// Handle nested message fields dynamically
		if msg, exists := extractedFields["message"]; exists && msg != "" {
			messageFields := make(map[string]string)
			flattenJSONString(msg, "message", messageFields, maxDepth, 1)
			for k, v := range messageFields {
				if slices.Contains(fields, k) { // Only add if it's in the --fields list
					extractedFields[k] = v
				}
			}
		}

		// Apply filters (only show matching logs)
		if len(filters) > 0 && !matchFilters(extractedFields, filters) {
			continue
		}

		// Apply excludes (hide matching logs)
		if len(excludes) > 0 && matchFilters(extractedFields, excludes) {
			continue
		}

		fmt.Println(formatter.FormatLog(extractedFields, format, fields, hideMissing))
	}
}

// flattenJSONString tries to decode nested JSON strings recursively
func flattenJSONString(jsonStr string, prefix string, result map[string]string, maxDepth int, currentDepth int) {
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

// matchFilters checks if all filter conditions are met
func matchFilters(fields map[string]string, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	for key, expectedValue := range filters {
		if actualValue, exists := fields[key]; !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}
