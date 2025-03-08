package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
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

			// Parse JSON
			var data map[string]any
			if err := json.Unmarshal([]byte(scanner.Text()), &data); err != nil {
				return fmt.Errorf("invalid JSON: %v", err)
			}

			// Get all fields
			fields := make(map[string]fieldInfo)
			for key, value := range data {
				example := fmt.Sprintf("%v", value)
				if cmd.Bool("basename") && key == "file" {
					example = filepath.Base(example)
				}
				fields[key] = fieldInfo{
					Type:    fmt.Sprintf("%T", value),
					Example: example,
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

type fieldInfo struct {
	Type    string
	Example string
}

func printFields(fields map[string]fieldInfo) {
	// Get sorted field names
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print field information
	for i, key := range keys {
		info := fields[key]
		prefix := "├──"
		if i == len(keys)-1 {
			prefix = "└──"
		}

		// Check for aliases
		aliases := getAliases(key)
		aliasStr := ""
		if len(aliases) > 0 {
			aliasStr = fmt.Sprintf(" (alias: %s)", strings.Join(aliases, ", "))
		}

		// Print field name and aliases
		fmt.Printf("%s %s%s\n", prefix, color.BlueString(key), aliasStr)

		// Print type and example value
		valuePrefix := "│   └──"
		if i == len(keys)-1 {
			valuePrefix = "    └──"
		}
		fmt.Printf("%s Type: %s\n", valuePrefix, info.Type)
		fmt.Printf("%s Example: %s\n", valuePrefix, color.YellowString("%q", info.Example))
	}
}

func printSuggestedFormats(fields map[string]fieldInfo, useBasename bool) {
	// 1. Show all fields in alphabetical order
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	allFieldsFormat := buildFormat(fields, keys, false, useBasename)
	fmt.Println("\n1. 📝 All Fields Format (Alphabetically Ordered):")
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
