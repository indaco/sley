package clix

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"

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
	return getOrInitVersionFile(cmd.String("path"), cmd.Bool("strict"))
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

// getOrInitVersionFile is the internal implementation that wraps errors for CLI display.
//
// Deprecated: Use GetOrInitVersionFile and handle errors at the CLI layer.
func getOrInitVersionFile(path string, strict bool) (bool, error) {
	created, err := GetOrInitVersionFile(path, strict)
	if err != nil {
		// Convert typed errors to CLI exits for backward compatibility
		var vfErr *apperrors.VersionFileNotFoundError
		if stderrors.As(err, &vfErr) {
			return false, cli.Exit(vfErr.Error(), 1)
		}
		return false, err
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
	options := &executionOptions{}
	for _, opt := range opts {
		opt(options)
	}
	// Check if --path flag is provided
	if cmd.IsSet("path") {
		path := cmd.String("path")
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: path,
		}, nil
	}

	// Check if .sley.yaml has an explicit path configured (not default ".version")
	// This takes precedence over multi-module detection
	if cfg != nil && cfg.Path != "" && cfg.Path != ".version" {
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: cfg.Path,
		}, nil
	}

	// Check for multi-module flags
	hasAll := cmd.Bool("all")
	hasModule := cmd.IsSet("module")

	// If explicit multi-module flags are set, detect modules
	if hasAll || hasModule {
		return getMultiModuleContext(ctx, cmd, cfg, options, true)
	}

	// Detect workspace context
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)
	detectedCtx, err := detector.DetectContext(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to detect workspace context: %w", err)
	}

	// Handle different detection modes
	switch detectedCtx.Mode {
	case workspace.SingleModule:
		// Single module found, use it
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: detectedCtx.Path,
		}, nil

	case workspace.MultiModule:
		// Multiple modules found, determine if we should prompt
		return getMultiModuleContext(ctx, cmd, cfg, options, false)

	case workspace.NoModules:
		// No modules found, fall back to default path
		path := cmd.String("path")
		if path == "" {
			path = cfg.Path
		}
		return &ExecutionContext{
			Mode: SingleModuleMode,
			Path: path,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected detection mode: %v", detectedCtx.Mode)
	}
}

// getMultiModuleContext handles multi-module execution context setup.
// It discovers modules, filters based on flags, and optionally shows TUI.
func getMultiModuleContext(ctx context.Context, cmd *cli.Command, cfg *config.Config, options *executionOptions, skipDetection bool) (*ExecutionContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)

	// Discover modules
	modules, err := detector.DiscoverModules(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules: %w", err)
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules found in workspace")
	}

	// Filter modules based on --module flag
	if cmd.IsSet("module") {
		moduleName := cmd.String("module")
		modules = filterModulesByName(modules, moduleName)
		if len(modules) == 0 {
			return nil, fmt.Errorf("module %q not found", moduleName)
		}
	}

	// Check if we should skip TUI prompt
	// Skip prompting for read-only commands (defaultToAll) or when explicit flags are set
	shouldPrompt := tui.IsInteractive() && !cmd.Bool("yes") && !cmd.Bool("non-interactive") && !cmd.Bool("all") && !options.defaultToAll

	if shouldPrompt {
		// Show TUI module selection
		prompter := tui.NewModulePrompt(modules)
		selection, err := prompter.PromptModuleSelection(modules)
		if err != nil {
			return nil, fmt.Errorf("module selection failed: %w", err)
		}

		if selection.Canceled {
			return nil, fmt.Errorf("operation canceled by user")
		}

		// Filter modules based on selection
		if !selection.All {
			modules = filterModulesBySelection(modules, selection.Modules)
		}

		return &ExecutionContext{
			Mode:      MultiModuleMode,
			Modules:   modules,
			Selection: selection,
		}, nil
	}

	// Non-interactive or --all flag: select all modules
	return &ExecutionContext{
		Mode:      MultiModuleMode,
		Modules:   modules,
		Selection: tui.AllModules(),
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
