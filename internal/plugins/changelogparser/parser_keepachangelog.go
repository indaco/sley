package changelogparser

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

// Keep a Changelog section patterns for parsing.
var (
	sectionHeaderRe    = regexp.MustCompile(`^##\s+\[([^\]]+)\]`)
	subsectionHeaderRe = regexp.MustCompile(`^###\s+(.+)$`)
)

// keepAChangelogParser parses Keep a Changelog format.
type keepAChangelogParser struct {
	config *Config
}

func newKeepAChangelogParser(cfg *Config) *keepAChangelogParser {
	return &keepAChangelogParser{config: cfg}
}

func (p *keepAChangelogParser) Format() string {
	return "keepachangelog"
}

func (p *keepAChangelogParser) ParseUnreleased(reader io.Reader) (*ParsedSection, error) {
	section, err := parseKeepAChangelogUnreleased(reader)
	if err != nil {
		return nil, err
	}
	return section.ToParsedSection(), nil
}

// parseKeepAChangelogUnreleased reads and extracts the Unreleased section.
func parseKeepAChangelogUnreleased(reader io.Reader) (*UnreleasedSection, error) {
	scanner := bufio.NewScanner(reader)
	inUnreleased := false
	currentSubsection := ""
	section := &UnreleasedSection{
		Subsections: make(map[string][]string),
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if matches := sectionHeaderRe.FindStringSubmatch(line); matches != nil {
			versionName := matches[1]
			if strings.EqualFold(versionName, "Unreleased") {
				inUnreleased = true
				continue
			} else if inUnreleased {
				break
			}
		}

		if !inUnreleased {
			continue
		}

		if matches := subsectionHeaderRe.FindStringSubmatch(line); matches != nil {
			currentSubsection = strings.TrimSpace(matches[1])
			section.Subsections[currentSubsection] = []string{}
			continue
		}

		if currentSubsection != "" && strings.HasPrefix(trimmedLine, "- ") {
			entry := strings.TrimPrefix(trimmedLine, "- ")
			section.Subsections[currentSubsection] = append(section.Subsections[currentSubsection], entry)
			section.HasEntries = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !inUnreleased {
		return nil, errors.New("unreleased section not found in changelog")
	}

	mapSubsectionsToFields(section)
	return section, nil
}

// mapSubsectionsToFields maps the generic subsections map to specific fields.
func mapSubsectionsToFields(section *UnreleasedSection) {
	for name, entries := range section.Subsections {
		normalized := strings.ToLower(strings.TrimSpace(name))
		switch normalized {
		case "added":
			section.Added = entries
		case "changed":
			section.Changed = entries
		case "deprecated":
			section.Deprecated = entries
		case "removed":
			section.Removed = entries
		case "fixed":
			section.Fixed = entries
		case "security":
			section.Security = entries
		}
	}
}
