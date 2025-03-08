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

// Predefined format templates
var builtinTemplates = map[string]string{
	"basic":    "{timestamp} [{level}] {message}",
	"detailed": "{timestamp} [{level}] {message} (service={service})",
	"compact":  "[{level}] {message}",
	"debug":    "{timestamp} [{level}] {message} (file={caller}:{line})",
	"json":     "{timestamp} [{level}] {message} {data}",
	"metrics":  "{timestamp} {service} CPU:{cpu_usage}% MEM:{memory_usage}%",
}

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
			&cli.StringFlag{
				Name:  "template",
				Usage: "Use predefined format template",
			},
			&cli.IntFlag{
				Name:  "max-depth",
				Usage: "Maximum depth for JSON parsing inside message field",
				Value: 2,
			},
			&cli.BoolFlag{
				Name:  "hide-missing",
				Usage: "Hide missing fields when --format is specified",
				Value: false,
			},
			&cli.BoolFlag{
				Name:    "auto-convert-level",
				Aliases: []string{"c"},
				Usage:   "Automatically convert numeric log levels to text format",
				Value:   false,
			},
			&cli.StringSliceFlag{
				Name:  "filter",
				Usage: "Only show logs that match the specified field=value conditions",
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "Hide logs that match the specified field=value conditions",
			},
		},
		Commands: []*cli.Command{
			NewVersionCommand(),
			NewConfigCommand(),
			NewInspectCommand(),
			NewTemplateCommand(),
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

			// Get format from template or format flag
			format := cmd.String("format")
			if template := cmd.String("template"); template != "" {
				if tmpl, ok := builtinTemplates[template]; ok {
					format = tmpl
				} else {
					return fmt.Errorf("unknown template: %s", template)
				}
			}
			if format == "" {
				format = activeProfile.Format
			}
			if format == "" {
				format = builtinTemplates["basic"] // Use default template
			}

			maxDepth := int(cmd.Int("max-depth"))
			if !cmd.IsSet("max-depth") {
				maxDepth = activeProfile.MaxDepth
			}

			hideMissing := cmd.Bool("hide-missing")
			if !cmd.IsSet("hide-missing") {
				hideMissing = activeProfile.HideMissing
			}

			autoConvertLevel := cmd.Bool("auto-convert-level")
			if !cmd.IsSet("auto-convert-level") {
				autoConvertLevel = activeProfile.AutoConvertLevel
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
			logparser.ProcessLog(scanner, format, maxDepth, hideMissing, filters, excludes, activeProfile.LevelMappings, autoConvertLevel)
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
