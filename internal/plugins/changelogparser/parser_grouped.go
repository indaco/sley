package changelogparser

import (
	"bufio"
	"errors"
	"io"
	"maps"
	"regexp"
	"strings"
	"unicode"
)

// Grouped format patterns.
var (
	groupedVersionRe      = regexp.MustCompile(`^##\s+(v?\d+\.\d+\.\d+[^\s]*)\s*(?:-\s*(.+))?$`)
	groupedSectionRe      = regexp.MustCompile(`^###\s+(.+)$`)
	groupedScopeEntryRe   = regexp.MustCompile(`^\*?\s*-?\s*\*\*([^:]+):\*\*\s*(.+)$`)
	groupedCommitLinkRe   = regexp.MustCompile(`\s*\(\[([a-f0-9]+)\]\([^)]+\)\)`)
	groupedPRLinkRe       = regexp.MustCompile(`\s*\(\[#(\d+)\]\([^)]+\)\)`)
	groupedMarkdownLinkRe = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
)

// defaultGroupedSectionMap maps common section names to semantic categories.
var defaultGroupedSectionMap = map[string]string{
	"breaking changes": "Removed",
	"features":         "Added",
	"enhancements":     "Added",
	"bug fixes":        "Fixed",
	"fixes":            "Fixed",
	"performance":      "Changed",
	"refactors":        "Changed",
	"documentation":    "",
	"styling":          "",
	"tests":            "",
	"chores":           "",
	"ci":               "",
	"build":            "",
	"reverts":          "Removed",
	"other":            "",
}

// groupedParser parses grouped changelog format.
type groupedParser struct {
	config     *Config
	sectionMap map[string]string
}

func newGroupedParser(cfg *Config) *groupedParser {
	p := &groupedParser{
		config:     cfg,
		sectionMap: make(map[string]string),
	}

	maps.Copy(p.sectionMap, defaultGroupedSectionMap)

	if cfg != nil && cfg.GroupedSectionMap != nil {
		for k, v := range cfg.GroupedSectionMap {
			p.sectionMap[strings.ToLower(k)] = v
		}
	}

	return p
}

func (p *groupedParser) Format() string {
	return "grouped"
}

func (p *groupedParser) ParseUnreleased(reader io.Reader) (*ParsedSection, error) {
	scanner := bufio.NewScanner(reader)
	section := &ParsedSection{
		Entries: make([]ParsedEntry, 0),
	}

	inFirstVersion := false
	foundAnyVersion := false
	currentSection := ""
	isBreakingSection := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if matches := groupedVersionRe.FindStringSubmatch(trimmed); matches != nil {
			if inFirstVersion {
				break
			}
			section.Version = matches[1]
			if len(matches) > 2 {
				section.Date = strings.TrimSpace(matches[2])
			}
			inFirstVersion = true
			foundAnyVersion = true
			continue
		}

		if !inFirstVersion {
			continue
		}

		if matches := groupedSectionRe.FindStringSubmatch(trimmed); matches != nil {
			currentSection = stripSectionIcon(strings.TrimSpace(matches[1]))
			isBreakingSection = strings.EqualFold(currentSection, "Breaking Changes")
			continue
		}

		if currentSection != "" && (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) {
			entry := p.parseGroupedEntry(trimmed, currentSection, isBreakingSection)
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

func (p *groupedParser) parseGroupedEntry(line, sectionName string, isBreaking bool) ParsedEntry {
	entry := ParsedEntry{
		OriginalSection: sectionName,
		IsBreaking:      isBreaking,
	}

	sectionKey := strings.ToLower(sectionName)
	if cat, ok := p.sectionMap[sectionKey]; ok {
		entry.Category = cat
	}

	content := strings.TrimLeft(line, "-* ")

	if matches := groupedScopeEntryRe.FindStringSubmatch(line); matches != nil {
		entry.Scope = strings.TrimSpace(matches[1])
		content = matches[2]
	}

	content = groupedCommitLinkRe.ReplaceAllString(content, "")
	content = groupedPRLinkRe.ReplaceAllString(content, "")
	content = groupedMarkdownLinkRe.ReplaceAllString(content, "$1")
	entry.Description = strings.TrimSpace(content)

	return entry
}

func (p *groupedParser) inferBumpType(section *ParsedSection) {
	if len(section.Entries) == 0 {
		section.BumpTypeConfidence = "none"
		return
	}

	hasBreaking := false
	hasRemoved := false
	hasChanged := false
	hasAdded := false
	hasPatch := false

	for _, e := range section.Entries {
		if e.IsBreaking {
			hasBreaking = true
		}
		switch e.Category {
		case "Removed":
			hasRemoved = true
		case "Changed":
			hasChanged = true
		case "Added":
			hasAdded = true
		case "Fixed", "Security", "Deprecated":
			hasPatch = true
		}
	}

	if hasBreaking || hasRemoved {
		section.InferredBumpType = "major"
		section.BumpTypeConfidence = "high"
		return
	}
	if hasChanged {
		section.InferredBumpType = "major"
		section.BumpTypeConfidence = "medium"
		return
	}
	if hasAdded {
		section.InferredBumpType = "minor"
		section.BumpTypeConfidence = "high"
		return
	}
	if hasPatch {
		section.InferredBumpType = "patch"
		section.BumpTypeConfidence = "high"
		return
	}

	section.BumpTypeConfidence = "low"
}

// stripSectionIcon removes leading emoji/icon from section name.
func stripSectionIcon(s string) string {
	s = strings.TrimSpace(s)

	// Handle GitHub-style emoji codes like :sparkles:
	if strings.HasPrefix(s, ":") {
		idx := strings.Index(s[1:], ":")
		if idx > 0 {
			s = strings.TrimSpace(s[idx+2:])
		}
	}

	// Handle actual unicode emoji at start
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}

	start := 0
	for i, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			start = i
			break
		}
	}

	return strings.TrimSpace(string(runes[start:]))
}

// stripMarkdownLinks removes [text](url) patterns, keeping text.
func stripMarkdownLinks(s string) string {
	return groupedMarkdownLinkRe.ReplaceAllString(s, "$1")
}
