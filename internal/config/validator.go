package config

import (
	"context"

	"github.com/indaco/sley/internal/core"
)

// ValidationResult represents the result of a validation check.
type ValidationResult struct {
	// Category is the validation category (e.g., "YAML Syntax", "Plugin Config").
	Category string

	// Passed indicates if the check passed.
	Passed bool

	// Message provides details about the validation result.
	Message string

	// Warning indicates if this is a warning rather than an error.
	Warning bool
}

// Validator validates configuration files and settings.
type Validator struct {
	fs          core.FileSystem
	cfg         *Config
	configPath  string
	rootDir     string
	validations []ValidationResult
}

// NewValidator creates a new configuration validator.
// The rootDir parameter is the directory where .sley.yaml is located.
func NewValidator(fs core.FileSystem, cfg *Config, configPath string, rootDir string) *Validator {
	return &Validator{
		fs:          fs,
		cfg:         cfg,
		configPath:  configPath,
		rootDir:     rootDir,
		validations: make([]ValidationResult, 0),
	}
}

// Validate runs all validation checks and returns the results.
func (v *Validator) Validate(ctx context.Context) ([]ValidationResult, error) {
	// Reset validations
	v.validations = make([]ValidationResult, 0)

	// Validate YAML syntax (by trying to load it)
	v.validateYAMLSyntax(ctx)

	// Validate plugin configurations
	v.validatePluginConfigs(ctx)

	// Validate workspace configuration
	v.validateWorkspaceConfig(ctx)

	// Validate extension configurations
	v.validateExtensionConfigs(ctx)

	return v.validations, nil
}

// addValidation adds a validation result to the list.
func (v *Validator) addValidation(category string, passed bool, message string, warning bool) {
	v.validations = append(v.validations, ValidationResult{
		Category: category,
		Passed:   passed,
		Message:  message,
		Warning:  warning,
	})
}

// HasErrors returns true if any validation failed.
func HasErrors(results []ValidationResult) bool {
	for _, r := range results {
		if !r.Passed && !r.Warning {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of failed validations.
func ErrorCount(results []ValidationResult) int {
	count := 0
	for _, r := range results {
		if !r.Passed && !r.Warning {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warnings.
func WarningCount(results []ValidationResult) int {
	count := 0
	for _, r := range results {
		if r.Warning {
			count++
		}
	}
	return count
}
