package initcmd

import (
	"fmt"
	"slices"
	"strings"
)

// Template represents a pre-configured set of plugins for common use cases.
type Template struct {
	Name        string
	Description string
	Plugins     []string
}

// AllTemplates returns all available templates.
func AllTemplates() []Template {
	return []Template{
		{
			Name:        "basic",
			Description: "Minimal setup with commit analysis only",
			Plugins:     []string{"commit-parser"},
		},
		{
			Name:        "git",
			Description: "Standard git workflow with tagging",
			Plugins:     []string{"commit-parser", "tag-manager"},
		},
		{
			Name:        "automation",
			Description: "Automated releases with changelog generation",
			Plugins:     []string{"commit-parser", "tag-manager", "changelog-generator"},
		},
		{
			Name:        "strict",
			Description: "Enforced policies with release gates",
			Plugins:     []string{"commit-parser", "tag-manager", "version-validator", "release-gate"},
		},
		{
			Name:        "full",
			Description: "All plugins enabled for maximum automation",
			Plugins: []string{
				"commit-parser",
				"tag-manager",
				"version-validator",
				"dependency-check",
				"changelog-generator",
				"release-gate",
				"audit-log",
			},
		},
	}
}

// TemplateNames returns the names of all available templates.
func TemplateNames() []string {
	templates := AllTemplates()
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name
	}
	return names
}

// GetTemplate returns the template with the given name, or an error if not found.
func GetTemplate(name string) (*Template, error) {
	for _, t := range AllTemplates() {
		if t.Name == name {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unknown template %q (available: %s)", name, strings.Join(TemplateNames(), ", "))
}

// IsValidTemplate checks if the given name is a valid template.
func IsValidTemplate(name string) bool {
	return slices.Contains(TemplateNames(), name)
}
