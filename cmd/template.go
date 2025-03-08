package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

// NewTemplateCommand creates a new template command
func NewTemplateCommand() *cli.Command {
	return &cli.Command{
		Name:  "template",
		Usage: "Manage format templates",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List available templates",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Available Templates:")
					for name, format := range builtinTemplates {
						fmt.Printf("- %s:\n  %s\n", color.BlueString(name), format)
					}
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Show template details",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if cmd.Args().Len() == 0 {
						return fmt.Errorf("template name is required")
					}

					name := cmd.Args().Get(0)
					if format, ok := builtinTemplates[name]; ok {
						fmt.Printf("Template: %s\n", color.BlueString(name))
						fmt.Printf("Format: %s\n", format)
						return nil
					}

					return fmt.Errorf("unknown template: %s", name)
				},
			},
		},
	}
}
