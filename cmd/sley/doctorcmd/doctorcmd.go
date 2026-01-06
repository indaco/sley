package doctorcmd

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

// Run returns the "doctor" command (alias "validate").
func Run(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "doctor",
		Aliases:   []string{"validate"},
		Usage:     "Validate .version file(s) and configuration",
		UsageText: "sley doctor [--all] [--module name] [--format text|json|table]",
		Flags:     flags.MultiModuleFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runDoctorCmd(ctx, cmd, cfg)
		},
	}
}

// runDoctorCmd checks that the .version file is valid.
func runDoctorCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModuleValidate(cmd, execCtx.Path)
	}

	// Handle multi-module mode
	return runMultiModuleValidate(ctx, cmd, execCtx)
}

// runSingleModuleValidate handles the single-module validate operation.
func runSingleModuleValidate(cmd *cli.Command, path string) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	_, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("invalid version file at %s: %w", path, err)
	}

	printer.PrintSuccess(fmt.Sprintf("Valid version file at %s", path))
	return nil
}

// runMultiModuleValidate handles the multi-module validate operation.
func runMultiModuleValidate(ctx context.Context, cmd *cli.Command, execCtx *clix.ExecutionContext) error {
	fs := core.NewOSFileSystem()
	operation := operations.NewValidateOperation(fs)

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

	formatter := workspace.GetFormatter(format, "Validation Summary")

	if quiet {
		// In quiet mode, just show summary
		printQuietSummary(results)
	} else {
		fmt.Println(formatter.FormatResults(results))
	}

	// Return error if any failures occurred
	if workspace.HasErrors(results) {
		return fmt.Errorf("%d module(s) failed validation", workspace.ErrorCount(results))
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
		printer.PrintSuccess(fmt.Sprintf("Success: %d module(s) validated", success))
	}
}
