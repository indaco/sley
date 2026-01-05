package setcmd

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

// Run returns the "set" command.
func Run(cfg *config.Config) *cli.Command {
	cmdFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  "pre",
			Usage: "Optional pre-release label",
		},
		&cli.StringFlag{
			Name:  "meta",
			Usage: "Optional build metadata",
		},
	}
	cmdFlags = append(cmdFlags, flags.MultiModuleFlags()...)

	return &cli.Command{
		Name:      "set",
		Usage:     "Set the version manually",
		UsageText: "sley set <version> [--pre label] [--all] [--module name]",
		Flags:     cmdFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runSetCmd(ctx, cmd, cfg)
		},
	}
}

// runSetCmd manually sets the version.
func runSetCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	args := cmd.Args()
	if args.Len() < 1 {
		return cli.Exit("missing required version argument", 1)
	}

	raw := args.Get(0)
	pre := cmd.String("pre")
	meta := cmd.String("meta")

	// Parse and validate the version first
	version, err := semver.ParseVersion(raw)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	version.PreRelease = pre
	version.Build = meta

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModuleSet(execCtx.Path, version)
	}

	// Handle multi-module mode
	return runMultiModuleSet(ctx, cmd, execCtx, version.String())
}

// runSingleModuleSet handles the single-module set operation.
func runSingleModuleSet(path string, version semver.SemVersion) error {
	if err := semver.SaveVersion(path, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	printer.PrintSuccess(fmt.Sprintf("Set version to %s in %s", version.String(), path))
	return nil
}

// runMultiModuleSet handles the multi-module set operation.
func runMultiModuleSet(ctx context.Context, cmd *cli.Command, execCtx *clix.ExecutionContext, version string) error {
	fs := core.NewOSFileSystem()
	operation := operations.NewSetOperation(fs, version)

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

	formatter := workspace.GetFormatter(format, fmt.Sprintf("Set version to %s", version))

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
		printer.PrintSuccess(fmt.Sprintf("Success: %d module(s) updated", success))
	}
}
