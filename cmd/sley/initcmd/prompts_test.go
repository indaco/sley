package initcmd

import (
	"testing"
)

func TestAllPluginOptions(t *testing.T) {
	options := AllPluginOptions()

	if len(options) == 0 {
		t.Fatal("expected non-empty plugin options")
	}

	// Verify expected plugins are present
	expectedPlugins := map[string]bool{
		"commit-parser":       true,
		"tag-manager":         true,
		"version-validator":   true,
		"dependency-check":    true,
		"changelog-parser":    true,
		"changelog-generator": true,
		"release-gate":        true,
		"audit-log":           true,
	}

	for _, opt := range options {
		if !expectedPlugins[opt.Name] {
			t.Errorf("unexpected plugin: %s", opt.Name)
		}
		delete(expectedPlugins, opt.Name)

		// Verify all options have descriptions
		if opt.Description == "" {
			t.Errorf("plugin %s missing description", opt.Name)
		}
	}

	// Verify no expected plugins are missing
	for name := range expectedPlugins {
		t.Errorf("missing expected plugin: %s", name)
	}
}

func TestDefaultPluginNames(t *testing.T) {
	defaults := DefaultPluginNames()

	if len(defaults) == 0 {
		t.Fatal("expected at least one default plugin")
	}

	// Verify commit-parser and tag-manager are defaults
	expectedDefaults := map[string]bool{
		"commit-parser": false,
		"tag-manager":   false,
	}

	for _, name := range defaults {
		if _, ok := expectedDefaults[name]; ok {
			expectedDefaults[name] = true
		}
	}

	for name, found := range expectedDefaults {
		if !found {
			t.Errorf("expected %s to be in defaults", name)
		}
	}
}

func TestPluginOption_Defaults(t *testing.T) {
	options := AllPluginOptions()

	// Verify commit-parser and tag-manager have Default=true
	for _, opt := range options {
		if opt.Name == "commit-parser" || opt.Name == "tag-manager" {
			if !opt.Default {
				t.Errorf("expected %s to have Default=true", opt.Name)
			}
		}
	}
}
