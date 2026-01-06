package main

import (
	"context"
	"fmt"

	"github.com/indaco/sley/cmd/sley/bumpcmd"
	"github.com/indaco/sley/cmd/sley/changelogcmd"
	"github.com/indaco/sley/cmd/sley/doctorcmd"
	"github.com/indaco/sley/cmd/sley/extensioncmd"
	"github.com/indaco/sley/cmd/sley/initcmd"
	"github.com/indaco/sley/cmd/sley/modulescmd"
	"github.com/indaco/sley/cmd/sley/precmd"
	"github.com/indaco/sley/cmd/sley/setcmd"
	"github.com/indaco/sley/cmd/sley/showcmd"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/console"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/version"
	"github.com/urfave/cli/v3"
)

var noColorFlag bool

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the sley cli.
func newCLI(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:                  "sley",
		Version:               fmt.Sprintf("v%s", version.GetVersion()),
		Usage:                 "Version orchestrator for semantic versioning",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   cfg.Path,
			},
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"no-auto-init"},
				Usage:   "Fail if .version file is missing (disable auto-initialization)",
			},
			&cli.BoolFlag{
				Name:        "no-color",
				Usage:       "Disable colored output",
				Destination: &noColorFlag,
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			console.SetNoColor(noColorFlag)
			printer.SetNoColor(noColorFlag)
			return ctx, nil
		},
		Commands: []*cli.Command{
			showcmd.Run(cfg),
			setcmd.Run(cfg),
			bumpcmd.Run(cfg),
			changelogcmd.Run(cfg),
			precmd.Run(),
			doctorcmd.Run(cfg),
			initcmd.Run(),
			extensioncmd.Run(),
			modulescmd.Run(),
		},
	}
}
