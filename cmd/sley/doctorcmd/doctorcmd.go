package doctorcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// runDoctorCmd validates both configuration and .version files.
func runDoctorCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	// First, validate the configuration file
	if err := validateConfiguration(cmd, cfg); err != nil {
		return err
	}

	// Get execution context to determine single vs multi-module mode
	// Use WithDefaultAll since doctor is a read-only command - no need for TUI prompt
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg, clix.WithDefaultAll())
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

// validateConfiguration validates the .sley.yaml configuration file.
func validateConfiguration(cmd *cli.Command, cfg *config.Config) error {
	// Determine config file path and root directory
	configPath := ".sley.yaml"
	rootDir, err := os.Getwd()
	if err != nil {
		rootDir = "."
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			configPath = "" // No config file
		}
	}

	// Create validator
	fs := core.NewOSFileSystem()
	validator := config.NewValidator(fs, cfg, configPath, rootDir)

	// Run validation
	results, err := validator.Validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Format and display results
	format := cmd.String("format")
	quiet := cmd.Bool("quiet")

	if quiet {
		// In quiet mode, only show summary
		printConfigValidationSummary(results)
	} else {
		printConfigValidationResults(results, format)
	}

	// Return error if any validation failed
	if config.HasErrors(results) {
		return fmt.Errorf("configuration validation failed with %d error(s)", config.ErrorCount(results))
	}

	return nil
}

// printConfigValidationResults prints detailed validation results.
func printConfigValidationResults(results []config.ValidationResult, format string) {
	if format == "json" {
		printConfigValidationJSON(results)
		return
	}

	// Text/table format
	fmt.Println() // Empty line before header
	printer.PrintInfo("Configuration Validation:")
	printer.PrintFaint(strings.Repeat("-", 70))

	for _, result := range results {
		var formatted string
		switch {
		case !result.Passed:
			// FAIL - bold red symbol and badge, normal category, faint message
			formatted = printer.FormatValidationFail("✗", "[FAIL]", result.Category, result.Message)
		case result.Warning:
			// WARN - bold yellow symbol and badge, normal category, faint message
			formatted = printer.FormatValidationWarn("⚠", "[WARN]", result.Category, result.Message)
		default:
			// PASS - bold green symbol and badge, normal category, faint message
			formatted = printer.FormatValidationPass("✓", "[PASS]", result.Category, result.Message)
		}
		fmt.Println(formatted)
	}

	printer.PrintFaint(strings.Repeat("-", 70))
	printConfigValidationSummary(results)
	fmt.Println()
}

// printConfigValidationJSON prints validation results in JSON format.
func printConfigValidationJSON(results []config.ValidationResult) {
	type jsonResult struct {
		Category string `json:"category"`
		Status   string `json:"status"`
		Message  string `json:"message"`
	}

	output := make([]jsonResult, len(results))
	for i, r := range results {
		status := "pass"
		if !r.Passed {
			status = "fail"
		} else if r.Warning {
			status = "warning"
		}

		output[i] = jsonResult{
			Category: r.Category,
			Status:   status,
			Message:  r.Message,
		}
	}

	data, err := json.Marshal(map[string]any{
		"validations": output,
		"summary": map[string]int{
			"total":    len(results),
			"passed":   len(results) - config.ErrorCount(results) - config.WarningCount(results),
			"errors":   config.ErrorCount(results),
			"warnings": config.WarningCount(results),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// printConfigValidationSummary prints a summary of validation results.
func printConfigValidationSummary(results []config.ValidationResult) {
	total := len(results)
	errors := config.ErrorCount(results)
	warnings := config.WarningCount(results)
	passed := total - errors - warnings

	switch {
	case errors > 0:
		printer.PrintError(fmt.Sprintf("Summary: %d passed, %d errors, %d warnings", passed, errors, warnings))
	case warnings > 0:
		printer.PrintWarning(fmt.Sprintf("Summary: %d passed, %d warnings", passed, warnings))
	default:
		printer.PrintSuccess(fmt.Sprintf("Summary: All %d validation(s) passed", total))
	}
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

	formatter := workspace.GetFormatterWithVerb(format, "Validation Summary", "validated")

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
