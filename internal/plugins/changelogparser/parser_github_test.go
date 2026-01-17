package changelogparser

import (
	"strings"
	"testing"
)

func TestGitHubParser_Format(t *testing.T) {
	p := newGitHubParser(nil)
	if p.Format() != "github" {
		t.Errorf("expected 'github', got %s", p.Format())
	}
}

func TestGitHubParser_ParseUnreleased(t *testing.T) {
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
			name: "standard github format with breaking",
			content: `## v1.2.0 - 2024-01-15

### Breaking Changes

* Removed deprecated API by @johndoe in #100

### What's Changed

* Added new feature by @janedoe in #101
* Fixed bug by @contributor in #102

## v1.1.0 - 2024-01-01
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "high",
			wantEntryCount: 3,
			wantVersion:    "v1.2.0",
			wantDate:       "2024-01-15",
		},
		{
			name: "whats changed only - low confidence",
			content: `## v1.2.0 - 2024-01-15

### What's Changed

* Added feature by @johndoe in #100
* Fixed something by @janedoe in #101

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "",
			wantConfidence: "low",
			wantEntryCount: 2,
			wantVersion:    "v1.2.0",
		},
		{
			name: "breaking changes only",
			content: `## v2.0.0 - 2024-01-15

### Breaking Changes

* Major API change by @dev in #50

## v1.0.0
`,
			wantHasEntries: true,
			wantBumpType:   "major",
			wantConfidence: "high",
			wantEntryCount: 1,
			wantVersion:    "v2.0.0",
		},
		{
			name: "entries with scope",
			content: `## v1.2.0

### What's Changed

* **api:** New endpoint by @dev in #100

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "",
			wantConfidence: "low",
			wantEntryCount: 1,
			wantVersion:    "v1.2.0",
		},
		{
			name: "no version found",
			content: `# Release Notes

Some text without version.
`,
			wantErr:     true,
			errContains: "no version section found",
		},
		{
			name: "empty changelog",
			content: `## v1.2.0

## v1.1.0
`,
			wantHasEntries: false,
			wantBumpType:   "",
			wantConfidence: "none",
			wantEntryCount: 0,
			wantVersion:    "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newGitHubParser(nil)
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

func TestGitHubParser_EntryParsing(t *testing.T) {
	content := `## v1.2.0 - 2024-01-15

### Breaking Changes

* **api:** Removed v1 endpoints by @johndoe in #100

### What's Changed

* Added OAuth support by @janedoe in #101
`
	p := newGitHubParser(nil)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(section.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(section.Entries))
	}

	// Check breaking entry
	breaking := section.Entries[0]
	if !breaking.IsBreaking {
		t.Error("first entry should be breaking")
	}
	if breaking.Category != "Removed" {
		t.Errorf("breaking entry category = %q, want %q", breaking.Category, "Removed")
	}
	if breaking.Scope != "api" {
		t.Errorf("breaking entry scope = %q, want %q", breaking.Scope, "api")
	}
	if strings.Contains(breaking.Description, "@johndoe") {
		t.Error("description should not contain author")
	}
	if strings.Contains(breaking.Description, "#100") {
		t.Error("description should not contain PR reference")
	}

	// Check regular entry
	regular := section.Entries[1]
	if regular.IsBreaking {
		t.Error("second entry should not be breaking")
	}
	if regular.Category != "" {
		t.Errorf("regular entry category should be empty, got %q", regular.Category)
	}
}

func TestGitHubParser_LimitationsDocumented(t *testing.T) {
	// This test documents that GitHub format cannot reliably distinguish
	// between minor and patch changes in "What's Changed" section.
	content := `## v1.2.0

### What's Changed

* Added feature by @dev in #1
* Fixed bug by @dev in #2

## v1.1.0
`
	p := newGitHubParser(nil)
	section, err := p.ParseUnreleased(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both entries have empty Category because we cannot determine type
	for _, e := range section.Entries {
		if e.Category != "" {
			t.Errorf("What's Changed entry should have empty category, got %q", e.Category)
		}
	}

	// Confidence should be low
	if section.BumpTypeConfidence != "low" {
		t.Errorf("confidence should be 'low', got %q", section.BumpTypeConfidence)
	}
}
