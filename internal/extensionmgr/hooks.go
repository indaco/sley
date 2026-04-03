package extensionmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
	"github.com/indaco/sley/internal/printer"
)

// HookType represents the different hook points available in the extension system
type HookType string

const (
	// PreBumpHook is called before any version bump operation
	PreBumpHook HookType = "pre-bump"

	// PostBumpHook is called after a version bump completes successfully
	PostBumpHook HookType = "post-bump"

	// PreReleaseHook is called before pre-release changes are applied
	PreReleaseHook HookType = "pre-release"

	// ValidateHook is called to validate version changes
	ValidateHook HookType = "validate"
)

// ExtensionHookRunner manages the execution of extension hooks
type ExtensionHookRunner struct {
	Config         *config.Config
	Executor       Executor
	ManifestLoader ManifestLoader
}

// NewExtensionHookRunner creates a new ExtensionHookRunner
func NewExtensionHookRunner(cfg *config.Config) *ExtensionHookRunner {
	return &ExtensionHookRunner{
		Config:         cfg,
		Executor:       NewScriptExecutor(),
		ManifestLoader: &DefaultManifestLoader{},
	}
}

// RunHooks executes all enabled extensions for the specified hook point
func (r *ExtensionHookRunner) RunHooks(ctx context.Context, hookType HookType, input *HookInput) error {
	if r.Config == nil || len(r.Config.Extensions) == 0 {
		return nil
	}

	// Track if any hooks were executed
	hooksExecuted := 0

	for _, extCfg := range r.Config.Extensions {
		// Skip disabled extensions
		if !extCfg.Enabled {
			continue
		}

		// Load extension manifest
		manifest, err := r.ManifestLoader.Load(extCfg.Path)
		if err != nil {
			return fmt.Errorf("failed to load extension %q: %w", extCfg.Name, err)
		}

		// Check if extension supports this hook
		if !hasHook(manifest.Hooks, string(hookType)) {
			continue
		}

		// Resolve script path
		scriptPath := filepath.Join(extCfg.Path, manifest.Entry)

		// Create extension-specific input with config
		extInput := *input
		extInput.Config = extCfg.Config

		// Execute the hook
		ty := printer.Typography()
		fmt.Printf("Running extension %s (%s)... ", printer.Info(extCfg.Name), ty.Small(string(hookType)))

		output, err := r.Executor.Execute(ctx, scriptPath, &extInput)
		if err != nil {
			fmt.Println(ty.ErrorBadge("FAIL"))
			return fmt.Errorf("extension %q hook %q failed: %w", extCfg.Name, hookType, err)
		}

		fmt.Println(ty.SuccessBadge("OK"))

		if output.Message != "" {
			fmt.Printf("  %s\n", ty.Small(output.Message))
		}

		hooksExecuted++
	}

	return nil
}

// hasHook checks if a hook type is present in the hooks slice
func hasHook(hooks []string, hookType string) bool {
	return slices.Contains(hooks, hookType)
}

// LoadExtensionsForHook returns all enabled extensions that support the specified hook.
// It uses the runner's ManifestLoader for dependency injection.
func (r *ExtensionHookRunner) LoadExtensionsForHook(hookType HookType) ([]*extensions.ExtensionManifest, error) {
	if r.Config == nil || len(r.Config.Extensions) == 0 {
		return nil, nil
	}

	var result []*extensions.ExtensionManifest

	for _, extCfg := range r.Config.Extensions {
		if !extCfg.Enabled {
			continue
		}

		manifest, err := r.ManifestLoader.Load(extCfg.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to load extension %q: %w", extCfg.Name, err)
		}

		if hasHook(manifest.Hooks, string(hookType)) {
			result = append(result, manifest)
		}
	}

	return result, nil
}

// LoadExtensionsForHook is a package-level convenience that creates a default runner.
func LoadExtensionsForHook(cfg *config.Config, hookType HookType) ([]*extensions.ExtensionManifest, error) {
	runner := NewExtensionHookRunner(cfg)
	return runner.LoadExtensionsForHook(hookType)
}

// ValidateExtensionHook validates that a hook type is valid
func ValidateExtensionHook(hookType string) error {
	validHooks := []string{
		string(PreBumpHook),
		string(PostBumpHook),
		string(PreReleaseHook),
		string(ValidateHook),
	}

	if slices.Contains(validHooks, hookType) {
		return nil
	}

	return fmt.Errorf("invalid hook type %q, must be one of: %v", hookType, validHooks)
}

// ModuleInfo contains optional module context for monorepo support.
type ModuleInfo struct {
	Dir  string // Directory containing the .version file
	Name string // Module identifier
}

// RunPreBumpHooks is a convenience function to run pre-bump hooks
func RunPreBumpHooks(ctx context.Context, cfg *config.Config, version, previousVersion, bumpType string, moduleInfo *ModuleInfo) error {
	if cfg == nil {
		return nil
	}

	runner := NewExtensionHookRunner(cfg)

	// Get project root (current directory)
	projectRoot, err := os.Getwd()
	if err != nil {
		projectRoot = "."
	}

	input := HookInput{
		Hook:            string(PreBumpHook),
		Version:         version,
		PreviousVersion: previousVersion,
		BumpType:        bumpType,
		ProjectRoot:     projectRoot,
	}

	// Add module info if provided (monorepo support)
	if moduleInfo != nil {
		input.ModuleDir = moduleInfo.Dir
		input.ModuleName = moduleInfo.Name
	}

	return runner.RunHooks(ctx, PreBumpHook, &input)
}

// RunPostBumpHooks is a convenience function to run post-bump hooks
func RunPostBumpHooks(ctx context.Context, cfg *config.Config, version, previousVersion, bumpType string, prerelease, metadata *string, moduleInfo *ModuleInfo) error {
	if cfg == nil {
		return nil
	}

	runner := NewExtensionHookRunner(cfg)

	// Get project root (current directory)
	projectRoot, err := os.Getwd()
	if err != nil {
		projectRoot = "."
	}

	input := HookInput{
		Hook:            string(PostBumpHook),
		Version:         version,
		PreviousVersion: previousVersion,
		BumpType:        bumpType,
		Prerelease:      prerelease,
		Metadata:        metadata,
		ProjectRoot:     projectRoot,
	}

	// Add module info if provided (monorepo support)
	if moduleInfo != nil {
		input.ModuleDir = moduleInfo.Dir
		input.ModuleName = moduleInfo.Name
	}

	return runner.RunHooks(ctx, PostBumpHook, &input)
}
