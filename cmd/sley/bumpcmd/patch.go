package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// patchCmd returns the "patch" subcommand.
func patchCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Increment patch version",
		UsageText: "sley bump patch [--pre label] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpPatch(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpPatch executes the patch bump logic.
func runBumpPatch(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	if err := hooks.RunPreReleaseHooksFn(ctx, cmd.Bool("skip-hooks")); err != nil {
		return err
	}

	params := extractBumpParams(cmd, "patch")
	params.versionCalc = makePatchCalculator(params.pre, params.meta, params.preserveMeta)

	return executeStandardBump(ctx, cmd, cfg, registry, params, operations.BumpPatch)
}
