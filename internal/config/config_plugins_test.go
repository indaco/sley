package config

import (
	"testing"

	"github.com/indaco/sley/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* TAG MANAGER CONFIG                                                        */
/* ------------------------------------------------------------------------- */

func TestTagManagerConfig_GetAutoCreate(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "nil AutoCreate - defaults to true",
			config:   &TagManagerConfig{AutoCreate: nil},
			expected: true,
		},
		{
			name:     "explicit true",
			config:   &TagManagerConfig{AutoCreate: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicit false",
			config:   &TagManagerConfig{AutoCreate: boolPtr(false)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAutoCreate()
			if result != tt.expected {
				t.Errorf("GetAutoCreate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetAnnotate(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "nil Annotate - defaults to true",
			config:   &TagManagerConfig{Annotate: nil},
			expected: true,
		},
		{
			name:     "explicit true",
			config:   &TagManagerConfig{Annotate: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicit false",
			config:   &TagManagerConfig{Annotate: boolPtr(false)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAnnotate()
			if result != tt.expected {
				t.Errorf("GetAnnotate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected string
	}{
		{
			name:     "empty prefix - defaults to v",
			config:   &TagManagerConfig{Prefix: ""},
			expected: "v",
		},
		{
			name:     "custom prefix",
			config:   &TagManagerConfig{Prefix: "release-"},
			expected: "release-",
		},
		{
			name:     "v prefix explicit",
			config:   &TagManagerConfig{Prefix: "v"},
			expected: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPrefix()
			if result != tt.expected {
				t.Errorf("GetPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetTagPrereleases(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "nil TagPrereleases - defaults to true",
			config:   &TagManagerConfig{TagPrereleases: nil},
			expected: true,
		},
		{
			name:     "explicit true",
			config:   &TagManagerConfig{TagPrereleases: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicit false",
			config:   &TagManagerConfig{TagPrereleases: boolPtr(false)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTagPrereleases()
			if result != tt.expected {
				t.Errorf("GetTagPrereleases() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetSign(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "sign false",
			config:   &TagManagerConfig{Sign: false},
			expected: false,
		},
		{
			name:     "sign true",
			config:   &TagManagerConfig{Sign: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetSign()
			if result != tt.expected {
				t.Errorf("GetSign() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetSigningKey(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected string
	}{
		{
			name:     "empty signing key",
			config:   &TagManagerConfig{SigningKey: ""},
			expected: "",
		},
		{
			name:     "custom signing key",
			config:   &TagManagerConfig{SigningKey: "ABC123DEF456"},
			expected: "ABC123DEF456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetSigningKey()
			if result != tt.expected {
				t.Errorf("GetSigningKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetMessageTemplate(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected string
	}{
		{
			name:     "empty message template - defaults to Release {version}",
			config:   &TagManagerConfig{MessageTemplate: ""},
			expected: "Release {version}",
		},
		{
			name:     "custom message template",
			config:   &TagManagerConfig{MessageTemplate: "{tag}: Release version {version}"},
			expected: "{tag}: Release version {version}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetMessageTemplate()
			if result != tt.expected {
				t.Errorf("GetMessageTemplate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// boolPtr is a helper to create a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

/* ------------------------------------------------------------------------- */
/* TAG MANAGER TAG-PRERELEASES CONFIG YAML LOADING                          */
/* ------------------------------------------------------------------------- */

func TestTagManagerConfig_TagPrereleases_YAMLLoading(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantValue bool
	}{
		{
			name: "tag-prereleases explicitly true",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    tag-prereleases: true
`,
			wantValue: true,
		},
		{
			name: "tag-prereleases explicitly false",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    tag-prereleases: false
`,
			wantValue: false,
		},
		{
			name: "tag-prereleases omitted (defaults to true)",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
`,
			wantValue: true,
		},
		{
			name: "tag-prereleases with other options",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: false
    tag-prereleases: false
`,
			wantValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cfg.Plugins == nil || cfg.Plugins.TagManager == nil {
					t.Fatal("expected tag-manager plugin config")
				}

				got := cfg.Plugins.TagManager.GetTagPrereleases()
				if got != tt.wantValue {
					t.Errorf("GetTagPrereleases() = %v, want %v", got, tt.wantValue)
				}
			})
		})
	}
}

/* ------------------------------------------------------------------------- */
/* TAG MANAGER SIGNING AND MESSAGE TEMPLATE YAML LOADING                    */
/* ------------------------------------------------------------------------- */

func TestTagManagerConfig_SignAndMessageTemplate_YAMLLoading(t *testing.T) {
	tests := []struct {
		name            string
		yamlInput       string
		wantSign        bool
		wantSigningKey  string
		wantMsgTemplate string
	}{
		{
			name: "sign enabled without key",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    sign: true
`,
			wantSign:        true,
			wantSigningKey:  "",
			wantMsgTemplate: "Release {version}",
		},
		{
			name: "sign enabled with key",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    sign: true
    signing-key: "ABC123DEF456"
`,
			wantSign:        true,
			wantSigningKey:  "ABC123DEF456",
			wantMsgTemplate: "Release {version}",
		},
		{
			name: "custom message template",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    message-template: "{tag}: Release version {version} on {date}"
`,
			wantSign:        false,
			wantSigningKey:  "",
			wantMsgTemplate: "{tag}: Release version {version} on {date}",
		},
		{
			name: "all new options",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
    sign: true
    signing-key: "MYKEY123"
    message-template: "Version {version}"
`,
			wantSign:        true,
			wantSigningKey:  "MYKEY123",
			wantMsgTemplate: "Version {version}",
		},
		{
			name: "sign disabled by default",
			yamlInput: `path: .version
plugins:
  tag-manager:
    enabled: true
`,
			wantSign:        false,
			wantSigningKey:  "",
			wantMsgTemplate: "Release {version}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cfg.Plugins == nil || cfg.Plugins.TagManager == nil {
					t.Fatal("expected tag-manager plugin config")
				}

				gotSign := cfg.Plugins.TagManager.GetSign()
				if gotSign != tt.wantSign {
					t.Errorf("GetSign() = %v, want %v", gotSign, tt.wantSign)
				}

				gotSigningKey := cfg.Plugins.TagManager.GetSigningKey()
				if gotSigningKey != tt.wantSigningKey {
					t.Errorf("GetSigningKey() = %q, want %q", gotSigningKey, tt.wantSigningKey)
				}

				gotMsgTemplate := cfg.Plugins.TagManager.GetMessageTemplate()
				if gotMsgTemplate != tt.wantMsgTemplate {
					t.Errorf("GetMessageTemplate() = %q, want %q", gotMsgTemplate, tt.wantMsgTemplate)
				}
			})
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BACKWARD COMPATIBILITY                                                    */
/* ------------------------------------------------------------------------- */

func TestConfig_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		check     func(t *testing.T, cfg *Config)
	}{
		{
			name: "existing config without workspace - still works",
			yamlInput: `path: .version
plugins:
  commit-parser: true
`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != ".version" {
					t.Errorf("expected path to be '.version', got %q", cfg.Path)
				}
				if cfg.Plugins == nil || !cfg.Plugins.CommitParser {
					t.Error("expected plugins.commit-parser to be true")
				}
				if cfg.Workspace != nil {
					t.Error("expected Workspace to be nil for legacy configs")
				}
			},
		},
		{
			name: "config with extensions and no workspace",
			yamlInput: `path: custom.version
extensions:
  - name: test-ext
    path: /path/to/ext
    enabled: true
`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != "custom.version" {
					t.Errorf("expected path to be 'custom.version', got %q", cfg.Path)
				}
				if len(cfg.Extensions) != 1 {
					t.Fatalf("expected 1 extension, got %d", len(cfg.Extensions))
				}
				if cfg.Extensions[0].Name != "test-ext" {
					t.Errorf("expected extension name 'test-ext', got %q", cfg.Extensions[0].Name)
				}
			},
		},
		{
			name:      "minimal config - just path",
			yamlInput: `path: minimal.version`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != "minimal.version" {
					t.Errorf("expected path to be 'minimal.version', got %q", cfg.Path)
				}
				if cfg.Workspace != nil {
					t.Error("expected Workspace to be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error loading legacy config: %v", err)
				}

				tt.check(t, cfg)
			})
		})
	}
}

/* ------------------------------------------------------------------------- */
/* AUDIT LOG CONFIG GETTER TESTS                                             */
/* ------------------------------------------------------------------------- */

func TestAuditLogConfig_GetPath(t *testing.T) {
	tests := []struct {
		name     string
		config   *AuditLogConfig
		expected string
	}{
		{
			name:     "empty path returns default",
			config:   &AuditLogConfig{Path: ""},
			expected: ".version-history.json",
		},
		{
			name:     "custom path returns custom",
			config:   &AuditLogConfig{Path: "custom-history.json"},
			expected: "custom-history.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPath()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAuditLogConfig_GetFormat(t *testing.T) {
	tests := []struct {
		name     string
		config   *AuditLogConfig
		expected string
	}{
		{
			name:     "empty format returns default",
			config:   &AuditLogConfig{Format: ""},
			expected: "json",
		},
		{
			name:     "custom format returns custom",
			config:   &AuditLogConfig{Format: "yaml"},
			expected: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetFormat()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* CHANGELOG GENERATOR CONFIG GETTER TESTS                                   */
/* ------------------------------------------------------------------------- */

func TestChangelogGeneratorConfig_GetChangesDir(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty changes dir returns default",
			config:   &ChangelogGeneratorConfig{ChangesDir: ""},
			expected: ".changes",
		},
		{
			name:     "custom changes dir returns custom",
			config:   &ChangelogGeneratorConfig{ChangesDir: "changelog-entries"},
			expected: "changelog-entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetChangesDir()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChangelogGeneratorConfig_GetChangelogPath(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty changelog path returns default",
			config:   &ChangelogGeneratorConfig{ChangelogPath: ""},
			expected: "CHANGELOG.md",
		},
		{
			name:     "custom changelog path returns custom",
			config:   &ChangelogGeneratorConfig{ChangelogPath: "docs/HISTORY.md"},
			expected: "docs/HISTORY.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetChangelogPath()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChangelogGeneratorConfig_GetMode(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty mode returns default",
			config:   &ChangelogGeneratorConfig{Mode: ""},
			expected: "versioned",
		},
		{
			name:     "unified mode returns unified",
			config:   &ChangelogGeneratorConfig{Mode: "unified"},
			expected: "unified",
		},
		{
			name:     "both mode returns both",
			config:   &ChangelogGeneratorConfig{Mode: "both"},
			expected: "both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetMode()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* CONTRIBUTORS CONFIG TESTS                                                 */
/* ------------------------------------------------------------------------- */

func TestContributorsConfig_Icon(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantIcon  string
	}{
		{
			name: "contributors with icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
      icon: "❤️"
`,
			wantIcon: "❤️",
		},
		{
			name: "contributors without icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
`,
			wantIcon: "",
		},
		{
			name: "contributors with custom format and icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
      format: "- {{.Name}}"
      icon: "👥"
`,
			wantIcon: "👥",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cfg.Plugins == nil || cfg.Plugins.ChangelogGenerator == nil {
					t.Fatal("expected changelog-generator plugin config")
				}

				if cfg.Plugins.ChangelogGenerator.Contributors == nil {
					t.Fatal("expected contributors config")
				}

				if cfg.Plugins.ChangelogGenerator.Contributors.Icon != tt.wantIcon {
					t.Errorf("expected icon %q, got %q",
						tt.wantIcon, cfg.Plugins.ChangelogGenerator.Contributors.Icon)
				}
			})
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BREAKING CHANGES ICON CONFIG TESTS                                        */
/* ------------------------------------------------------------------------- */

func TestChangelogGeneratorConfig_BreakingChangesIcon(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantIcon  string
	}{
		{
			name: "breaking-changes-icon with custom value",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    breaking-changes-icon: "BOOM"
`,
			wantIcon: "BOOM",
		},
		{
			name: "breaking-changes-icon omitted",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
`,
			wantIcon: "",
		},
		{
			name: "breaking-changes-icon with emoji",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    breaking-changes-icon: "💥"
`,
			wantIcon: "💥",
		},
		{
			name: "breaking-changes-icon with other options",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    format: github
    use-default-icons: false
    breaking-changes-icon: "WARNING"
`,
			wantIcon: "WARNING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cfg.Plugins == nil || cfg.Plugins.ChangelogGenerator == nil {
					t.Fatal("expected changelog-generator plugin config")
				}

				if cfg.Plugins.ChangelogGenerator.BreakingChangesIcon != tt.wantIcon {
					t.Errorf("expected breaking-changes-icon %q, got %q",
						tt.wantIcon, cfg.Plugins.ChangelogGenerator.BreakingChangesIcon)
				}
			})
		})
	}
}
