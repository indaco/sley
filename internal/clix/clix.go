package clix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/apperrors"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/tui"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

var FromCommandFn = fromCommand

// fromCommand extracts the --path and --strict flags from a cli.Command,
// and passes them to GetOrInitVersionFile.
func fromCommand(cmd *cli.Command) (bool, error) {
	return GetOrInitVersionFile(cmd.String("path"), cmd.Bool("strict"))
}

// GetOrInitVersionFile initializes the version file at the given path
// or checks for its existence based on the strict flag.
// It returns true if the file was created, false if it already existed.
// Returns a typed error (*apperrors.VersionFileNotFoundError) instead of cli.Exit.
func GetOrInitVersionFile(path string, strict bool) (bool, error) {
	if strict {
		if _, err := os.Stat(path); err != nil {
			return false, &apperrors.VersionFileNotFoundError{Path: path}
		}
		return false, nil
	}

	created, err := semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return false, err
	}
	if created {
		fmt.Printf("Auto-initialized %s with default version\n", path)
	}
	return created, nil
}

// ExecutionMode indicates whether to operate on a single module or multiple modules.
type ExecutionMode int

const (
	// SingleModuleMode indicates operating on a single .version file.
	SingleModuleMode ExecutionMode = iota

	// MultiModuleMode indicates operating on multiple .version files.
	MultiModuleMode
)

// String returns a string representation of the execution mode.
func (m ExecutionMode) String() string {
	switch m {
	case SingleModuleMode:
		return "SingleModule"
	case MultiModuleMode:
		return "MultiModule"
	default:
		return fmt.Sprintf("Unknown(%d)", m)
	}
}

// ExecutionContext encapsulates the execution context for a command.
// It determines whether to operate on a single module or multiple modules
// based on flags, configuration, and user interaction.
type ExecutionContext struct {
	// Mode indicates whether this is single or multi-module execution.
	Mode ExecutionMode

	// Path is the .version file path for single-module mode.
	// Empty for multi-module mode.
	Path string

	// Modules contains the selected modules for multi-module mode.
	// Nil for single-module mode.
	Modules []*workspace.Module

	// Selection contains the TUI selection result for multi-module mode.
	// Used to determine if user selected all or specific modules.
	Selection tui.Selection
}

// IsSingleModule returns true if this is single-module execution.
func (ec *ExecutionContext) IsSingleModule() bool {
	return ec.Mode == SingleModuleMode
}

// IsMultiModule returns true if this is multi-module execution.
func (ec *ExecutionContext) IsMultiModule() bool {
	return ec.Mode == MultiModuleMode
}

// ExecutionOption configures execution context behavior.
type ExecutionOption func(*executionOptions)

type executionOptions struct {
	defaultToAll bool // Skip TUI prompt and default to all modules (for read-only commands)
}

// WithDefaultAll configures the execution context to default to all modules
// without showing a TUI prompt. Use this for read-only commands like "show"
// and "doctor" where prompting adds friction without value.
func WithDefaultAll() ExecutionOption {
	return func(o *executionOptions) {
		o.defaultToAll = true
	}
}

// GetExecutionContext determines the execution context for a command.
// It follows this logic:
//  1. If --path flag provided -> single-module mode
//  2. If .sley.yaml has explicit path (not default) -> single-module mode
//  3. If --all or --module flags -> multi-module mode (skip TUI)
//  4. Detect context using workspace.Detector
//  5. If MultiModule detected and interactive -> show TUI prompt
//  6. If MultiModule detected and non-interactive (CI or --yes) -> auto-select all
//
// The context parameter is used for cancellation and timeouts.
// The cmd parameter provides access to CLI flags.
// The cfg parameter provides workspace configuration.
// Optional ExecutionOption can be passed to modify behavior (e.g., WithDefaultAll).
func GetExecutionContext(ctx context.Context, cmd *cli.Command, cfg *config.Config, opts ...ExecutionOption) (*ExecutionContext, error) {
	options := applyExecutionOptions(opts)

	// Check for explicit single-module mode
	if execCtx := getSingleModuleFromFlags(cmd, cfg); execCtx != nil {
		return execCtx, nil
	}

	// Check for explicit multi-module flags
	if hasMultiModuleFlags(cmd) {
		return getMultiModuleContext(ctx, cmd, cfg, options, true)
	}

	// Auto-detect workspace context
	return detectAndBuildContext(ctx, cmd, cfg, options)
}

// applyExecutionOptions applies option functions to create execution options.
func applyExecutionOptions(opts []ExecutionOption) *executionOptions {
	options := &executionOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// getSingleModuleFromFlags checks if explicit single-module mode is requested.
func getSingleModuleFromFlags(cmd *cli.Command, cfg *config.Config) *ExecutionContext {
	if cmd.IsSet("path") {
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: cmd.String("path"),
		}
	}

	if cfg != nil && cfg.Path != "" && cfg.Path != ".version" {
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: cfg.Path,
		}
	}

	return nil
}

// hasMultiModuleFlags checks if any multi-module flags are set.
func hasMultiModuleFlags(cmd *cli.Command) bool {
	return cmd.Bool("all") ||
		cmd.IsSet("module") ||
		len(cmd.StringSlice("modules")) > 0 ||
		cmd.IsSet("pattern")
}

// detectAndBuildContext auto-detects workspace mode and builds context.
func detectAndBuildContext(ctx context.Context, cmd *cli.Command, cfg *config.Config, options *executionOptions) (*ExecutionContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)
	detectedCtx, err := detector.DetectContext(ctx, cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to detect workspace context: %w", err)
	}

	return handleDetectedMode(ctx, cmd, cfg, options, detectedCtx)
}

// handleDetectedMode builds execution context based on detected workspace mode.
func handleDetectedMode(ctx context.Context, cmd *cli.Command, cfg *config.Config, options *executionOptions, detectedCtx *workspace.Context) (*ExecutionContext, error) {
	switch detectedCtx.Mode {
	case workspace.SingleModule:
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: detectedCtx.Path,
		}, nil

	case workspace.MultiModule:
		return getMultiModuleContext(ctx, cmd, cfg, options, false)

	case workspace.NoModules:
		return buildNoModulesContext(cmd, cfg), nil

	default:
		return nil, fmt.Errorf("unexpected detection mode: %v", detectedCtx.Mode)
	}
}

// buildNoModulesContext creates context when no modules are found.
func buildNoModulesContext(cmd *cli.Command, cfg *config.Config) *ExecutionContext {
	path := cmd.String("path")
	if path == "" {
		path = cfg.Path
	}
	return &ExecutionContext{
		Mode: SingleModuleMode,
		Path: path,
	}
}

// getMultiModuleContext handles multi-module execution context setup.
// It discovers modules, filters based on flags, and optionally shows TUI.
func getMultiModuleContext(ctx context.Context, cmd *cli.Command, cfg *config.Config, options *executionOptions, _ bool) (*ExecutionContext, error) {
	modules, err := discoverModules(ctx, cfg)
	if err != nil {
		return nil, err
	}

	modules, err = applyModuleFilters(cmd, modules)
	if err != nil {
		return nil, err
	}

	return buildMultiModuleContext(cmd, options, modules)
}

// discoverModules finds all modules in the workspace.
func discoverModules(ctx context.Context, cfg *config.Config) ([]*workspace.Module, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)

	modules, err := detector.DiscoverModules(ctx, cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules: %w", err)
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules found in workspace")
	}

	return modules, nil
}

// applyModuleFilters applies --module, --modules, and --pattern filters.
func applyModuleFilters(cmd *cli.Command, modules []*workspace.Module) ([]*workspace.Module, error) {
	var err error

	if cmd.IsSet("module") {
		modules = filterModulesByName(modules, cmd.String("module"))
		if len(modules) == 0 {
			return nil, fmt.Errorf("module %q not found", cmd.String("module"))
		}
	}

	if names := cmd.StringSlice("modules"); len(names) > 0 {
		modules = filterModulesByNames(modules, names)
		if len(modules) == 0 {
			return nil, fmt.Errorf("no modules found matching names: %v", names)
		}
	}

	if cmd.IsSet("pattern") {
		pattern := cmd.String("pattern")
		modules, err = filterModulesByPattern(modules, pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		if len(modules) == 0 {
			return nil, fmt.Errorf("no modules found matching pattern: %s", pattern)
		}
	}

	return modules, nil
}

// buildMultiModuleContext creates the execution context, optionally showing TUI.
func buildMultiModuleContext(cmd *cli.Command, options *executionOptions, modules []*workspace.Module) (*ExecutionContext, error) {
	if !shouldShowTUIPrompt(cmd, options) {
		return &ExecutionContext{
			Mode:      MultiModuleMode,
			Modules:   modules,
			Selection: tui.AllModules(),
		}, nil
	}

	return promptForModuleSelection(modules)
}

// shouldShowTUIPrompt determines if interactive module selection should be shown.
func shouldShowTUIPrompt(cmd *cli.Command, options *executionOptions) bool {
	if !tui.IsInteractive() {
		return false
	}
	if cmd.Bool("yes") || cmd.Bool("non-interactive") || cmd.Bool("all") {
		return false
	}
	return !options.defaultToAll
}

// promptForModuleSelection shows the TUI and returns the selected modules.
func promptForModuleSelection(modules []*workspace.Module) (*ExecutionContext, error) {
	prompter := tui.NewModulePrompt(modules)
	selection, err := prompter.PromptModuleSelection(modules)
	if err != nil {
		return nil, fmt.Errorf("module selection failed: %w", err)
	}

	if selection.Canceled {
		return nil, fmt.Errorf("operation canceled by user")
	}

	if !selection.All {
		modules = filterModulesBySelection(modules, selection.Modules)
	}

	return &ExecutionContext{
		Mode:      MultiModuleMode,
		Modules:   modules,
		Selection: selection,
	}, nil
}

// filterModulesByName filters modules to include all with the given name.
// When multiple modules share the same name (common in monorepos), all are returned.
func filterModulesByName(modules []*workspace.Module, name string) []*workspace.Module {
	var filtered []*workspace.Module
	for _, mod := range modules {
		if mod.Name == name {
			filtered = append(filtered, mod)
		}
	}
	return filtered
}

// filterModulesBySelection filters modules based on user selection.
func filterModulesBySelection(modules []*workspace.Module, selected []string) []*workspace.Module {
	selectedMap := make(map[string]bool)
	for _, name := range selected {
		selectedMap[name] = true
	}

	var filtered []*workspace.Module
	for _, mod := range modules {
		if selectedMap[mod.Name] {
			filtered = append(filtered, mod)
		}
	}
	return filtered
}

// filterModulesByNames filters modules to include those matching any of the given names.
// Names can be provided as a slice (from --modules flag).
func filterModulesByNames(modules []*workspace.Module, names []string) []*workspace.Module {
	if len(names) == 0 {
		return modules
	}

	// Build a set of names for O(1) lookup
	nameSet := make(map[string]bool)
	for _, name := range names {
		// Handle comma-separated values within a single argument
		for n := range strings.SplitSeq(name, ",") {
			n = strings.TrimSpace(n)
			if n != "" {
				nameSet[n] = true
			}
		}
	}

	var filtered []*workspace.Module
	for _, mod := range modules {
		if nameSet[mod.Name] {
			filtered = append(filtered, mod)
		}
	}
	return filtered
}

// filterModulesByPattern filters modules whose paths match the given glob pattern.
// The pattern is matched against the module's directory path relative to the workspace root.
// Examples: "services/*", "packages/shared", "**/api"
func filterModulesByPattern(modules []*workspace.Module, pattern string) ([]*workspace.Module, error) {
	if pattern == "" {
		return modules, nil
	}

	var filtered []*workspace.Module
	for _, mod := range modules {
		// Get the directory containing the .version file
		dir := filepath.Dir(mod.Path)

		// Try matching against the directory path
		matched, err := filepath.Match(pattern, dir)
		if err != nil {
			return nil, err
		}

		// Also try matching against just the module name for simple patterns
		if !matched {
			matched, _ = filepath.Match(pattern, mod.Name)
		}

		// Also try matching against the full path
		if !matched {
			matched, _ = filepath.Match(pattern, mod.Path)
		}

		if matched {
			filtered = append(filtered, mod)
		}
	}
	return filtered, nil
}
