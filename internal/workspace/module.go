// Package workspace provides types and operations for managing multiple modules
// in a monorepo or multi-module context.
package workspace

import (
	"fmt"
)

// Module represents a single versioned module within a workspace.
// Each module has its own .version file and can be operated on independently.
type Module struct {
	// Name is the module identifier, typically derived from directory name.
	Name string

	// Path is the absolute path to the module's .version file.
	Path string

	// RelPath is the relative path from workspace root to the .version file.
	RelPath string

	// CurrentVersion is the current version string for display purposes.
	// This is loaded lazily to improve discovery performance.
	CurrentVersion string

	// Dir is the directory containing the .version file.
	Dir string
}

// DisplayName returns a formatted name suitable for TUI display.
// Format: "module-name (version)"
func (m *Module) DisplayName() string {
	if m.CurrentVersion == "" {
		return m.Name
	}
	return fmt.Sprintf("%s (%s)", m.Name, m.CurrentVersion)
}

// DisplayNameWithPath returns a formatted name with path for disambiguation.
// Format: "module-name (version) - path/to/module"
// Useful in TUI when multiple modules have the same name.
func (m *Module) DisplayNameWithPath() string {
	base := m.DisplayName()
	if m.Dir == "" || m.Dir == "." {
		return base
	}
	return fmt.Sprintf("%s - %s", base, m.Dir)
}

// String returns a string representation of the module.
func (m *Module) String() string {
	return m.DisplayName()
}
