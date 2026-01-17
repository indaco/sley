package changelogparser

import (
	"strings"
	"testing"
)

func TestGroupedParser_Format(t *testing.T) {
	p := newGroupedParser(nil)
	if p.Format() != "grouped" {
		t.Errorf("expected 'grouped', got %s", p.Format())
	}
}

func TestGroupedParser_ParseUnreleased(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantErr        bool
		errContains    string
		wantHasEntries bool
		wantBumpType   string
		wantConfidence string
		wantEntryCount int
		wantVersion    string
		wantDate       string
	}{
		{
			name: "standard grouped format",
			content: `# Changelog

## v1.2.0 - 2024-01-15

### Features

- **auth:** New OAuth2 support
- **core:** Improved performance

### Bug Fixes

- Fixed memory leak

## v1.1.0 - 2024-01-01
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 3,
			wantVersion:    "v1.2.0",
			wantDate:       "2024-01-15",
		},
		{
			name: "breaking changes triggers major",
			content: `## v2.0.0 - 2024-01-15

### Breaking Changes

- Removed deprecated API

### Features

- New feature

## v1.0.0
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "high",
			wantEntryCount: 2,
			wantVersion:    "v2.0.0",
		},
		{
			name: "fixes only triggers patch",
			content: `## v1.0.1 - 2024-01-15

### Bug Fixes

- Fixed bug

## v1.0.0
`,
			wantHasEntries: true,
			wantBumpType:   "patch",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "v1.0.1",
		},
		{
			name: "sections with icons",
			content: `## v1.2.0 - 2024-01-15

### :sparkles: Features

- New feature

### :bug: Bug Fixes

- Bug fix

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 2,
			wantVersion:    "v1.2.0",
		},
		{
			name: "entries with commit links",
			content: `## v1.2.0

### Features

- **auth:** OAuth2 support ([abc123](https://github.com/owner/repo/commit/abc123))

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "v1.2.0",
		},
		{
			name: "version without v prefix",
			content: `## 1.2.0 - 2024-01-15

### Features

- Feature

## 1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "1.2.0",
		},
		{
			name: "no version found",
			content: `# Changelog

Some text without version.
`,
			wantErr:     true,
			errContains: "no version section found",
		},
		{
			name: "asterisk bullet points",
			content: `## v1.2.0

### Features

* Feature A
* Feature B

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 2,
			wantVersion:    "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newGroupedParser(nil)
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

			if tt.wantVersion != "" && section.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", section.Version, tt.wantVersion)
			}

			if tt.wantDate != "" && section.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", section.Date, tt.wantDate)
			}
		})
	}
}

func TestGroupedParser_CustomSectionMap(t *testing.T) {
	content := `## v1.2.0

### New Features

- Feature A

### Bugfixes

- Fix B

## v1.1.0
`
	cfg := &Config{
		GroupedSectionMap: map[string]string{
			"New Features": "Added",
			"Bugfixes":     "Fixed",
		},
	}

	p := newGroupedParser(cfg)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(section.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(section.Entries))
	}

	categories := make(map[string]int)
	for _, e := range section.Entries {
		categories[e.Category]++
	}

	if categories["Added"] != 1 {
		t.Errorf("expected 1 Added entry, got %d", categories["Added"])
	}
	if categories["Fixed"] != 1 {
		t.Errorf("expected 1 Fixed entry, got %d", categories["Fixed"])
	}
}

func TestGroupedParser_EntryCleaning(t *testing.T) {
	content := `## v1.2.0

### Features

- **auth:** OAuth2 support ([abc123](https://github.com/owner/repo/commit/abc123)) ([#42](https://github.com/owner/repo/pull/42))

## v1.1.0
`
	p := newGroupedParser(nil)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(section.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(section.Entries))
	}

	e := section.Entries[0]
	if e.Scope != "auth" {
		t.Errorf("Scope = %q, want %q", e.Scope, "auth")
	}

	if strings.Contains(e.Description, "[abc123]") {
		t.Error("Description should not contain commit link")
	}

	if strings.Contains(e.Description, "[#42]") {
		t.Error("Description should not contain PR link")
	}
}

func TestStripSectionIcon(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{":sparkles: Features", "Features"},
		{":bug: Bug Fixes", "Bug Fixes"},
		{"Features", "Features"},
		{":boom: Breaking Changes", "Breaking Changes"},
		{"  :rocket: Enhancements", "Enhancements"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripSectionIcon(tt.input)
			if result != tt.expected {
				t.Errorf("stripSectionIcon(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripMarkdownLinks(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"text [link](url) more", "text link more"},
		{"[abc123](https://example.com)", "abc123"},
		{"no links here", "no links here"},
		{"[a](b) and [c](d)", "a and c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripMarkdownLinks(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdownLinks(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
