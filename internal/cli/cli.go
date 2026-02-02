package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/indaco/sley/internal/commands/bump"
	"github.com/indaco/sley/internal/commands/changelog"
	"github.com/indaco/sley/internal/commands/discover"
	"github.com/indaco/sley/internal/commands/doctor"
	"github.com/indaco/sley/internal/commands/extension"
	"github.com/indaco/sley/internal/commands/initialize"
	"github.com/indaco/sley/internal/commands/pre"
	"github.com/indaco/sley/internal/commands/set"
	"github.com/indaco/sley/internal/commands/show"
	"github.com/indaco/sley/internal/commands/tag"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/console"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/tui"
	"github.com/indaco/sley/internal/version"
	urfavecli "github.com/urfave/cli/v3"
)

var (
	noColorFlag bool
	themeFlag   string
)

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
			&urfavecli.StringFlag{
				Name:        "theme",
				Usage:       "TUI theme (sley, base, base16, catppuccin, charm, dracula)",
				Value:       cfg.GetTheme(),
				Destination: &themeFlag,
			},
		},
		Before: func(ctx context.Context, cmd *urfavecli.Command) (context.Context, error) {
			console.SetNoColor(noColorFlag)
			printer.SetNoColor(noColorFlag)

			// Theme priority: CLI flag > env var > config file > default
			theme := themeFlag
			if envTheme := os.Getenv("SLEY_THEME"); envTheme != "" && themeFlag == cfg.GetTheme() {
				// Only use env var if CLI flag wasn't explicitly set
				theme = envTheme
			}

			if theme != "" && !tui.IsValidTheme(theme) {
				return ctx, fmt.Errorf("invalid theme %q: valid themes are %s",
					theme, strings.Join(tui.ValidThemes, ", "))
			}

			tui.SetTheme(theme)
			return ctx, nil
		},
		Commands: []*urfavecli.Command{
			initialize.Run(),
			discover.Run(cfg),
			show.Run(cfg),
			set.Run(cfg),
			bump.Run(cfg, registry),
			pre.Run(cfg, registry),
			doctor.Run(cfg),
			tag.Run(cfg),
			changelog.Run(cfg),
			extension.Run(),
		},
	}
}
