package showcmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/cmd/sley/flags"
	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// Run returns the "show" command.
func Run(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Display current version",
		UsageText: "sley show [--all] [--module name] [--format text|json|table]",
		Flags:     flags.MultiModuleFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runShowCmd(ctx, cmd, cfg)
		},
	}
}

// runShowCmd prints the current version.
func runShowCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	// Get execution context to determine single vs multi-module mode
	// Use WithDefaultAll since show is a read-only command - no need for TUI prompt
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg, clix.WithDefaultAll())
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModuleShow(cmd, execCtx.Path)
	}

	// Handle multi-module mode
	return runMultiModuleShow(ctx, cmd, execCtx)
}

// runSingleModuleShow handles the single-module show operation.
func runSingleModuleShow(cmd *cli.Command, path string) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version file at %s: %w", path, err)
	}

	fmt.Println(version.String())
	return nil
}

// runMultiModuleShow handles the multi-module show operation.
func runMultiModuleShow(ctx context.Context, cmd *cli.Command, execCtx *clix.ExecutionContext) error {
	fs := core.NewOSFileSystem()
	operation := operations.NewShowOperation(fs)

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

	formatter := workspace.GetFormatter(format, "Version Summary")

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

	return nil
}

// printQuietSummary prints a minimal summary of results.
func printQuietSummary(results []workspace.ExecutionResult) {
	success := workspace.SuccessCount(results)
	errors := workspace.ErrorCount(results)
	if errors > 0 {
		printer.PrintWarning(fmt.Sprintf("Completed: %d succeeded, %d failed", success, errors))
	} else {
		printer.PrintInfo(fmt.Sprintf("Success: %d module(s)", success))
	}
}
