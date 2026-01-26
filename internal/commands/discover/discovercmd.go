package discover

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/discovery"
	"github.com/urfave/cli/v3"
)

// Run returns the "discover" command.
func Run(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "discover",
		Aliases: []string{"scan"},
		Usage:   "Scan for version sources and suggest configuration",
		UsageText: `sley discover [options]

Scans the current directory for:
  - .version files (sley modules)
  - Manifest files (package.json, Cargo.toml, pyproject.toml, etc.)

Shows discovered version sources and suggests dependency-check configuration
for keeping versions in sync.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: text, json, table",
				Value:   "text",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Only show summary",
			},
			&cli.BoolFlag{
				Name:  "no-interactive",
				Usage: "Skip interactive prompts",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runDiscoverCmd(ctx, cmd, cfg)
		},
	}
}

// runDiscoverCmd executes the discover command.
func runDiscoverCmd(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	// Get current working directory
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Run discovery
	fs := core.NewOSFileSystem()
	svc := discovery.NewService(fs, cfg)

	result, err := svc.Discover(ctx, rootDir)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	// Format and display results
	format := ParseOutputFormat(cmd.String("format"))
	quiet := cmd.Bool("quiet")

	formatter := NewFormatter(format)

	if quiet {
		// In quiet mode, only show summary
		printQuietSummary(result)
	} else {
		formatter.PrintResult(result)
	}

	// Run interactive workflow if not disabled
	noInteractive := cmd.Bool("no-interactive")
	if !noInteractive && format == FormatText {
		prompter := NewPrompter()
		workflow := NewWorkflow(prompter, result, rootDir)
		if _, err := workflow.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}

// printQuietSummary prints a minimal summary of discovery results.
func printQuietSummary(result *discovery.Result) {
	moduleCount := len(result.Modules)
	manifestCount := len(result.Manifests)
	mismatchCount := len(result.Mismatches)

	fmt.Printf("Mode: %s | Modules: %d | Manifests: %d", result.Mode, moduleCount, manifestCount)

	if mismatchCount > 0 {
		fmt.Printf(" | Mismatches: %d", mismatchCount)
	}

	if result.PrimaryVersion() != "" {
		fmt.Printf(" | Version: %s", result.PrimaryVersion())
	}

	fmt.Println()
}

// DiscoverAndSuggest is a helper function that performs discovery and returns
// suggested dependency-check configuration.
func DiscoverAndSuggest(ctx context.Context, cfg *config.Config, rootDir string) (*discovery.Result, *config.DependencyCheckConfig, error) {
	fs := core.NewOSFileSystem()
	svc := discovery.NewService(fs, cfg)

	result, err := svc.Discover(ctx, rootDir)
	if err != nil {
		return nil, nil, err
	}

	suggestion := SuggestDependencyCheckFromDiscovery(result)

	return result, suggestion, nil
}
