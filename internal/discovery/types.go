package discovery

import "github.com/indaco/sley/internal/parser"

// DetectionMode indicates the type of project structure detected.
type DetectionMode int

const (
	// NoModules indicates no .version files were found.
	NoModules DetectionMode = iota

	// SingleModule indicates a single .version file was found.
	SingleModule

	// MultiModule indicates multiple .version files were found.
	MultiModule
)

// String returns a human-readable representation of the detection mode.
func (m DetectionMode) String() string {
	switch m {
	case SingleModule:
		return "SingleModule"
	case MultiModule:
		return "MultiModule"
	case NoModules:
		return "NoModules"
	default:
		return "Unknown"
	}
}

// Result represents the complete discovery result for a project.
type Result struct {
	// Mode indicates the detected project mode.
	Mode DetectionMode

	// Modules contains discovered .version files (monorepo modules).
	Modules []Module

	// Manifests contains discovered manifest files with version information.
	Manifests []ManifestSource

	// SyncCandidates contains files suitable for dependency-check sync.
	SyncCandidates []SyncCandidate

	// Mismatches contains detected version mismatches.
	Mismatches []Mismatch
}

// HasModules returns true if any .version files were found.
func (r *Result) HasModules() bool {
	return len(r.Modules) > 0
}

// HasManifests returns true if any manifest files were found.
func (r *Result) HasManifests() bool {
	return len(r.Manifests) > 0
}

// HasMismatches returns true if version mismatches were detected.
func (r *Result) HasMismatches() bool {
	return len(r.Mismatches) > 0
}

// IsEmpty returns true if no version sources were found.
func (r *Result) IsEmpty() bool {
	return len(r.Modules) == 0 && len(r.Manifests) == 0
}

// PrimaryVersion returns the recommended primary version based on discovery.
// Priority: .version in cwd > first module > first manifest
func (r *Result) PrimaryVersion() string {
	for _, m := range r.Modules {
		if m.RelPath == ".version" {
			return m.Version
		}
	}
	if len(r.Modules) > 0 {
		return r.Modules[0].Version
	}
	if len(r.Manifests) > 0 {
		return r.Manifests[0].Version
	}
	return ""
}

// Module represents a discovered .version file.
type Module struct {
	// Name is the module name (typically the directory name).
	Name string

	// Path is the absolute path to the .version file.
	Path string

	// RelPath is the relative path from the discovery root.
	RelPath string

	// Version is the current version string.
	Version string

	// Dir is the directory containing the .version file.
	Dir string
}

// ManifestSource represents a discovered manifest file with version information.
type ManifestSource struct {
	// Path is the absolute path to the manifest file.
	Path string

	// RelPath is the relative path from the discovery root.
	RelPath string

	// Filename is the base name of the file (e.g., "package.json").
	Filename string

	// Version is the extracted version string.
	Version string

	// Format is the file format (json, yaml, toml, etc.).
	Format parser.Format

	// Field is the dot-notation path to the version field.
	Field string

	// Description is a human-readable description of the file type.
	Description string
}

// SyncCandidate represents a file that can be synced via dependency-check.
type SyncCandidate struct {
	// Path is the relative path to the file.
	Path string

	// Format is the file format.
	Format parser.Format

	// Field is the dot-notation path to the version field.
	Field string

	// Pattern is the regex pattern (for regex format).
	Pattern string

	// Version is the current version in the file.
	Version string

	// Description is a human-readable description.
	Description string
}

// ToFileConfig converts a SyncCandidate to a parser.FileConfig.
func (s SyncCandidate) ToFileConfig() parser.FileConfig {
	return parser.FileConfig{
		Path:    s.Path,
		Format:  s.Format,
		Field:   s.Field,
		Pattern: s.Pattern,
	}
}

// Mismatch represents a version mismatch between sources.
type Mismatch struct {
	// Source is the path of the file with the mismatched version.
	Source string

	// ExpectedVersion is what the version should be.
	ExpectedVersion string

	// ActualVersion is the version found in the file.
	ActualVersion string
}

// KnownManifest describes a known manifest file type for discovery.
type KnownManifest struct {
	// Filename is the expected filename.
	Filename string

	// Format is the file format.
	Format parser.Format

	// Field is the dot-notation path to the version field.
	Field string

	// Description is a human-readable description.
	Description string

	// Priority determines discovery order (lower = higher priority).
	Priority int
}

// DefaultKnownManifests returns the list of known manifest files to discover.
func DefaultKnownManifests() []KnownManifest {
	return []KnownManifest{
		{
			Filename:    "package.json",
			Format:      parser.FormatJSON,
			Field:       "version",
			Description: "Node.js (package.json)",
			Priority:    1,
		},
		{
			Filename:    "Cargo.toml",
			Format:      parser.FormatTOML,
			Field:       "package.version",
			Description: "Rust (Cargo.toml)",
			Priority:    2,
		},
		{
			Filename:    "pyproject.toml",
			Format:      parser.FormatTOML,
			Field:       "project.version",
			Description: "Python (pyproject.toml)",
			Priority:    3,
		},
		{
			Filename:    "Chart.yaml",
			Format:      parser.FormatYAML,
			Field:       "version",
			Description: "Helm (Chart.yaml)",
			Priority:    4,
		},
		{
			Filename:    "pubspec.yaml",
			Format:      parser.FormatYAML,
			Field:       "version",
			Description: "Dart/Flutter (pubspec.yaml)",
			Priority:    5,
		},
		{
			Filename:    "composer.json",
			Format:      parser.FormatJSON,
			Field:       "version",
			Description: "PHP (composer.json)",
			Priority:    6,
		},
		{
			Filename:    "version.txt",
			Format:      parser.FormatRaw,
			Field:       "",
			Description: "Plain text (version.txt)",
			Priority:    10,
		},
		{
			Filename:    "VERSION",
			Format:      parser.FormatRaw,
			Field:       "",
			Description: "Plain text (VERSION)",
			Priority:    11,
		},
	}
}
