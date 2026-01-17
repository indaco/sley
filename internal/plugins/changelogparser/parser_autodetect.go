package changelogparser

import (
	"bytes"
	"io"
	"strings"
)

// autoDetectParser automatically detects and uses the appropriate parser.
type autoDetectParser struct {
	config *Config
}

func newAutoDetectParser(cfg *Config) *autoDetectParser {
	return &autoDetectParser{config: cfg}
}

func (p *autoDetectParser) Format() string {
	return "auto"
}

func (p *autoDetectParser) ParseUnreleased(reader io.Reader) (*ParsedSection, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	format := detectFormat(string(content))

	parser, err := NewParser(format, p.config)
	if err != nil {
		return nil, err
	}

	return parser.ParseUnreleased(bytes.NewReader(content))
}

// detectFormat analyzes content to determine the changelog format.
func detectFormat(content string) string {
	lines := strings.Split(content, "\n")

	hasKeepAChangelogHeader := false
	hasMinimalEntry := false
	hasWhatsChanged := false
	hasVersionWithV := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## [") {
			hasKeepAChangelogHeader = true
		}

		if minimalEntryRe.MatchString(trimmed) {
			hasMinimalEntry = true
		}

		if strings.HasPrefix(trimmed, "### What's Changed") {
			hasWhatsChanged = true
		}

		if strings.HasPrefix(trimmed, "## v") || strings.HasPrefix(trimmed, "## V") {
			hasVersionWithV = true
		}
	}

	if hasKeepAChangelogHeader {
		return "keepachangelog"
	}

	if hasMinimalEntry {
		return "minimal"
	}

	if hasWhatsChanged {
		return "github"
	}

	if hasVersionWithV {
		return "grouped"
	}

	return "keepachangelog"
}

// DetectFormat is exported for testing purposes.
func DetectFormat(content string) string {
	return detectFormat(content)
}
