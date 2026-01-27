package extensionmgr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
	"github.com/indaco/sley/internal/printer"
)

var (
	userHomeDirFn            = os.UserHomeDir
	RegisterLocalExtensionFn = registerLocalExtension
)

// registerLocalExtension installs an extension from a local directory into the extension
// directory and registers it in the project's configuration file.
//
// The function performs the following steps:
//  1. Validates that localPath is a directory containing a valid extension
//  2. Loads and validates the extension manifest (extension.yaml)
//  3. Resolves the destination directory based on extensionDirectory parameter
//  4. Copies extension files to the destination
//  5. Updates the configuration file to register the extension
//
// Parameters:
//   - localPath: Path to the extension source directory (must contain extension.yaml)
//   - configPath: Path to the sley configuration file (e.g., ".sley.yaml")
//   - extensionDirectory: Base directory for extension installation (see below)
//
// Extension Directory Resolution:
//
//   - If extensionDirectory is "" (empty): Global installation
//     Uses ~/.sley-extensions as the base directory
//     Example: ~/.sley-extensions/extension-name
//
//   - If extensionDirectory is specified: Project-local installation
//     Uses <extensionDirectory>/.sley-extensions as the base directory
//     Examples:
//     extensionDirectory="."  -> ./.sley-extensions/extension-name
//     extensionDirectory="/path/to/project" -> /path/to/project/.sley-extensions/extension-name
//
// This allows users to:
//  1. Install extensions globally (shared across projects) using "" or by omitting the flag
//  2. Install extensions locally (project-specific) by specifying a directory path
//
// Returns an error if:
//   - localPath is not a directory
//   - extension.yaml is missing or invalid
//   - configuration file doesn't exist
//   - file operations fail (permissions, disk space, etc.)
func registerLocalExtension(localPath, configPath, extensionDirectory string) error {
	// 1. Validate source path (ensure it's a directory)
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("extension path %q error: %w", localPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("extension path %q must be a directory", localPath)
	}

	// 2. Load and validate the extension manifest
	manifest, err := extensions.LoadExtensionManifestFn(localPath)
	if err != nil {
		return fmt.Errorf("failed to load extension manifest from %q: %w", localPath, err)
	}

	// 3. Resolve base extension directory
	// Determine installation location based on extensionDirectory parameter:
	//   - Empty string: Global installation at ~/.sley-extensions
	//   - Specified path: Project-local installation at <path>/.sley-extensions
	baseDir := extensionDirectory
	if baseDir == "" {
		// Global installation: use user's home directory
		homeDir, err := userHomeDirFn()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".sley-extensions")
	} else {
		// Project-local installation: use specified directory
		baseDir = filepath.Join(baseDir, ".sley-extensions")
	}

	// Final destination path includes the extension name
	destPath := filepath.Join(baseDir, manifest.Name)

	// 4. Resolve and validate config path
	if configPath == "" {
		configPath = ".sley.yaml"
	}
	absConfigPath, _ := filepath.Abs(configPath)

	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		// Print user-friendly guidance for missing config file
		printer.PrintError(fmt.Sprintf("Config file not found: %s", absConfigPath))
		fmt.Println()
		printer.PrintInfo("To enable extension support, create a .sley.yaml file:")
		fmt.Println()
		fmt.Println("    echo 'extensions: []' > .sley.yaml")
		fmt.Println()
		printer.PrintInfo("Then run this command again.")
		return fmt.Errorf("config file not found at %s", absConfigPath)
	}

	// 5. Copy the extension files to the destination directory
	if err := copyDirFn(localPath, destPath); err != nil {
		return fmt.Errorf("failed to copy extension files from %q to %q: %w", localPath, destPath, err)
	}

	// 6. Update the config
	extensionCfg := config.ExtensionConfig{
		Name:    manifest.Name,
		Path:    destPath,
		Enabled: true,
	}

	// 7. Add the extension to the config file
	if err := AddExtensionToConfigFn(absConfigPath, extensionCfg); err != nil {
		return fmt.Errorf("failed to update config %q: %w", absConfigPath, err)
	}

	// 8. Success message
	printer.PrintSuccess(fmt.Sprintf("Extension %q registered successfully.", manifest.Name))
	return nil
}
