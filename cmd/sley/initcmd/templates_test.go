package initcmd

import (
	"testing"
)

func TestAllTemplates(t *testing.T) {
	templates := AllTemplates()

	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

	// Verify each template has required fields
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template has empty name")
		}
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", tmpl.Name)
		}
		if len(tmpl.Plugins) == 0 {
			t.Errorf("template %q has no plugins", tmpl.Name)
		}
	}
}

func TestTemplateNames(t *testing.T) {
	names := TemplateNames()

	expectedNames := []string{"basic", "git", "automation", "strict", "full"}
	if len(names) != len(expectedNames) {
		t.Errorf("expected %d templates, got %d", len(expectedNames), len(names))
	}

	for i, expected := range expectedNames {
		if names[i] != expected {
			t.Errorf("template[%d]: expected %q, got %q", i, expected, names[i])
		}
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name            string
		templateName    string
		expectedPlugins []string
		expectError     bool
	}{
		{
			name:            "basic template",
			templateName:    "basic",
			expectedPlugins: []string{"commit-parser"},
		},
		{
			name:            "git template",
			templateName:    "git",
			expectedPlugins: []string{"commit-parser", "tag-manager"},
		},
		{
			name:            "automation template",
			templateName:    "automation",
			expectedPlugins: []string{"commit-parser", "tag-manager", "changelog-generator"},
		},
		{
			name:            "strict template",
			templateName:    "strict",
			expectedPlugins: []string{"commit-parser", "tag-manager", "version-validator", "release-gate"},
		},
		{
			name:         "full template has 7 plugins",
			templateName: "full",
			expectedPlugins: []string{
				"commit-parser",
				"tag-manager",
				"version-validator",
				"dependency-check",
				"changelog-generator",
				"release-gate",
				"audit-log",
			},
		},
		{
			name:         "unknown template returns error",
			templateName: "unknown",
			expectError:  true,
		},
		{
			name:         "empty template name returns error",
			templateName: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.templateName)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tmpl.Name != tt.templateName {
				t.Errorf("expected template name %q, got %q", tt.templateName, tmpl.Name)
			}

			if len(tmpl.Plugins) != len(tt.expectedPlugins) {
				t.Errorf("expected %d plugins, got %d: %v", len(tt.expectedPlugins), len(tmpl.Plugins), tmpl.Plugins)
				return
			}

			for i, expected := range tt.expectedPlugins {
				if tmpl.Plugins[i] != expected {
					t.Errorf("plugin[%d]: expected %q, got %q", i, expected, tmpl.Plugins[i])
				}
			}
		})
	}
}

func TestIsValidTemplate(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"basic", true},
		{"git", true},
		{"automation", true},
		{"strict", true},
		{"full", true},
		{"unknown", false},
		{"", false},
		{"BASIC", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTemplate(tt.name)
			if got != tt.expected {
				t.Errorf("IsValidTemplate(%q): expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}
