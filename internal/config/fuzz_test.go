package config

import (
	"bytes"
	"testing"

	"github.com/goccy/go-yaml"
)

// FuzzConfigParsing tests YAML config parsing with random inputs.
// Run with: go test -fuzz=FuzzConfigParsing -fuzztime=30s
func FuzzConfigParsing(f *testing.F) {
	// Seed corpus with valid configs
	seeds := []string{
		// Minimal config
		`path: .version`,
		// Config with plugins
		`path: .version
plugins:
  commit-parser: true`,
		// Config with tag manager
		`path: .version
plugins:
  tag-manager:
    enabled: true
    prefix: v`,
		// Config with extensions
		`path: .version
extensions:
  - name: my-ext
    path: /path/to/ext
    enabled: true`,
		// Config with workspace
		`path: .version
workspace:
  discovery:
    enabled: true
    recursive: true
    max_depth: 10`,
		// Config with modules
		`path: .version
workspace:
  modules:
    - name: core
      path: ./core/.version
      enabled: true`,
		// Full config
		`path: .version
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: v
    annotate: true
  changelog-generator:
    enabled: true
    mode: unified
extensions:
  - name: test
    path: ./ext
    enabled: false`,
		// Edge cases
		``,
		`path: ""`,
		`plugins: null`,
		`extensions: []`,
		// Invalid YAML (for error handling)
		`path: [invalid`,
		`plugins: {unclosed`,
		`: invalid`,
		`	tabs: not: valid`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		fuzzConfigParsingInput(t, input)
	})
}

// fuzzConfigParsingInput tests a single YAML input for config parsing.
func fuzzConfigParsingInput(t *testing.T, input string) {
	t.Helper()

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(input)), yaml.Strict())

	// The parser should never panic
	err := decoder.Decode(&cfg)

	if err == nil {
		// If parsing succeeded, verify the config is usable
		verifyConfigUsable(t, &cfg)
	}

	// Verify consistency: parsing the same input should give same result
	verifyConfigConsistency(t, input)
}

// verifyConfigUsable checks that a parsed config can be used without panics.
func verifyConfigUsable(t *testing.T, cfg *Config) {
	t.Helper()

	// These should never panic
	_ = cfg.Path
	_ = cfg.GetDiscoveryConfig()
	_ = cfg.GetExcludePatterns()
	_ = cfg.HasExplicitModules()

	if cfg.Plugins != nil {
		if cfg.Plugins.TagManager != nil {
			_ = cfg.Plugins.TagManager.GetAutoCreate()
			_ = cfg.Plugins.TagManager.GetAnnotate()
			_ = cfg.Plugins.TagManager.GetPrefix()
		}
		if cfg.Plugins.ChangelogGenerator != nil {
			_ = cfg.Plugins.ChangelogGenerator.GetChangesDir()
			_ = cfg.Plugins.ChangelogGenerator.GetChangelogPath()
			_ = cfg.Plugins.ChangelogGenerator.GetMode()
			_ = cfg.Plugins.ChangelogGenerator.GetFormat()
		}
		if cfg.Plugins.AuditLog != nil {
			_ = cfg.Plugins.AuditLog.GetPath()
			_ = cfg.Plugins.AuditLog.GetFormat()
		}
	}

	if cfg.Workspace != nil {
		for _, mod := range cfg.Workspace.Modules {
			_ = mod.IsEnabled()
			_ = cfg.IsModuleEnabled(mod.Name)
		}
	}
}

// verifyConfigConsistency checks that parsing is deterministic.
func verifyConfigConsistency(t *testing.T, input string) {
	t.Helper()

	var cfg1, cfg2 Config
	decoder1 := yaml.NewDecoder(bytes.NewReader([]byte(input)), yaml.Strict())
	decoder2 := yaml.NewDecoder(bytes.NewReader([]byte(input)), yaml.Strict())

	err1 := decoder1.Decode(&cfg1)
	err2 := decoder2.Decode(&cfg2)

	// Both should either succeed or fail
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("inconsistent error state: first=%v, second=%v", err1, err2)
	}

	// If both succeeded, basic fields should match
	if err1 == nil && err2 == nil {
		if cfg1.Path != cfg2.Path {
			t.Errorf("inconsistent path: %q vs %q", cfg1.Path, cfg2.Path)
		}
	}
}

// FuzzDiscoveryConfig tests discovery config parsing and defaults.
func FuzzDiscoveryConfig(f *testing.F) {
	seeds := []string{
		`workspace:
  discovery:
    enabled: true`,
		`workspace:
  discovery:
    enabled: false
    recursive: false`,
		`workspace:
  discovery:
    max_depth: 5`,
		`workspace:
  discovery:
    exclude:
      - node_modules
      - vendor`,
		`workspace: null`,
		``,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var cfg Config
		decoder := yaml.NewDecoder(bytes.NewReader([]byte(input)), yaml.Strict())

		if err := decoder.Decode(&cfg); err == nil {
			// GetDiscoveryConfig should never panic and always return valid config
			discovery := cfg.GetDiscoveryConfig()
			if discovery == nil {
				t.Error("GetDiscoveryConfig returned nil")
				return
			}

			// Enabled should always have a value (defaults to true)
			if discovery.Enabled == nil {
				t.Error("discovery.Enabled should not be nil after GetDiscoveryConfig")
			}

			// Recursive should always have a value (defaults to true)
			if discovery.Recursive == nil {
				t.Error("discovery.Recursive should not be nil after GetDiscoveryConfig")
			}

			// MaxDepth should always have a value (defaults to 10)
			if discovery.MaxDepth == nil {
				t.Error("discovery.MaxDepth should not be nil after GetDiscoveryConfig")
			}
		}
	})
}

// FuzzModuleConfig tests module config parsing.
func FuzzModuleConfig(f *testing.F) {
	type seedInput struct {
		name    string
		path    string
		enabled bool
	}

	seeds := []seedInput{
		{"core", "./core/.version", true},
		{"api", "./api/.version", false},
		{"", "", true},
		{"test-module", "/absolute/path/.version", true},
		{"module_with_underscore", "./path/.version", false},
	}

	for _, seed := range seeds {
		f.Add(seed.name, seed.path, seed.enabled)
	}

	f.Fuzz(func(t *testing.T, name, path string, enabled bool) {
		mod := ModuleConfig{
			Name:    name,
			Path:    path,
			Enabled: &enabled,
		}

		// IsEnabled should match the enabled value
		if mod.IsEnabled() != enabled {
			t.Errorf("IsEnabled() = %v, want %v", mod.IsEnabled(), enabled)
		}

		// Test with nil Enabled (should default to true)
		modNil := ModuleConfig{
			Name:    name,
			Path:    path,
			Enabled: nil,
		}
		if !modNil.IsEnabled() {
			t.Error("IsEnabled() with nil should return true")
		}
	})
}
