package plugins

import (
	"testing"
)

func TestGetBuiltinPlugins(t *testing.T) {
	t.Parallel()
	plugins := GetBuiltinPlugins()

	// Should return all 8 built-in plugins
	if len(plugins) != 8 {
		t.Errorf("expected 8 built-in plugins, got %d", len(plugins))
	}

	// Verify expected plugin types are present
	expectedTypes := map[PluginType]bool{
		TypeCommitParser:       false,
		TypeTagManager:         false,
		TypeVersionValidator:   false,
		TypeDependencyChecker:  false,
		TypeChangelogParser:    false,
		TypeChangelogGenerator: false,
		TypeReleaseGate:        false,
		TypeAuditLog:           false,
	}

	for _, p := range plugins {
		if _, ok := expectedTypes[p.Type]; ok {
			expectedTypes[p.Type] = true
		}
	}

	for pluginType, found := range expectedTypes {
		if !found {
			t.Errorf("expected plugin type %q not found", pluginType)
		}
	}
}

func TestGetBuiltinPlugins_ReturnsCopy(t *testing.T) {
	t.Parallel()
	plugins1 := GetBuiltinPlugins()
	plugins2 := GetBuiltinPlugins()

	// Modify the first slice
	if len(plugins1) > 0 {
		plugins1[0].Name = "modified"
	}

	// Second slice should not be affected
	if len(plugins2) > 0 && plugins2[0].Name == "modified" {
		t.Error("GetBuiltinPlugins should return a copy, not a reference")
	}
}

func TestBuiltinPluginMetadata(t *testing.T) {
	t.Parallel()
	plugins := GetBuiltinPlugins()

	for _, p := range plugins {
		t.Run(p.Name, func(t *testing.T) {
			t.Parallel(
			// All plugins should have required fields
			)

			if p.Name == "" {
				t.Error("plugin name should not be empty")
			}
			if p.Type == "" {
				t.Error("plugin type should not be empty")
			}
			if p.Description == "" {
				t.Error("plugin description should not be empty")
			}
			if p.Version == "" {
				t.Error("plugin version should not be empty")
			}
			if p.ConfigPath == "" {
				t.Error("plugin config path should not be empty")
			}

			// Config path should follow expected pattern
			expectedConfigPath := "plugins." + p.Name
			if p.ConfigPath != expectedConfigPath {
				t.Errorf("expected config path %q, got %q", expectedConfigPath, p.ConfigPath)
			}
		})
	}
}

func TestPluginTypeConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pluginType PluginType
		expected   string
	}{
		{TypeCommitParser, "commit-parser"},
		{TypeTagManager, "tag-manager"},
		{TypeVersionValidator, "version-validator"},
		{TypeDependencyChecker, "dependency-check"},
		{TypeChangelogParser, "changelog-parser"},
		{TypeChangelogGenerator, "changelog-generator"},
		{TypeReleaseGate, "release-gate"},
		{TypeAuditLog, "audit-log"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if string(tt.pluginType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.pluginType))
			}
		})
	}
}
