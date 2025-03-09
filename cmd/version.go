package cmd

import (
	"context"
	"fmt"

	"github.com/techarm/jclog/internal/meta"
	"github.com/urfave/cli/v3"
)

// NewVersionCommand defines the `--version` flag behavior
func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(context.Context, *cli.Command) error {
			fmt.Println("jclog version", meta.GetMetadata().Version)
			return nil
		},
	}
}
