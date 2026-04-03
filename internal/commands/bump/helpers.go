package bump

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
)

// moduleInfoFromPath derives module info from a .version file path.
// For single-module projects, this provides the directory context for extensions.
func moduleInfoFromPath(versionPath string) *extensionmgr.ModuleInfo {
	dir := filepath.Dir(versionPath)
	name := filepath.Base(dir)
	// If the .version is in the current directory, use "." as name
	if dir == "." || name == "." {
		return nil
	}
	return &extensionmgr.ModuleInfo{
		Dir:  dir,
		Name: name,
	}
}

// runPreBumpExtensionHooks runs pre-bump extension hooks if not skipped.
func runPreBumpExtensionHooks(ctx context.Context, cfg *config.Config, path, newVersion, prevVersion, bumpType string, skipHooks bool) error {
	if skipHooks {
		return nil
	}
	moduleInfo := moduleInfoFromPath(path)
	return extensionmgr.RunPreBumpHooks(ctx, cfg, newVersion, prevVersion, bumpType, moduleInfo)
}

// runPostBumpExtensionHooks runs post-bump extension hooks if not skipped.
func runPostBumpExtensionHooks(ctx context.Context, cfg *config.Config, path, prevVersion, bumpType string, skipHooks bool) error {
	if skipHooks {
		return nil
	}

	currentVersion, err := semver.ReadVersion(path)
	if err != nil {
		return err
	}

	prereleasePtr, metadataPtr := extractVersionPointers(currentVersion)
	moduleInfo := moduleInfoFromPath(path)
	return extensionmgr.RunPostBumpHooks(ctx, cfg, currentVersion.String(), prevVersion, bumpType, prereleasePtr, metadataPtr, moduleInfo)
}

// extractVersionPointers extracts prerelease and metadata as pointers (nil if empty).
func extractVersionPointers(v semver.SemVersion) (*string, *string) {
	var prereleasePtr, metadataPtr *string
	if v.PreRelease != "" {
		prereleasePtr = &v.PreRelease
	}
	if v.Build != "" {
		metadataPtr = &v.Build
	}
	return prereleasePtr, metadataPtr
}

// calculateNewBuild determines the build metadata for a new version.
func calculateNewBuild(meta string, preserveMeta bool, currentBuild string) string {
	if meta != "" {
		return meta
	}
	if preserveMeta {
		return currentBuild
	}
	return ""
}

// validateTagAvailable checks if a tag can be created for the version.
// Returns nil if tag manager is not enabled or tag is available.
func validateTagAvailable(registry *plugins.PluginRegistry, version semver.SemVersion) error {
	tm := registry.GetTagManager()
	if tm == nil {
		return nil
	}

	if !tm.IsAutoCreateEnabled() {
		return nil
	}

	return tm.ValidateTagAvailable(version)
}

// applyModuleTagPrefix resolves the effective tag prefix for a module and
// overrides the tag manager's prefix if it differs from root. Returns a
// cleanup function that restores the original prefix (caller should defer it).
func applyModuleTagPrefix(tm tagmanager.TagManager, bumpedPath string, cfg *config.Config) (cleanup func(), err error) {
	noop := func() {}
	moduleDir := "."
	if bumpedPath != "" {
		moduleDir = filepath.Dir(bumpedPath)
		// Make relative to CWD so tag prefixes use relative paths
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			if relDir, relErr := filepath.Rel(cwd, moduleDir); relErr == nil {
				moduleDir = relDir
			}
		}
	}
	if moduleDir == "." || moduleDir == "" || cfg == nil {
		return noop, nil
	}
	moduleCfg, err := config.LoadConfigFromDir(moduleDir)
	if err != nil {
		return noop, err
	}
	if moduleCfg == nil {
		return noop, nil
	}
	mergedCfg := config.MergeConfig(cfg, moduleCfg)
	if mergedCfg.Plugins == nil || mergedCfg.Plugins.TagManager == nil {
		return noop, nil
	}
	effectivePrefix := tagmanager.InterpolatePrefix(
		mergedCfg.Plugins.TagManager.GetPrefix(), moduleDir,
	)
	originalPrefix := tm.GetConfig().Prefix
	if effectivePrefix != originalPrefix {
		tm.(*tagmanager.TagManagerPlugin).SetPrefix(effectivePrefix)
		return func() { tm.(*tagmanager.TagManagerPlugin).SetPrefix(originalPrefix) }, nil
	}
	return noop, nil
}

// createTagAfterBump creates a git tag for the version if tag manager is enabled.
func createTagAfterBump(registry *plugins.PluginRegistry, version semver.SemVersion, bumpType string, cfg *config.Config) error {
	return commitAndTagAfterBump(registry, version, bumpType, "", cfg)
}

// commitAndTagAfterBump commits bump-modified files and creates a git tag.
// When auto-create is enabled, it stages and commits the bumpedPath and any other
// modified files before creating the tag so the tag points to the correct release commit.
func commitAndTagAfterBump(registry *plugins.PluginRegistry, version semver.SemVersion, bumpType string, bumpedPath string, cfg *config.Config) error {
	tm := registry.GetTagManager()
	if tm == nil {
		return nil
	}

	if !tm.IsAutoCreateEnabled() {
		return nil
	}

	// Apply per-module tag prefix if bumping a subdirectory module
	restorePrefix, err := applyModuleTagPrefix(tm, bumpedPath, cfg)
	if err != nil {
		return err
	}
	defer restorePrefix()

	// Commit bump-modified files before creating the tag
	var extraFiles []string
	if bumpedPath != "" {
		extraFiles = []string{bumpedPath}
	}
	if err := tm.CommitChanges(version, extraFiles); err != nil {
		return fmt.Errorf("failed to commit release changes: %w", err)
	}
	printer.PrintFaint(fmt.Sprintf("Committed release changes for %s", printer.Info(version.String())))

	// Create tag on the new commit
	message := fmt.Sprintf("Release %s (%s bump)", version.String(), bumpType)
	if err := tm.CreateTag(version, message); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tagName := tm.FormatTagName(version)
	printer.PrintFaint(fmt.Sprintf("Created tag: %s", printer.Info(tagName)))

	if tm.GetConfig().Push {
		printer.PrintFaint(fmt.Sprintf("Pushed tag: %s", printer.Info(tagName)))
	}

	return nil
}

// validateVersionPolicy checks if the version bump is allowed by configured policies.
// Returns nil if version validator is not enabled or validation passes.
func validateVersionPolicy(registry *plugins.PluginRegistry, newVersion, previousVersion semver.SemVersion, bumpType string) error {
	vv := registry.GetVersionValidator()
	if vv == nil {
		return nil
	}

	if !vv.IsEnabled() {
		return nil
	}

	return vv.Validate(newVersion, previousVersion, bumpType)
}

// validateReleaseGate checks if quality gates pass before allowing the bump.
// Returns nil if release gate is not enabled or all gates pass.
func validateReleaseGate(registry *plugins.PluginRegistry, newVersion, previousVersion semver.SemVersion, bumpType string) error {
	rg := registry.GetReleaseGate()
	if rg == nil {
		return nil
	}

	if !rg.IsEnabled() {
		return nil
	}

	return rg.ValidateRelease(newVersion, previousVersion, bumpType)
}

// validateDependencyConsistency checks if all dependency files match the current version.
// Returns nil if dependency checker is not enabled or all files are consistent.
func validateDependencyConsistency(registry *plugins.PluginRegistry, version semver.SemVersion) error {
	dc := registry.GetDependencyChecker()
	if dc == nil {
		return nil
	}

	if !dc.IsEnabled() {
		return nil
	}

	inconsistencies, err := dc.CheckConsistency(version.String())
	if err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	if len(inconsistencies) > 0 {
		// If auto-sync is enabled, skip the error - inconsistencies will be fixed after the bump
		if dc.GetConfig().AutoSync {
			return nil
		}

		var details strings.Builder
		details.WriteString("version inconsistencies detected:\n")
		for _, inc := range inconsistencies {
			fmt.Fprintf(&details, "  - %s\n", inc.String())
		}
		details.WriteString("\nRun with auto-sync enabled to fix automatically, or update files manually.")
		return fmt.Errorf("%s", details.String())
	}

	return nil
}

// applyModuleChangelog temporarily overrides the changelog generator's output
// directories and git scoping per module. Returns a cleanup function that restores
// the originals.
//
// moduleName is used for unified mode section headers (always set in multi-module mode).
// modulePath scopes versioned output dirs and git log (empty for root module).
// tagPrefix scopes tag resolution (empty for root module).
func applyModuleChangelog(cg changeloggenerator.ChangelogGenerator, moduleName, modulePath, tagPrefix string) func() {
	noop := func() {}

	plugin, ok := cg.(*changeloggenerator.ChangelogGeneratorPlugin)
	if !ok {
		return noop
	}

	cfg := cg.GetConfig()
	originalChangesDir := cfg.ChangesDir

	// Always set module name for unified mode heading
	if moduleName != "" {
		plugin.SetModuleName(moduleName)
	}

	// Scope versioned output dir and git operations for non-root modules
	if modulePath != "" {
		plugin.SetChangesDir(filepath.Join(originalChangesDir, modulePath))
		plugin.SetModulePath(modulePath)
		plugin.SetTagPrefix(tagPrefix)
	}

	return func() {
		plugin.SetChangesDir(originalChangesDir)
		plugin.SetModuleName("")
		plugin.SetModulePath("")
		plugin.SetTagPrefix("")
	}
}

// resolveTagPrefix returns the effective tag prefix for a module.
// If tag-manager is enabled, it interpolates the prefix template with the module path.
// Otherwise returns an empty string (no prefix filtering).
func resolveTagPrefix(registry *plugins.PluginRegistry, modulePath string) string {
	if modulePath == "" {
		return ""
	}
	tm := registry.GetTagManager()
	if tm == nil || !tm.IsAutoCreateEnabled() {
		return ""
	}
	return tagmanager.InterpolatePrefix(tm.GetConfig().Prefix, modulePath)
}

// generateChangelogAfterBump generates changelog entries if changelog generator is enabled.
// Returns nil if changelog generator is not enabled.
// moduleName identifies the module in unified changelog headings (empty for single-module).
// modulePath scopes versioned output dirs and git log (empty for root or single-module).
func generateChangelogAfterBump(registry *plugins.PluginRegistry, version, _ semver.SemVersion, bumpType, moduleName, modulePath string) error {
	cg := registry.GetChangelogGenerator()
	if cg == nil {
		return nil
	}

	if !cg.IsEnabled() {
		return nil
	}

	// Resolve the effective tag prefix for this module
	tagPrefix := resolveTagPrefix(registry, modulePath)

	// Apply per-module changelog and git scoping
	restore := applyModuleChangelog(cg, moduleName, modulePath, tagPrefix)
	defer restore()

	versionStr := "v" + version.String()

	// Use actual git tag for commit range, not version file content
	// The version file may contain pre-release/metadata that doesn't match a real tag
	prevVersionStr, err := changeloggenerator.GetLatestTag()
	if err != nil {
		// If no tags exist, generate from all commits
		prevVersionStr = ""
	}

	if err := cg.GenerateForVersion(versionStr, prevVersionStr, bumpType); err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	cfg := cg.GetConfig()
	switch cfg.Mode {
	case "versioned":
		printer.PrintFaint(fmt.Sprintf("Generated changelog: %s", printer.Info(fmt.Sprintf("%s/%s.md", cfg.ChangesDir, versionStr))))
	case "unified":
		printer.PrintFaint(fmt.Sprintf("Updated changelog: %s", printer.Info(cfg.ChangelogPath)))
	case "both":
		printer.PrintFaint(fmt.Sprintf("Generated changelog: %s and %s",
			printer.Info(fmt.Sprintf("%s/%s.md", cfg.ChangesDir, versionStr)), printer.Info(cfg.ChangelogPath)))
	}

	return nil
}

// recordAuditLogEntry records the version bump to the audit log if enabled.
// Returns nil if audit log is not enabled or if logging fails (doesn't block the bump).
func recordAuditLogEntry(registry *plugins.PluginRegistry, version, previousVersion semver.SemVersion, bumpType string) error {
	al := registry.GetAuditLog()
	if al == nil {
		return nil
	}

	if !al.IsEnabled() {
		return nil
	}

	entry := &auditlog.Entry{
		PreviousVersion: previousVersion.String(),
		NewVersion:      version.String(),
		BumpType:        bumpType,
	}

	// RecordEntry handles errors gracefully and logs warnings
	return al.RecordEntry(entry)
}
