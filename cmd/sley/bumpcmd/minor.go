package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// minorCmd returns the "minor" subcommand.
func minorCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name:      "minor",
		Usage:     "Increment minor version and reset patch",
		UsageText: "sley bump minor [--pre label] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpMinor(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpMinor increments the minor version and resets patch.
func runBumpMinor(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	if err := hooks.RunPreReleaseHooksFn(cmd.Bool("skip-hooks")); err != nil {
		return err
	}

	params := extractBumpParams(cmd, "minor")
	params.versionCalc = makeMinorCalculator(params.pre, params.meta, params.preserveMeta)

	return executeStandardBump(ctx, cmd, cfg, registry, params, operations.BumpMinor)
}
