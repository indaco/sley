package cli

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/commands/bump"
	"github.com/indaco/sley/internal/commands/changelog"
	"github.com/indaco/sley/internal/commands/doctor"
	"github.com/indaco/sley/internal/commands/extension"
	"github.com/indaco/sley/internal/commands/initialize"
	"github.com/indaco/sley/internal/commands/modules"
	"github.com/indaco/sley/internal/commands/pre"
	"github.com/indaco/sley/internal/commands/set"
	"github.com/indaco/sley/internal/commands/show"
	"github.com/indaco/sley/internal/commands/tag"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/console"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/version"
	urfavecli "github.com/urfave/cli/v3"
)

var noColorFlag bool

// New builds and returns the root CLI command,
// configuring all subcommands and flags for the sley cli.
func New(cfg *config.Config, registry *plugins.PluginRegistry) *urfavecli.Command {
	return &urfavecli.Command{
		Name:                  "sley",
		Version:               fmt.Sprintf("v%s", version.GetVersion()),
		Usage:                 "Version orchestrator for semantic versioning",
		EnableShellCompletion: true,
		Flags: []urfavecli.Flag{
			&urfavecli.StringFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Usage:       "Path to .version file",
				Value:       cfg.Path,
				DefaultText: ".version",
			},
			&urfavecli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"no-auto-init"},
				Usage:   "Fail if .version file is missing (disable auto-initialization)",
			},
			&urfavecli.BoolFlag{
				Name:        "no-color",
				Usage:       "Disable colored output",
				Destination: &noColorFlag,
			},
		},
		Before: func(ctx context.Context, cmd *urfavecli.Command) (context.Context, error) {
			console.SetNoColor(noColorFlag)
			printer.SetNoColor(noColorFlag)
			return ctx, nil
		},
		Commands: []*urfavecli.Command{
			initialize.Run(),
			show.Run(cfg),
			set.Run(cfg),
			bump.Run(cfg, registry),
			pre.Run(cfg),
			doctor.Run(cfg),
			changelog.Run(cfg),
			tag.Run(cfg),
			extension.Run(),
			modules.Run(),
		},
	}
}
