package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// majorCmd returns the "major" subcommand.
func majorCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name:      "major",
		Usage:     "Increment major version and reset minor and patch",
		UsageText: "sley bump major [--pre label] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpMajor(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpMajor increments the major version and resets minor and patch.
func runBumpMajor(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	if err := hooks.RunPreReleaseHooksFn(ctx, cmd.Bool("skip-hooks")); err != nil {
		return err
	}

	params := extractBumpParams(cmd, "major")
	params.versionCalc = makeMajorCalculator(params.pre, params.meta, params.preserveMeta)

	return executeStandardBump(ctx, cmd, cfg, registry, params, operations.BumpMajor)
}
