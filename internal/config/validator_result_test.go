package config

import (
	"context"
	"testing"

	"github.com/indaco/sley/internal/core"
)

/* ------------------------------------------------------------------------- */
/* RESULT HELPER TESTS                                                       */
/* ------------------------------------------------------------------------- */

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name    string
		results []ValidationResult
		want    bool
	}{
		{
			name:    "no results",
			results: []ValidationResult{},
			want:    false,
		},
		{
			name: "all passed",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: true, Warning: false},
			},
			want: false,
		},
		{
			name: "has errors",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: false, Warning: false},
			},
			want: true,
		},
		{
			name: "only warnings",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: true, Warning: true},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCount(t *testing.T) {
	tests := []struct {
		name    string
		results []ValidationResult
		want    int
	}{
		{
			name:    "no results",
			results: []ValidationResult{},
			want:    0,
		},
		{
			name: "no errors",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: true, Warning: true},
			},
			want: 0,
		},
		{
			name: "has errors",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: false, Warning: false},
				{Passed: false, Warning: false},
				{Passed: true, Warning: true},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorCount(tt.results); got != tt.want {
				t.Errorf("ErrorCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWarningCount(t *testing.T) {
	tests := []struct {
		name    string
		results []ValidationResult
		want    int
	}{
		{
			name:    "no results",
			results: []ValidationResult{},
			want:    0,
		},
		{
			name: "no warnings",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: false, Warning: false},
			},
			want: 0,
		},
		{
			name: "has warnings",
			results: []ValidationResult{
				{Passed: true, Warning: false},
				{Passed: true, Warning: true},
				{Passed: true, Warning: true},
				{Passed: false, Warning: false},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WarningCount(tt.results); got != tt.want {
				t.Errorf("WarningCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* NO PLUGINS / EDGE CASE TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestValidator_ValidateNoPlugins(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "no plugin config",
			config: &Config{
				Plugins: nil,
			},
		},
		{
			name:   "nil config",
			config: &Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Should have YAML syntax validation at minimum
			if len(results) == 0 {
				t.Error("expected at least YAML syntax validation result")
			}

			// Find plugin configuration validation
			for _, r := range results {
				if r.Category == "Plugin Configuration" {
					if !r.Passed {
						t.Errorf("expected plugin configuration to pass with no config")
					}
				}
			}
		})
	}
}

func TestValidator_ValidateVersionValidatorNoRules(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Plugins: &PluginConfig{
			VersionValidator: &VersionValidatorConfig{
				Enabled: true,
				Rules:   []ValidationRule{},
			},
		},
	}

	fs := core.NewMockFileSystem()
	validator := NewValidator(fs, config, "", ".")

	results, err := validator.Validate(ctx)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should have a warning about no rules
	hasWarning := false
	for _, r := range results {
		if r.Category == "Plugin: version-validator" && r.Warning {
			hasWarning = true
			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for version-validator with no rules")
	}
}

func TestValidator_ValidateDependencyCheckNoFiles(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Plugins: &PluginConfig{
			DependencyCheck: &DependencyCheckConfig{
				Enabled: true,
				Files:   []DependencyFileConfig{},
			},
		},
	}

	fs := core.NewMockFileSystem()
	validator := NewValidator(fs, config, "", ".")

	results, err := validator.Validate(ctx)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should have a warning about no files
	hasWarning := false
	for _, r := range results {
		if r.Category == "Plugin: dependency-check" && r.Warning {
			hasWarning = true
			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for dependency-check with no files")
	}
}
