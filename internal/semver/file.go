package semver

import "context"

// VersionFilePerm defines secure file permissions for version files (owner read/write only).
const VersionFilePerm = 0600

// ReadVersion reads a version string from the given file and parses it into a SemVersion.
// This is a convenience function that uses the default VersionManager with context.Background().
// For better testability and context control, use VersionManager.Read() instead.
func ReadVersion(path string) (SemVersion, error) {
	return defaultManager.Read(context.Background(), path)
}

// SaveVersion writes a SemVersion to the given file path.
// This is a convenience function that uses the default VersionManager with context.Background().
// For better testability and context control, use VersionManager.Save() instead.
func SaveVersion(path string, version SemVersion) error {
	return defaultManager.Save(context.Background(), path, version)
}

// UpdateVersion updates the semantic version in the given file based on the bump type (patch, minor, major),
// and optionally sets the pre-release and build metadata strings.
// If preserve is true and meta is empty, existing build metadata is retained.
// This is a convenience function that uses the default VersionManager with context.Background().
// For better testability and context control, use VersionManager.Update() instead.
func UpdateVersion(path string, bumpType string, pre string, meta string, preserve bool) error {
	return defaultManager.Update(context.Background(), path, bumpType, pre, meta, preserve)
}

// UpdatePreRelease updates only the pre-release portion of the version.
// If label is provided, it switches to that label (starting at .1).
// If label is empty, it increments the existing pre-release number.
// This is a convenience function that uses the default VersionManager with context.Background().
// For better testability and context control, use VersionManager.UpdatePreRelease() instead.
func UpdatePreRelease(path string, label string, meta string, preserve bool) error {
	return defaultManager.UpdatePreRelease(context.Background(), path, label, meta, preserve)
}
