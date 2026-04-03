package initialize

import (
	"fmt"
	"io"
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
	modules, err := discoverVersionFiles(path)
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

	// Step 4: Detect monorepo workspace markers
	monoInfo, err := DetectMonorepo()
	if err != nil {
		return fmt.Errorf("failed to detect monorepo: %w", err)
	}

	applyMonorepo := false
	if monoInfo != nil && len(monoInfo.Modules) > 0 {
		if yesFlag || !isTerminalInteractive() {
			applyMonorepo = true
		} else {
			confirmed, promptErr := ConfirmMonorepoDefaults(monoInfo)
			if promptErr != nil {
				return promptErr
			}
			applyMonorepo = confirmed
		}
	}

	// Step 5: Create .sley.yaml with workspace configuration
	var configCreated bool
	if applyMonorepo {
		configCreated, err = createWorkspaceConfigFileWithMonorepo(selectedPlugins, modules, monoInfo, forceFlag)
	} else {
		configCreated, err = createWorkspaceConfigFile(selectedPlugins, modules, forceFlag)
	}
	if err != nil {
		return err
	}

	// Step 6: Create .version files for detected monorepo modules
	var createdVersionFiles []string
	if applyMonorepo {
		createdVersionFiles = CreateMonorepoVersionFiles(monoInfo)
	}

	// Step 7: Print success messages
	printWorkspaceSuccessSummary(configCreated, selectedPlugins, modules, createdVersionFiles, projectCtx)

	if applyMonorepo {
		printMonorepoSummary(monoInfo)
	}

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

	// Resolve the scan directory: if root is a file path, use its parent directory.
	scanDir := root
	if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
		scanDir = filepath.Dir(root)
	}

	// Use root-scoped API to prevent symlink TOCTOU traversal (gosec G122).
	rootDir, err := os.OpenRoot(scanDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open root directory: %w", err)
	}
	defer rootDir.Close()

	err = filepath.WalkDir(scanDir, func(path string, d os.DirEntry, err error) error {
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
			if dir == "." || dir == scanDir {
				return nil
			}

			relPath, _ := filepath.Rel(scanDir, path)
			moduleName := filepath.Base(dir)

			// Read current version using root-scoped file access
			version := ""
			if f, openErr := rootDir.Open(relPath); openErr == nil {
				if data, readErr := io.ReadAll(f); readErr == nil {
					version = strings.TrimSpace(string(data))
				}
				f.Close()
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

// createWorkspaceConfigFileWithMonorepo generates and writes the .sley.yaml with monorepo defaults applied.
func createWorkspaceConfigFileWithMonorepo(plugins []string, modules []DiscoveredModule, monoInfo *MonorepoInfo, forceFlag bool) (bool, error) {
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

	// Generate config with workspace + monorepo section
	configData, err := GenerateWorkspaceConfigWithMonorepo(plugins, modules, monoInfo)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	return true, nil
}

// CreateMonorepoVersionFiles creates .version files in each detected module directory
// if one does not already exist. Each file is initialized with "0.0.0".
// CreateMonorepoVersionFiles creates .version files in module directories
// and returns the list of created file paths.
func CreateMonorepoVersionFiles(monoInfo *MonorepoInfo) []string {
	var created []string
	for _, modDir := range monoInfo.Modules {
		versionFile := filepath.Join(modDir, ".version")
		if _, err := os.Stat(versionFile); err == nil {
			// Already exists, skip
			continue
		}
		if err := os.WriteFile(versionFile, []byte("0.0.0\n"), 0o600); err != nil {
			printer.PrintWarning(fmt.Sprintf("Failed to create %s: %v", versionFile, err))
			continue
		}
		created = append(created, versionFile)
	}
	return created
}

// printMonorepoSummary prints additional information about the monorepo setup.
func printMonorepoSummary(monoInfo *MonorepoInfo) {
	ty := printer.Typography()
	fmt.Println(ty.Compose(
		printer.Info(fmt.Sprintf("Monorepo detected: %s workspace (%s)", monoInfo.Type, monoInfo.MarkerFile)),
		ty.Section(ty.H4("Applied monorepo defaults"), ty.KVGroup([][2]string{
			{"Versioning", "independent"},
			{"Tag prefix", "{module_path}/v"},
			{"Modules", fmt.Sprintf("%d", len(monoInfo.Modules))},
		})),
	))
}

// GenerateWorkspaceConfigWithMonorepo generates YAML config with workspace section and monorepo defaults.
// This sets versioning to "independent" and configures the tag-manager prefix.
func GenerateWorkspaceConfigWithMonorepo(plugins []string, modules []DiscoveredModule, monoInfo *MonorepoInfo) ([]byte, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# sley configuration file\n")
	sb.WriteString("# Documentation: https://github.com/indaco/sley\n")
	sb.WriteString("\n")

	// List enabled plugins in header
	if len(plugins) > 0 {
		sb.WriteString("# Enabled plugins:\n")
		for _, p := range plugins {
			fmt.Fprintf(&sb, "#   - %s\n", p)
		}
		sb.WriteString("\n")
	}

	// Workspace section first (structure before behavior)
	sb.WriteString("# Workspace configuration for monorepo support\n")
	sb.WriteString("workspace:\n")
	sb.WriteString("  # Versioning mode: \"independent\" (each module versioned separately)\n")
	sb.WriteString("  versioning: independent\n")
	sb.WriteString("  # Discovery settings for automatic module detection\n")
	sb.WriteString("  discovery:\n")
	sb.WriteString("    enabled: true\n")
	sb.WriteString("    recursive: true\n")
	sb.WriteString("    module_max_depth: 10\n")
	sb.WriteString("    exclude:\n")
	for _, pattern := range config.DefaultExcludePatterns {
		fmt.Fprintf(&sb, "      - %q\n", pattern)
	}

	// Add discovered modules as explicit entries
	if len(modules) > 0 {
		sb.WriteString("\n")
		sb.WriteString("  # Discovered modules (uncomment to use explicit configuration)\n")
		sb.WriteString("  # modules:\n")
		for _, mod := range modules {
			fmt.Fprintf(&sb, "  #   - name: %s\n", mod.Name)
			fmt.Fprintf(&sb, "  #     path: %s\n", mod.RelPath)
		}
	}

	// Add monorepo module list if discovery found modules not in existing .version files
	if monoInfo != nil && len(monoInfo.Modules) > 0 && len(modules) == 0 {
		sb.WriteString("\n")
		sb.WriteString("  # Detected monorepo modules (uncomment to use explicit configuration)\n")
		sb.WriteString("  # modules:\n")
		for _, modPath := range monoInfo.Modules {
			name := filepath.Base(modPath)
			fmt.Fprintf(&sb, "  #   - name: %s\n", name)
			fmt.Fprintf(&sb, "  #     path: %s/.version\n", modPath)
		}
	}

	sb.WriteString("\n")

	// Plugins section
	sb.WriteString("# Plugin configuration\n")
	sb.WriteString("plugins:\n")
	for _, pluginName := range plugins {
		writePluginConfigWithMonorepo(&sb, pluginName)
	}

	return []byte(sb.String()), nil
}

// writePluginConfigWithMonorepo writes a single plugin configuration to the builder,
// applying monorepo-specific defaults (e.g., tag-manager prefix).
func writePluginConfigWithMonorepo(sb *strings.Builder, pluginName string) {
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

	// tag-manager gets monorepo prefix
	if pluginName == "tag-manager" {
		sb.WriteString("  tag-manager:\n")
		sb.WriteString("    enabled: true\n")
		sb.WriteString("    prefix: \"{module_path}/v\"\n")
		return
	}

	// Other plugins use enabled: true format
	fmt.Fprintf(sb, "  %s:\n", pluginName)
	sb.WriteString("    enabled: true\n")
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
			fmt.Fprintf(&sb, "#   - %s\n", p)
		}
		sb.WriteString("\n")
	}

	// Workspace section first (structure before behavior)
	sb.WriteString("# Workspace configuration for monorepo support\n")
	sb.WriteString("workspace:\n")
	sb.WriteString("  # Discovery settings for automatic module detection\n")
	sb.WriteString("  discovery:\n")
	sb.WriteString("    enabled: true\n")
	sb.WriteString("    recursive: true\n")
	sb.WriteString("    module_max_depth: 10\n")
	sb.WriteString("    exclude:\n")
	for _, pattern := range config.DefaultExcludePatterns {
		fmt.Fprintf(&sb, "      - %q\n", pattern)
	}

	// If modules were discovered, add them as explicit modules
	if len(modules) > 0 {
		sb.WriteString("\n")
		sb.WriteString("  # Discovered modules (uncomment to use explicit configuration)\n")
		sb.WriteString("  # modules:\n")
		for _, mod := range modules {
			fmt.Fprintf(&sb, "  #   - name: %s\n", mod.Name)
			fmt.Fprintf(&sb, "  #     path: %s\n", mod.RelPath)
		}
	}

	sb.WriteString("\n")

	// Plugins section
	sb.WriteString("# Plugin configuration\n")
	sb.WriteString("plugins:\n")
	for _, pluginName := range plugins {
		writePluginConfig(&sb, pluginName)
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
func printWorkspaceSuccessSummary(configCreated bool, plugins []string, modules []DiscoveredModule, createdFiles []string, _ *ProjectContext) {
	ty := printer.Typography()
	var blocks []string

	if len(createdFiles) > 0 {
		faintFiles := make([]string, len(createdFiles))
		for i, f := range createdFiles {
			faintFiles[i] = ty.Small(f)
		}
		blocks = append(blocks, ty.Section(ty.H4("Created version files"), ty.UL(faintFiles...)))
	}

	if configCreated {
		blocks = append(blocks, printer.Success(fmt.Sprintf("Created .sley.yaml with %d plugin%s and workspace configuration",
			len(plugins), tui.Pluralize(len(plugins)))))
	}

	if len(modules) > 0 {
		modItems := make([]string, len(modules))
		for i, mod := range modules {
			version := mod.Version
			if version == "" {
				version = "unknown"
			}
			modItems[i] = fmt.Sprintf("%s (%s) at %s", mod.Name, version, mod.RelPath)
		}
		blocks = append(blocks, ty.Section(
			ty.H4(fmt.Sprintf("Discovered %d module%s", len(modules), tui.Pluralize(len(modules)))),
			ty.UL(modItems...),
		))
	} else {
		blocks = append(blocks,
			printer.Info("No existing .version files found in subdirectories"),
			ty.P("Create .version files in your module directories, then run:"),
			ty.Code("sley modules list"),
		)
	}

	// Next steps
	nextSteps := []string{"Review .sley.yaml and adjust settings"}
	if len(modules) == 0 {
		nextSteps = append(nextSteps, "Create .version files in your module directories")
	}
	nextSteps = append(nextSteps,
		"Run 'sley modules list' to see discovered modules",
		"Run 'sley bump patch --all' to bump all modules",
	)
	blocks = append(blocks, ty.Section(ty.H4("Next steps"), ty.UL(nextSteps...)))

	fmt.Println(ty.Compose(blocks...))
}
