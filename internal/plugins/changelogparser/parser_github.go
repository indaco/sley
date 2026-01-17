package changelogparser

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

// GitHub format patterns.
var (
	githubVersionRe    = regexp.MustCompile(`^##\s+(v?\d+\.\d+\.\d+[^\s]*)\s*(?:-\s*(.+))?$`)
	githubSectionRe    = regexp.MustCompile(`^###\s+(.+)$`)
	githubAuthorRe     = regexp.MustCompile(`\s+by\s+@[\w-]+`)
	githubPRRefRe      = regexp.MustCompile(`\s+in\s+#\d+`)
	githubScopeEntryRe = regexp.MustCompile(`^\*\s*\*\*([^:]+):\*\*\s*(.+)$`)
)

// githubParser parses GitHub release changelog format.
// Note: GitHub format has limited type information - it cannot reliably
// distinguish between minor and patch changes in "What's Changed" section.
type githubParser struct {
	config *Config
}

func newGitHubParser(cfg *Config) *githubParser {
	return &githubParser{config: cfg}
}

func (p *githubParser) Format() string {
	return "github"
}

func (p *githubParser) ParseUnreleased(reader io.Reader) (*ParsedSection, error) {
	scanner := bufio.NewScanner(reader)
	section := &ParsedSection{
		Entries: make([]ParsedEntry, 0),
	}

	inFirstVersion := false
	foundAnyVersion := false
	currentSection := ""
	isBreakingSection := false
	isWhatsChanged := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if matches := githubVersionRe.FindStringSubmatch(trimmed); matches != nil {
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

		if matches := githubSectionRe.FindStringSubmatch(trimmed); matches != nil {
			currentSection = strings.TrimSpace(matches[1])
			isBreakingSection = strings.EqualFold(currentSection, "Breaking Changes")
			isWhatsChanged = strings.EqualFold(currentSection, "What's Changed")
			continue
		}

		if currentSection != "" && strings.HasPrefix(trimmed, "* ") {
			entry := p.parseGitHubEntry(trimmed, currentSection, isBreakingSection, isWhatsChanged)
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

func (p *githubParser) parseGitHubEntry(line, sectionName string, isBreaking, isWhatsChanged bool) ParsedEntry {
	entry := ParsedEntry{
		OriginalSection: sectionName,
		IsBreaking:      isBreaking,
	}

	if isBreaking {
		entry.Category = "Removed"
	} else if isWhatsChanged {
		entry.Category = ""
	}

	content := strings.TrimPrefix(line, "* ")

	if matches := githubScopeEntryRe.FindStringSubmatch(line); matches != nil {
		entry.Scope = strings.TrimSpace(matches[1])
		content = matches[2]
	}

	content = githubAuthorRe.ReplaceAllString(content, "")
	content = githubPRRefRe.ReplaceAllString(content, "")
	entry.Description = strings.TrimSpace(content)

	return entry
}

func (p *githubParser) inferBumpType(section *ParsedSection) {
	if len(section.Entries) == 0 {
		section.BumpTypeConfidence = "none"
		return
	}

	hasBreaking := false
	hasUnknown := false

	for _, e := range section.Entries {
		if e.IsBreaking {
			hasBreaking = true
		} else if e.Category == "" {
			hasUnknown = true
		}
	}

	if hasBreaking {
		section.InferredBumpType = "major"
		section.BumpTypeConfidence = "high"
		return
	}

	if hasUnknown {
		section.BumpTypeConfidence = "low"
		return
	}

	section.BumpTypeConfidence = "none"
}
