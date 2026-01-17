package changelogparser

import (
	"strings"
	"testing"
)

func TestMinimalParser_Format(t *testing.T) {
	p := newMinimalParser(nil)
	if p.Format() != "minimal" {
		t.Errorf("expected 'minimal', got %s", p.Format())
	}
}

func TestMinimalParser_ParseUnreleased(t *testing.T) {
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
	}{
		{
			name: "standard minimal format",
			content: `# Changelog

## v1.2.0

- [Feat] New authentication
- [Fix] Memory leak fix
- [Docs] Updated docs

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 3,
			wantVersion:    "1.2.0",
		},
		{
			name: "breaking change triggers major",
			content: `## v2.0.0

- [Breaking] Removed deprecated API
- [Feat] New feature

## v1.0.0
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "high",
			wantEntryCount: 2,
			wantVersion:    "2.0.0",
		},
		{
			name: "fix only triggers patch",
			content: `## v1.0.1

- [Fix] Bug fix

## v1.0.0
`,
			wantHasEntries: true,
			wantBumpType:   "patch",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "1.0.1",
		},
		{
			name: "all type prefixes",
			content: `## v1.0.0

- [Feat] Feature
- [Fix] Fix
- [Docs] Documentation
- [Perf] Performance
- [Refactor] Refactoring
- [Style] Styling
- [Test] Testing
- [Chore] Chore
- [CI] CI changes
- [Build] Build changes
- [Revert] Revert
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 11,
			wantVersion:    "1.0.0",
		},
		{
			name: "no version found",
			content: `# Changelog

Just some text without version headers.
`,
			wantErr:     true,
			errContains: "no version section found",
		},
		{
			name: "empty version section",
			content: `## v1.0.0

## v0.9.0
`,
			wantHasEntries: false,
			wantBumpType:   "",
			wantConfidence: "none",
			wantEntryCount: 0,
			wantVersion:    "1.0.0",
		},
		{
			name: "version without v prefix",
			content: `## 1.2.3

- [Feat] Feature

## 1.2.2
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newMinimalParser(nil)
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
		})
	}
}

func TestMinimalParser_EntryParsing(t *testing.T) {
	content := `## v1.0.0

- [Feat] New authentication system
- [Breaking] Removed old API
`
	p := newMinimalParser(nil)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(section.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(section.Entries))
	}

	// Check first entry
	e1 := section.Entries[0]
	if e1.Description != "New authentication system" {
		t.Errorf("entry 1 description = %q, want %q", e1.Description, "New authentication system")
	}
	if e1.Category != "Added" {
		t.Errorf("entry 1 category = %q, want %q", e1.Category, "Added")
	}
	if e1.OriginalSection != "Feat" {
		t.Errorf("entry 1 original section = %q, want %q", e1.OriginalSection, "Feat")
	}
	if e1.CommitType != "feat" {
		t.Errorf("entry 1 commit type = %q, want %q", e1.CommitType, "feat")
	}
	if e1.IsBreaking {
		t.Error("entry 1 should not be breaking")
	}

	// Check second entry
	e2 := section.Entries[1]
	if e2.Description != "Removed old API" {
		t.Errorf("entry 2 description = %q, want %q", e2.Description, "Removed old API")
	}
	if e2.Category != "Removed" {
		t.Errorf("entry 2 category = %q, want %q", e2.Category, "Removed")
	}
	if !e2.IsBreaking {
		t.Error("entry 2 should be breaking")
	}
}
