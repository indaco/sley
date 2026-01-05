package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// majorCmd returns the "major" subcommand.
func majorCmd(cfg *config.Config) *cli.Command {
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
			return runBumpMajor(ctx, cmd, cfg)
		},
	}
}

// runBumpMajor increments the major version and resets minor and patch.
func runBumpMajor(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
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
		return runMultiModuleBump(ctx, cmd, execCtx, operations.BumpMajor, pre, meta, isPreserveMeta)
	}

	return runSingleModuleMajorBump(ctx, cmd, cfg, execCtx, pre, meta, isPreserveMeta, isSkipHooks)
}

// runSingleModuleMajorBump handles major bump for single-module mode.
func runSingleModuleMajorBump(ctx context.Context, cmd *cli.Command, cfg *config.Config, execCtx *clix.ExecutionContext, pre, meta string, isPreserveMeta, isSkipHooks bool) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	previousVersion, err := semver.ReadVersion(execCtx.Path)
	if err != nil {
		return err
	}

	// Calculate new version
	newVersion := previousVersion
	newVersion.Major++
	newVersion.Minor = 0
	newVersion.Patch = 0
	newVersion.PreRelease = pre
	newVersion.Build = calculateNewBuild(meta, isPreserveMeta, previousVersion.Build)

	// Validate release gates before bumping
	if err := validateReleaseGate(newVersion, previousVersion, "major"); err != nil {
		return err
	}

	// Validate version policy before bumping
	if err := validateVersionPolicy(newVersion, previousVersion, "major"); err != nil {
		return err
	}

	// Validate dependency consistency before bumping
	if err := validateDependencyConsistency(newVersion); err != nil {
		return err
	}

	// Validate tag availability before bumping
	if err := validateTagAvailable(newVersion); err != nil {
		return err
	}

	if err := runPreBumpExtensionHooks(ctx, cfg, execCtx.Path, newVersion.String(), previousVersion.String(), "major", isSkipHooks); err != nil {
		return err
	}

	if err := semver.UpdateVersion(execCtx.Path, "major", pre, meta, isPreserveMeta); err != nil {
		return err
	}

	// Sync dependency files after updating .version
	if err := syncDependencies(newVersion); err != nil {
		return err
	}

	// Generate changelog entry
	if err := generateChangelogAfterBump(newVersion, previousVersion, "major"); err != nil {
		return err
	}

	// Record audit log entry
	if err := recordAuditLogEntry(newVersion, previousVersion, "major"); err != nil {
		return err
	}

	if err := runPostBumpExtensionHooks(ctx, cfg, execCtx.Path, previousVersion.String(), "major", isSkipHooks); err != nil {
		return err
	}

	// Create tag after successful bump
	return createTagAfterBump(newVersion, "major")
}
