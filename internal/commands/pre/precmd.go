package pre

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/indaco/sley/internal/cliflags"
	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/commands/depsync"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// Run returns the "pre" command.
func Run(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	cmdFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     "label",
			Usage:    "Pre-release label to set",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "inc",
			Usage: "Increment numeric suffix if it exists or add '.1'",
		},
	}
	cmdFlags = append(cmdFlags, cliflags.MultiModuleFlags()...)

	return &cli.Command{
		Name:      "pre",
		Usage:     "Set pre-release label (e.g., alpha, beta.1)",
		UsageText: "sley pre --label <label> [--inc] [--all] [--module name]",
		Flags:     cmdFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runPreCmd(ctx, cmd, cfg, registry)
		},
	}
}

// runPreCmd sets or increments the pre-release label.
func runPreCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	label := cmd.String("label")
	isInc := cmd.Bool("inc")

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModulePre(execCtx.Path, label, isInc, registry)
	}

	// Handle multi-module mode
	return runMultiModulePre(ctx, cmd, execCtx, label, isInc, registry)
}

// runSingleModulePre handles the single-module pre-release operation.
func runSingleModulePre(path, label string, isInc bool, registry *plugins.PluginRegistry) error {
	// Auto-initialize if file doesn't exist
	var version semver.SemVersion
	version, err := semver.ReadVersion(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Auto-initialize with default version
			version = semver.SemVersion{Major: 0, Minor: 1, Patch: 0}
			if err := semver.SaveVersion(path, version); err != nil {
				return fmt.Errorf("failed to initialize version file: %w", err)
			}
			printer.PrintSuccess(fmt.Sprintf("Auto-initialized %s with default version", path))
		} else {
			return fmt.Errorf("failed to read version: %w", err)
		}
	}

	oldVersion := version.String()

	if isInc {
		version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
	} else {
		if version.PreRelease == "" {
			version.Patch++
		}
		version.PreRelease = label
	}

	if err := semver.SaveVersion(path, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	printer.PrintSuccess(fmt.Sprintf("Updated version from %s to %s", oldVersion, version.String()))

	// Sync dependencies if configured
	if err := depsync.SyncDependencies(registry, version, path); err != nil {
		return err
	}

	return nil
}

// runMultiModulePre handles the multi-module pre-release operation.
func runMultiModulePre(ctx context.Context, cmd *cli.Command, execCtx *clix.ExecutionContext, label string, isInc bool, registry *plugins.PluginRegistry) error {
	fs := core.NewOSFileSystem()
	operation := operations.NewPreOperation(fs, label, isInc)

	// Create executor with options from flags
	parallel := cmd.Bool("parallel")
	failFast := cmd.Bool("fail-fast") && !cmd.Bool("continue-on-error")

	executor := workspace.NewExecutor(
		workspace.WithParallel(parallel),
		workspace.WithFailFast(failFast),
	)

	// Execute the operation on all modules
	results, err := executor.Run(ctx, execCtx.Modules, operation)
	if err != nil && failFast {
		// In fail-fast mode, we may have partial results
		// Fall through to display what we have
		_ = err
	}

	// Format and display results
	format := cmd.String("format")
	quiet := cmd.Bool("quiet")

	actionVerb := "updated"
	if isInc {
		actionVerb = "incremented"
	}

	operationName := fmt.Sprintf("Set pre-release to %s", label)
	if isInc {
		operationName = fmt.Sprintf("Increment pre-release with %s", label)
	}

	formatter := workspace.GetFormatterWithVerb(format, operationName, actionVerb)

	if quiet {
		// In quiet mode, just show summary
		printQuietSummary(results)
	} else {
		fmt.Println(formatter.FormatResults(results))
	}

	// Sync dependencies if configured (use the first successful result's version)
	if !workspace.HasErrors(results) && len(results) > 0 {
		// Get the new version from the first successful result
		for _, result := range results {
			if result.Error == nil && result.NewVersion != "" {
				parsedVersion, parseErr := semver.ParseVersion(result.NewVersion)
				if parseErr == nil {
					bumpedPaths := getBumpedModulePaths(results)
					if syncErr := depsync.SyncDependencies(registry, parsedVersion, bumpedPaths...); syncErr != nil {
						return syncErr
					}
				}
				break
			}
		}
	}

	// Return error if any failures occurred
	if workspace.HasErrors(results) {
		return fmt.Errorf("%d module(s) failed", workspace.ErrorCount(results))
	}

	return nil
}

// getBumpedModulePaths extracts the paths of successfully bumped modules.
func getBumpedModulePaths(results []workspace.ExecutionResult) []string {
	var paths []string
	for _, r := range results {
		if r.Error == nil {
			paths = append(paths, r.Module.Path)
		}
	}
	return paths
}

// printQuietSummary prints a minimal summary of results.
func printQuietSummary(results []workspace.ExecutionResult) {
	success := workspace.SuccessCount(results)
	errors := workspace.ErrorCount(results)
	if errors > 0 {
		printer.PrintWarning(fmt.Sprintf("Completed: %d succeeded, %d failed", success, errors))
	} else {
		printer.PrintSuccess(fmt.Sprintf("Success: %d module(s) updated", success))
	}
}
