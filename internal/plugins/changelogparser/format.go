package changelogparser

import (
	"fmt"
	"io"
)

// Parser defines the interface for changelog format parsers.
type Parser interface {
	// ParseUnreleased extracts the Unreleased/latest version section.
	ParseUnreleased(reader io.Reader) (*ParsedSection, error)
	// Format returns the format name this parser handles.
	Format() string
}

// ParsedSection represents a parsed changelog section (format-agnostic).
type ParsedSection struct {
	Version            string
	Date               string
	HasEntries         bool
	Entries            []ParsedEntry
	InferredBumpType   string // major, minor, patch, or empty
	BumpTypeConfidence string // high, medium, low, none
}

// ParsedEntry represents a single changelog entry.
type ParsedEntry struct {
	Category        string // Semantic: Added, Changed, Fixed, Removed, etc.
	OriginalSection string // As it appeared in changelog
	Description     string // Entry text (cleaned)
	Scope           string // Optional scope
	IsBreaking      bool
	CommitType      string // Original type if parseable
}

// NewParser creates a parser for the specified format.
func NewParser(format string, cfg *Config) (Parser, error) {
	switch format {
	case "keepachangelog", "":
		return newKeepAChangelogParser(cfg), nil
	case "grouped":
		return newGroupedParser(cfg), nil
	case "github":
		return newGitHubParser(cfg), nil
	case "minimal":
		return newMinimalParser(cfg), nil
	case "auto":
		return newAutoDetectParser(cfg), nil
	default:
		return nil, fmt.Errorf("unknown changelog format: %s (supported: keepachangelog, grouped, github, minimal, auto)", format)
	}
}

// ValidFormats returns the list of valid parser format names.
func ValidFormats() []string {
	return []string{"keepachangelog", "grouped", "github", "minimal", "auto"}
}
