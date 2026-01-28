package extension

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// configFilePath is the default config file name. It is a package-level
// variable so tests can override it when running in temporary directories.
var configFilePath = ".sley.yaml"

// localExtensionsDir is the name of the local extensions directory.
const localExtensionsDir = ".sley-extensions"

// uninstallCmd returns the "uninstall" subcommand.
func uninstallCmd() *cli.Command {
	return &cli.Command{
		Name:    "uninstall",
		Aliases: []string{"remove"},
		Usage:   "Uninstall a registered extension",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the extension to uninstall",
			},
			&cli.BoolFlag{
				Name:  "delete-folder",
				Usage: "Delete the extension directory from the .sley-extensions folder",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtensionUninstall(cmd)
		},
	}
}

// runExtensionUninstall removes an installed extension.
// It uses a surgical YAML replacement so that comments and formatting in
// the configuration file are preserved.
func runExtensionUninstall(cmd *cli.Command) error {
	extensionName := cmd.String("name")
	if extensionName == "" {
		return fmt.Errorf("please provide an extension name to uninstall")
	}

	updater := extensionmgr.NewDefaultConfigUpdater(&extensionmgr.DefaultYAMLMarshaler{})
	if err := updater.RemoveExtension(configFilePath, extensionName); err != nil {
		// Distinguish "not found" from other errors so the user gets a
		// warning instead of a hard error.
		if strings.Contains(err.Error(), "not found in configuration") {
			printer.PrintWarning(fmt.Sprintf("extension %q not found", extensionName))
			return nil
		}
		printer.PrintError(fmt.Sprintf("failed to uninstall extension: %v", err))
		return nil
	}

	// Check if --delete-folder flag is set to remove the extension folder
	isDeleteFolder := cmd.Bool("delete-folder")
	if isDeleteFolder {
		// Remove the extension directory from ".sley-extensions"
		extensionDir := filepath.Join(localExtensionsDir, extensionName)
		if err := os.RemoveAll(extensionDir); err != nil {
			return fmt.Errorf("failed to remove extension directory: %w", err)
		}

		// Clean up the local .sley-extensions directory if it is now empty.
		// This only applies to the project-local directory, never the global
		// ~/.sley-extensions directory.
		if err := removeLocalExtensionsDirIfEmpty(localExtensionsDir); err != nil {
			return fmt.Errorf("failed to clean up extensions directory: %w", err)
		}

		printer.PrintSuccess(fmt.Sprintf("Extension %q and its directory uninstalled successfully.", extensionName))
	} else {
		printer.PrintInfo(fmt.Sprintf("Extension %q uninstalled, but its directory is preserved.", extensionName))
	}

	return nil
}

// removeLocalExtensionsDirIfEmpty removes the given directory if it exists
// and contains no entries. This is used to clean up the local
// .sley-extensions directory after the last extension has been uninstalled.
func removeLocalExtensionsDirIfEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory already gone; nothing to do.
			return nil
		}
		return fmt.Errorf("failed to read directory %q: %w", dir, err)
	}

	if len(entries) == 0 {
		if err := os.Remove(dir); err != nil {
			return fmt.Errorf("failed to remove empty directory %q: %w", dir, err)
		}
	}

	return nil
}
