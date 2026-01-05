package modulescmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// listCmd returns the "list" subcommand for showing discovered modules.
func listCmd() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all discovered modules in workspace",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed information",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format (text, json)",
				Value: "text",
			},
		},
		Action: runList,
	}
}

func runList(ctx context.Context, cmd *cli.Command) error {
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

	modules, err := detector.DiscoverModules(cwd)
	if err != nil {
		return fmt.Errorf("failed to discover modules: %w", err)
	}

	if len(modules) == 0 {
		printer.PrintInfo("No modules found in workspace")
		return nil
	}

	format := cmd.String("format")
	verbose := cmd.Bool("verbose")

	switch format {
	case "json":
		return outputJSON(modules)
	default:
		return outputText(modules, verbose)
	}
}

func outputText(modules []*workspace.Module, verbose bool) error {
	printer.PrintBold(fmt.Sprintf("Found %d module(s):", len(modules)))
	for _, mod := range modules {
		if verbose {
			fmt.Printf("  - %s\n", mod.Name)
			fmt.Printf("    %s\n", printer.Faint(fmt.Sprintf("Path: %s", mod.RelPath)))
			fmt.Printf("    %s\n", printer.Faint(fmt.Sprintf("Version: %s", mod.CurrentVersion)))
		} else {
			version := mod.CurrentVersion
			if version == "" {
				version = "unknown"
			}
			fmt.Printf("  - %s (%s)\n", mod.Name, version)

			// Display the directory path in faint style for disambiguation
			// RelPath includes the .version filename, so we extract just the directory
			dirPath := filepath.Dir(mod.RelPath)
			if dirPath != "." {
				fmt.Printf("    %s\n", printer.Faint(dirPath))
			}
		}
	}
	return nil
}

type moduleJSON struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

func outputJSON(modules []*workspace.Module) error {
	output := make([]moduleJSON, len(modules))
	for i, mod := range modules {
		output[i] = moduleJSON{
			Name:    mod.Name,
			Path:    mod.RelPath,
			Version: mod.CurrentVersion,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
