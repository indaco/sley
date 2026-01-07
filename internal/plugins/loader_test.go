package plugins

import (
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/changelogparser"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
)

func TestRegisterConfiguredPlugins_WithCommitParser(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: true,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	p := registry.GetCommitParser()
	if p == nil {
		t.Fatal("expected commit parser to be registered, got nil")
	}

	if p.Name() != "commit-parser" {
		t.Errorf("expected name 'commit-parser', got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_DisabledCommitParser(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: false,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if p := registry.GetCommitParser(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_NilConfig(t *testing.T) {
	commitparser.ResetCommitParser()

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(nil, registry)

	if p := registry.GetCommitParser(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_NilPluginsField(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: nil, // explicit nil
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if p := registry.GetCommitParser(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_WithTagManager(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	autoCreate := true
	annotate := true
	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled:    true,
				AutoCreate: &autoCreate,
				Prefix:     "v",
				Annotate:   &annotate,
				Push:       false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	tm := registry.GetTagManager()
	if tm == nil {
		t.Fatal("expected tag manager to be registered, got nil")
	}

	if tm.Name() != "tag-manager" {
		t.Errorf("expected name 'tag-manager', got %q", tm.Name())
	}
}

func TestRegisterConfiguredPlugins_TagManagerDisabled(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if tm := registry.GetTagManager(); tm != nil {
		t.Errorf("expected no tag manager to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_TagManagerNil(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if tm := registry.GetTagManager(); tm != nil {
		t.Errorf("expected no tag manager to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithVersionValidator(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: true,
				Rules: []config.ValidationRule{
					{Type: "major-version-max", Value: 10},
					{Type: "pre-release-format", Pattern: "^(alpha|beta)$"},
				},
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	vv := registry.GetVersionValidator()
	if vv == nil {
		t.Fatal("expected version validator to be registered, got nil")
	}

	if vv.Name() != "version-validator" {
		t.Errorf("expected name 'version-validator', got %q", vv.Name())
	}
}

func TestRegisterConfiguredPlugins_VersionValidatorDisabled(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if vv := registry.GetVersionValidator(); vv != nil {
		t.Errorf("expected no version validator to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_VersionValidatorNil(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if vv := registry.GetVersionValidator(); vv != nil {
		t.Errorf("expected no version validator to be registered when nil")
	}
}

func TestConvertValidationRules(t *testing.T) {
	configRules := []config.ValidationRule{
		{Type: "major-version-max", Value: 10},
		{Type: "pre-release-format", Pattern: "^alpha$"},
		{Type: "branch-constraint", Branch: "release/*", Allowed: []string{"patch"}, Enabled: true},
	}

	rules := convertValidationRules(configRules)

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	if rules[0].Type != versionvalidator.RuleMajorVersionMax {
		t.Errorf("expected rule type 'major-version-max', got %q", rules[0].Type)
	}
	if rules[0].Value != 10 {
		t.Errorf("expected value 10, got %d", rules[0].Value)
	}

	if rules[1].Pattern != "^alpha$" {
		t.Errorf("expected pattern '^alpha$', got %q", rules[1].Pattern)
	}

	if rules[2].Branch != "release/*" {
		t.Errorf("expected branch 'release/*', got %q", rules[2].Branch)
	}
	if len(rules[2].Allowed) != 1 || rules[2].Allowed[0] != "patch" {
		t.Errorf("expected allowed [patch], got %v", rules[2].Allowed)
	}
}

func TestConvertValidationRules_Empty(t *testing.T) {
	rules := convertValidationRules(nil)

	if len(rules) != 0 {
		t.Errorf("expected 0 rules for nil input, got %d", len(rules))
	}

	rules = convertValidationRules([]config.ValidationRule{})

	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty input, got %d", len(rules))
	}
}

func TestRegisterConfiguredPlugins_AllPlugins(t *testing.T) {
	commitparser.ResetCommitParser()
	tagmanager.ResetTagManager()
	versionvalidator.Unregister()
	defer func() {
		commitparser.ResetCommitParser()
		tagmanager.ResetTagManager()
		versionvalidator.Unregister()
	}()

	autoCreate := true
	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: true,
			TagManager: &config.TagManagerConfig{
				Enabled:    true,
				AutoCreate: &autoCreate,
			},
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: true,
				Rules:   []config.ValidationRule{{Type: "major-version-max", Value: 5}},
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if p := registry.GetCommitParser(); p == nil {
		t.Error("expected commit parser to be registered")
	}
	if tm := registry.GetTagManager(); tm == nil {
		t.Error("expected tag manager to be registered")
	}
	if vv := registry.GetVersionValidator(); vv == nil {
		t.Error("expected version validator to be registered")
	}
}

func TestRegisterConfiguredPlugins_WithDependencyCheck(t *testing.T) {
	dependencycheck.ResetDependencyChecker()
	defer dependencycheck.ResetDependencyChecker()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			DependencyCheck: &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
				Files: []config.DependencyFileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
				},
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	dc := registry.GetDependencyChecker()
	if dc == nil {
		t.Fatal("expected dependency check to be registered, got nil")
	}

	if dc.Name() != "dependency-check" {
		t.Errorf("expected name 'dependency-check', got %q", dc.Name())
	}
}

func TestRegisterConfiguredPlugins_DependencyCheckDisabled(t *testing.T) {
	dependencycheck.ResetDependencyChecker()
	defer dependencycheck.ResetDependencyChecker()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			DependencyCheck: &config.DependencyCheckConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if dc := registry.GetDependencyChecker(); dc != nil {
		t.Errorf("expected no dependency check to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_DependencyCheckNil(t *testing.T) {
	dependencycheck.ResetDependencyChecker()
	defer dependencycheck.ResetDependencyChecker()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			DependencyCheck: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if dc := registry.GetDependencyChecker(); dc != nil {
		t.Errorf("expected no dependency check to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithChangelogParser(t *testing.T) {
	changelogparser.ResetChangelogParser()
	defer changelogparser.ResetChangelogParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogParser: &config.ChangelogParserConfig{
				Enabled:                  true,
				Path:                     "CHANGELOG.md",
				RequireUnreleasedSection: true,
				InferBumpType:            true,
				Priority:                 "changelog",
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	cp := registry.GetChangelogParser()
	if cp == nil {
		t.Fatal("expected changelog parser to be registered, got nil")
	}

	if cp.Name() != "changelog-parser" {
		t.Errorf("expected name 'changelog-parser', got %q", cp.Name())
	}
}

func TestRegisterConfiguredPlugins_ChangelogParserDisabled(t *testing.T) {
	changelogparser.ResetChangelogParser()
	defer changelogparser.ResetChangelogParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogParser: &config.ChangelogParserConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if cp := registry.GetChangelogParser(); cp != nil {
		t.Errorf("expected no changelog parser to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_ChangelogParserNil(t *testing.T) {
	changelogparser.ResetChangelogParser()
	defer changelogparser.ResetChangelogParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogParser: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if cp := registry.GetChangelogParser(); cp != nil {
		t.Errorf("expected no changelog parser to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithChangelogGenerator(t *testing.T) {
	changeloggenerator.ResetChangelogGenerator()
	defer changeloggenerator.ResetChangelogGenerator()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled: true,
				Mode:    "versioned",
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	cg := registry.GetChangelogGenerator()
	if cg == nil {
		t.Fatal("expected changelog generator to be registered, got nil")
	}

	if cg.Name() != "changelog-generator" {
		t.Errorf("expected name 'changelog-generator', got %q", cg.Name())
	}
}

func TestRegisterConfiguredPlugins_ChangelogGeneratorDisabled(t *testing.T) {
	changeloggenerator.ResetChangelogGenerator()
	defer changeloggenerator.ResetChangelogGenerator()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogGenerator: &config.ChangelogGeneratorConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if cg := registry.GetChangelogGenerator(); cg != nil {
		t.Errorf("expected no changelog generator to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_ChangelogGeneratorNil(t *testing.T) {
	changeloggenerator.ResetChangelogGenerator()
	defer changeloggenerator.ResetChangelogGenerator()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ChangelogGenerator: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if cg := registry.GetChangelogGenerator(); cg != nil {
		t.Errorf("expected no changelog generator to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithReleaseGate(t *testing.T) {
	releasegate.Unregister()
	defer releasegate.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ReleaseGate: &config.ReleaseGateConfig{
				Enabled:              true,
				RequireCleanWorktree: true,
				AllowedBranches:      []string{"main", "release/*"},
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	rg := registry.GetReleaseGate()
	if rg == nil {
		t.Fatal("expected release gate to be registered, got nil")
	}

	if rg.Name() != "release-gate" {
		t.Errorf("expected name 'release-gate', got %q", rg.Name())
	}
}

func TestRegisterConfiguredPlugins_ReleaseGateDisabled(t *testing.T) {
	releasegate.Unregister()
	defer releasegate.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ReleaseGate: &config.ReleaseGateConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if rg := registry.GetReleaseGate(); rg != nil {
		t.Errorf("expected no release gate to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_ReleaseGateNil(t *testing.T) {
	releasegate.Unregister()
	defer releasegate.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			ReleaseGate: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if rg := registry.GetReleaseGate(); rg != nil {
		t.Errorf("expected no release gate to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithAuditLog(t *testing.T) {
	auditlog.ResetAuditLog()
	defer auditlog.ResetAuditLog()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			AuditLog: &config.AuditLogConfig{
				Enabled: true,
				Path:    ".version-history.json",
				Format:  "json",
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	al := registry.GetAuditLog()
	if al == nil {
		t.Fatal("expected audit log to be registered, got nil")
	}

	if al.Name() != "audit-log" {
		t.Errorf("expected name 'audit-log', got %q", al.Name())
	}
}

func TestRegisterConfiguredPlugins_AuditLogDisabled(t *testing.T) {
	auditlog.ResetAuditLog()
	defer auditlog.ResetAuditLog()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			AuditLog: &config.AuditLogConfig{
				Enabled: false,
			},
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if al := registry.GetAuditLog(); al != nil {
		t.Errorf("expected no audit log to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_AuditLogNil(t *testing.T) {
	auditlog.ResetAuditLog()
	defer auditlog.ResetAuditLog()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			AuditLog: nil,
		},
	}

	registry := NewPluginRegistry()
	RegisterBuiltinPlugins(cfg, registry)

	if al := registry.GetAuditLog(); al != nil {
		t.Errorf("expected no audit log to be registered when nil")
	}
}

func TestConvertDependencyCheckConfig(t *testing.T) {
	tests := []struct {
		name         string
		input        *config.DependencyCheckConfig
		wantEnabled  bool
		wantAutoSync bool
		wantFiles    int
	}{
		{
			name: "full config",
			input: &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
				Files: []config.DependencyFileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
					{Path: "go.mod", Pattern: "module version (.*)"},
				},
			},
			wantEnabled:  true,
			wantAutoSync: true,
			wantFiles:    2,
		},
		{
			name: "empty files",
			input: &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: false,
				Files:    nil,
			},
			wantEnabled:  true,
			wantAutoSync: false,
			wantFiles:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDependencyCheckConfig(tt.input)
			if result.Enabled != tt.wantEnabled {
				t.Errorf("expected Enabled %v, got %v", tt.wantEnabled, result.Enabled)
			}
			if result.AutoSync != tt.wantAutoSync {
				t.Errorf("expected AutoSync %v, got %v", tt.wantAutoSync, result.AutoSync)
			}
			if len(result.Files) != tt.wantFiles {
				t.Errorf("expected %d files, got %d", tt.wantFiles, len(result.Files))
			}
		})
	}
}

func TestConvertDependencyCheckConfig_FileDetails(t *testing.T) {
	input := &config.DependencyCheckConfig{
		Enabled: true,
		Files: []config.DependencyFileConfig{
			{Path: "package.json", Field: "version", Format: "json", Pattern: ""},
		},
	}

	result := convertDependencyCheckConfig(input)
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	file := result.Files[0]
	if file.Path != "package.json" {
		t.Errorf("expected path 'package.json', got %q", file.Path)
	}
	if file.Field != "version" {
		t.Errorf("expected field 'version', got %q", file.Field)
	}
	if file.Format != "json" {
		t.Errorf("expected format 'json', got %q", file.Format)
	}
}

func TestConvertChangelogParserConfig(t *testing.T) {
	tests := []struct {
		name                  string
		input                 *config.ChangelogParserConfig
		wantPath              string
		wantRequireUnreleased bool
		wantInferBumpType     bool
		wantPriority          string
	}{
		{
			name: "full config",
			input: &config.ChangelogParserConfig{
				Enabled:                  true,
				Path:                     "CHANGELOG.md",
				RequireUnreleasedSection: true,
				InferBumpType:            true,
				Priority:                 "changelog",
			},
			wantPath:              "CHANGELOG.md",
			wantRequireUnreleased: true,
			wantInferBumpType:     true,
			wantPriority:          "changelog",
		},
		{
			name: "minimal config",
			input: &config.ChangelogParserConfig{
				Enabled: true,
			},
			wantPath:              "",
			wantRequireUnreleased: false,
			wantInferBumpType:     false,
			wantPriority:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertChangelogParserConfig(tt.input)
			if result.Path != tt.wantPath {
				t.Errorf("expected path %q, got %q", tt.wantPath, result.Path)
			}
			if result.RequireUnreleasedSection != tt.wantRequireUnreleased {
				t.Errorf("expected RequireUnreleasedSection %v, got %v", tt.wantRequireUnreleased, result.RequireUnreleasedSection)
			}
			if result.InferBumpType != tt.wantInferBumpType {
				t.Errorf("expected InferBumpType %v, got %v", tt.wantInferBumpType, result.InferBumpType)
			}
			if result.Priority != tt.wantPriority {
				t.Errorf("expected priority %q, got %q", tt.wantPriority, result.Priority)
			}
		})
	}
}

func TestConvertReleaseGateConfig(t *testing.T) {
	tests := []struct {
		name                     string
		input                    *config.ReleaseGateConfig
		wantEnabled              bool
		wantRequireCleanWorktree bool
		wantRequireCIPass        bool
		wantBlockedOnWIP         bool
		wantAllowedBranches      int
		wantBlockedBranches      int
	}{
		{
			name: "full config",
			input: &config.ReleaseGateConfig{
				Enabled:              true,
				RequireCleanWorktree: true,
				RequireCIPass:        true,
				BlockedOnWIPCommits:  true,
				AllowedBranches:      []string{"main", "release/*"},
				BlockedBranches:      []string{"feature/*"},
			},
			wantEnabled:              true,
			wantRequireCleanWorktree: true,
			wantRequireCIPass:        true,
			wantBlockedOnWIP:         true,
			wantAllowedBranches:      2,
			wantBlockedBranches:      1,
		},
		{
			name: "minimal config",
			input: &config.ReleaseGateConfig{
				Enabled: true,
			},
			wantEnabled:              true,
			wantRequireCleanWorktree: false,
			wantRequireCIPass:        false,
			wantBlockedOnWIP:         false,
			wantAllowedBranches:      0,
			wantBlockedBranches:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertReleaseGateConfig(tt.input)
			if result.Enabled != tt.wantEnabled {
				t.Errorf("expected Enabled %v, got %v", tt.wantEnabled, result.Enabled)
			}
			if result.RequireCleanWorktree != tt.wantRequireCleanWorktree {
				t.Errorf("expected RequireCleanWorktree %v, got %v", tt.wantRequireCleanWorktree, result.RequireCleanWorktree)
			}
			if result.RequireCIPass != tt.wantRequireCIPass {
				t.Errorf("expected RequireCIPass %v, got %v", tt.wantRequireCIPass, result.RequireCIPass)
			}
			if result.BlockedOnWIPCommits != tt.wantBlockedOnWIP {
				t.Errorf("expected BlockedOnWIPCommits %v, got %v", tt.wantBlockedOnWIP, result.BlockedOnWIPCommits)
			}
			if len(result.AllowedBranches) != tt.wantAllowedBranches {
				t.Errorf("expected %d allowed branches, got %d", tt.wantAllowedBranches, len(result.AllowedBranches))
			}
			if len(result.BlockedBranches) != tt.wantBlockedBranches {
				t.Errorf("expected %d blocked branches, got %d", tt.wantBlockedBranches, len(result.BlockedBranches))
			}
		})
	}
}
