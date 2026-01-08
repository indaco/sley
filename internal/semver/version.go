package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVersion represents a semantic version (major.minor.patch-preRelease+build).
type SemVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

var (
	// versionRegex matches semantic version strings with optional "v" prefix,
	// optional pre-release (e.g., "-beta.1"), and optional build metadata (e.g., "+build.123").
	// It captures:
	//   1. Major version
	//   2. Minor version
	//   3. Patch version
	//   4. (optional) Pre-release identifier
	//   5. (optional) Build metadata
	versionRegex = regexp.MustCompile(
		`^v?([^\.\-+]+)\.([^\.\-+]+)\.([^\.\-+]+)` + // major.minor.patch
			`(?:-([0-9A-Za-z\-\.]+))?` + // optional pre-release
			`(?:\+([0-9A-Za-z\-\.]+))?$`, // optional build metadata
	)

	// errInvalidVersion is returned when a version string does not conform
	// to the expected semantic version format.
	errInvalidVersion = errors.New("invalid version format")

	// BumpNextFunc is a function variable for performing heuristic-based version bumps.
	// It defaults to BumpNext but can be overridden in tests to simulate errors.
	BumpNextFunc = BumpNext

	// BumpByLabelFunc is a function variable for bumping a version using an explicit label (patch, minor, major).
	// It defaults to BumpByLabel but can be overridden in tests to simulate errors.
	BumpByLabelFunc = BumpByLabel
)

// String returns the string representation of the semantic version.
func (v SemVersion) String() string {
	var sb strings.Builder
	sb.Grow(20) // Pre-allocate for typical version string length
	sb.WriteString(strconv.Itoa(v.Major))
	sb.WriteByte('.')
	sb.WriteString(strconv.Itoa(v.Minor))
	sb.WriteByte('.')
	sb.WriteString(strconv.Itoa(v.Patch))
	if v.PreRelease != "" {
		sb.WriteByte('-')
		sb.WriteString(v.PreRelease)
	}
	if v.Build != "" {
		sb.WriteByte('+')
		sb.WriteString(v.Build)
	}
	return sb.String()
}

// maxVersionLength is the maximum allowed length for a version string.
// This prevents potential ReDoS attacks on the regex parser.
const maxVersionLength = 128

// ParseVersion parses a semantic version string and returns a SemVersion.
// Returns an error if the version format is invalid or exceeds maxVersionLength.
func ParseVersion(s string) (SemVersion, error) {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) > maxVersionLength {
		return SemVersion{}, fmt.Errorf("%w: version string exceeds maximum length of %d", errInvalidVersion, maxVersionLength)
	}

	matches := versionRegex.FindStringSubmatch(trimmed)
	if len(matches) < 4 {
		return SemVersion{}, errInvalidVersion
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid major version: %v", errInvalidVersion, err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid minor version: %v", errInvalidVersion, err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid patch version: %v", errInvalidVersion, err)
	}

	pre := matches[4]
	build := matches[5]

	return SemVersion{Major: major, Minor: minor, Patch: patch, PreRelease: pre, Build: build}, nil
}

// BumpNext applies heuristic-based smart bump logic.
// - If it's a pre-release (e.g., alpha.1, rc.1), it promotes to final version.
// - If it's a final release, it bumps patch by default.
func BumpNext(v SemVersion) (SemVersion, error) {
	// If the version has a pre-release label, strip it (promote to final)
	if v.PreRelease != "" {
		promoted := v
		promoted.PreRelease = ""
		return promoted, nil
	}

	if v.Major == 0 && v.Minor == 9 && v.Patch == 0 {
		return SemVersion{Major: v.Major, Minor: v.Minor + 1, Patch: 0}, nil
	}

	// Default case: bump patch
	return SemVersion{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
}

// BumpByLabel bumps the version using an explicit label (patch, minor, major).
func BumpByLabel(v SemVersion, label string) (SemVersion, error) {
	switch label {
	case "patch":
		return SemVersion{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
	case "minor":
		return SemVersion{Major: v.Major, Minor: v.Minor + 1, Patch: 0}, nil
	case "major":
		return SemVersion{Major: v.Major + 1, Minor: 0, Patch: 0}, nil
	default:
		return SemVersion{}, fmt.Errorf("invalid bump label: %s", label)
	}
}

// IncrementPreRelease increments the numeric suffix of a pre-release label.
// Preserves the original separator style:
// - "rc.1" -> "rc.2" (dot separator)
// - "rc-1" -> "rc-2" (dash separator)
// - "rc1" -> "rc2" (no separator)
// - "rc" -> "rc.1" (no number, defaults to dot)
// If current doesn't match base, returns base.1.
func IncrementPreRelease(current, base string) string {
	if current == base {
		return formatPreReleaseWithSep(base, 1, ".")
	}

	// Check if current starts with base
	if !strings.HasPrefix(current, base) {
		return formatPreReleaseWithSep(base, 1, ".")
	}

	// Get the suffix after base
	suffix := current[len(base):]
	if suffix == "" {
		return formatPreReleaseWithSep(base, 1, ".")
	}

	// Determine separator and parse number
	var sep string
	var numStr string

	switch suffix[0] {
	case '.':
		sep = "."
		numStr = suffix[1:]
	case '-':
		sep = "-"
		numStr = suffix[1:]
	default:
		// No separator, digits start immediately
		sep = ""
		numStr = suffix
	}

	// Validate numStr is all digits
	if numStr == "" || !isAllDigits(numStr) {
		return formatPreReleaseWithSep(base, 1, ".")
	}

	n, err := strconv.Atoi(numStr)
	if err != nil {
		return formatPreReleaseWithSep(base, 1, ".")
	}

	return formatPreReleaseWithSep(base, n+1, sep)
}

// isAllDigits returns true if s consists entirely of ASCII digits.
func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func formatPreReleaseWithSep(base string, num int, sep string) string {
	return fmt.Sprintf("%s%s%d", base, sep, num)
}
