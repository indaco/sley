package changeloggenerator

import "fmt"

// Formatter defines the interface for changelog formatters.
// Different formatters can produce different styles of changelogs
// from the same commit data.
type Formatter interface {
	// FormatChangelog generates the changelog content for a version.
	// It receives the version information, grouped commits, and remote info
	// and returns the formatted markdown content.
	FormatChangelog(
		version string,
		previousVersion string,
		grouped map[string][]*GroupedCommit,
		sortedKeys []string,
		remote *RemoteInfo,
	) string
}

// NewFormatter creates a new formatter based on the format type.
func NewFormatter(format string, config *Config) (Formatter, error) {
	switch format {
	case "grouped":
		return &GroupedFormatter{config: config}, nil
	case "keepachangelog":
		return &KeepAChangelogFormatter{config: config}, nil
	case "github":
		return &GitHubFormatter{config: config}, nil
	default:
		return nil, fmt.Errorf("unknown changelog format: %s (supported: grouped, keepachangelog, github)", format)
	}
}
