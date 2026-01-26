package discover

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/printer"
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

	// Suggest dependency-check configuration if we have sync candidates
	if len(w.result.SyncCandidates) > 0 {
		return w.runDependencyCheckSetup(ctx)
	}

	printer.PrintInfo("Run 'sley init' to complete setup with plugin selection.")
	return true, nil
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
func (w *Workflow) runMismatchWorkflow(ctx context.Context) (bool, error) {
	fmt.Println()
	printer.PrintWarning(fmt.Sprintf("Found %d version mismatch(es).", len(w.result.Mismatches)))
	printer.PrintFaint("Consider enabling the dependency-check plugin with auto-sync to keep versions in sync.")
	printer.PrintFaint("Run 'sley bump auto --sync' to sync versions during bumps.")
	return false, nil
}

// suggestAdditionalSyncFiles suggests files that could be added to dependency-check.
func (w *Workflow) suggestAdditionalSyncFiles(ctx context.Context) (bool, error) {
	// This is informational only - don't prompt in normal flow
	return false, nil
}

// runDependencyCheckSetup guides the user through dependency-check configuration.
func (w *Workflow) runDependencyCheckSetup(ctx context.Context) (bool, error) {
	fmt.Println()
	printer.PrintInfo("Suggested dependency-check configuration:")
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
		printer.PrintFaint("No files selected. You can configure dependency-check later in .sley.yaml")
		return false, nil
	}

	// Generate and show the configuration
	configSnippet := w.generateDependencyCheckConfig(selected)
	fmt.Println()
	printer.PrintInfo("Add this to your .sley.yaml:")
	fmt.Println()
	printer.PrintFaint("```yaml")
	fmt.Print(configSnippet)
	printer.PrintFaint("```")
	fmt.Println()

	return true, nil
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
