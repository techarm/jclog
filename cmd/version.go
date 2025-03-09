package cmd

import (
	"context"
	"fmt"

	"github.com/techarm/jclog/internal"
	"github.com/urfave/cli/v3"
)

// NewVersionCommand defines the `--version` flag behavior
func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("jclog version", internal.Version)
			return nil
		},
	}
}
