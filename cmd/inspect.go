package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/techarm/jclog/internal/config"
	"github.com/techarm/jclog/internal/logparser"
	"github.com/urfave/cli/v3"
)

// Default field order for log output
var defaultFieldOrder = []string{
	"time", "level", "http_code", "http_method", "uri", "route", "pid",
	"target", "file", "line", "message", "error",
}

// Suggested fields for common use cases
var suggestedFields = []string{
	"time", "level", "http_code", "http_method", "uri", "file", "line", "message", "error",
}

type fieldInfo struct {
	Type    string
	Example string
	Order   int
}

// orderedJSON represents a JSON object with ordered fields
type orderedJSON struct {
	fields     map[string]any
	fieldOrder []string
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (o *orderedJSON) UnmarshalJSON(data []byte) error {
	// First, get the field order by decoding tokens
	dec := json.NewDecoder(bytes.NewReader(data))

	// Read the opening brace
	if _, err := dec.Token(); err != nil {
		return err
	}

	o.fields = make(map[string]any)
	o.fieldOrder = make([]string, 0)

	// Read field names in order
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return err
		}

		key := token.(string)
		o.fieldOrder = append(o.fieldOrder, key)

		// Skip the value for now
		var value any
		if err := dec.Decode(&value); err != nil {
			return err
		}
		o.fields[key] = value
	}

	return nil
}

// NewInspectCommand creates a new inspect command
func NewInspectCommand() *cli.Command {
	return &cli.Command{
		Name:  "inspect",
		Usage: "Analyze log file and show available fields",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "basename",
				Aliases: []string{"b"},
				Usage:   "Show only the base name of file paths",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("log file path is required")
			}

			// Load configuration
			cfg, err := config.LoadConfig(config.GetDefaultConfigPath())
			if err != nil {
				return fmt.Errorf("failed to load config: %v", err)
			}
			activeProfile := cfg.GetActiveProfile()

			filePath := cmd.Args().Get(0)
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer file.Close()

			// Analyze the first log entry
			scanner := bufio.NewScanner(file)
			if !scanner.Scan() {
				return fmt.Errorf("empty log file")
			}

			// Parse JSON with field order preservation
			var orderedData orderedJSON
			if err := json.Unmarshal([]byte(scanner.Text()), &orderedData); err != nil {
				return fmt.Errorf("invalid JSON: %v", err)
			}

			// Get all fields with their original order
			fields := make(map[string]fieldInfo)
			for i, key := range orderedData.fieldOrder {
				value := orderedData.fields[key]
				example := fmt.Sprintf("%v", value)

				// Handle special cases
				if cmd.Bool("basename") && key == "file" {
					example = filepath.Base(example)
				} else if key == "time" || key == "timestamp" {
					if t, err := time.Parse(time.RFC3339Nano, example); err == nil {
						example = t.Format(activeProfile.TimeFormat)
					}
				}

				fields[key] = fieldInfo{
					Type:    fmt.Sprintf("%T", value),
					Example: example,
					Order:   i,
				}
			}

			// Print field information
			fmt.Println("📋 Available Fields:")
			printFields(fields)

			// Print format suggestions
			fmt.Println("\n🎨 Format Suggestions:")
			printSuggestedFormats(fields, cmd.Bool("basename"))

			return nil
		},
	}
}

func printFields(fields map[string]fieldInfo) {
	// Get field names in original order
	type fieldWithName struct {
		name string
		info fieldInfo
	}
	orderedFields := make([]fieldWithName, 0, len(fields))
	for name, info := range fields {
		orderedFields = append(orderedFields, fieldWithName{name, info})
	}
	sort.Slice(orderedFields, func(i, j int) bool {
		return orderedFields[i].info.Order < orderedFields[j].info.Order
	})

	// Print field information
	for i, field := range orderedFields {
		prefix := "├──"
		if i == len(orderedFields)-1 {
			prefix = "└──"
		}

		// Check for aliases
		aliases := getAliases(field.name)
		aliasStr := ""
		if len(aliases) > 0 {
			aliasStr = fmt.Sprintf(" (alias: %s)", strings.Join(aliases, ", "))
		}

		// Print field name and aliases
		fmt.Printf("%s %s%s\n", prefix, color.BlueString(field.name), aliasStr)

		// Print type and example value
		valuePrefix := "│   └──"
		if i == len(orderedFields)-1 {
			valuePrefix = "    └──"
		}
		fmt.Printf("%s Type: %s\n", valuePrefix, field.info.Type)
		fmt.Printf("%s Example: %s\n", valuePrefix, color.YellowString("%q", field.info.Example))
	}
}

func printSuggestedFormats(fields map[string]fieldInfo, useBasename bool) {
	// 1. Show all fields in original order
	type fieldWithName struct {
		name string
		info fieldInfo
	}
	orderedFields := make([]fieldWithName, 0, len(fields))
	for name, info := range fields {
		orderedFields = append(orderedFields, fieldWithName{name, info})
	}
	sort.Slice(orderedFields, func(i, j int) bool {
		return orderedFields[i].info.Order < orderedFields[j].info.Order
	})

	keys := make([]string, len(orderedFields))
	for i, field := range orderedFields {
		keys[i] = field.name
	}

	allFieldsFormat := buildFormat(fields, keys, false, useBasename)
	fmt.Println("\n1. 📝 All Fields Format (Original Order):")
	fmt.Printf("   %s\n", color.GreenString(allFieldsFormat))

	// 2. Show fields in recommended order
	recommendedFormat := buildFormat(fields, defaultFieldOrder, false, useBasename)
	fmt.Println("\n2. ⭐ Recommended Format (Default Order):")
	fmt.Printf("   %s\n", color.GreenString(recommendedFormat))

	// 3. Show suggested fields
	suggestedFormat := buildFormat(fields, suggestedFields, false, useBasename)
	fmt.Println("\n3. 💡 Suggested Format (Common Fields):")
	fmt.Printf("   %s\n", color.GreenString(suggestedFormat))

	// 4. Help tip
	fmt.Println("\n💪 Tip: Feel free to customize your format by removing or reordering fields from Format 1 above.")
}

func buildFormat(fields map[string]fieldInfo, fieldOrder []string, includeExtra bool, useBasename bool) string {
	parts := []string{}
	usedFields := make(map[string]bool)

	// Add ordered fields first
	for _, field := range fieldOrder {
		if _, exists := fields[field]; exists {
			if field == "level" {
				parts = append(parts, "[{level}]")
			} else if field == "http_code" {
				parts = append(parts, "[{http_code}]")
			} else if field == "file" && useBasename {
				parts = append(parts, "{file|basename}")
			} else {
				parts = append(parts, "{"+field+"}")
			}
			usedFields[field] = true
		}
	}

	// Add remaining fields if includeExtra is true
	if includeExtra {
		for field := range fields {
			if !usedFields[field] {
				parts = append(parts, "{"+field+"}")
				usedFields[field] = true
			}
		}
	}

	return strings.Join(parts, " ")
}

// getAliases returns the aliases for a given field
func getAliases(field string) []string {
	for key, aliases := range logparser.FieldAliases {
		if field == key {
			return aliases[1:] // Exclude the first one (primary name)
		}
		for _, alias := range aliases {
			if field == alias {
				return []string{key}
			}
		}
	}
	return nil
}
