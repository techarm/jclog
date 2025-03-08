package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/techarm/jclog/internal/formatter"
)

// FieldAliases defines field name aliases for better flexibility
var FieldAliases = map[string][]string{
	"timestamp": {"timestamp", "time", "ts"},
	"level":     {"level", "lvl", "severity"},
	"message":   {"message", "msg", "text"},
}

// Pattern for extracting field names from format string
var fieldPattern = regexp.MustCompile(`{([^}]+)}`)

// ProcessLog parses JSON logs and outputs formatted results
func ProcessLog(scanner *bufio.Scanner, format string, maxDepth int, hideMissing bool, filters map[string]string, excludes map[string]string, levelMappings map[string]string, autoConvertLevel bool) {
	// Extract fields from format string
	fields := extractFields(format)

	for scanner.Scan() {
		// Parse the log line as JSON
		raw := make(map[string]any)
		if err := json.Unmarshal([]byte(scanner.Text()), &raw); err != nil {
			fmt.Println("Invalid JSON:", scanner.Text())
			continue
		}

		// Extract fields
		extractedFields := make(map[string]string)
		for _, field := range fields {
			// Check if field has a modifier
			fieldName := field
			modifier := ""
			if strings.Contains(field, "|") {
				parts := strings.Split(field, "|")
				fieldName = parts[0]
				if len(parts) > 1 {
					modifier = parts[1]
				}
			}

			value := getFieldValue(raw, fieldName)
			// Apply level mappings if available
			if fieldName == "level" && autoConvertLevel && levelMappings != nil {
				if mapped, ok := levelMappings[value]; ok {
					value = mapped
				}
			}
			// Apply modifiers
			if modifier == "basename" && fieldName == "file" {
				value = filepath.Base(value)
			}
			extractedFields[field] = value
		}

		// Handle nested message fields dynamically
		if msg, exists := extractedFields["message"]; exists && msg != "" {
			messageFields := make(map[string]string)
			flattenJSONString(msg, "message", messageFields, maxDepth, 1)
			for k, v := range messageFields {
				if slices.Contains(fields, k) {
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

		// Format output with unknown field handling
		output := format
		for _, field := range fields {
			value := extractedFields[field]
			placeholder := "{" + field + "}"

			if value == "" {
				if hideMissing {
					// Remove the placeholder and any surrounding brackets
					output = removeFieldAndBrackets(output, field)
				} else {
					// Mark unknown field in gray with warning symbol
					unknownValue := color.New(color.FgHiBlack).Sprintf("❓%s", field)
					output = strings.Replace(output, placeholder, unknownValue, -1)
				}
			} else {
				output = strings.Replace(output, placeholder, value, -1)
			}
		}

		// Apply color based on log level
		if level, exists := extractedFields["level"]; exists {
			output = formatter.ColorizeByLevel(output, level)
		}

		fmt.Println(output)
	}
}

// extractFields extracts field names from format string
func extractFields(format string) []string {
	matches := fieldPattern.FindAllStringSubmatch(format, -1)
	fields := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			fields = append(fields, match[1])
		}
	}
	return fields
}

// removeFieldAndBrackets removes a field placeholder and its surrounding brackets
func removeFieldAndBrackets(format, field string) string {
	// Remove [field] pattern
	bracketPattern := regexp.MustCompile(`\[[^]]*{` + field + `}[^]]*\]`)
	format = bracketPattern.ReplaceAllString(format, "")

	// Remove field pattern
	fieldPattern := regexp.MustCompile(`{` + field + `}`)
	format = fieldPattern.ReplaceAllString(format, "")

	// Clean up extra spaces
	format = strings.ReplaceAll(format, "  ", " ")
	return strings.TrimSpace(format)
}

// flattenJSONString tries to decode nested JSON strings recursively
func flattenJSONString(jsonStr string, prefix string, result map[string]string, maxDepth int, currentDepth int) {
	if currentDepth > maxDepth {
		result[prefix] = jsonStr // Store as a string if max depth is reached
		return
	}

	// Try to parse the string as JSON
	parsedJSON, err := tryParseJSON(jsonStr)
	if err != nil {
		result[prefix] = jsonStr // Store as a normal string if it's not JSON
		return
	}

	// Recursively flatten the JSON object
	for key, val := range parsedJSON {
		newKey := prefix
		if prefix != "" {
			newKey = prefix + "." + key
		} else {
			newKey = key
		}

		switch v := val.(type) {
		case string:
			// Try to parse the string value as JSON
			if _, err := tryParseJSON(v); err == nil {
				flattenJSONString(v, newKey, result, maxDepth, currentDepth+1)
			} else {
				result[newKey] = v
			}
		case float64:
			result[newKey] = fmt.Sprintf("%.0f", v)
		case bool:
			result[newKey] = fmt.Sprintf("%v", v)
		case map[string]any:
			if currentDepth < maxDepth {
				for k, v := range v {
					flattenJSONString(fmt.Sprintf("%v", v), newKey+"."+k, result, maxDepth, currentDepth+1)
				}
			} else {
				jsonBytes, _ := json.Marshal(v)
				result[newKey] = string(jsonBytes)
			}
		case nil:
			result[newKey] = ""
		default:
			result[newKey] = fmt.Sprintf("%v", v)
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
			if field == "level" {
				return fmt.Sprintf("%d", int(v))
			}
			return fmt.Sprintf("%.0f", v)
		}
		if v, ok := data[alias].(int); ok {
			if field == "level" {
				return fmt.Sprintf("%d", v)
			}
			return fmt.Sprintf("%d", v)
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
