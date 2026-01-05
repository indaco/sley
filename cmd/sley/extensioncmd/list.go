package extensioncmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// listCmd returns the "list" subcommand.
func listCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List installed extensions",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtenstionList()
		},
	}
}

// runExtenstionList lists installed extensions.
func runExtenstionList() error {
	cfg, err := config.LoadConfigFn()
	if err != nil {
		printer.PrintError(fmt.Sprintf("failed to load configuration: %v", err))
		return nil
	}

	if len(cfg.Extensions) == 0 {
		printer.PrintInfo("No extensions registered.")
		return nil
	}

	printer.PrintBold("List of Registered Extensions:")
	fmt.Println()
	fmt.Printf("  %s\n", printer.Faint("NAME              VERSION     ENABLED   DESCRIPTION"))
	fmt.Printf("  %s\n", printer.Faint("----------------------------------------------------------"))

	for _, ext := range cfg.Extensions {
		version := "?"
		desc := "(no manifest)"

		if ext.Path != "" {
			if manifest, err := extensions.LoadExtensionManifestFn(ext.Path); err == nil {
				version = manifest.Version
				desc = manifest.Description
			}
		}

		fmt.Printf("  %-17s %-10s %-9v %s\n", ext.Name, version, ext.Enabled, desc)
	}

	fmt.Println()

	return nil
}
