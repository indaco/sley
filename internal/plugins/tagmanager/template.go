package tagmanager

import (
	"fmt"
	"strings"
	"time"

	"github.com/indaco/sley/internal/semver"
)

// TemplatePlaceholders defines the available placeholders for message templates.
var TemplatePlaceholders = []string{
	"{version}",    // Full version string (e.g., "1.2.3-alpha.1+build.123")
	"{tag}",        // Full tag name with prefix (e.g., "v1.2.3")
	"{prefix}",     // Tag prefix (e.g., "v")
	"{date}",       // Current date in YYYY-MM-DD format
	"{major}",      // Major version number
	"{minor}",      // Minor version number
	"{patch}",      // Patch version number
	"{prerelease}", // Pre-release identifier (empty if none)
	"{build}",      // Build metadata (empty if none)
}

// TemplateData holds values for template placeholder substitution.
type TemplateData struct {
	Version    string
	Tag        string
	Prefix     string
	Date       string
	Major      string
	Minor      string
	Patch      string
	PreRelease string
	Build      string
}

// NowFunc returns the current time. Used to allow deterministic testing.
type NowFunc func() time.Time

// NewTemplateData creates TemplateData from a version and prefix.
// An optional NowFunc can be provided to override time.Now (for testing).
func NewTemplateData(version semver.SemVersion, prefix string, opts ...NowFunc) TemplateData {
	now := time.Now
	if len(opts) > 0 && opts[0] != nil {
		now = opts[0]
	}
	return TemplateData{
		Version:    version.String(),
		Tag:        prefix + version.String(),
		Prefix:     prefix,
		Date:       now().Format("2006-01-02"),
		Major:      fmt.Sprintf("%d", version.Major),
		Minor:      fmt.Sprintf("%d", version.Minor),
		Patch:      fmt.Sprintf("%d", version.Patch),
		PreRelease: version.PreRelease,
		Build:      version.Build,
	}
}

// FormatMessage applies template substitution to a message template.
func FormatMessage(template string, data TemplateData) string {
	result := template
	result = strings.ReplaceAll(result, "{version}", data.Version)
	result = strings.ReplaceAll(result, "{tag}", data.Tag)
	result = strings.ReplaceAll(result, "{prefix}", data.Prefix)
	result = strings.ReplaceAll(result, "{date}", data.Date)
	result = strings.ReplaceAll(result, "{major}", data.Major)
	result = strings.ReplaceAll(result, "{minor}", data.Minor)
	result = strings.ReplaceAll(result, "{patch}", data.Patch)
	result = strings.ReplaceAll(result, "{prerelease}", data.PreRelease)
	result = strings.ReplaceAll(result, "{build}", data.Build)
	return result
}
