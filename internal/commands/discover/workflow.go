package discover

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/tui"
)

// Workflow handles the interactive discovery workflow.
type Workflow struct {
	prompter Prompter
	result   *discovery.Result
	rootDir  string
}

// NewWorkflow creates a new workflow handler.
func NewWorkflow(prompter Prompter, result *discovery.Result, rootDir string) *Workflow {
	return &Workflow{
		prompter: prompter,
		result:   result,
		rootDir:  rootDir,
	}
}

// Run executes the interactive workflow if appropriate.
// Returns true if the workflow completed with actions taken.
func (w *Workflow) Run(ctx context.Context) (bool, error) {
	// Check if we're in interactive mode
	if !tui.IsInteractive() {
		return false, nil
	}

	// If .sley.yaml exists, skip init workflow
	if configExists() {
		return w.runExistingConfigWorkflow(ctx)
	}

	// No config exists - offer to initialize
	return w.runInitWorkflow(ctx)
}

// runInitWorkflow handles the case when no .sley.yaml exists.
func (w *Workflow) runInitWorkflow(ctx context.Context) (bool, error) {
	fmt.Println()
	printer.PrintInfo("No .sley.yaml configuration found.")

	// Check if we have useful suggestions
	if len(w.result.SyncCandidates) == 0 && len(w.result.Modules) == 0 {
		printer.PrintFaint("Run 'sley init' to create a configuration file.")
		return false, nil
	}

	// Offer to initialize
	initConfig, err := w.prompter.Confirm(
		"Would you like to initialize sley configuration?",
		"This will create .sley.yaml with suggested settings based on discovered files.",
	)
	if err != nil {
		return false, err
	}

	if !initConfig {
		printer.PrintFaint("You can run 'sley init' later to create configuration.")
		return false, nil
	}

	// Handle multi-module projects specially
	if w.result.Mode == discovery.MultiModule {
		return w.runMultiModuleSetup(ctx)
	}

	// If we have sync candidates (manifest files), run the dependency-check setup
	if len(w.result.SyncCandidates) > 0 {
		return w.runDependencyCheckSetup(ctx)
	}

	// No sync candidates - create config with default plugins only
	return w.createConfigWithDefaults(ctx)
}

// WorkspaceChoice represents the user's choice for multi-module configuration.
type WorkspaceChoice string

const (
	// WorkspaceChoiceCoordinated syncs all .version files to root (coordinated versioning).
	WorkspaceChoiceCoordinated WorkspaceChoice = "coordinated"
	// WorkspaceChoiceWorkspace configures independent module versions (workspace mode).
	WorkspaceChoiceWorkspace WorkspaceChoice = "workspace"
	// WorkspaceChoiceSingleRoot uses only the root .version for all manifests.
	WorkspaceChoiceSingleRoot WorkspaceChoice = "single"
)

// runMultiModuleSetup handles configuration for multi-module/monorepo projects.
func (w *Workflow) runMultiModuleSetup(ctx context.Context) (bool, error) {
	fmt.Println()
	printer.PrintInfo(fmt.Sprintf("Found %d modules - this appears to be a monorepo.", len(w.result.Modules)))
	fmt.Println()

	// Show discovered modules
	printer.PrintFaint("Discovered modules:")
	for _, m := range w.result.Modules {
		fmt.Printf("  - %s (%s)\n", m.Name, m.RelPath)
	}
	fmt.Println()

	// Ask how to configure the project
	choice, err := w.prompter.Select(
		"How would you like to configure versioning?",
		"Choose how to manage versions in this monorepo.",
		[]huh.Option[string]{
			huh.NewOption("Coordinated versioning (recommended) - all .version files sync to root", string(WorkspaceChoiceCoordinated)),
			huh.NewOption("Independent workspace - each module versioned separately", string(WorkspaceChoiceWorkspace)),
			huh.NewOption("Single root only - ignore submodule .version files", string(WorkspaceChoiceSingleRoot)),
		},
	)
	if err != nil {
		return false, err
	}

	if choice == "" {
		printer.PrintFaint("Configuration canceled.")
		return false, nil
	}

	switch WorkspaceChoice(choice) {
	case WorkspaceChoiceCoordinated:
		return w.createConfigWithCoordinatedVersioning(ctx)
	case WorkspaceChoiceWorkspace:
		return w.createConfigWithWorkspace(ctx)
	case WorkspaceChoiceSingleRoot:
		// Only configure dependency-check for manifest files found near root
		return w.runDependencyCheckSetup(ctx)
	default:
		return w.createConfigWithDefaults(ctx)
	}
}

// createConfigWithCoordinatedVersioning creates .sley.yaml with coordinated versioning.
// All submodule .version files and manifest files sync to the root .version.
func (w *Workflow) createConfigWithCoordinatedVersioning(ctx context.Context) (bool, error) {
	// Ensure root .version exists
	if err := w.ensureVersionFile(ctx); err != nil {
		return false, err
	}

	// Combine: submodule .version files + manifest files as sync candidates
	var allSyncCandidates []discovery.SyncCandidate

	// Add submodule .version files (excluding root)
	for _, m := range w.result.Modules {
		if m.RelPath == ".version" {
			continue // Skip root
		}
		allSyncCandidates = append(allSyncCandidates, discovery.SyncCandidate{
			Path:        m.RelPath,
			Format:      parser.FormatRaw,
			Field:       "",
			Version:     m.Version,
			Description: "Version file (" + m.RelPath + ")",
		})
	}

	// Add manifest files
	allSyncCandidates = append(allSyncCandidates, w.result.SyncCandidates...)

	// Create config with dependency-check for ALL files
	return w.createConfigWithDependencyCheck(ctx, allSyncCandidates)
}

// createConfigWithWorkspace creates .sley.yaml with workspace configuration for multi-module projects.
func (w *Workflow) createConfigWithWorkspace(ctx context.Context) (bool, error) {
	// Initialize .version file if needed
	if err := w.ensureVersionFile(ctx); err != nil {
		return false, err
	}

	// Default plugins: commit-parser and tag-manager
	selectedPlugins := []string{"commit-parser", "tag-manager"}

	// Generate config with workspace discovery enabled
	configData, err := generateConfigYAMLWithWorkspace(defaultVersionPath(), selectedPlugins, w.result.SyncCandidates)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(".sley.yaml", configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	w.printWorkspaceInitSuccess(selectedPlugins)
	return true, nil
}

// printWorkspaceInitSuccess prints success messages after workspace initialization.
func (w *Workflow) printWorkspaceInitSuccess(plugins []string) {
	fmt.Println()
	printer.PrintSuccess(fmt.Sprintf("Created .sley.yaml with workspace configuration and %d plugin(s)", len(plugins)))

	// Show enabled plugins
	fmt.Println()
	printer.PrintInfo("Enabled plugins:")
	for _, p := range plugins {
		fmt.Printf("  - %s\n", p)
	}

	// Show workspace info
	fmt.Println()
	printer.PrintInfo("Workspace configuration:")
	fmt.Println("  - Auto-discovery enabled")
	fmt.Println("  - Each module manages its own .version file")

	// Show discovered modules
	fmt.Println()
	printer.PrintInfo(fmt.Sprintf("Discovered %d module(s):", len(w.result.Modules)))
	for _, m := range w.result.Modules {
		fmt.Printf("  - %s (%s)\n", m.Name, m.RelPath)
	}

	// Note about per-module dependency-check
	if len(w.result.SyncCandidates) > 0 {
		fmt.Println()
		printer.PrintFaint("Tip: Each module can have its own .sley.yaml with dependency-check")
		printer.PrintFaint("     configured for manifests in that module's directory.")
	}

	// Next steps
	fmt.Println()
	printer.PrintInfo("Next steps:")
	fmt.Println("  - Review .sley.yaml and adjust settings")
	fmt.Println("  - Run 'sley bump patch' to see available modules")
	fmt.Println("  - Run 'sley doctor' to verify setup")
}

// generateConfigYAMLWithWorkspace generates the YAML configuration content with workspace settings.
func generateConfigYAMLWithWorkspace(versionPath string, plugins []string, syncCandidates []discovery.SyncCandidate) ([]byte, error) {
	cfg := &config.Config{
		Path: versionPath,
	}

	// Create workspace config with discovery enabled
	enabled := true
	recursive := true
	moduleMaxDepth := 10
	cfg.Workspace = &config.WorkspaceConfig{
		Discovery: &config.DiscoveryConfig{
			Enabled:        &enabled,
			Recursive:      &recursive,
			ModuleMaxDepth: &moduleMaxDepth,
			Exclude:        []string{"testdata", "node_modules"},
		},
	}

	// Create plugins config based on selections
	pluginsCfg := &config.PluginConfig{}

	for _, name := range plugins {
		switch name {
		case "commit-parser":
			pluginsCfg.CommitParser = true
		case "tag-manager":
			pluginsCfg.TagManager = &config.TagManagerConfig{
				Enabled: true,
			}
		case "dependency-check":
			depCheck := &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
			}
			if len(syncCandidates) > 0 {
				depCheck.Files = make([]config.DependencyFileConfig, len(syncCandidates))
				for i, c := range syncCandidates {
					depCheck.Files[i] = config.DependencyFileConfig{
						Path:    c.Path,
						Format:  c.Format.String(),
						Field:   c.Field,
						Pattern: c.Pattern,
					}
				}
			}
			pluginsCfg.DependencyCheck = depCheck
		}
	}

	cfg.Plugins = pluginsCfg

	return marshalConfigWithWorkspaceComments(cfg, plugins)
}

// marshalConfigWithWorkspaceComments marshals config to YAML with helpful comments for workspace.
func marshalConfigWithWorkspaceComments(cfg *config.Config, plugins []string) ([]byte, error) {
	data, err := marshalToYAML(cfg)
	if err != nil {
		return nil, err
	}

	// Add header comments
	var result strings.Builder
	result.WriteString("# sley configuration file\n")
	result.WriteString("# Documentation: https://github.com/indaco/sley\n")
	result.WriteString("# Generated by 'sley discover'\n")
	result.WriteString("\n")
	result.WriteString("# This is a workspace configuration for a multi-module project.\n")
	result.WriteString("# Each module with a .version file is discovered automatically.\n")
	result.WriteString("# Modules can have their own .sley.yaml for module-specific settings.\n")
	result.WriteString("\n")

	if len(plugins) > 0 {
		result.WriteString("# Enabled plugins:\n")
		for _, name := range plugins {
			result.WriteString(fmt.Sprintf("#   - %s\n", name))
		}
		result.WriteString("\n")
	}

	result.Write(data)
	return []byte(result.String()), nil
}

// runExistingConfigWorkflow handles the case when .sley.yaml already exists.
func (w *Workflow) runExistingConfigWorkflow(ctx context.Context) (bool, error) {
	// Check for mismatches and offer to fix
	if w.result.HasMismatches() {
		return w.runMismatchWorkflow(ctx)
	}

	// Check if dependency-check could benefit from additional files
	if len(w.result.SyncCandidates) > 0 {
		return w.suggestAdditionalSyncFiles(ctx)
	}

	return false, nil
}

// runMismatchWorkflow offers to help resolve version mismatches.
func (w *Workflow) runMismatchWorkflow(_ context.Context) (bool, error) {
	fmt.Println()
	printer.PrintWarning(fmt.Sprintf("Found %d version mismatch(es).", len(w.result.Mismatches)))
	printer.PrintFaint("Consider enabling the dependency-check plugin with auto-sync to keep versions in sync.")
	printer.PrintFaint("Run 'sley bump auto --sync' to sync versions during bumps.")
	return false, nil
}

// suggestAdditionalSyncFiles suggests files that could be added to dependency-check.
func (w *Workflow) suggestAdditionalSyncFiles(_ context.Context) (bool, error) {
	// This is informational only - don't prompt in normal flow
	return false, nil
}

// runDependencyCheckSetup guides the user through dependency-check configuration
// and creates the config file with dependency-check plugin enabled.
func (w *Workflow) runDependencyCheckSetup(ctx context.Context) (bool, error) {
	fmt.Println()
	printer.PrintInfo("Discovered files that can sync with .version:")
	fmt.Println()

	// Show discovered files that can be synced
	for _, c := range w.result.SyncCandidates {
		fmt.Printf("  - %s (%s)\n", c.Path, c.Description)
	}
	fmt.Println()

	// Ask which files to include
	selected, err := w.selectSyncFiles()
	if err != nil {
		return false, err
	}

	if len(selected) == 0 {
		printer.PrintFaint("No files selected. Creating config with default plugins only.")
		return w.createConfigWithDefaults(ctx)
	}

	// Create config file with dependency-check plugin enabled
	return w.createConfigWithDependencyCheck(ctx, selected)
}

// selectSyncFiles prompts the user to select which files to sync.
func (w *Workflow) selectSyncFiles() ([]discovery.SyncCandidate, error) {
	if len(w.result.SyncCandidates) == 0 {
		return nil, nil
	}

	options, defaults := buildSyncFileOptions(w.result.SyncCandidates)

	selectedPaths, err := w.prompter.MultiSelect(
		"Select files to sync with .version:",
		"These files will be updated when you bump the version.",
		options,
		defaults,
	)
	if err != nil {
		return nil, err
	}

	return filterCandidatesByPaths(w.result.SyncCandidates, selectedPaths), nil
}

// buildSyncFileOptions creates huh options and default selections from sync candidates.
func buildSyncFileOptions(candidates []discovery.SyncCandidate) ([]huh.Option[string], []string) {
	options := make([]huh.Option[string], len(candidates))
	defaults := make([]string, len(candidates))

	for i, c := range candidates {
		label := fmt.Sprintf("%s - %s", c.Path, c.Description)
		options[i] = huh.NewOption(label, c.Path)
		defaults[i] = c.Path
	}

	return options, defaults
}

// filterCandidatesByPaths returns candidates whose paths are in the selected list.
func filterCandidatesByPaths(candidates []discovery.SyncCandidate, selectedPaths []string) []discovery.SyncCandidate {
	selected := make([]discovery.SyncCandidate, 0, len(selectedPaths))
	for _, path := range selectedPaths {
		for _, c := range candidates {
			if c.Path == path {
				selected = append(selected, c)
				break
			}
		}
	}
	return selected
}

// createConfigWithDefaults creates .sley.yaml with default plugins (commit-parser, tag-manager).
func (w *Workflow) createConfigWithDefaults(ctx context.Context) (bool, error) {
	// Initialize .version file if needed
	if err := w.ensureVersionFile(ctx); err != nil {
		return false, err
	}

	// Default plugins: commit-parser and tag-manager
	selectedPlugins := []string{"commit-parser", "tag-manager"}

	// Generate config
	configData, err := generateConfigYAML(defaultVersionPath(), selectedPlugins, nil)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(".sley.yaml", configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	w.printInitSuccess(selectedPlugins, nil)
	return true, nil
}

// createConfigWithDependencyCheck creates .sley.yaml with default plugins plus dependency-check.
func (w *Workflow) createConfigWithDependencyCheck(ctx context.Context, syncCandidates []discovery.SyncCandidate) (bool, error) {
	// Initialize .version file if needed
	if err := w.ensureVersionFile(ctx); err != nil {
		return false, err
	}

	// Plugins: default plugins + dependency-check
	selectedPlugins := []string{"commit-parser", "tag-manager", "dependency-check"}

	// Generate config with discovery
	configData, err := generateConfigYAML(defaultVersionPath(), selectedPlugins, syncCandidates)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(".sley.yaml", configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	w.printInitSuccess(selectedPlugins, syncCandidates)
	return true, nil
}

// ensureVersionFile creates the .version file if it doesn't exist.
func (w *Workflow) ensureVersionFile(_ context.Context) error {
	versionPath := defaultVersionPath()

	// Check if .version already exists
	if _, err := os.Stat(versionPath); err == nil {
		return nil // File exists, nothing to do
	}

	// Create .version file
	created, err := semver.InitializeVersionFileWithFeedback(versionPath)
	if err != nil {
		return fmt.Errorf("failed to initialize .version file: %w", err)
	}

	if created {
		version, err := semver.ReadVersion(versionPath)
		if err == nil {
			printer.PrintSuccess(fmt.Sprintf("Created %s with version %s", versionPath, version.String()))
		} else {
			printer.PrintSuccess(fmt.Sprintf("Created %s", versionPath))
		}
	}

	return nil
}

// defaultVersionPath returns the default .version file path.
func defaultVersionPath() string {
	return ".version"
}

// printInitSuccess prints success messages after initialization.
func (w *Workflow) printInitSuccess(plugins []string, syncCandidates []discovery.SyncCandidate) {
	fmt.Println()
	printer.PrintSuccess(fmt.Sprintf("Created .sley.yaml with %d plugin(s) enabled", len(plugins)))

	// Show enabled plugins
	fmt.Println()
	printer.PrintInfo("Enabled plugins:")
	for _, p := range plugins {
		fmt.Printf("  - %s\n", p)
	}

	// Show sync files if dependency-check is configured
	if len(syncCandidates) > 0 {
		fmt.Println()
		printer.PrintInfo("Configured sync files:")
		for _, c := range syncCandidates {
			fmt.Printf("  - %s\n", c.Path)
		}
	}

	// Next steps
	fmt.Println()
	printer.PrintInfo("Next steps:")
	fmt.Println("  - Review .sley.yaml and adjust settings")
	fmt.Println("  - Run 'sley bump patch' to increment version")
	fmt.Println("  - Run 'sley doctor' to verify setup")
}

// generateConfigYAML generates the YAML configuration content.
func generateConfigYAML(versionPath string, plugins []string, syncCandidates []discovery.SyncCandidate) ([]byte, error) {
	cfg := &config.Config{
		Path: versionPath,
	}

	// Create plugins config based on selections
	pluginsCfg := &config.PluginConfig{}

	for _, name := range plugins {
		switch name {
		case "commit-parser":
			pluginsCfg.CommitParser = true
		case "tag-manager":
			pluginsCfg.TagManager = &config.TagManagerConfig{
				Enabled: true,
			}
		case "dependency-check":
			depCheck := &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
			}
			if len(syncCandidates) > 0 {
				depCheck.Files = make([]config.DependencyFileConfig, len(syncCandidates))
				for i, c := range syncCandidates {
					depCheck.Files[i] = config.DependencyFileConfig{
						Path:    c.Path,
						Format:  c.Format.String(),
						Field:   c.Field,
						Pattern: c.Pattern,
					}
				}
			}
			pluginsCfg.DependencyCheck = depCheck
		}
	}

	cfg.Plugins = pluginsCfg

	return marshalConfigWithComments(cfg, plugins)
}

// marshalConfigWithComments marshals config to YAML with helpful comments.
func marshalConfigWithComments(cfg *config.Config, plugins []string) ([]byte, error) {
	// Import yaml package for marshaling
	data, err := marshalToYAML(cfg)
	if err != nil {
		return nil, err
	}

	// Add header comments
	var result strings.Builder
	result.WriteString("# sley configuration file\n")
	result.WriteString("# Documentation: https://github.com/indaco/sley\n")
	result.WriteString("# Generated by 'sley discover'\n")
	result.WriteString("\n")

	if len(plugins) > 0 {
		result.WriteString("# Enabled plugins:\n")
		for _, name := range plugins {
			result.WriteString(fmt.Sprintf("#   - %s\n", name))
		}
		result.WriteString("\n")
	}

	result.Write(data)
	return []byte(result.String()), nil
}

// marshalToYAML marshals a config to YAML bytes.
func marshalToYAML(cfg *config.Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// generateDependencyCheckConfig generates YAML configuration for dependency-check.
func (w *Workflow) generateDependencyCheckConfig(candidates []discovery.SyncCandidate) string {
	var sb strings.Builder

	sb.WriteString("plugins:\n")
	sb.WriteString("  dependency-check:\n")
	sb.WriteString("    enabled: true\n")
	sb.WriteString("    auto-sync: true\n")
	sb.WriteString("    files:\n")

	for _, c := range candidates {
		sb.WriteString(fmt.Sprintf("      - path: %s\n", c.Path))
		sb.WriteString(fmt.Sprintf("        format: %s\n", c.Format.String()))
		if c.Field != "" {
			sb.WriteString(fmt.Sprintf("        field: %s\n", c.Field))
		}
		if c.Pattern != "" {
			sb.WriteString(fmt.Sprintf("        pattern: '%s'\n", c.Pattern))
		}
	}

	return sb.String()
}

// GenerateDependencyCheckFileConfig generates a config.DependencyFileConfig from a SyncCandidate.
func GenerateDependencyCheckFileConfig(c discovery.SyncCandidate) config.DependencyFileConfig {
	return config.DependencyFileConfig{
		Path:    c.Path,
		Format:  c.Format.String(),
		Field:   c.Field,
		Pattern: c.Pattern,
	}
}

// GenerateDependencyCheckConfig generates the full dependency check config from candidates.
func GenerateDependencyCheckConfig(candidates []discovery.SyncCandidate) *config.DependencyCheckConfig {
	files := make([]config.DependencyFileConfig, len(candidates))
	for i, c := range candidates {
		files[i] = GenerateDependencyCheckFileConfig(c)
	}

	return &config.DependencyCheckConfig{
		Enabled:  true,
		AutoSync: true,
		Files:    files,
	}
}

// configExists checks if .sley.yaml exists in the current directory.
func configExists() bool {
	_, err := os.Stat(".sley.yaml")
	return err == nil
}

// SuggestDependencyCheckFromDiscovery analyzes discovery results and suggests
// dependency-check configuration if appropriate.
func SuggestDependencyCheckFromDiscovery(result *discovery.Result) *config.DependencyCheckConfig {
	if result == nil || len(result.SyncCandidates) == 0 {
		return nil
	}

	return GenerateDependencyCheckConfig(result.SyncCandidates)
}

// GetFieldForManifest returns the appropriate field path for a manifest file.
func GetFieldForManifest(filename string, format parser.Format) string {
	return parser.FieldForFormat(filename)
}
