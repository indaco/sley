package changelogparser

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

// Minimal format patterns.
var (
	minimalVersionRe = regexp.MustCompile(`^##\s+v?(\d+\.\d+\.\d+.*)$`)
	minimalEntryRe   = regexp.MustCompile(`^-\s+\[(\w+)\]\s+(.+)$`)
)

// minimalTypeMapping maps minimal format type prefixes to semantic categories.
var minimalTypeMapping = map[string]struct {
	category string
	bumpType string
}{
	"Feat":     {"Added", "minor"},
	"Fix":      {"Fixed", "patch"},
	"Breaking": {"Removed", "major"},
	"Perf":     {"Changed", "patch"},
	"Refactor": {"Changed", ""},
	"Docs":     {"", ""},
	"Test":     {"", ""},
	"Chore":    {"", ""},
	"CI":       {"", ""},
	"Build":    {"", ""},
	"Revert":   {"Removed", "patch"},
	"Style":    {"", ""},
	"Other":    {"", ""},
}

// minimalParser parses minimal changelog format.
type minimalParser struct {
	config *Config
}

func newMinimalParser(cfg *Config) *minimalParser {
	return &minimalParser{config: cfg}
}

func (p *minimalParser) Format() string {
	return "minimal"
}

func (p *minimalParser) ParseUnreleased(reader io.Reader) (*ParsedSection, error) {
	scanner := bufio.NewScanner(reader)
	section := &ParsedSection{
		Entries: make([]ParsedEntry, 0),
	}

	inFirstSection := false
	foundAnyVersion := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if matches := minimalVersionRe.FindStringSubmatch(trimmed); matches != nil {
			if inFirstSection {
				break
			}
			section.Version = matches[1]
			inFirstSection = true
			foundAnyVersion = true
			continue
		}

		if !inFirstSection {
			continue
		}

		if matches := minimalEntryRe.FindStringSubmatch(trimmed); matches != nil {
			typePrefix := matches[1]
			description := matches[2]

			entry := p.parseMinimalEntry(typePrefix, description)
			section.Entries = append(section.Entries, entry)
			section.HasEntries = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !foundAnyVersion {
		return nil, errors.New("no version section found in changelog")
	}

	p.inferBumpType(section)
	return section, nil
}

func (p *minimalParser) parseMinimalEntry(typePrefix, description string) ParsedEntry {
	entry := ParsedEntry{
		Description:     description,
		OriginalSection: typePrefix,
		CommitType:      strings.ToLower(typePrefix),
	}

	if mapping, ok := minimalTypeMapping[typePrefix]; ok {
		entry.Category = mapping.category
		if typePrefix == "Breaking" {
			entry.IsBreaking = true
		}
	}

	return entry
}

func (p *minimalParser) inferBumpType(section *ParsedSection) {
	if len(section.Entries) == 0 {
		section.BumpTypeConfidence = "none"
		return
	}

	var maxBump string
	hasBreaking := false

	for _, e := range section.Entries {
		if e.IsBreaking {
			hasBreaking = true
			break
		}

		mapping, ok := minimalTypeMapping[e.OriginalSection]
		if !ok {
			continue
		}

		switch mapping.bumpType {
		case "major":
			maxBump = "major"
		case "minor":
			if maxBump != "major" {
				maxBump = "minor"
			}
		case "patch":
			if maxBump == "" {
				maxBump = "patch"
			}
		}
	}

	if hasBreaking {
		section.InferredBumpType = "major"
		section.BumpTypeConfidence = "high"
		return
	}

	if maxBump != "" {
		section.InferredBumpType = maxBump
		section.BumpTypeConfidence = "high"
	} else {
		section.BumpTypeConfidence = "none"
	}
}
