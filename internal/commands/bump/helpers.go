package bump

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
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

	// Check if the plugin is enabled and auto-create is on
	if plugin, ok := tm.(*tagmanager.TagManagerPlugin); ok {
		if !plugin.IsAutoCreateEnabled() {
			return nil
		}
	}

	return tm.ValidateTagAvailable(version)
}

// createTagAfterBump creates a git tag for the version if tag manager is enabled.
func createTagAfterBump(registry *plugins.PluginRegistry, version semver.SemVersion, bumpType string) error {
	return commitAndTagAfterBump(registry, version, bumpType, "")
}

// commitAndTagAfterBump commits bump-modified files and creates a git tag.
// When auto-create is enabled, it stages and commits the bumpedPath and any other
// modified files before creating the tag so the tag points to the correct release commit.
func commitAndTagAfterBump(registry *plugins.PluginRegistry, version semver.SemVersion, bumpType string, bumpedPath string) error {
	tm := registry.GetTagManager()
	if tm == nil {
		return nil
	}

	// Check if the plugin is enabled and auto-create is on
	plugin, ok := tm.(*tagmanager.TagManagerPlugin)
	if !ok || !plugin.IsAutoCreateEnabled() {
		return nil
	}

	// Commit bump-modified files before creating the tag
	var extraFiles []string
	if bumpedPath != "" {
		extraFiles = []string{bumpedPath}
	}
	if err := plugin.CommitChanges(version, extraFiles); err != nil {
		return fmt.Errorf("failed to commit release changes: %w", err)
	}
	printer.PrintSuccess(fmt.Sprintf("Committed release changes for %s", version.String()))

	// Create tag on the new commit
	message := fmt.Sprintf("Release %s (%s bump)", version.String(), bumpType)
	if err := tm.CreateTag(version, message); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tagName := tm.FormatTagName(version)
	printer.PrintSuccess(fmt.Sprintf("Created tag: %s", tagName))

	if plugin.GetConfig().Push {
		printer.PrintSuccess(fmt.Sprintf("Pushed tag: %s", tagName))
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

	// Check if the plugin is enabled
	if plugin, ok := vv.(*versionvalidator.VersionValidatorPlugin); ok {
		if !plugin.IsEnabled() {
			return nil
		}
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

	// Check if the plugin is enabled
	if plugin, ok := rg.(*releasegate.ReleaseGatePlugin); ok {
		if !plugin.IsEnabled() {
			return nil
		}
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

	plugin, ok := dc.(*dependencycheck.DependencyCheckerPlugin)
	if !ok || !plugin.IsEnabled() {
		return nil
	}

	inconsistencies, err := dc.CheckConsistency(version.String())
	if err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	if len(inconsistencies) > 0 {
		// If auto-sync is enabled, skip the error - inconsistencies will be fixed after the bump
		if plugin.GetConfig().AutoSync {
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

// generateChangelogAfterBump generates changelog entries if changelog generator is enabled.
// Returns nil if changelog generator is not enabled.
func generateChangelogAfterBump(registry *plugins.PluginRegistry, version, _ semver.SemVersion, bumpType string) error {
	cg := registry.GetChangelogGenerator()
	if cg == nil {
		return nil
	}

	plugin, ok := cg.(*changeloggenerator.ChangelogGeneratorPlugin)
	if !ok || !plugin.IsEnabled() {
		return nil
	}

	versionStr := "v" + version.String()

	// Use actual git tag for commit range, not version file content
	// The version file may contain pre-release/metadata that doesn't match a real tag
	prevVersionStr, err := changeloggenerator.GetLatestTagFn()
	if err != nil {
		// If no tags exist, generate from all commits
		prevVersionStr = ""
	}

	if err := cg.GenerateForVersion(versionStr, prevVersionStr, bumpType); err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	mode := plugin.GetConfig().Mode
	switch mode {
	case "versioned":
		printer.PrintSuccess(fmt.Sprintf("Generated changelog: %s/%s.md", plugin.GetConfig().ChangesDir, versionStr))
	case "unified":
		printer.PrintSuccess(fmt.Sprintf("Updated changelog: %s", plugin.GetConfig().ChangelogPath))
	case "both":
		printer.PrintSuccess(fmt.Sprintf("Generated changelog: %s/%s.md and %s",
			plugin.GetConfig().ChangesDir, versionStr, plugin.GetConfig().ChangelogPath))
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

	plugin, ok := al.(*auditlog.AuditLogPlugin)
	if !ok || !plugin.IsEnabled() {
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
