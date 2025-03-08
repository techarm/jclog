package cmd

import (
	"context"
	"fmt"

	"github.com/techarm/jclog/internal/config"
	"github.com/urfave/cli/v3"
)

// NewConfigCommand creates a new command for managing configuration
func NewConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration profiles",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize default configuration file",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath := config.GetDefaultConfigPath()
					cfg := config.DefaultConfig()
					if err := config.SaveConfig(cfg, configPath); err != nil {
						return fmt.Errorf("failed to save config: %v", err)
					}
					fmt.Printf("Created default configuration at %s\n", configPath)
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Show current configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath := config.GetDefaultConfigPath()
					cfg, err := config.LoadConfig(configPath)
					if err != nil {
						return fmt.Errorf("failed to load config: %v", err)
					}
					fmt.Printf("Active profile: %s\n", cfg.ActiveProfile)
					fmt.Println("\nAvailable profiles:")
					for name, profile := range cfg.Profiles {
						fmt.Printf("\n[%s]\n", name)
						fmt.Printf("  Format: %s\n", profile.Format)
						fmt.Printf("  Fields: %v\n", profile.Fields)
						fmt.Printf("  MaxDepth: %d\n", profile.MaxDepth)
						fmt.Printf("  HideMissing: %v\n", profile.HideMissing)
						fmt.Printf("  Filters: %v\n", profile.Filters)
						fmt.Printf("  Excludes: %v\n", profile.Excludes)
					}
					return nil
				},
			},
			{
				Name:  "add-profile",
				Usage: "Add a new profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Profile name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "format",
						Usage: "Output format",
					},
					&cli.StringSliceFlag{
						Name:  "fields",
						Usage: "Fields to display",
					},
					&cli.IntFlag{
						Name:  "max-depth",
						Usage: "Maximum depth for JSON parsing",
						Value: 2,
					},
					&cli.BoolFlag{
						Name:  "hide-missing",
						Usage: "Hide missing fields",
					},
					&cli.StringSliceFlag{
						Name:  "filter",
						Usage: "Filter conditions",
					},
					&cli.StringSliceFlag{
						Name:  "exclude",
						Usage: "Exclude conditions",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath := config.GetDefaultConfigPath()
					cfg, err := config.LoadConfig(configPath)
					if err != nil {
						return fmt.Errorf("failed to load config: %v", err)
					}

					name := cmd.String("name")
					maxDepth := int(cmd.Int("max-depth"))
					profile := config.Profile{
						Format:      cmd.String("format"),
						Fields:      cmd.StringSlice("fields"),
						MaxDepth:    maxDepth,
						HideMissing: cmd.Bool("hide-missing"),
						Filters:     cmd.StringSlice("filter"),
						Excludes:    cmd.StringSlice("exclude"),
					}

					cfg.Profiles[name] = profile
					if err := config.SaveConfig(cfg, configPath); err != nil {
						return fmt.Errorf("failed to save config: %v", err)
					}

					fmt.Printf("Added profile '%s'\n", name)
					return nil
				},
			},
			{
				Name:  "remove-profile",
				Usage: "Remove a profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Profile name",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath := config.GetDefaultConfigPath()
					cfg, err := config.LoadConfig(configPath)
					if err != nil {
						return fmt.Errorf("failed to load config: %v", err)
					}

					name := cmd.String("name")
					if name == "default" {
						return fmt.Errorf("cannot remove default profile")
					}

					if _, exists := cfg.Profiles[name]; !exists {
						return fmt.Errorf("profile '%s' does not exist", name)
					}

					delete(cfg.Profiles, name)
					if err := config.SaveConfig(cfg, configPath); err != nil {
						return fmt.Errorf("failed to save config: %v", err)
					}

					fmt.Printf("Removed profile '%s'\n", name)
					return nil
				},
			},
			{
				Name:  "set-active",
				Usage: "Set the active profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Profile name",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath := config.GetDefaultConfigPath()
					cfg, err := config.LoadConfig(configPath)
					if err != nil {
						return fmt.Errorf("failed to load config: %v", err)
					}

					name := cmd.String("name")
					if _, exists := cfg.Profiles[name]; !exists {
						return fmt.Errorf("profile '%s' does not exist", name)
					}

					cfg.ActiveProfile = name
					if err := config.SaveConfig(cfg, configPath); err != nil {
						return fmt.Errorf("failed to save config: %v", err)
					}

					fmt.Printf("Set active profile to '%s'\n", name)
					return nil
				},
			},
		},
	}
}
