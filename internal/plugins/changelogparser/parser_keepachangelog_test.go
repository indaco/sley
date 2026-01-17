package changelogparser

import (
	"strings"
	"testing"
)

func TestKeepAChangelogParser_Format(t *testing.T) {
	p := newKeepAChangelogParser(nil)
	if p.Format() != "keepachangelog" {
		t.Errorf("expected 'keepachangelog', got %s", p.Format())
	}
}

func TestKeepAChangelogParser_ParseUnreleased(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantErr        bool
		errContains    string
		wantHasEntries bool
		wantBumpType   string
		wantConfidence string
		wantEntryCount int
	}{
		{
			name: "standard format with all sections",
			content: `# Changelog

## [Unreleased]

### Added
- Feature A
- Feature B

### Fixed
- Bug fix

## [1.0.0] - 2024-01-01
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 3,
		},
		{
			name: "removed section triggers major",
			content: `# Changelog

## [Unreleased]

### Removed
- Deprecated API

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "high",
			wantEntryCount: 1,
		},
		{
			name: "changed section triggers major",
			content: `# Changelog

## [Unreleased]

### Changed
- Breaking change

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "medium",
			wantEntryCount: 1,
		},
		{
			name: "fixed only triggers patch",
			content: `# Changelog

## [Unreleased]

### Fixed
- Bug fix

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "patch",
			wantConfidence: "high",
			wantEntryCount: 1,
		},
		{
			name: "security triggers patch",
			content: `# Changelog

## [Unreleased]

### Security
- Security fix

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "patch",
			wantConfidence: "high",
			wantEntryCount: 1,
		},
		{
			name: "empty unreleased section",
			content: `# Changelog

## [Unreleased]

## [1.0.0]
`,
			wantHasEntries: false,
			wantBumpType:   "",
			wantConfidence: "none",
			wantEntryCount: 0,
		},
		{
			name: "no unreleased section",
			content: `# Changelog

## [1.0.0] - 2024-01-01

### Added
- Feature
`,
			wantErr:     true,
			errContains: "unreleased section not found",
		},
		{
			name: "case insensitive unreleased",
			content: `# Changelog

## [unreleased]

### added
- Feature

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newKeepAChangelogParser(nil)
			section, err := p.ParseUnreleased(strings.NewReader(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if section.HasEntries != tt.wantHasEntries {
				t.Errorf("HasEntries = %v, want %v", section.HasEntries, tt.wantHasEntries)
			}

			if section.InferredBumpType != tt.wantBumpType {
				t.Errorf("InferredBumpType = %q, want %q", section.InferredBumpType, tt.wantBumpType)
			}

			if section.BumpTypeConfidence != tt.wantConfidence {
				t.Errorf("BumpTypeConfidence = %q, want %q", section.BumpTypeConfidence, tt.wantConfidence)
			}

			if len(section.Entries) != tt.wantEntryCount {
				t.Errorf("entry count = %d, want %d", len(section.Entries), tt.wantEntryCount)
			}
		})
	}
}

func TestKeepAChangelogParser_EntryCategories(t *testing.T) {
	content := `# Changelog

## [Unreleased]

### Added
- New feature

### Changed
- Modified behavior

### Deprecated
- Old function

### Removed
- Legacy code

### Fixed
- Bug fix

### Security
- Security patch

## [1.0.0]
`
	p := newKeepAChangelogParser(nil)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	categories := make(map[string]int)
	for _, e := range section.Entries {
		categories[e.Category]++
	}

	expected := map[string]int{
		"Added":      1,
		"Changed":    1,
		"Deprecated": 1,
		"Removed":    1,
		"Fixed":      1,
		"Security":   1,
	}

	for cat, count := range expected {
		if categories[cat] != count {
			t.Errorf("category %q: got %d, want %d", cat, categories[cat], count)
		}
	}
}
