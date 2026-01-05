package extensioncmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// listCmd returns the "list" subcommand.
func removeCmd() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove a registered extension",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the extension to remove",
			},
			&cli.BoolFlag{
				Name:  "delete-folder",
				Usage: "Delete the extension directory from the .sley-extensions folder",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtenstionRemove(cmd)
		},
	}
}

// runExtenstionRemove removes an installed extension.
func runExtenstionRemove(cmd *cli.Command) error {
	extensionName := cmd.String("name")
	if extensionName == "" {
		return fmt.Errorf("please provide an extension name to remove")
	}

	cfg, err := config.LoadConfigFn()
	if err != nil {
		printer.PrintError(fmt.Sprintf("failed to load configuration: %v", err))
		return nil
	}

	var extensionToRemove *config.ExtensionConfig
	for i, extension := range cfg.Extensions {
		if extension.Name == extensionName {
			extensionToRemove = &cfg.Extensions[i]
			break
		}
	}

	if extensionToRemove == nil {
		printer.PrintWarning(fmt.Sprintf("extension %q not found", extensionName))
		return nil
	}

	// Disable the plugin in the configuration (set Enabled to false)
	extensionToRemove.Enabled = false

	// Save the updated config back to the file
	if err := config.SaveConfigFn(cfg); err != nil {
		printer.PrintError(fmt.Sprintf("failed to save updated configuration: %v", err))
		return nil
	}

	// Check if --delete-folder flag is set to remove the extension folder
	isDeleteFolder := cmd.Bool("delete-folder")
	if isDeleteFolder {
		// Remove the extension directory from ".sley-extensions"
		extensionDir := filepath.Join(".sley-extensions", extensionName)
		if err := os.RemoveAll(extensionDir); err != nil {
			return fmt.Errorf("failed to remove extension directory: %w", err)
		}
		printer.PrintSuccess(fmt.Sprintf("Extension %q and its directory removed successfully.", extensionName))
	} else {
		printer.PrintInfo(fmt.Sprintf("Extension %q removed, but its directory is preserved.", extensionName))
	}

	return nil
}
