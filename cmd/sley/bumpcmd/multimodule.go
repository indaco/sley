package bumpcmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// runMultiModuleBump executes a bump operation on multiple modules.
func runMultiModuleBump(
	ctx context.Context,
	cmd *cli.Command,
	execCtx *clix.ExecutionContext,
	registry *plugins.PluginRegistry,
	bumpType operations.BumpType,
	preRelease, metadata string,
	preserveMetadata bool,
) error {
	fs := core.NewOSFileSystem()
	operation := operations.NewBumpOperation(fs, bumpType, preRelease, metadata, preserveMetadata)

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

	formatter := workspace.GetFormatter(format, fmt.Sprintf("Bump %s", bumpType))

	if quiet {
		// In quiet mode, just show summary
		printQuietSummary(results)
	} else {
		fmt.Println(formatter.FormatResults(results))
	}

	// Return error if any failures occurred
	if workspace.HasErrors(results) {
		return fmt.Errorf("%d module(s) failed", workspace.ErrorCount(results))
	}

	// Sync dependencies after all modules are bumped successfully.
	// The dependency-check plugin syncs files globally, so we call it once
	// using the new version from the first successful result.
	// We pass the paths of all bumped modules to avoid showing them twice.
	if newVersion := getFirstSuccessfulVersion(results); newVersion != "" {
		parsedVersion, err := semver.ParseVersion(newVersion)
		if err == nil {
			bumpedPaths := getBumpedModulePaths(results)
			if err := syncDependencies(registry, parsedVersion, bumpedPaths...); err != nil {
				return err
			}
		}
	}

	return nil
}

// getBumpedModulePaths extracts the paths of all successfully bumped modules.
func getBumpedModulePaths(results []workspace.ExecutionResult) []string {
	var paths []string
	for _, r := range results {
		if r.Success && r.Module != nil {
			paths = append(paths, r.Module.Path)
		}
	}
	return paths
}

// getFirstSuccessfulVersion returns the new version from the first successful result.
// Returns empty string if no successful results exist.
func getFirstSuccessfulVersion(results []workspace.ExecutionResult) string {
	for _, r := range results {
		if r.Success && r.NewVersion != "" {
			return r.NewVersion
		}
	}
	return ""
}

// printQuietSummary prints a minimal summary of results.
func printQuietSummary(results []workspace.ExecutionResult) {
	success := workspace.SuccessCount(results)
	errors := workspace.ErrorCount(results)
	if errors > 0 {
		printer.PrintWarning(fmt.Sprintf("Completed: %d succeeded, %d failed", success, errors))
	} else {
		printer.PrintSuccess(fmt.Sprintf("Success: %d module(s) bumped", success))
	}
}
