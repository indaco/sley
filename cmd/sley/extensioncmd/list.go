package extensioncmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
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
		fmt.Println("failed to load configuration:", err)
		return nil
	}

	if len(cfg.Extensions) == 0 {
		fmt.Println("No extensions registered.")
		return nil
	}

	fmt.Println("List of Registered Extensions:")
	fmt.Println()
	fmt.Println("  NAME              VERSION     ENABLED   DESCRIPTION")
	fmt.Println("  ----------------------------------------------------------")

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
