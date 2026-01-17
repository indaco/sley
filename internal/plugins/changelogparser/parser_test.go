package changelogparser

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNewChangelogFileParser(t *testing.T) {
	parser := newChangelogFileParser("CHANGELOG.md")
	if parser.path != "CHANGELOG.md" {
		t.Errorf("expected path 'CHANGELOG.md', got %s", parser.path)
	}
}

func TestParseUnreleasedSection(t *testing.T) {
	tests := []struct {
		name          string
		changelog     string
		wantErr       bool
		errMsg        string
		checkEntries  bool
		expectedAdded int
		expectedFixed int
	}{
		{
			name: "valid unreleased section with all subsections",
			changelog: `# Changelog

## [Unreleased]

### Added
- New feature X
- New feature Y

### Changed
- Modified behavior Z

### Fixed
- Bug fix A
- Bug fix B

### Security
- Security patch C

## [1.0.0] - 2024-01-01

### Added
- Previous feature
`,
			wantErr:       false,
			checkEntries:  true,
			expectedAdded: 2,
			expectedFixed: 2,
		},
		{
			name: "unreleased section with only Added",
			changelog: `# Changelog

## [Unreleased]

### Added
- New feature X

## [1.0.0] - 2024-01-01
`,
			wantErr:       false,
			checkEntries:  true,
			expectedAdded: 1,
			expectedFixed: 0,
		},
		{
			name: "empty unreleased section",
			changelog: `# Changelog

## [Unreleased]

## [1.0.0] - 2024-01-01
`,
			wantErr:      false,
			checkEntries: false,
		},
		{
			name: "no unreleased section",
			changelog: `# Changelog

## [1.0.0] - 2024-01-01

### Added
- Some feature
`,
			wantErr: true,
			errMsg:  "unreleased section not found",
		},
		{
			name: "unreleased with case variations",
			changelog: `# Changelog

## [unreleased]

### added
- Feature A
`,
			wantErr:       false,
			checkEntries:  true,
			expectedAdded: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.changelog)

			section, err := parseKeepAChangelogUnreleased(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkEntries {
				if !section.HasEntries {
					t.Error("expected HasEntries to be true")
				}
				if len(section.Added) != tt.expectedAdded {
					t.Errorf("expected %d Added entries, got %d", tt.expectedAdded, len(section.Added))
				}
				if len(section.Fixed) != tt.expectedFixed {
					t.Errorf("expected %d Fixed entries, got %d", tt.expectedFixed, len(section.Fixed))
				}
			} else if section.HasEntries {
				t.Error("expected HasEntries to be false")
			}
		})
	}
}

func TestInferBumpType(t *testing.T) {
	tests := []struct {
		name     string
		section  *UnreleasedSection
		expected string
		wantErr  bool
	}{
		{
			name: "removed triggers major",
			section: &UnreleasedSection{
				HasEntries: true,
				Removed:    []string{"Deprecated API"},
			},
			expected: "major",
			wantErr:  false,
		},
		{
			name: "changed triggers major",
			section: &UnreleasedSection{
				HasEntries: true,
				Changed:    []string{"Breaking change"},
			},
			expected: "major",
			wantErr:  false,
		},
		{
			name: "added triggers minor",
			section: &UnreleasedSection{
				HasEntries: true,
				Added:      []string{"New feature"},
			},
			expected: "minor",
			wantErr:  false,
		},
		{
			name: "fixed triggers patch",
			section: &UnreleasedSection{
				HasEntries: true,
				Fixed:      []string{"Bug fix"},
			},
			expected: "patch",
			wantErr:  false,
		},
		{
			name: "security triggers patch",
			section: &UnreleasedSection{
				HasEntries: true,
				Security:   []string{"Security fix"},
			},
			expected: "patch",
			wantErr:  false,
		},
		{
			name: "deprecated triggers patch",
			section: &UnreleasedSection{
				HasEntries: true,
				Deprecated: []string{"Old method"},
			},
			expected: "patch",
			wantErr:  false,
		},
		{
			name: "removed takes precedence over added",
			section: &UnreleasedSection{
				HasEntries: true,
				Removed:    []string{"Old API"},
				Added:      []string{"New feature"},
			},
			expected: "major",
			wantErr:  false,
		},
		{
			name: "changed takes precedence over added",
			section: &UnreleasedSection{
				HasEntries: true,
				Changed:    []string{"Breaking change"},
				Added:      []string{"New feature"},
			},
			expected: "major",
			wantErr:  false,
		},
		{
			name: "added takes precedence over fixed",
			section: &UnreleasedSection{
				HasEntries: true,
				Added:      []string{"New feature"},
				Fixed:      []string{"Bug fix"},
			},
			expected: "minor",
			wantErr:  false,
		},
		{
			name: "no entries error",
			section: &UnreleasedSection{
				HasEntries: false,
			},
			wantErr: true,
		},
		{
			name: "no recognized sections error",
			section: &UnreleasedSection{
				HasEntries:  true,
				Subsections: map[string][]string{"Unknown": {"something"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.section.InferBumpType()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseUnreleased_FileOperations(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		origOpenFile := openFileFn
		defer func() { openFileFn = origOpenFile }()

		openFileFn = func(name string) (*os.File, error) {
			return nil, os.ErrNotExist
		}

		parser := newChangelogFileParser("nonexistent.md")
		_, err := parser.ParseUnreleased()

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "changelog file not found") {
			t.Errorf("expected 'changelog file not found' error, got: %v", err)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		origOpenFile := openFileFn
		defer func() { openFileFn = origOpenFile }()

		openFileFn = func(name string) (*os.File, error) {
			return nil, errors.New("permission denied")
		}

		parser := newChangelogFileParser("test.md")
		_, err := parser.ParseUnreleased()

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}
	})
}

func TestParseUnreleasedSection_ComplexScenarios(t *testing.T) {
	t.Run("multiple version sections", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Added
- Unreleased feature

## [2.0.0] - 2024-02-01

### Added
- Version 2 feature

## [1.0.0] - 2024-01-01

### Added
- Version 1 feature
`
		reader := strings.NewReader(changelog)

		section, err := parseKeepAChangelogUnreleased(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(section.Added) != 1 {
			t.Errorf("expected 1 Added entry, got %d", len(section.Added))
		}
		if section.Added[0] != "Unreleased feature" {
			t.Errorf("expected 'Unreleased feature', got %q", section.Added[0])
		}
	})

	t.Run("entries with different bullet formats", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Added
- Feature A
- Feature B with multiple words
-Feature C without space (should be ignored)

### Fixed
- Bug fix
`
		reader := strings.NewReader(changelog)

		section, err := parseKeepAChangelogUnreleased(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(section.Added) != 2 {
			t.Errorf("expected 2 Added entries, got %d", len(section.Added))
		}
	})

	t.Run("subsection without entries", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Added

### Fixed
- Bug fix
`
		reader := strings.NewReader(changelog)

		section, err := parseKeepAChangelogUnreleased(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(section.Added) != 0 {
			t.Errorf("expected 0 Added entries, got %d", len(section.Added))
		}
		if len(section.Fixed) != 1 {
			t.Errorf("expected 1 Fixed entry, got %d", len(section.Fixed))
		}
	})
}

func TestSectionRegexPatterns(t *testing.T) {
	t.Run("section header regex", func(t *testing.T) {
		tests := []struct {
			line     string
			expected string
			matches  bool
		}{
			{"## [Unreleased]", "Unreleased", true},
			{"## [1.2.3]", "1.2.3", true},
			{"## [1.2.3] - 2024-01-01", "1.2.3", true},
			{"## Unreleased", "", false},
			{"### [Unreleased]", "", false},
			{"# [Unreleased]", "", false},
		}

		for _, tt := range tests {
			matches := sectionHeaderRe.FindStringSubmatch(tt.line)
			if tt.matches {
				if matches == nil {
					t.Errorf("line %q should match", tt.line)
				} else if matches[1] != tt.expected {
					t.Errorf("line %q: expected %q, got %q", tt.line, tt.expected, matches[1])
				}
			} else if matches != nil {
				t.Errorf("line %q should not match", tt.line)
			}
		}
	})

	t.Run("subsection header regex", func(t *testing.T) {
		tests := []struct {
			line     string
			expected string
			matches  bool
		}{
			{"### Added", "Added", true},
			{"### Fixed", "Fixed", true},
			{"### Security", "Security", true},
			{"## Added", "", false},
			{"#### Added", "", false},
		}

		for _, tt := range tests {
			matches := subsectionHeaderRe.FindStringSubmatch(tt.line)
			if tt.matches {
				if matches == nil {
					t.Errorf("line %q should match", tt.line)
				} else if matches[1] != tt.expected {
					t.Errorf("line %q: expected %q, got %q", tt.line, tt.expected, matches[1])
				}
			} else if matches != nil {
				t.Errorf("line %q should not match", tt.line)
			}
		}
	})
}

func TestParseUnreleasedSection_ScannerError(t *testing.T) {
	errorReader := &erroringReader{}

	_, err := parseKeepAChangelogUnreleased(errorReader)
	if err == nil {
		t.Error("expected error from scanner, got nil")
	}
}

type erroringReader struct{}

func (e *erroringReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestToParsedSection(t *testing.T) {
	section := &UnreleasedSection{
		HasEntries: true,
		Added:      []string{"Feature A", "Feature B"},
		Fixed:      []string{"Bug fix"},
		Removed:    []string{"Old API"},
	}

	parsed := section.ToParsedSection()

	if !parsed.HasEntries {
		t.Error("expected HasEntries to be true")
	}

	if parsed.InferredBumpType != "major" {
		t.Errorf("expected InferredBumpType 'major', got %q", parsed.InferredBumpType)
	}

	if len(parsed.Entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(parsed.Entries))
	}

	categories := make(map[string]int)
	for _, e := range parsed.Entries {
		categories[e.Category]++
	}

	if categories["Added"] != 2 {
		t.Errorf("expected 2 Added entries, got %d", categories["Added"])
	}
	if categories["Fixed"] != 1 {
		t.Errorf("expected 1 Fixed entry, got %d", categories["Fixed"])
	}
	if categories["Removed"] != 1 {
		t.Errorf("expected 1 Removed entry, got %d", categories["Removed"])
	}
}

func TestFromParsedSection(t *testing.T) {
	parsed := &ParsedSection{
		HasEntries: true,
		Entries: []ParsedEntry{
			{Category: "Added", Description: "Feature A"},
			{Category: "Added", Description: "Feature B"},
			{Category: "Fixed", Description: "Bug fix"},
			{Category: "Removed", Description: "Old API"},
		},
	}

	section := FromParsedSection(parsed)

	if !section.HasEntries {
		t.Error("expected HasEntries to be true")
	}

	if len(section.Added) != 2 {
		t.Errorf("expected 2 Added entries, got %d", len(section.Added))
	}
	if len(section.Fixed) != 1 {
		t.Errorf("expected 1 Fixed entry, got %d", len(section.Fixed))
	}
	if len(section.Removed) != 1 {
		t.Errorf("expected 1 Removed entry, got %d", len(section.Removed))
	}
}

func TestInferBumpFromEntries(t *testing.T) {
	tests := []struct {
		name           string
		entries        []ParsedEntry
		wantBump       string
		wantConfidence string
	}{
		{
			name:           "empty entries",
			entries:        []ParsedEntry{},
			wantBump:       "",
			wantConfidence: "none",
		},
		{
			name: "breaking entry",
			entries: []ParsedEntry{
				{Category: "Added", IsBreaking: true},
			},
			wantBump:       "major",
			wantConfidence: "high",
		},
		{
			name: "removed category",
			entries: []ParsedEntry{
				{Category: "Removed"},
			},
			wantBump:       "major",
			wantConfidence: "high",
		},
		{
			name: "changed category",
			entries: []ParsedEntry{
				{Category: "Changed"},
			},
			wantBump:       "major",
			wantConfidence: "medium",
		},
		{
			name: "added category",
			entries: []ParsedEntry{
				{Category: "Added"},
			},
			wantBump:       "minor",
			wantConfidence: "high",
		},
		{
			name: "fixed category",
			entries: []ParsedEntry{
				{Category: "Fixed"},
			},
			wantBump:       "patch",
			wantConfidence: "high",
		},
		{
			name: "unknown category",
			entries: []ParsedEntry{
				{Category: "Unknown"},
			},
			wantBump:       "",
			wantConfidence: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bumpType, confidence := inferBumpFromEntries(tt.entries)
			if bumpType != tt.wantBump {
				t.Errorf("bumpType = %q, want %q", bumpType, tt.wantBump)
			}
			if confidence != tt.wantConfidence {
				t.Errorf("confidence = %q, want %q", confidence, tt.wantConfidence)
			}
		})
	}
}
