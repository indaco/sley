// Package depsync provides dependency synchronization for CLI commands.
package depsync

import (
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/semver"
)

// SyncDependencies delegates to operations.SyncDependencies.
// Kept for backward compatibility with existing callers outside the bump command.
func SyncDependencies(registry *plugins.PluginRegistry, version semver.SemVersion, bumpedPaths ...string) error {
	return operations.SyncDependencies(registry, version, bumpedPaths...)
}
