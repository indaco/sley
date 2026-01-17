package changelogparser

import (
	"strings"
	"testing"
)

func TestAutoDetectParser_Format(t *testing.T) {
	p := newAutoDetectParser(nil)
	if p.Format() != "auto" {
		t.Errorf("expected 'auto', got %s", p.Format())
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "keepachangelog format",
			content: `# Changelog

## [Unreleased]

### Added
- Feature

## [1.0.0] - 2024-01-01
`,
			expected: "keepachangelog",
		},
		{
			name: "minimal format",
			content: `# Changelog

## v1.2.0

- [Feat] New feature
- [Fix] Bug fix

## v1.1.0
`,
			expected: "minimal",
		},
		{
			name: "github format",
			content: `## v1.2.0 - 2024-01-15

### What's Changed

* Feature by @dev in #100

## v1.1.0
`,
			expected: "github",
		},
		{
			name: "grouped format",
			content: `## v1.2.0 - 2024-01-15

### Features

- New feature

### Bug Fixes

- Bug fix

## v1.1.0
`,
			expected: "grouped",
		},
		{
			name: "ambiguous defaults to keepachangelog",
			content: `# Changelog

Some introductory text.
`,
			expected: "keepachangelog",
		},
		{
			name: "keepachangelog takes precedence",
			content: `# Changelog

## [Unreleased]

### Added
- Feature

## v1.0.0 - 2024-01-01
`,
			expected: "keepachangelog",
		},
		{
			name: "minimal entry in github-like structure",
			content: `## v1.2.0

- [Feat] Feature

### What's Changed

* Something
`,
			expected: "minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFormat(tt.content)
			if result != tt.expected {
				t.Errorf("DetectFormat() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAutoDetectParser_ParseUnreleased(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantErr        bool
		wantHasEntries bool
		wantBumpType   string
	}{
		{
			name: "auto detects keepachangelog",
			content: `## [Unreleased]

### Added
- New feature

## [1.0.0]
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
		},
		{
			name: "auto detects minimal",
			content: `## v1.2.0

- [Feat] Feature
- [Fix] Fix

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
		},
		{
			name: "auto detects github",
			content: `## v1.2.0

### Breaking Changes

* Breaking change by @dev in #1

### What's Changed

* Feature by @dev in #2

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "major",
		},
		{
			name: "auto detects grouped",
			content: `## v1.2.0

### Features

- New feature

## v1.1.0
`,
			wantHasEntries: true,
			wantBumpType:   "minor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newAutoDetectParser(nil)
			section, err := p.ParseUnreleased(strings.NewReader(tt.content))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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
		})
	}
}
