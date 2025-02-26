package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/techarm/jclog/internal/logparser"
	"github.com/urfave/cli/v3"
)

// NewRootCommand defines the CLI root command
func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "jclog",
		Usage: "Parse JSON log files and display them with colors",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Specify output format (e.g., \"{timestamp} [{level}] {message}\")",
			},
			&cli.StringSliceFlag{
				Name:  "fields",
				Usage: "Additional JSON fields to display (e.g., --fields=service,user)",
			},
			&cli.IntFlag{
				Name:  "max-depth",
				Usage: "Maximum depth for JSON parsing inside message field",
				Value: 2, // Default depth is 2
			},
			&cli.BoolFlag{
				Name:  "hide-missing",
				Usage: "Hide missing fields when --format is specified",
				Value: false,
			},
			&cli.StringSliceFlag{
				Name:  "filter",
				Usage: "Only show logs that match the specified field=value conditions (e.g., --filter=level=INFO)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "Hide logs that match the specified field=value conditions (e.g., --exclude=level=DEBUG)",
			},
		},
		Commands: []*cli.Command{
			NewVersionCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var scanner *bufio.Scanner

			// Read from file if provided
			if cmd.Args().Len() > 0 {
				filePath := cmd.Args().Get(0)
				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("failed to open file: %v", err)
				}
				defer file.Close()
				scanner = bufio.NewScanner(file)
			} else {
				// Read from standard input (pipe)
				scanner = bufio.NewScanner(os.Stdin)
			}

			// Get parameters
			format := cmd.String("format")
			fields := cmd.StringSlice("fields")
			maxDepth := cmd.Int("max-depth")
			hideMissing := cmd.Bool("hide-missing")
			filters := parseFilterArgs(cmd.StringSlice("filter"))
			excludes := parseFilterArgs(cmd.StringSlice("exclude"))

			// Process logs
			fmt.Println(filters)
			fmt.Println(excludes)
			logparser.ProcessLog(scanner, format, fields, maxDepth, hideMissing, filters, excludes)
			return nil
		},
	}
}

// parseFilterArgs converts "key=value" strings into a map
func parseFilterArgs(args []string) map[string]string {
	filters := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}
	return filters
}
