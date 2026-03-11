package bump

import (
	"context"

	"github.com/indaco/sley/internal/cliflags"
	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/urfave/cli/v3"
)

// preCmd returns the "pre" subcommand for incrementing pre-release versions.
func preCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "label",
			Aliases: []string{"l"},
			Usage:   "Pre-release label (e.g., alpha, beta, rc). If omitted, increments existing pre-release",
		},
		&cli.StringFlag{
			Name:  "meta",
			Usage: "Optional build metadata",
		},
		&cli.BoolFlag{
			Name:  "preserve-meta",
			Usage: "Preserve existing build metadata when bumping",
		},
		&cli.BoolFlag{
			Name:  "skip-hooks",
			Usage: "Skip pre-release hooks",
		},
	}
	flags = append(flags, cliflags.MultiModuleFlags()...)

	return &cli.Command{
		Name:      "pre",
		Usage:     "Increment pre-release version (e.g., rc.1 -> rc.2)",
		UsageText: "sley bump pre [--label name] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags:     flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpPre(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpPre executes the pre-release bump logic.
func runBumpPre(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	label := cmd.String("label")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if err := hooks.RunPreReleaseHooks(ctx, isSkipHooks); err != nil {
		return err
	}

	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	if !execCtx.IsSingleModule() {
		return runMultiModuleBump(ctx, cmd, execCtx, registry, operations.BumpPre, label, meta, isPreserveMeta)
	}

	// Single-module: use the unified path via BumpOperation
	params := bumpParams{
		pre:          label,
		meta:         meta,
		preserveMeta: isPreserveMeta,
		skipHooks:    isSkipHooks,
		bumpType:     "pre",
		opBumpType:   operations.BumpPre,
	}
	return executeSingleModuleBump(ctx, cmd, cfg, registry, execCtx, params)
}
