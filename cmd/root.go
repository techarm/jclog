package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/techarm/jclog/internal/config"
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
				Name:  "config",
				Usage: "Path to config file (default: ~/.jclog.json)",
			},
			&cli.StringFlag{
				Name:  "profile",
				Usage: "Configuration profile to use (default: default)",
			},
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
			NewConfigCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Load configuration
			configPath := cmd.String("config")
			if configPath == "" {
				configPath = config.GetDefaultConfigPath()
			}

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %v", err)
			}

			// Get active profile
			if profile := cmd.String("profile"); profile != "" {
				cfg.ActiveProfile = profile
			}
			activeProfile := cfg.GetActiveProfile()

			// Command line arguments take precedence over config file
			format := cmd.String("format")
			if format == "" {
				format = activeProfile.Format
			}

			fields := cmd.StringSlice("fields")
			if len(fields) == 0 {
				fields = activeProfile.Fields
			}

			maxDepth := int(cmd.Int("max-depth"))
			if !cmd.IsSet("max-depth") {
				maxDepth = activeProfile.MaxDepth
			}

			hideMissing := cmd.Bool("hide-missing")
			if !cmd.IsSet("hide-missing") {
				hideMissing = activeProfile.HideMissing
			}

			filters := parseFilterArgs(cmd.StringSlice("filter"))
			if len(filters) == 0 {
				filters = parseFilterArgs(activeProfile.Filters)
			}

			excludes := parseFilterArgs(cmd.StringSlice("exclude"))
			if len(excludes) == 0 {
				excludes = parseFilterArgs(activeProfile.Excludes)
			}

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

			// Process logs
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
