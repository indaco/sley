package bumpcmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// releaseCmd returns the "release" subcommand.
func releaseCmd(cfg *config.Config) *cli.Command {
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
			return runBumpRelease(ctx, cmd, cfg)
		},
	}
}

// runBumpRelease promotes a pre-release version to a final release.
func runBumpRelease(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	// Run pre-release hooks first (before any version operations)
	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModuleRelease(cmd, execCtx.Path, isPreserveMeta)
	}

	// Handle multi-module mode
	// For release, we pass empty pre-release and metadata, preserveMetadata flag controls the behavior
	meta := ""
	if isPreserveMeta {
		// The BumpOperation will handle preserve-meta correctly
		meta = ""
	}
	return runMultiModuleBump(ctx, cmd, execCtx, operations.BumpRelease, "", meta, isPreserveMeta)
}

// runSingleModuleRelease handles the single-module release operation.
func runSingleModuleRelease(cmd *cli.Command, path string, isPreserveMeta bool) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	previousVersion, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	newVersion := previousVersion
	newVersion.PreRelease = ""
	if !isPreserveMeta {
		newVersion.Build = ""
	}

	// Validate release gates before bumping
	if err := validateReleaseGate(newVersion, previousVersion, "release"); err != nil {
		return err
	}

	// Validate version policy before bumping
	if err := validateVersionPolicy(newVersion, previousVersion, "release"); err != nil {
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

	if err := semver.SaveVersion(path, newVersion); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	// Generate changelog entry
	if err := generateChangelogAfterBump(newVersion, previousVersion, "release"); err != nil {
		return err
	}

	// Record audit log entry
	if err := recordAuditLogEntry(newVersion, previousVersion, "release"); err != nil {
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Promoted to release version: %s", newVersion.String()))
	return nil
}
