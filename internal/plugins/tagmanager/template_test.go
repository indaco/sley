package tagmanager

import (
	"slices"
	"testing"
	"time"

	"github.com/indaco/sley/internal/semver"
)

func TestNewTemplateData(t *testing.T) {
	// Save and restore nowFunc
	originalNowFunc := nowFunc
	defer func() { nowFunc = originalNowFunc }()

	// Set a fixed time for deterministic testing
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return fixedTime }

	tests := []struct {
		name     string
		version  semver.SemVersion
		prefix   string
		wantData TemplateData
	}{
		{
			name:    "basic version",
			version: semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			prefix:  "v",
			wantData: TemplateData{
				Version:    "1.2.3",
				Tag:        "v1.2.3",
				Prefix:     "v",
				Date:       "2024-06-15",
				Major:      "1",
				Minor:      "2",
				Patch:      "3",
				PreRelease: "",
				Build:      "",
			},
		},
		{
			name:    "version with prerelease",
			version: semver.SemVersion{Major: 2, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			prefix:  "v",
			wantData: TemplateData{
				Version:    "2.0.0-alpha.1",
				Tag:        "v2.0.0-alpha.1",
				Prefix:     "v",
				Date:       "2024-06-15",
				Major:      "2",
				Minor:      "0",
				Patch:      "0",
				PreRelease: "alpha.1",
				Build:      "",
			},
		},
		{
			name:    "version with build metadata",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0, Build: "build.123"},
			prefix:  "release-",
			wantData: TemplateData{
				Version:    "1.0.0+build.123",
				Tag:        "release-1.0.0+build.123",
				Prefix:     "release-",
				Date:       "2024-06-15",
				Major:      "1",
				Minor:      "0",
				Patch:      "0",
				PreRelease: "",
				Build:      "build.123",
			},
		},
		{
			name:    "version with prerelease and build",
			version: semver.SemVersion{Major: 3, Minor: 1, Patch: 4, PreRelease: "rc.2", Build: "build.456"},
			prefix:  "",
			wantData: TemplateData{
				Version:    "3.1.4-rc.2+build.456",
				Tag:        "3.1.4-rc.2+build.456",
				Prefix:     "",
				Date:       "2024-06-15",
				Major:      "3",
				Minor:      "1",
				Patch:      "4",
				PreRelease: "rc.2",
				Build:      "build.456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTemplateData(tt.version, tt.prefix)

			if got.Version != tt.wantData.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.wantData.Version)
			}
			if got.Tag != tt.wantData.Tag {
				t.Errorf("Tag = %q, want %q", got.Tag, tt.wantData.Tag)
			}
			if got.Prefix != tt.wantData.Prefix {
				t.Errorf("Prefix = %q, want %q", got.Prefix, tt.wantData.Prefix)
			}
			if got.Date != tt.wantData.Date {
				t.Errorf("Date = %q, want %q", got.Date, tt.wantData.Date)
			}
			if got.Major != tt.wantData.Major {
				t.Errorf("Major = %q, want %q", got.Major, tt.wantData.Major)
			}
			if got.Minor != tt.wantData.Minor {
				t.Errorf("Minor = %q, want %q", got.Minor, tt.wantData.Minor)
			}
			if got.Patch != tt.wantData.Patch {
				t.Errorf("Patch = %q, want %q", got.Patch, tt.wantData.Patch)
			}
			if got.PreRelease != tt.wantData.PreRelease {
				t.Errorf("PreRelease = %q, want %q", got.PreRelease, tt.wantData.PreRelease)
			}
			if got.Build != tt.wantData.Build {
				t.Errorf("Build = %q, want %q", got.Build, tt.wantData.Build)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     TemplateData
		want     string
	}{
		{
			name:     "default template",
			template: "Release {version}",
			data: TemplateData{
				Version: "1.2.3",
			},
			want: "Release 1.2.3",
		},
		{
			name:     "template with tag",
			template: "{tag}: Release version {version}",
			data: TemplateData{
				Version: "1.2.3",
				Tag:     "v1.2.3",
			},
			want: "v1.2.3: Release version 1.2.3",
		},
		{
			name:     "template with date",
			template: "Released {version} on {date}",
			data: TemplateData{
				Version: "2.0.0",
				Date:    "2024-06-15",
			},
			want: "Released 2.0.0 on 2024-06-15",
		},
		{
			name:     "template with version components",
			template: "Version {major}.{minor}.{patch}",
			data: TemplateData{
				Major: "1",
				Minor: "2",
				Patch: "3",
			},
			want: "Version 1.2.3",
		},
		{
			name:     "template with prerelease",
			template: "{version} (pre-release: {prerelease})",
			data: TemplateData{
				Version:    "1.0.0-alpha.1",
				PreRelease: "alpha.1",
			},
			want: "1.0.0-alpha.1 (pre-release: alpha.1)",
		},
		{
			name:     "template with build metadata",
			template: "{version} build={build}",
			data: TemplateData{
				Version: "1.0.0+build.123",
				Build:   "build.123",
			},
			want: "1.0.0+build.123 build=build.123",
		},
		{
			name:     "template with all placeholders",
			template: "{prefix}{version} ({tag}) - {major}.{minor}.{patch}-{prerelease}+{build} on {date}",
			data: TemplateData{
				Version:    "1.2.3-rc.1+build.456",
				Tag:        "v1.2.3-rc.1+build.456",
				Prefix:     "v",
				Date:       "2024-06-15",
				Major:      "1",
				Minor:      "2",
				Patch:      "3",
				PreRelease: "rc.1",
				Build:      "build.456",
			},
			want: "v1.2.3-rc.1+build.456 (v1.2.3-rc.1+build.456) - 1.2.3-rc.1+build.456 on 2024-06-15",
		},
		{
			name:     "template with empty prerelease",
			template: "Release {version}{prerelease}",
			data: TemplateData{
				Version:    "1.0.0",
				PreRelease: "",
			},
			want: "Release 1.0.0",
		},
		{
			name:     "template with no placeholders",
			template: "Static message",
			data: TemplateData{
				Version: "1.0.0",
			},
			want: "Static message",
		},
		{
			name:     "template with repeated placeholders",
			template: "{version} - {version} - {version}",
			data: TemplateData{
				Version: "1.2.3",
			},
			want: "1.2.3 - 1.2.3 - 1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMessage(tt.template, tt.data)
			if got != tt.want {
				t.Errorf("FormatMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplatePlaceholders(t *testing.T) {
	// Verify all expected placeholders are defined
	expectedPlaceholders := []string{
		"{version}",
		"{tag}",
		"{prefix}",
		"{date}",
		"{major}",
		"{minor}",
		"{patch}",
		"{prerelease}",
		"{build}",
	}

	if len(TemplatePlaceholders) != len(expectedPlaceholders) {
		t.Errorf("TemplatePlaceholders has %d items, want %d", len(TemplatePlaceholders), len(expectedPlaceholders))
	}

	for _, expected := range expectedPlaceholders {
		if !slices.Contains(TemplatePlaceholders, expected) {
			t.Errorf("TemplatePlaceholders missing %q", expected)
		}
	}
}
