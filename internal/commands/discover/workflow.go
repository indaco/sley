package discover

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/commands/initialize"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/tui"
)

// Workflow handles the interactive discovery workflow.
type Workflow struct {
	prompter           Prompter
	result             *discovery.Result
	rootDir            string
	cfg                *config.Config
	rootVersionCreated bool
}

// confirmAction prompts for confirmation in interactive mode.
// Returns true automatically in non-interactive environments or when prompter is nil.
func (w *Workflow) confirmAction(title, description string) (bool, error) {
	if w.prompter == nil || !tui.IsInteractive() {
		return true, nil
	}
	return w.prompter.Confirm(title, description)
}

// selectPlugins prompts for plugin selection in interactive mode,
// or returns defaults in non-interactive (CI/test) environments.
func (w *Workflow) selectPlugins() ([]string, error) {
	if !tui.IsInteractive() {
		return initialize.DefaultPluginNames(), nil
	}
	projectCtx := initialize.DetectProjectContext()
	selected, err := initialize.PromptPluginSelection(projectCtx.FormatDetectionSummary())
	if err != nil {
		return initialize.DefaultPluginNames(), nil //nolint:nilerr // Fall back to defaults on prompt error
	}
	if len(selected) == 0 {
		return initialize.DefaultPluginNames(), nil
	}
	return selected, nil
}

// NewWorkflow creates a new workflow handler.
func NewWorkflow(prompter Prompter, result *discovery.Result, rootDir string) *Workflow {
	return &Workflow{
		prompter: prompter,
		result:   result,
		rootDir:  rootDir,
	}
}

// NewWorkflowWithConfig creates a new workflow handler with config awareness.
func NewWorkflowWithConfig(prompter Prompter, result *discovery.Result, rootDir string, cfg *config.Config) *Workflow {
	return &Workflow{
		prompter: prompter,
		result:   result,
		rootDir:  rootDir,
		cfg:      cfg,
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
	printer.PrintFaint("No .sley.yaml configuration found.")

	// Check if we have useful suggestions
	if len(w.result.SyncCandidates) == 0 && len(w.result.Modules) == 0 {
		// No .version files found — check for monorepo workspace markers
		// (go.work, pnpm-workspace.yaml, package.json workspaces, Cargo.toml [workspace])
		if monoInfo, err := initialize.DetectMonorepo(); err == nil && monoInfo != nil {
			return w.runMonorepoInitWorkflow(ctx, monoInfo)
		}
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

// runMonorepoInitWorkflow handles the case when no .version files exist but
// a monorepo workspace marker (go.work, pnpm-workspace.yaml, etc.) is found.
// It shows the detected workspace info and offers to run the full workspace
// initialization flow.
func (w *Workflow) runMonorepoInitWorkflow(_ context.Context, monoInfo *initialize.MonorepoInfo) (bool, error) {
	ty := printer.Typography()
	moduleNames := make([]string, len(monoInfo.Modules))
	for i, m := range monoInfo.Modules {
		moduleNames[i] = m + "/"
	}
	fmt.Println(ty.Compose(
		printer.Info(fmt.Sprintf("Monorepo detected: %s workspace (%s) with %d module(s)",
			monoInfo.Type, monoInfo.MarkerFile, len(monoInfo.Modules))),
		ty.UL(moduleNames...),
	))

	fmt.Println()
	confirmed, err := w.prompter.Confirm(
		"Would you like to initialize as a workspace project?",
		"This will create .sley.yaml with workspace discovery, set tag prefix to {module_path}/v,\ncreate .version files in each module, and set versioning to independent.",
	)
	if err != nil {
		return false, err
	}

	if !confirmed {
		printer.PrintFaint("Run 'sley init --workspace' when ready.")
		return false, nil
	}

	// Prompt for plugin selection (same as sley init)
	detectionSummary := fmt.Sprintf("Detected: %s workspace (%s)", monoInfo.Type, monoInfo.MarkerFile)
	plugins, err := initialize.PromptPluginSelection(detectionSummary)
	if err != nil {
		return false, err
	}
	if len(plugins) == 0 {
		plugins = initialize.DefaultPluginNames()
	}

	// Confirm before writing
	proceed, err := w.prompter.Confirm(
		"Proceed with initialization?",
		fmt.Sprintf("This will create .sley.yaml and .version files in %d module directories.", len(monoInfo.Modules)),
	)
	if err != nil {
		return false, err
	}
	if !proceed {
		printer.PrintFaint("Initialization canceled.")
		return false, nil
	}

	// Ensure root .version exists
	if err := w.ensureVersionFile(context.Background()); err != nil {
		return false, err
	}

	// Build DiscoveredModule list from detected monorepo modules
	var modules []initialize.DiscoveredModule
	for _, m := range monoInfo.Modules {
		modules = append(modules, initialize.DiscoveredModule{
			Name:    m,
			RelPath: m + "/.version",
		})
	}
	configData, err := initialize.GenerateWorkspaceConfigWithMonorepo(plugins, modules, monoInfo)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}
	if err := os.WriteFile(".sley.yaml", configData, config.ConfigFilePerm); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	// Create .version files in module directories
	createdFiles := initialize.CreateMonorepoVersionFiles(monoInfo)

	// Include root .version if it was just created
	if w.rootVersionCreated {
		createdFiles = append([]string{".version"}, createdFiles...)
	}

	// Print summary
	blocks := []string{}
	if len(createdFiles) > 0 {
		faintFiles := make([]string, len(createdFiles))
		for i, f := range createdFiles {
			faintFiles[i] = ty.Small(f)
		}
		blocks = append(blocks, ty.Section(ty.H4("Created version files"), ty.UL(faintFiles...)))
	}
	blocks = append(blocks,
		printer.Success(fmt.Sprintf("Created .sley.yaml with workspace configuration and %d plugin(s)", len(plugins))),
		ty.Section(ty.H4("Enabled plugins"), ty.UL(plugins...)),
		ty.Section(ty.H4("Applied monorepo defaults"), ty.KVGroup([][2]string{
			{"Versioning", "independent"},
			{"Tag prefix", "{module_path}/v"},
			{"Modules", fmt.Sprintf("%d", len(monoInfo.Modules))},
		})),
		ty.Section(ty.H4("Next steps"), ty.UL(
			"Review .sley.yaml and adjust settings",
			"Run 'sley bump patch --all' to bump all modules",
			"Run 'sley doctor' to verify setup",
		)),
	)
	fmt.Println(ty.Compose(blocks...))

	return true, nil
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
	ty := printer.Typography()
	modItems := make([]string, len(w.result.Modules))
	for i, m := range w.result.Modules {
		modItems[i] = fmt.Sprintf("%s (%s)", m.Name, m.RelPath)
	}
	fmt.Println(ty.Compose(
		printer.Info(fmt.Sprintf("Found %d modules - this appears to be a monorepo.", len(w.result.Modules))),
		ty.Section(ty.H4("Discovered modules"), ty.UL(modItems...)),
	))

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

	// Prompt for plugin selection (same as init --workspace)
	selectedPlugins, err := w.selectPlugins()
	if err != nil {
		return false, err
	}

	// Confirm before writing
	proceed, err := w.confirmAction(
		"Proceed with initialization?",
		fmt.Sprintf("This will create .sley.yaml with workspace discovery and %d plugin(s).", len(selectedPlugins)),
	)
	if err != nil {
		return false, err
	}
	if !proceed {
		printer.PrintFaint("Initialization canceled.")
		return false, nil
	}

	// Generate config with workspace discovery enabled
	configData, err := generateConfigYAMLWithWorkspace(defaultVersionPath(), selectedPlugins, w.result.Modules, w.result.SyncCandidates)
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
	ty := printer.Typography()

	moduleItems := make([]string, len(w.result.Modules))
	for i, m := range w.result.Modules {
		moduleItems[i] = fmt.Sprintf("%s (%s)", m.Name, m.RelPath)
	}

	var blocks []string

	if w.rootVersionCreated {
		blocks = append(blocks, ty.Section(ty.H4("Created version files"), ty.UL(ty.Small(".version"))))
	}

	blocks = append(blocks,
		printer.Success(fmt.Sprintf("Created .sley.yaml with workspace configuration and %d plugin(s)", len(plugins))),
		ty.Section(ty.H4("Enabled plugins"), ty.UL(plugins...)),
		ty.Section(ty.H4("Workspace configuration"), ty.UL("Auto-discovery enabled", "Each module manages its own .version file")),
		ty.Section(ty.H4(fmt.Sprintf("Discovered %d module(s)", len(w.result.Modules))), ty.UL(moduleItems...)),
	)

	if len(w.result.SyncCandidates) > 0 {
		blocks = append(blocks, ty.Small(
			"Tip: Each module can have its own .sley.yaml with dependency-check\n"+
				"     configured for manifests in that module's directory."))
	}

	blocks = append(blocks, ty.Section(ty.H4("Next steps"), ty.UL(
		"Review .sley.yaml and adjust settings",
		"Run 'sley bump patch' to see available modules",
		"Run 'sley doctor' to verify setup",
	)))

	fmt.Println(ty.Compose(blocks...))
}

// generateConfigYAMLWithWorkspace generates workspace config by converting
// discovery.Module to initialize.DiscoveredModule and delegating to the
// shared generator in the initialize package.
func generateConfigYAMLWithWorkspace(_ string, plugins []string, modules []discovery.Module, _ []discovery.SyncCandidate) ([]byte, error) {
	initModules := make([]initialize.DiscoveredModule, len(modules))
	for i, m := range modules {
		initModules[i] = initialize.DiscoveredModule{
			Name:    m.Name,
			RelPath: m.RelPath,
		}
	}
	return initialize.GenerateWorkspaceConfigWithComments(plugins, initModules)
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
	ty := printer.Typography()
	if w.cfg != nil && w.cfg.Workspace != nil && w.cfg.Workspace.IsIndependentVersioning() {
		fmt.Println(ty.Compose(
			printer.Info(fmt.Sprintf("Version summary: %d module(s) at different versions (independent versioning).", len(w.result.Mismatches))),
			ty.Small("Each module manages its own version independently."),
		))
	} else {
		fmt.Println(ty.Compose(
			printer.Warning(fmt.Sprintf("Found %d version mismatch(es).", len(w.result.Mismatches))),
			ty.Small("Consider enabling the dependency-check plugin with auto-sync to keep versions in sync."),
			ty.Code("sley bump auto --sync"),
		))
	}
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
	ty := printer.Typography()
	syncItems := make([]string, len(w.result.SyncCandidates))
	for i, c := range w.result.SyncCandidates {
		syncItems[i] = fmt.Sprintf("%s (%s)", c.Path, c.Description)
	}
	fmt.Println(ty.Section(ty.H4("Discovered files that can sync with .version"), ty.UL(syncItems...)))

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

// createConfigWithDefaults creates .sley.yaml after plugin selection.
func (w *Workflow) createConfigWithDefaults(ctx context.Context) (bool, error) {
	// Initialize .version file if needed
	if err := w.ensureVersionFile(ctx); err != nil {
		return false, err
	}

	// Prompt for plugin selection (same as init --workspace)
	selectedPlugins, err := w.selectPlugins()
	if err != nil {
		return false, err
	}

	// Confirm before writing
	proceed, err := w.confirmAction(
		"Proceed with initialization?",
		fmt.Sprintf("This will create .sley.yaml with %d plugin(s).", len(selectedPlugins)),
	)
	if err != nil {
		return false, err
	}
	if !proceed {
		printer.PrintFaint("Initialization canceled.")
		return false, nil
	}

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

	// Prompt for plugin selection (same as init --workspace)
	selectedPlugins, err := w.selectPlugins()
	if err != nil {
		return false, err
	}

	// Ensure dependency-check is included since we have sync candidates
	if !containsPlugin(selectedPlugins, "dependency-check") {
		selectedPlugins = append(selectedPlugins, "dependency-check")
	}

	// Confirm before writing
	proceed, err := w.confirmAction(
		"Proceed with initialization?",
		fmt.Sprintf("This will create .sley.yaml with %d plugin(s) and %d sync file(s).", len(selectedPlugins), len(syncCandidates)),
	)
	if err != nil {
		return false, err
	}
	if !proceed {
		printer.PrintFaint("Initialization canceled.")
		return false, nil
	}

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

// containsPlugin checks if a plugin name exists in the list.
func containsPlugin(plugins []string, name string) bool {
	return slices.Contains(plugins, name)
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
		w.rootVersionCreated = true
	}

	return nil
}

// defaultVersionPath returns the default .version file path.
func defaultVersionPath() string {
	return ".version"
}

// printInitSuccess prints success messages after initialization.
func (w *Workflow) printInitSuccess(plugins []string, syncCandidates []discovery.SyncCandidate) {
	ty := printer.Typography()

	var blocks []string

	if w.rootVersionCreated {
		blocks = append(blocks, ty.Section(ty.H4("Created version files"), ty.UL(ty.Small(".version"))))
	}

	blocks = append(blocks,
		printer.Success(fmt.Sprintf("Created .sley.yaml with %d plugin(s) enabled", len(plugins))),
		ty.Section(ty.H4("Enabled plugins"), ty.UL(plugins...)),
	)

	if len(syncCandidates) > 0 {
		syncPaths := make([]string, len(syncCandidates))
		for i, c := range syncCandidates {
			syncPaths[i] = c.Path
		}
		blocks = append(blocks, ty.Section(ty.H4("Configured sync files"), ty.UL(syncPaths...)))
	}

	blocks = append(blocks, ty.Section(ty.H4("Next steps"), ty.UL(
		"Review .sley.yaml and adjust settings",
		"Run 'sley bump patch' to increment version",
		"Run 'sley doctor' to verify setup",
	)))

	fmt.Println(ty.Compose(blocks...))
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
			fmt.Fprintf(&result, "#   - %s\n", name)
		}
		result.WriteString("\n")
	}

	result.Write(data)
	return []byte(result.String()), nil
}

// marshalToYAML marshals a config to YAML bytes.
func marshalToYAML(cfg *config.Config) ([]byte, error) {
	return yaml.MarshalWithOptions(cfg, yaml.Indent(2), yaml.IndentSequence(true))
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
		fmt.Fprintf(&sb, "      - path: %s\n", c.Path)
		fmt.Fprintf(&sb, "        format: %s\n", c.Format.String())
		if c.Field != "" {
			fmt.Fprintf(&sb, "        field: %s\n", c.Field)
		}
		if c.Pattern != "" {
			fmt.Fprintf(&sb, "        pattern: '%s'\n", c.Pattern)
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
