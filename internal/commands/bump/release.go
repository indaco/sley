package bump

import (
	"context"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// releaseCmd returns the "release" subcommand.
func releaseCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "Promote pre-release to final version (e.g. 1.2.3-alpha -> 1.2.3)",
		UsageText: "sley bump release [--preserve-meta] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpRelease(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpRelease promotes a pre-release version to a final release.
func runBumpRelease(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	// Run pre-release hooks first (before any version operations)
	if err := hooks.RunPreReleaseHooks(ctx, isSkipHooks); err != nil {
		return err
	}

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	if !execCtx.IsSingleModule() {
		return runMultiModuleBump(ctx, cmd, execCtx, registry, operations.BumpRelease, "", "", isPreserveMeta)
	}

	// Single-module: use the unified path
	params := bumpParams{
		preserveMeta: isPreserveMeta,
		skipHooks:    isSkipHooks,
		bumpType:     "release",
		opBumpType:   operations.BumpRelease,
	}
	return executeSingleModuleBump(ctx, cmd, cfg, registry, execCtx, params)
}
