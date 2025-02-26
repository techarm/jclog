package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// Version information
const VERSION = "0.1.1"

// NewVersionCommand defines the `--version` flag behavior
func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("jclog version", VERSION)
			return nil
		},
	}
}
