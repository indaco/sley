package changelogparser

import (
	"errors"
	"os"
)

// Function variables for testability.
var openFileFn = os.Open

// ChangelogSection represents a parsed section from CHANGELOG.md.
type ChangelogSection struct {
	Version     string
	Date        string
	Subsections map[string][]string
}

// UnreleasedSection represents the parsed Unreleased section with change types.
type UnreleasedSection struct {
	HasEntries  bool
	Added       []string
	Changed     []string
	Deprecated  []string
	Removed     []string
	Fixed       []string
	Security    []string
	Subsections map[string][]string
}

// changelogFileParser parses CHANGELOG.md files in Keep a Changelog format.
type changelogFileParser struct {
	path string
}

// newChangelogFileParser creates a new changelog parser for the given file path.
func newChangelogFileParser(path string) *changelogFileParser {
	return &changelogFileParser{path: path}
}

// ParseUnreleased extracts and parses the Unreleased section from CHANGELOG.md.
func (p *changelogFileParser) ParseUnreleased() (*UnreleasedSection, error) {
	file, err := openFileFn(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("changelog file not found")
		}
		return nil, err
	}
	defer file.Close()
	return parseKeepAChangelogUnreleased(file)
}

// InferBumpType determines the bump type based on changelog entries.
func (s *UnreleasedSection) InferBumpType() (string, error) {
	if !s.HasEntries {
		return "", errors.New("no changelog entries found in unreleased section")
	}

	if len(s.Removed) > 0 {
		return "major", nil
	}

	if len(s.Changed) > 0 {
		return "major", nil
	}

	if len(s.Added) > 0 {
		return "minor", nil
	}

	if len(s.Fixed) > 0 || len(s.Security) > 0 || len(s.Deprecated) > 0 {
		return "patch", nil
	}

	return "", errors.New("no bump type could be inferred from changelog")
}

// ToParsedSection converts UnreleasedSection to format-agnostic ParsedSection.
func (s *UnreleasedSection) ToParsedSection() *ParsedSection {
	ps := &ParsedSection{
		Version:    "Unreleased",
		HasEntries: s.HasEntries,
		Entries:    make([]ParsedEntry, 0),
	}

	addEntries := func(entries []string, category string) {
		for _, e := range entries {
			ps.Entries = append(ps.Entries, ParsedEntry{
				Category:        category,
				OriginalSection: category,
				Description:     e,
			})
		}
	}

	addEntries(s.Added, "Added")
	addEntries(s.Changed, "Changed")
	addEntries(s.Deprecated, "Deprecated")
	addEntries(s.Removed, "Removed")
	addEntries(s.Fixed, "Fixed")
	addEntries(s.Security, "Security")

	bumpType, confidence := inferBumpFromEntries(ps.Entries)
	ps.InferredBumpType = bumpType
	ps.BumpTypeConfidence = confidence

	return ps
}

// FromParsedSection creates UnreleasedSection from ParsedSection.
func FromParsedSection(ps *ParsedSection) *UnreleasedSection {
	s := &UnreleasedSection{
		HasEntries:  ps.HasEntries,
		Subsections: make(map[string][]string),
	}

	for _, e := range ps.Entries {
		cat := e.Category
		s.Subsections[cat] = append(s.Subsections[cat], e.Description)

		switch cat {
		case "Added":
			s.Added = append(s.Added, e.Description)
		case "Changed":
			s.Changed = append(s.Changed, e.Description)
		case "Deprecated":
			s.Deprecated = append(s.Deprecated, e.Description)
		case "Removed":
			s.Removed = append(s.Removed, e.Description)
		case "Fixed":
			s.Fixed = append(s.Fixed, e.Description)
		case "Security":
			s.Security = append(s.Security, e.Description)
		}
	}

	return s
}

// inferBumpFromEntries determines bump type and confidence from parsed entries.
func inferBumpFromEntries(entries []ParsedEntry) (bumpType, confidence string) {
	if len(entries) == 0 {
		return "", "none"
	}

	hasBreaking := false
	hasRemoved := false
	hasChanged := false
	hasAdded := false
	hasPatch := false

	for _, e := range entries {
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
		return "major", "high"
	}
	if hasChanged {
		return "major", "medium"
	}
	if hasAdded {
		return "minor", "high"
	}
	if hasPatch {
		return "patch", "high"
	}

	return "", "none"
}
