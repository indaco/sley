package bump

import (
	"github.com/indaco/sley/internal/cli/flags"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// Run returns the "bump" parent command.
func Run(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	cmdFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  "pre",
			Usage: "Optional pre-release label",
		},
		&cli.StringFlag{
			Name:  "meta",
			Usage: "Optional build metadata",
		},
		&cli.BoolFlag{
			Name:  "preserve-meta",
			Usage: "Preserve existing build metadata when bumping",
		},
	}
	cmdFlags = append(cmdFlags, flags.MultiModuleFlags()...)

	return &cli.Command{
		Name:      "bump",
		Usage:     "Bump semantic version (patch, minor, major)",
		UsageText: "sley bump <subcommand> [--flags]",
		Flags:     cmdFlags,
		Commands: []*cli.Command{
			patchCmd(cfg, registry),
			minorCmd(cfg, registry),
			majorCmd(cfg, registry),
			preCmd(cfg, registry),
			releaseCmd(cfg, registry),
			autoCmd(cfg, registry),
		},
	}
}
