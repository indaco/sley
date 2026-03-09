package semver

// VersionBumper abstracts version bumping operations for testability.
// This enables dependency injection instead of mutable package-level function variables.
type VersionBumper interface {
	// BumpNext applies heuristic-based smart bump logic.
	BumpNext(v SemVersion) (SemVersion, error)

	// BumpByLabel bumps the version using an explicit label (patch, minor, major).
	BumpByLabel(v SemVersion, label string) (SemVersion, error)
}

// DefaultBumper is the standard VersionBumper implementation
// that delegates to the package-level BumpNext and BumpByLabel functions.
type DefaultBumper struct{}

// NewDefaultBumper creates a new DefaultBumper.
func NewDefaultBumper() DefaultBumper {
	return DefaultBumper{}
}

// BumpNext applies heuristic-based smart bump logic.
func (DefaultBumper) BumpNext(v SemVersion) (SemVersion, error) {
	return BumpNext(v)
}

// BumpByLabel bumps the version using an explicit label.
func (DefaultBumper) BumpByLabel(v SemVersion, label string) (SemVersion, error) {
	return BumpByLabel(v, label)
}
