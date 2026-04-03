package bump

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

// runMultiModuleBump executes a bump operation on multiple modules.
func runMultiModuleBump(
	ctx context.Context,
	cmd *cli.Command,
	cfg *config.Config,
	execCtx *clix.ExecutionContext,
	registry *plugins.PluginRegistry,
	deps *bumpDeps,
	bumpType operations.BumpType,
	preRelease, metadata string,
	preserveMetadata bool,
) error {
	fs := core.NewOSFileSystem()
	bumperFn := func() semver.VersionBumper { return semver.NewDefaultBumper() }
	if deps != nil && deps.newBumper != nil {
		bumperFn = deps.newBumper
	}
	bumper := bumperFn()
	operation := operations.NewBumpOperation(fs, bumper, bumpType, preRelease, metadata, preserveMetadata)

	skipHooks := cmd.Bool("skip-hooks")

	// Pre-bump phase: run extension hooks and validations per module before any writes.
	if err := runPreBumpPhase(ctx, cfg, registry, operation, execCtx.Modules, string(bumpType), skipHooks); err != nil {
		return err
	}

	// Create executor with options from flags
	parallel := cmd.Bool("parallel")
	failFast := cmd.Bool("fail-fast") && !cmd.Bool("continue-on-error")

	executor := workspace.NewExecutor(
		workspace.WithParallel(parallel),
		workspace.WithFailFast(failFast),
	)

	// Execute the operation on all modules (write versions)
	results, err := executor.Run(ctx, execCtx.Modules, operation)
	if err != nil && failFast {
		// In fail-fast mode, we may have partial results
		// Fall through to display what we have
		_ = err
	}

	// Format and display results
	format := cmd.String("format")
	quiet := cmd.Bool("quiet")

	formatter := workspace.GetFormatter(format, fmt.Sprintf("Bump %s", bumpType))

	if quiet {
		// In quiet mode, just show summary
		printQuietSummary(results)
	} else {
		fmt.Println(formatter.FormatResults(results))
	}

	// Return error if any failures occurred during version bumps
	if workspace.HasErrors(results) {
		return fmt.Errorf("%d module(s) failed", workspace.ErrorCount(results))
	}

	// Run post-bump actions sequentially per module.
	// This loop is ALWAYS sequential regardless of --parallel, because
	// post-bump actions mutate shared plugin state (e.g. tag prefix).
	return runPerModulePostBump(ctx, results, registry, cfg, string(bumpType), skipHooks)
}

// runPerModulePostBump executes post-bump actions, extension hooks, and commit/tag
// sequentially for each successfully bumped module.
func runPerModulePostBump(ctx context.Context, results []workspace.ExecutionResult, registry *plugins.PluginRegistry, cfg *config.Config, bumpTypeStr string, skipHooks bool) error {
	var postBumpErrors []error
	for _, result := range results {
		if !result.Success || result.Module == nil {
			continue
		}
		if err := postBumpForModule(ctx, result, registry, cfg, bumpTypeStr, skipHooks); err != nil {
			postBumpErrors = append(postBumpErrors, err)
		}
	}
	if len(postBumpErrors) > 0 {
		return fmt.Errorf("%d module(s) had post-bump errors: %w", len(postBumpErrors), postBumpErrors[0])
	}
	return nil
}

// postBumpForModule runs post-bump actions, extension hooks, and commit/tag for a single module.
func postBumpForModule(ctx context.Context, result workspace.ExecutionResult, registry *plugins.PluginRegistry, cfg *config.Config, bumpTypeStr string, skipHooks bool) error {
	newVer, err := semver.ParseVersion(result.NewVersion)
	if err != nil {
		return fmt.Errorf("module %s: failed to parse new version %q: %w", result.Module.Name, result.NewVersion, err)
	}
	oldVer, err := semver.ParseVersion(result.OldVersion)
	if err != nil {
		return fmt.Errorf("module %s: failed to parse old version %q: %w", result.Module.Name, result.OldVersion, err)
	}

	modulePath := deriveModulePath(result.Module.RelPath)
	moduleName := resolveModuleName(result.Module.Name)
	effectiveCfg := resolveModuleConfig(cfg, modulePath, result.Module.Dir)

	// Post-bump actions (dep-sync, changelog, audit-log)
	if err := executePostBumpActions(registry, newVer, oldVer, bumpTypeStr, result.Module.Path, moduleName, modulePath); err != nil {
		return fmt.Errorf("module %s: post-bump actions: %w", result.Module.Name, err)
	}

	// Post-bump extension hooks
	if err := runPostBumpExtensionHooks(ctx, effectiveCfg, result.Module.Path, oldVer.String(), bumpTypeStr, skipHooks); err != nil {
		return fmt.Errorf("module %s: post-bump hooks: %w", result.Module.Name, err)
	}

	// Commit and tag
	if err := commitAndTagAfterBump(registry, newVer, bumpTypeStr, result.Module.Path, effectiveCfg); err != nil {
		return fmt.Errorf("module %s: commit/tag: %w", result.Module.Name, err)
	}
	return nil
}

// runPreBumpPhase runs extension hooks and validations for all modules before any
// version writes. This ensures all modules pass validation before committing to bumps.
func runPreBumpPhase(ctx context.Context, cfg *config.Config, registry *plugins.PluginRegistry, op *operations.BumpOperation, modules []*workspace.Module, bumpTypeStr string, skipHooks bool) error {
	for _, mod := range modules {
		effectiveCfg := resolveModuleConfig(cfg, deriveModulePath(mod.RelPath), mod.Dir)

		// Preview to get new/old versions for validation
		result, err := op.Preview(ctx, mod.Path)
		if err != nil {
			return fmt.Errorf("module %s: preview failed: %w", mod.Name, err)
		}

		// Pre-bump extension hooks (may modify .version file)
		if err := runPreBumpExtensionHooks(ctx, effectiveCfg, mod.Path, result.NewVersion.String(), result.PreviousVersion.String(), bumpTypeStr, skipHooks); err != nil {
			return fmt.Errorf("module %s: pre-bump hooks: %w", mod.Name, err)
		}

		// Pre-bump validations (release gate, version policy, dep consistency, tag availability)
		if err := executePreBumpValidations(registry, result.NewVersion, result.PreviousVersion, bumpTypeStr); err != nil {
			return fmt.Errorf("module %s: validation failed: %w", mod.Name, err)
		}
	}
	return nil
}

// resolveModuleName returns a display name for the module.
// For root modules (name "."), it uses the current working directory basename.
func resolveModuleName(name string) string {
	if name != "." {
		return name
	}
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Base(cwd)
	}
	return name
}

// deriveModulePath extracts the module directory path from a .version file's relative path.
// Returns "" for root module (where RelPath is ".version" or the dir is ".").
func deriveModulePath(relPath string) string {
	dir := filepath.Dir(relPath)
	if dir == "." {
		return ""
	}
	return dir
}

// resolveModuleConfig loads per-module config and merges with root.
// For root modules (empty modulePath), returns root config as-is.
func resolveModuleConfig(rootCfg *config.Config, modulePath, moduleDir string) *config.Config {
	if modulePath == "" {
		return rootCfg
	}
	moduleCfg, err := config.LoadConfigFromDir(moduleDir)
	if err != nil || moduleCfg == nil {
		return rootCfg
	}
	return config.MergeConfig(rootCfg, moduleCfg)
}

// getFirstSuccessfulVersion returns the new version from the first successful result.
// Returns empty string if no successful results exist.
func getFirstSuccessfulVersion(results []workspace.ExecutionResult) string {
	for _, r := range results {
		if r.Success && r.NewVersion != "" {
			return r.NewVersion
		}
	}
	return ""
}

// printQuietSummary prints a minimal summary of results.
func printQuietSummary(results []workspace.ExecutionResult) {
	success := workspace.SuccessCount(results)
	errors := workspace.ErrorCount(results)
	if errors > 0 {
		printer.PrintWarning(fmt.Sprintf("Completed: %d succeeded, %d failed", success, errors))
	} else {
		printer.PrintSuccess(fmt.Sprintf("Success: %d module(s) bumped", success))
	}
}
