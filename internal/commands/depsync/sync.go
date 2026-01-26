// Package depsync provides dependency synchronization for CLI commands.
package depsync

import (
	"fmt"
	"path/filepath"

	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
)

// SyncDependencies updates all configured dependency files to match the new version.
// Returns nil if dependency checker is not enabled or auto-sync is disabled.
// The bumpedPaths parameter contains paths that were already bumped as modules
// and should be excluded from the output (they're still synced but not displayed twice).
func SyncDependencies(registry *plugins.PluginRegistry, version semver.SemVersion, bumpedPaths ...string) error {
	dc := registry.GetDependencyChecker()
	if dc == nil {
		return nil
	}

	plugin, ok := dc.(*dependencycheck.DependencyCheckerPlugin)
	if !ok || !plugin.IsEnabled() || !plugin.GetConfig().AutoSync {
		return nil
	}

	files := plugin.GetConfig().Files
	if len(files) == 0 {
		return nil
	}

	if err := dc.SyncVersions(version.String()); err != nil {
		return fmt.Errorf("failed to sync dependency versions: %w", err)
	}

	// Build set of bumped paths for quick lookup
	bumpedSet := make(map[string]bool, len(bumpedPaths))
	for _, p := range bumpedPaths {
		bumpedSet[p] = true
	}

	// Filter files to only show ones not already bumped as modules
	var additionalFiles []dependencycheck.FileConfig
	for _, file := range files {
		if !bumpedSet[file.Path] {
			additionalFiles = append(additionalFiles, file)
		}
	}

	// Only print section if there are additional files to show
	if len(additionalFiles) > 0 {
		fmt.Println("Sync dependencies")
		for _, file := range additionalFiles {
			name := deriveDependencyName(file.Path)
			fmt.Printf("  %s %s %s%s\n", printer.SuccessBadge("✓"), name, printer.Faint("("+file.Path+")"), printer.Faint(": "+version.String()))
		}
	}

	return nil
}

// deriveDependencyName extracts a display name from a file path.
// For .version files, uses the parent directory name.
// For other files (package.json, etc.), uses the filename.
func deriveDependencyName(path string) string {
	base := filepath.Base(path)
	if base == ".version" {
		dir := filepath.Dir(path)
		return filepath.Base(dir)
	}
	return base
}
