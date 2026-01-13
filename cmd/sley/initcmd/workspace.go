package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/tui"
)

// DiscoveredModule represents a .version file found during workspace discovery.
type DiscoveredModule struct {
	Name    string
	Path    string
	RelPath string
	Version string
}

// runWorkspaceInit initializes a monorepo/workspace configuration.
func runWorkspaceInit(path string, yesFlag bool, templateFlag, enableFlag string, forceFlag bool) error {
	// Step 1: Discover existing .version files in subdirectories
	modules, err := discoverVersionFiles(".")
	if err != nil {
		return fmt.Errorf("failed to discover modules: %w", err)
	}

	// Step 2: Detect project context
	projectCtx := DetectProjectContext()

	// Step 3: Determine which plugins to enable
	selectedPlugins, err := determinePlugins(projectCtx, yesFlag, templateFlag, enableFlag)
	if err != nil {
		return err
	}

	// If no plugins selected (user canceled), skip config creation
	if len(selectedPlugins) == 0 {
		return nil
	}

	// Step 4: Create .sley.yaml with workspace configuration
	configCreated, err := createWorkspaceConfigFile(selectedPlugins, modules, forceFlag)
	if err != nil {
		return err
	}

	// Step 5: Print success messages
	printWorkspaceSuccessSummary(configCreated, selectedPlugins, modules, projectCtx)

	return nil
}

// discoverVersionFiles searches for .version files in subdirectories.
func discoverVersionFiles(root string) ([]DiscoveredModule, error) {
	var modules []DiscoveredModule

	excludeDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"tmp":          true,
		"build":        true,
		"dist":         true,
		".cache":       true,
		"__pycache__":  true,
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible directories
		}

		// Skip excluded directories
		if d.IsDir() {
			if excludeDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for .version files
		if d.Name() == ".version" {
			// Skip root .version file
			dir := filepath.Dir(path)
			if dir == "." || dir == root {
				return nil
			}

			relPath, _ := filepath.Rel(root, path)
			moduleName := filepath.Base(dir)

			// Read current version
			version := ""
			if data, err := os.ReadFile(path); err == nil {
				version = strings.TrimSpace(string(data))
			}

			modules = append(modules, DiscoveredModule{
				Name:    moduleName,
				Path:    path,
				RelPath: relPath,
				Version: version,
			})
		}

		return nil
	})

	return modules, err
}

// createWorkspaceConfigFile generates and writes the .sley.yaml with workspace configuration.
func createWorkspaceConfigFile(plugins []string, modules []DiscoveredModule, forceFlag bool) (bool, error) {
	configPath := ".sley.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !forceFlag {
		if !isTerminalInteractive() {
			return false, nil
		}

		confirmed, err := ConfirmOverwrite()
		if err != nil {
			return false, err
		}
		if !confirmed {
			return false, nil
		}
	}

	// Generate config with workspace section
	configData, err := GenerateWorkspaceConfigWithComments(plugins, modules)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	return true, nil
}

// GenerateWorkspaceConfigWithComments generates YAML config with workspace section.
// In workspace mode, the root path field is omitted since each module defines its own path.
func GenerateWorkspaceConfigWithComments(plugins []string, modules []DiscoveredModule) ([]byte, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# sley configuration file\n")
	sb.WriteString("# Documentation: https://github.com/indaco/sley\n")
	sb.WriteString("\n")

	// List enabled plugins in header
	if len(plugins) > 0 {
		sb.WriteString("# Enabled plugins:\n")
		for _, p := range plugins {
			sb.WriteString(fmt.Sprintf("#   - %s\n", p))
		}
		sb.WriteString("\n")
	}

	// Plugins section
	sb.WriteString("plugins:\n")
	for _, pluginName := range plugins {
		writePluginConfig(&sb, pluginName)
	}
	sb.WriteString("\n")

	// Workspace section
	sb.WriteString("# Workspace configuration for monorepo support\n")
	sb.WriteString("workspace:\n")
	sb.WriteString("  # Discovery settings for automatic module detection\n")
	sb.WriteString("  discovery:\n")
	sb.WriteString("    enabled: true\n")
	sb.WriteString("    recursive: true\n")
	sb.WriteString("    max_depth: 10\n")
	sb.WriteString("    exclude:\n")
	for _, pattern := range config.DefaultExcludePatterns {
		sb.WriteString(fmt.Sprintf("      - %q\n", pattern))
	}

	// If modules were discovered, add them as explicit modules
	if len(modules) > 0 {
		sb.WriteString("\n")
		sb.WriteString("  # Discovered modules (uncomment to use explicit configuration)\n")
		sb.WriteString("  # modules:\n")
		for _, mod := range modules {
			sb.WriteString(fmt.Sprintf("  #   - name: %s\n", mod.Name))
			sb.WriteString(fmt.Sprintf("  #     path: %s\n", mod.RelPath))
		}
	}

	return []byte(sb.String()), nil
}

// writePluginConfig writes a single plugin configuration to the builder.
func writePluginConfig(sb *strings.Builder, pluginName string) {
	descriptions := map[string]string{
		"commit-parser":       "Analyzes conventional commits to suggest version bumps",
		"tag-manager":         "Automatically creates git tags after version changes",
		"version-validator":   "Enforces versioning policies and constraints",
		"dependency-check":    "Syncs version to package.json and other files",
		"changelog-parser":    "Infers bump type from CHANGELOG.md entries",
		"changelog-generator": "Generates changelogs from git commits",
		"release-gate":        "Pre-bump validation (clean worktree, branch checks)",
		"audit-log":           "Records version history with metadata",
	}

	desc := descriptions[pluginName]
	if desc != "" {
		fmt.Fprintf(sb, "  # %s\n", desc)
	}

	// commit-parser is a simple boolean
	if pluginName == "commit-parser" {
		sb.WriteString("  commit-parser: true\n")
		return
	}

	// Other plugins use enabled: true format
	fmt.Fprintf(sb, "  %s:\n", pluginName)
	sb.WriteString("    enabled: true\n")
}

// printWorkspaceSuccessSummary prints the success message for workspace init.
func printWorkspaceSuccessSummary(configCreated bool, plugins []string, modules []DiscoveredModule, ctx *ProjectContext) {
	if configCreated {
		printer.PrintSuccess(fmt.Sprintf("Created .sley.yaml with %d plugin%s and workspace configuration",
			len(plugins), tui.Pluralize(len(plugins))))
	}

	if len(modules) > 0 {
		printer.PrintInfo(fmt.Sprintf("Discovered %d module%s:", len(modules), tui.Pluralize(len(modules))))
		for _, mod := range modules {
			version := mod.Version
			if version == "" {
				version = "unknown"
			}
			fmt.Printf("  - %s (%s) at %s\n", mod.Name, version, mod.RelPath)
		}
	} else {
		printer.PrintInfo("No existing .version files found in subdirectories")
		fmt.Println("  Create .version files in your module directories, then run:")
		fmt.Println("    sley modules list")
	}

	// Print next steps
	fmt.Println()
	printer.PrintInfo("Next steps:")
	fmt.Println("  - Review .sley.yaml and adjust settings")
	if len(modules) == 0 {
		fmt.Println("  - Create .version files in your module directories")
	}
	fmt.Println("  - Run 'sley modules list' to see discovered modules")
	fmt.Println("  - Run 'sley bump patch --all' to bump all modules")
}
