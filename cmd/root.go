package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/techarm/json-log-viewer/internal/logparser"
	"github.com/urfave/cli/v3"
)

// NewRootCommand defines the CLI root command
func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "logview",
		Usage: "Parse JSON log files and display them with colors",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Specify output format (e.g., \"{timestamp} [{level}] {message}\")",
				Value: "{timestamp} [{level}] {message}",
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

			// Process logs
			logparser.ProcessLog(scanner, format, fields, maxDepth)
			return nil
		},
	}
}
