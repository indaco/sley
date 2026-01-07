package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/semver"
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
	pre := cmd.String("pre")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	if !execCtx.IsSingleModule() {
		return runMultiModuleBump(ctx, cmd, execCtx, operations.BumpMinor, pre, meta, isPreserveMeta)
	}

	return runSingleModuleMinorBump(ctx, cmd, cfg, registry, execCtx, pre, meta, isPreserveMeta, isSkipHooks)
}

// runSingleModuleMinorBump handles minor bump for single-module mode.
func runSingleModuleMinorBump(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry, execCtx *clix.ExecutionContext, pre, meta string, isPreserveMeta, isSkipHooks bool) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	previousVersion, err := semver.ReadVersion(execCtx.Path)
	if err != nil {
		return err
	}

	// Calculate new version
	newVersion := previousVersion
	newVersion.Minor++
	newVersion.Patch = 0
	newVersion.PreRelease = pre
	newVersion.Build = calculateNewBuild(meta, isPreserveMeta, previousVersion.Build)

	// Validate release gates before bumping
	if err := validateReleaseGate(registry, newVersion, previousVersion, "minor"); err != nil {
		return err
	}

	// Validate version policy before bumping
	if err := validateVersionPolicy(registry, newVersion, previousVersion, "minor"); err != nil {
		return err
	}

	// Validate dependency consistency before bumping
	if err := validateDependencyConsistency(registry, newVersion); err != nil {
		return err
	}

	// Validate tag availability before bumping
	if err := validateTagAvailable(registry, newVersion); err != nil {
		return err
	}

	if err := runPreBumpExtensionHooks(ctx, cfg, execCtx.Path, newVersion.String(), previousVersion.String(), "minor", isSkipHooks); err != nil {
		return err
	}

	if err := semver.UpdateVersion(execCtx.Path, "minor", pre, meta, isPreserveMeta); err != nil {
		return err
	}

	// Sync dependency files after updating .version
	if err := syncDependencies(registry, newVersion); err != nil {
		return err
	}

	// Generate changelog entry
	if err := generateChangelogAfterBump(registry, newVersion, previousVersion, "minor"); err != nil {
		return err
	}

	// Record audit log entry
	if err := recordAuditLogEntry(registry, newVersion, previousVersion, "minor"); err != nil {
		return err
	}

	if err := runPostBumpExtensionHooks(ctx, cfg, execCtx.Path, previousVersion.String(), "minor", isSkipHooks); err != nil {
		return err
	}

	// Create tag after successful bump
	return createTagAfterBump(registry, newVersion, "minor")
}
