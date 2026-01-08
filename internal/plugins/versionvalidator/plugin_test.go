package versionvalidator

import (
	"context"
	"fmt"
	"testing"

	"github.com/indaco/sley/internal/semver"
)

func TestVersionValidatorPlugin_Name(t *testing.T) {
	vv := NewVersionValidator(nil)
	if got := vv.Name(); got != "version-validator" {
		t.Errorf("Name() = %q, want %q", got, "version-validator")
	}
}

func TestVersionValidatorPlugin_Description(t *testing.T) {
	vv := NewVersionValidator(nil)
	if got := vv.Description(); got == "" {
		t.Error("Description() should not be empty")
	}
}

func TestVersionValidatorPlugin_Version(t *testing.T) {
	vv := NewVersionValidator(nil)
	if got := vv.Version(); got != "v0.1.0" {
		t.Errorf("Version() = %q, want %q", got, "v0.1.0")
	}
}

func TestVersionValidatorPlugin_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{
			name: "enabled",
			cfg:  &Config{Enabled: true},
			want: true,
		},
		{
			name: "disabled",
			cfg:  &Config{Enabled: false},
			want: false,
		},
		{
			name: "nil config",
			cfg:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vv := NewVersionValidator(tt.cfg)
			if got := vv.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionValidatorPlugin_PreReleaseFormat(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		preRelease string
		wantErr    bool
	}{
		{
			name:       "valid alpha format",
			pattern:    `^(alpha|beta|rc)(\.[0-9]+)?$`,
			preRelease: "alpha.1",
			wantErr:    false,
		},
		{
			name:       "valid beta format",
			pattern:    `^(alpha|beta|rc)(\.[0-9]+)?$`,
			preRelease: "beta",
			wantErr:    false,
		},
		{
			name:       "valid rc format",
			pattern:    `^(alpha|beta|rc)(\.[0-9]+)?$`,
			preRelease: "rc.2",
			wantErr:    false,
		},
		{
			name:       "invalid format",
			pattern:    `^(alpha|beta|rc)(\.[0-9]+)?$`,
			preRelease: "preview.1",
			wantErr:    true,
		},
		{
			name:       "empty pre-release passes",
			pattern:    `^(alpha|beta|rc)(\.[0-9]+)?$`,
			preRelease: "",
			wantErr:    false,
		},
		{
			name:       "empty pattern passes",
			pattern:    "",
			preRelease: "anything",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules: []Rule{
					{Type: RulePreReleaseFormat, Pattern: tt.pattern},
				},
			}
			vv := NewVersionValidator(cfg)

			version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: tt.preRelease}
			err := vv.Validate(version, semver.SemVersion{}, "patch")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_MajorVersionMax(t *testing.T) {
	tests := []struct {
		name    string
		maxVal  int
		major   int
		wantErr bool
	}{
		{
			name:    "within limit",
			maxVal:  10,
			major:   5,
			wantErr: false,
		},
		{
			name:    "at limit",
			maxVal:  10,
			major:   10,
			wantErr: false,
		},
		{
			name:    "exceeds limit",
			maxVal:  10,
			major:   11,
			wantErr: true,
		},
		{
			name:    "no limit configured",
			maxVal:  0,
			major:   100,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules: []Rule{
					{Type: RuleMajorVersionMax, Value: tt.maxVal},
				},
			}
			vv := NewVersionValidator(cfg)

			version := semver.SemVersion{Major: tt.major, Minor: 0, Patch: 0}
			err := vv.Validate(version, semver.SemVersion{}, "major")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_MinorVersionMax(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: RuleMinorVersionMax, Value: 99},
		},
	}
	vv := NewVersionValidator(cfg)

	tests := []struct {
		name    string
		minor   int
		wantErr bool
	}{
		{"within limit", 50, false},
		{"at limit", 99, false},
		{"exceeds limit", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := semver.SemVersion{Major: 1, Minor: tt.minor, Patch: 0}
			err := vv.Validate(version, semver.SemVersion{}, "minor")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_PatchVersionMax(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: RulePatchVersionMax, Value: 50},
		},
	}
	vv := NewVersionValidator(cfg)

	tests := []struct {
		name    string
		patch   int
		wantErr bool
	}{
		{"within limit", 25, false},
		{"at limit", 50, false},
		{"exceeds limit", 51, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := semver.SemVersion{Major: 1, Minor: 0, Patch: tt.patch}
			err := vv.Validate(version, semver.SemVersion{}, "patch")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_RequirePreRelease0x(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		major      int
		preRelease string
		wantErr    bool
	}{
		{
			name:       "0.x without pre-release fails",
			enabled:    true,
			major:      0,
			preRelease: "",
			wantErr:    true,
		},
		{
			name:       "0.x with pre-release passes",
			enabled:    true,
			major:      0,
			preRelease: "alpha.1",
			wantErr:    false,
		},
		{
			name:       "1.x without pre-release passes",
			enabled:    true,
			major:      1,
			preRelease: "",
			wantErr:    false,
		},
		{
			name:       "rule disabled",
			enabled:    false,
			major:      0,
			preRelease: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules: []Rule{
					{Type: RuleRequirePreRelease0x, Enabled: tt.enabled},
				},
			}
			vv := NewVersionValidator(cfg)

			version := semver.SemVersion{Major: tt.major, Minor: 1, Patch: 0, PreRelease: tt.preRelease}
			err := vv.Validate(version, semver.SemVersion{}, "minor")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_NoBumpType(t *testing.T) {
	tests := []struct {
		name     string
		ruleType RuleType
		enabled  bool
		bumpType string
		wantErr  bool
	}{
		{
			name:     "no major bump - attempting major",
			ruleType: RuleNoMajorBump,
			enabled:  true,
			bumpType: "major",
			wantErr:  true,
		},
		{
			name:     "no major bump - attempting minor",
			ruleType: RuleNoMajorBump,
			enabled:  true,
			bumpType: "minor",
			wantErr:  false,
		},
		{
			name:     "no minor bump - attempting minor",
			ruleType: RuleNoMinorBump,
			enabled:  true,
			bumpType: "minor",
			wantErr:  true,
		},
		{
			name:     "no patch bump - attempting patch",
			ruleType: RuleNoPatchBump,
			enabled:  true,
			bumpType: "patch",
			wantErr:  true,
		},
		{
			name:     "rule disabled",
			ruleType: RuleNoMajorBump,
			enabled:  false,
			bumpType: "major",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules: []Rule{
					{Type: tt.ruleType, Enabled: tt.enabled},
				},
			}
			vv := NewVersionValidator(cfg)

			version := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}
			err := vv.Validate(version, semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, tt.bumpType)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_BranchConstraint(t *testing.T) {
	// Save original and restore after test
	original := getCurrentBranchFn
	defer func() { getCurrentBranchFn = original }()

	tests := []struct {
		name      string
		branch    string
		pattern   string
		allowed   []string
		bumpType  string
		wantErr   bool
		branchErr bool
	}{
		{
			name:     "allowed bump on matching branch",
			branch:   "release/1.0",
			pattern:  "release/*",
			allowed:  []string{"patch"},
			bumpType: "patch",
			wantErr:  false,
		},
		{
			name:     "disallowed bump on matching branch",
			branch:   "release/1.0",
			pattern:  "release/*",
			allowed:  []string{"patch"},
			bumpType: "minor",
			wantErr:  true,
		},
		{
			name:     "any bump on non-matching branch",
			branch:   "feature/new-thing",
			pattern:  "release/*",
			allowed:  []string{"patch"},
			bumpType: "major",
			wantErr:  false,
		},
		{
			name:      "branch check fails - skip validation",
			branch:    "",
			pattern:   "release/*",
			allowed:   []string{"patch"},
			bumpType:  "major",
			wantErr:   false,
			branchErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCurrentBranchFn = func(_ context.Context) (string, error) {
				if tt.branchErr {
					return "", nil
				}
				return tt.branch, nil
			}

			cfg := &Config{
				Enabled: true,
				Rules: []Rule{
					{Type: RuleBranchConstraint, Branch: tt.pattern, Allowed: tt.allowed},
				},
			}
			vv := NewVersionValidator(cfg)

			version := semver.SemVersion{Major: 1, Minor: 1, Patch: 1}
			err := vv.Validate(version, semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, tt.bumpType)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_ValidateSet(t *testing.T) {
	tests := []struct {
		name    string
		rules   []Rule
		version semver.SemVersion
		wantErr bool
	}{
		{
			name: "valid version passes all rules",
			rules: []Rule{
				{Type: RuleMajorVersionMax, Value: 10},
				{Type: RulePreReleaseFormat, Pattern: `^(alpha|beta|rc)(\.[0-9]+)?$`},
			},
			version: semver.SemVersion{Major: 5, Minor: 0, Patch: 0, PreRelease: "beta.1"},
			wantErr: false,
		},
		{
			name: "major exceeds max",
			rules: []Rule{
				{Type: RuleMajorVersionMax, Value: 10},
			},
			version: semver.SemVersion{Major: 15, Minor: 0, Patch: 0},
			wantErr: true,
		},
		{
			name: "invalid pre-release format",
			rules: []Rule{
				{Type: RulePreReleaseFormat, Pattern: `^(alpha|beta|rc)$`},
			},
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "dev"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules:   tt.rules,
			}
			vv := NewVersionValidator(cfg)

			err := vv.ValidateSet(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_MultipleRules(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: RuleMajorVersionMax, Value: 10},
			{Type: RulePreReleaseFormat, Pattern: `^(alpha|beta|rc)(\.[0-9]+)?$`},
			{Type: RuleRequirePreRelease0x, Enabled: true},
		},
	}
	vv := NewVersionValidator(cfg)

	tests := []struct {
		name    string
		version semver.SemVersion
		wantErr bool
	}{
		{
			name:    "passes all rules",
			version: semver.SemVersion{Major: 5, Minor: 0, Patch: 0, PreRelease: "beta.1"},
			wantErr: false,
		},
		{
			name:    "fails major max",
			version: semver.SemVersion{Major: 15, Minor: 0, Patch: 0, PreRelease: "beta.1"},
			wantErr: true,
		},
		{
			name:    "fails pre-release format",
			version: semver.SemVersion{Major: 5, Minor: 0, Patch: 0, PreRelease: "dev"},
			wantErr: true,
		},
		{
			name:    "fails require pre-release for 0.x",
			version: semver.SemVersion{Major: 0, Minor: 1, Patch: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vv.Validate(tt.version, semver.SemVersion{}, "minor")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_DisabledSkipsValidation(t *testing.T) {
	cfg := &Config{
		Enabled: false,
		Rules: []Rule{
			{Type: RuleMajorVersionMax, Value: 1},
		},
	}
	vv := NewVersionValidator(cfg)

	// This would fail if validation was enabled
	version := semver.SemVersion{Major: 100, Minor: 0, Patch: 0}
	err := vv.Validate(version, semver.SemVersion{}, "major")

	if err != nil {
		t.Errorf("Validate() should skip when disabled, got error: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled != false {
		t.Errorf("DefaultConfig().Enabled = %v, want false", cfg.Enabled)
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("DefaultConfig().Rules length = %d, want 0", len(cfg.Rules))
	}
}

func TestMatchBranchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		branch  string
		want    bool
	}{
		{"release/*", "release/1.0", true},
		{"release/*", "release/2.0.0", true},
		{"release/*", "feature/new", false},
		{"main", "main", true},
		{"main", "master", false},
		{"*-release", "hotfix-release", true},
		{"feature/*", "feature/my-feature", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.branch, func(t *testing.T) {
			got, err := matchBranchPattern(tt.pattern, tt.branch)
			if err != nil {
				t.Fatalf("matchBranchPattern() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("matchBranchPattern(%q, %q) = %v, want %v", tt.pattern, tt.branch, got, tt.want)
			}
		})
	}
}

func TestVersionValidatorPlugin_GetConfig(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules:   []Rule{{Type: RuleMajorVersionMax, Value: 10}},
	}
	vv := NewVersionValidator(cfg)

	got := vv.GetConfig()
	if got != cfg {
		t.Error("GetConfig() should return the same config passed to NewVersionValidator")
	}
	if !got.Enabled {
		t.Error("GetConfig().Enabled should be true")
	}
	if len(got.Rules) != 1 {
		t.Errorf("GetConfig().Rules length = %d, want 1", len(got.Rules))
	}
}

func TestRegister(t *testing.T) {
	Unregister()
	defer Unregister()

	cfg := &Config{Enabled: true}
	Register(cfg)

	vv := GetVersionValidatorFn()
	if vv == nil {
		t.Fatal("Register() should register the version validator")
	}
	if vv.Name() != "version-validator" {
		t.Errorf("Register() vv.Name() = %q, want %q", vv.Name(), "version-validator")
	}
}

func TestUnregister(t *testing.T) {
	Register(&Config{Enabled: true})

	if vv := GetVersionValidatorFn(); vv == nil {
		t.Fatal("Expected version validator to be registered")
	}

	Unregister()

	if vv := GetVersionValidatorFn(); vv != nil {
		t.Error("Unregister() should clear the registered validator")
	}
}

func TestVersionValidatorPlugin_InvalidRegexPattern(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: RulePreReleaseFormat, Pattern: "[invalid"},
		},
	}
	vv := NewVersionValidator(cfg)

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}
	err := vv.Validate(version, semver.SemVersion{}, "minor")

	if err == nil {
		t.Error("Validate() should return error for invalid regex pattern")
	}
}

func TestVersionValidatorPlugin_UnknownRuleType(t *testing.T) {
	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: "unknown-rule-type"},
		},
	}
	vv := NewVersionValidator(cfg)

	version := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}
	err := vv.Validate(version, semver.SemVersion{}, "minor")

	if err == nil {
		t.Error("Validate() should return error for unknown rule type")
	}
}

func TestVersionValidatorPlugin_ValidateSet_AllRuleTypes(t *testing.T) {
	tests := []struct {
		name    string
		rules   []Rule
		version semver.SemVersion
		wantErr bool
	}{
		{
			name: "minor version max - within limit",
			rules: []Rule{
				{Type: RuleMinorVersionMax, Value: 50},
			},
			version: semver.SemVersion{Major: 1, Minor: 25, Patch: 0},
			wantErr: false,
		},
		{
			name: "minor version max - exceeds limit",
			rules: []Rule{
				{Type: RuleMinorVersionMax, Value: 50},
			},
			version: semver.SemVersion{Major: 1, Minor: 51, Patch: 0},
			wantErr: true,
		},
		{
			name: "patch version max - within limit",
			rules: []Rule{
				{Type: RulePatchVersionMax, Value: 100},
			},
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 50},
			wantErr: false,
		},
		{
			name: "patch version max - exceeds limit",
			rules: []Rule{
				{Type: RulePatchVersionMax, Value: 100},
			},
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 101},
			wantErr: true,
		},
		{
			name: "require pre-release for 0.x",
			rules: []Rule{
				{Type: RuleRequirePreRelease0x, Enabled: true},
			},
			version: semver.SemVersion{Major: 0, Minor: 5, Patch: 0},
			wantErr: true,
		},
		{
			name: "require pre-release for 0.x - has pre-release",
			rules: []Rule{
				{Type: RuleRequirePreRelease0x, Enabled: true},
			},
			version: semver.SemVersion{Major: 0, Minor: 5, Patch: 0, PreRelease: "alpha"},
			wantErr: false,
		},
		{
			name: "branch constraint rule ignored in ValidateSet",
			rules: []Rule{
				{Type: RuleBranchConstraint, Branch: "release/*", Allowed: []string{"patch"}},
			},
			version: semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name: "no-bump rules ignored in ValidateSet",
			rules: []Rule{
				{Type: RuleNoMajorBump, Enabled: true},
			},
			version: semver.SemVersion{Major: 10, Minor: 0, Patch: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules:   tt.rules,
			}
			vv := NewVersionValidator(cfg)

			err := vv.ValidateSet(tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_ValidateSetDisabled(t *testing.T) {
	cfg := &Config{
		Enabled: false,
		Rules: []Rule{
			{Type: RuleMajorVersionMax, Value: 1},
		},
	}
	vv := NewVersionValidator(cfg)

	version := semver.SemVersion{Major: 100, Minor: 0, Patch: 0}
	err := vv.ValidateSet(version)

	if err != nil {
		t.Errorf("ValidateSet() should skip when disabled, got error: %v", err)
	}
}

func TestVersionValidatorPlugin_BranchConstraint_EmptyFields(t *testing.T) {
	original := getCurrentBranchFn
	defer func() { getCurrentBranchFn = original }()

	getCurrentBranchFn = func(_ context.Context) (string, error) {
		return "main", nil
	}

	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name:    "empty branch pattern",
			rule:    Rule{Type: RuleBranchConstraint, Branch: "", Allowed: []string{"patch"}},
			wantErr: false,
		},
		{
			name:    "empty allowed list",
			rule:    Rule{Type: RuleBranchConstraint, Branch: "main", Allowed: []string{}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled: true,
				Rules:   []Rule{tt.rule},
			}
			vv := NewVersionValidator(cfg)

			err := vv.Validate(semver.SemVersion{Major: 1}, semver.SemVersion{}, "major")

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidatorPlugin_BranchConstraint_GetBranchError(t *testing.T) {
	original := getCurrentBranchFn
	defer func() { getCurrentBranchFn = original }()

	// Mock getCurrentBranchFn to return an error
	getCurrentBranchFn = func(_ context.Context) (string, error) {
		return "", fmt.Errorf("git not available")
	}

	cfg := &Config{
		Enabled: true,
		Rules: []Rule{
			{Type: RuleBranchConstraint, Branch: "main", Allowed: []string{"patch"}},
		},
	}
	vv := NewVersionValidator(cfg)

	// When getting branch fails, validation should pass (skip the check)
	err := vv.Validate(semver.SemVersion{Major: 1}, semver.SemVersion{}, "major")
	if err != nil {
		t.Errorf("Validate() should skip branch constraint when branch lookup fails, got error: %v", err)
	}
}

func TestMatchBranchPattern_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		branch  string
		want    bool
		wantErr bool
	}{
		{
			name:    "exact match main",
			pattern: "main",
			branch:  "main",
			want:    true,
			wantErr: false,
		},
		{
			name:    "wildcard match release",
			pattern: "release/*",
			branch:  "release/v1.0.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "no match develop vs main",
			pattern: "main",
			branch:  "develop",
			want:    false,
			wantErr: false,
		},
		{
			name:    "wildcard no match feature",
			pattern: "release/*",
			branch:  "feature/new",
			want:    false,
			wantErr: false,
		},
		{
			name:    "multiple wildcards in pattern",
			pattern: "feature/*/test",
			branch:  "feature/user/test",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchBranchPattern(tt.pattern, tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchBranchPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("matchBranchPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}
