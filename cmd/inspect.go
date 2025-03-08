package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/techarm/jclog/internal/logparser"
	"github.com/urfave/cli/v3"
)

// NewInspectCommand creates a new inspect command
func NewInspectCommand() *cli.Command {
	return &cli.Command{
		Name:  "inspect",
		Usage: "Analyze log file and show available fields",
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
				fields[key] = fieldInfo{
					Type:    fmt.Sprintf("%T", value),
					Example: fmt.Sprintf("%v", value),
				}
			}

			// Print field information
			fmt.Println("Available Fields:")
			printFields(fields)

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
