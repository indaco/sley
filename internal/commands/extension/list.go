package extension

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
			return runExtensionList()
		},
	}
}

// runExtensionList lists installed extensions.
func runExtensionList() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		printer.PrintError(fmt.Sprintf("failed to load configuration: %v", err))
		return nil
	}

	if len(cfg.Extensions) == 0 {
		printer.PrintFaint("No extensions registered.")
		return nil
	}

	ty := printer.Typography()

	rows := [][]string{
		{"Name", "Version", "Enabled", "Description"},
	}
	for _, ext := range cfg.Extensions {
		version := "?"
		desc := "(no manifest)"

		if ext.Path != "" {
			if manifest, err := extensions.LoadExtensionManifest(ext.Path); err == nil {
				version = manifest.Version
				desc = manifest.Description
			}
		}

		rows = append(rows, []string{ext.Name, version, fmt.Sprintf("%v", ext.Enabled), desc})
	}

	fmt.Println(ty.Section(ty.H4("Registered Extensions"), ty.Table(rows)))

	return nil
}
