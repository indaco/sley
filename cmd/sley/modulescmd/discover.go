package modulescmd

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// discoverCmd returns the "discover" subcommand for testing discovery settings.
func discoverCmd() *cli.Command {
	return &cli.Command{
		Name:  "discover",
		Usage: "Test module discovery without running operations",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be discovered",
				Value: true,
			},
		},
		Action: runDiscover,
	}
}

func runDiscover(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadConfigFn()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)

	// Show discovery settings
	discovery := cfg.GetDiscoveryConfig()
	printer.PrintBold("Discovery settings:")
	fmt.Printf("  %s\n", printer.Faint(fmt.Sprintf("Enabled: %v", *discovery.Enabled)))
	fmt.Printf("  %s\n", printer.Faint(fmt.Sprintf("Recursive: %v", *discovery.Recursive)))
	fmt.Printf("  %s\n", printer.Faint(fmt.Sprintf("Max depth: %d", *discovery.MaxDepth)))
	fmt.Printf("  %s\n", printer.Faint(fmt.Sprintf("Exclude patterns: %v", cfg.GetExcludePatterns())))
	fmt.Println()

	// Detect context
	detectedCtx, err := detector.DetectContext(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect context: %w", err)
	}

	printer.PrintInfo(fmt.Sprintf("Detection mode: %s", detectedCtx.Mode))
	fmt.Println()

	switch detectedCtx.Mode {
	case workspace.SingleModule:
		printer.PrintSuccess("Single module found:")
		fmt.Printf("  %s\n", printer.Faint(fmt.Sprintf("Path: %s", detectedCtx.Path)))

	case workspace.MultiModule:
		printer.PrintSuccess(fmt.Sprintf("Multiple modules found (%d):", len(detectedCtx.Modules)))
		for _, mod := range detectedCtx.Modules {
			version := mod.CurrentVersion
			if version == "" {
				version = "unknown"
			}
			fmt.Printf("  - %s (%s)\n", mod.Name, version)
			fmt.Printf("    %s\n", printer.Faint(fmt.Sprintf("Path: %s", mod.RelPath)))
		}

	case workspace.NoModules:
		printer.PrintInfo("No modules found in workspace")
	}

	return nil
}
