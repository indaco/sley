package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
func (v *Validator) Validate() ([]ValidationResult, error) {
	// Reset validations
	v.validations = make([]ValidationResult, 0)

	// Validate YAML syntax (by trying to load it)
	v.validateYAMLSyntax()

	// Validate plugin configurations
	v.validatePluginConfigs()

	// Validate workspace configuration
	v.validateWorkspaceConfig()

	// Validate extension configurations
	v.validateExtensionConfigs()

	return v.validations, nil
}

// validateYAMLSyntax checks if the config file is valid YAML.
func (v *Validator) validateYAMLSyntax() {
	if v.configPath == "" {
		// No config file, use defaults
		v.addValidation("YAML Syntax", true, "No .sley.yaml file found, using defaults", false)
		return
	}

	// Check if file exists
	if _, err := v.fs.Stat(v.configPath); err != nil {
		if os.IsNotExist(err) {
			v.addValidation("YAML Syntax", true, "No .sley.yaml file found, using defaults", false)
		} else {
			v.addValidation("YAML Syntax", false, fmt.Sprintf("Failed to access config file: %v", err), false)
		}
		return
	}

	// If we got here, the config was successfully loaded (validated in LoadConfigFn)
	v.addValidation("YAML Syntax", true, "Configuration file is valid YAML", false)
}

// validatePluginConfigs validates plugin-specific configurations.
func (v *Validator) validatePluginConfigs() {
	if v.cfg == nil || v.cfg.Plugins == nil {
		v.addValidation("Plugin Configuration", true, "No plugin configuration found (using defaults)", false)
		return
	}

	// Validate tag-manager plugin
	v.validateTagManagerConfig()

	// Validate version-validator plugin
	v.validateVersionValidatorConfig()

	// Validate dependency-check plugin
	v.validateDependencyCheckConfig()

	// Validate changelog-parser plugin
	v.validateChangelogParserConfig()

	// Validate changelog-generator plugin
	v.validateChangelogGeneratorConfig()

	// Validate release-gate plugin
	v.validateReleaseGateConfig()

	// Validate audit-log plugin
	v.validateAuditLogConfig()
}

// validateTagManagerConfig validates the tag-manager plugin configuration.
func (v *Validator) validateTagManagerConfig() {
	if v.cfg.Plugins.TagManager == nil || !v.cfg.Plugins.TagManager.Enabled {
		return
	}

	cfg := v.cfg.Plugins.TagManager

	// Validate prefix pattern (should be a valid tag prefix)
	prefix := cfg.GetPrefix()
	if prefix != "" {
		// Check if prefix contains invalid characters
		if strings.ContainsAny(prefix, " \t\n\r/\\") {
			v.addValidation("Plugin: tag-manager", false,
				fmt.Sprintf("Invalid prefix '%s': contains whitespace or path separators", prefix), false)
		} else {
			v.addValidation("Plugin: tag-manager", true,
				fmt.Sprintf("Tag prefix '%s' is valid", prefix), false)
		}
	}
}

// validateVersionValidatorConfig validates the version-validator plugin configuration.
func (v *Validator) validateVersionValidatorConfig() {
	if v.cfg.Plugins.VersionValidator == nil || !v.cfg.Plugins.VersionValidator.Enabled {
		return
	}

	cfg := v.cfg.Plugins.VersionValidator

	// Validate rules
	if len(cfg.Rules) == 0 {
		v.addValidation("Plugin: version-validator", true,
			"Version validator enabled but no rules configured", true)
		return
	}

	validRuleTypes := map[string]bool{
		"pre-release-format": true,
		"major-version-max":  true,
		"branch-constraint":  true,
	}

	for i, rule := range cfg.Rules {
		// Check if rule type is valid
		if !validRuleTypes[rule.Type] {
			v.addValidation("Plugin: version-validator", false,
				fmt.Sprintf("Rule %d: unknown rule type '%s'", i+1, rule.Type), false)
			continue
		}

		// Validate pattern for pre-release-format rules
		if rule.Type == "pre-release-format" && rule.Pattern != "" {
			if _, err := regexp.Compile(rule.Pattern); err != nil {
				v.addValidation("Plugin: version-validator", false,
					fmt.Sprintf("Rule %d: invalid regex pattern: %v", i+1, err), false)
			}
		}

		// Validate branch-constraint rules
		if rule.Type == "branch-constraint" {
			if rule.Branch == "" {
				v.addValidation("Plugin: version-validator", false,
					fmt.Sprintf("Rule %d: branch-constraint requires 'branch' field", i+1), false)
			}
			if len(rule.Allowed) == 0 {
				v.addValidation("Plugin: version-validator", false,
					fmt.Sprintf("Rule %d: branch-constraint requires 'allowed' field", i+1), false)
			}
		}
	}

	v.addValidation("Plugin: version-validator", true,
		fmt.Sprintf("Configured with %d validation rule(s)", len(cfg.Rules)), false)
}

// validateDependencyCheckConfig validates the dependency-check plugin configuration.
func (v *Validator) validateDependencyCheckConfig() {
	if v.cfg.Plugins.DependencyCheck == nil || !v.cfg.Plugins.DependencyCheck.Enabled {
		return
	}

	cfg := v.cfg.Plugins.DependencyCheck

	if len(cfg.Files) == 0 {
		v.addValidation("Plugin: dependency-check", true,
			"Dependency check enabled but no files configured", true)
		return
	}

	validFormats := map[string]bool{
		"json":  true,
		"yaml":  true,
		"toml":  true,
		"raw":   true,
		"regex": true,
	}

	for i, file := range cfg.Files {
		// Check if file exists
		filePath := file.Path
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(v.rootDir, filePath)
		}

		if _, err := v.fs.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				v.addValidation("Plugin: dependency-check", false,
					fmt.Sprintf("File %d: '%s' does not exist", i+1, file.Path), false)
			} else {
				v.addValidation("Plugin: dependency-check", false,
					fmt.Sprintf("File %d: cannot access '%s': %v", i+1, file.Path, err), false)
			}
			continue
		}

		// Validate format
		if !validFormats[file.Format] {
			v.addValidation("Plugin: dependency-check", false,
				fmt.Sprintf("File %d: unknown format '%s'", i+1, file.Format), false)
		}

		// Validate regex pattern if format is regex
		if file.Format == "regex" && file.Pattern != "" {
			if _, err := regexp.Compile(file.Pattern); err != nil {
				v.addValidation("Plugin: dependency-check", false,
					fmt.Sprintf("File %d: invalid regex pattern: %v", i+1, err), false)
			}
		}
	}

	v.addValidation("Plugin: dependency-check", true,
		fmt.Sprintf("Configured to check %d file(s)", len(cfg.Files)), false)
}

// validateChangelogParserConfig validates the changelog-parser plugin configuration.
func (v *Validator) validateChangelogParserConfig() {
	if v.cfg.Plugins.ChangelogParser == nil || !v.cfg.Plugins.ChangelogParser.Enabled {
		return
	}

	cfg := v.cfg.Plugins.ChangelogParser

	// Validate changelog file path
	changelogPath := cfg.Path
	if changelogPath == "" {
		changelogPath = "CHANGELOG.md"
	}
	if !filepath.IsAbs(changelogPath) {
		changelogPath = filepath.Join(v.rootDir, changelogPath)
	}

	if _, err := v.fs.Stat(changelogPath); err != nil {
		if os.IsNotExist(err) {
			v.addValidation("Plugin: changelog-parser", false,
				fmt.Sprintf("Changelog file '%s' does not exist", cfg.Path), false)
		} else {
			v.addValidation("Plugin: changelog-parser", false,
				fmt.Sprintf("Cannot access changelog file: %v", err), false)
		}
		return
	}

	// Validate priority setting
	if cfg.Priority != "" && cfg.Priority != "changelog" && cfg.Priority != "commits" {
		v.addValidation("Plugin: changelog-parser", false,
			fmt.Sprintf("Invalid priority '%s': must be 'changelog' or 'commits'", cfg.Priority), false)
	} else {
		v.addValidation("Plugin: changelog-parser", true,
			fmt.Sprintf("Changelog file '%s' is accessible", cfg.Path), false)
	}
}

// validateChangelogGeneratorConfig validates the changelog-generator plugin configuration.
func (v *Validator) validateChangelogGeneratorConfig() {
	if v.cfg.Plugins.ChangelogGenerator == nil || !v.cfg.Plugins.ChangelogGenerator.Enabled {
		return
	}

	cfg := v.cfg.Plugins.ChangelogGenerator

	// Validate mode
	mode := cfg.GetMode()
	validModes := map[string]bool{
		"versioned": true,
		"unified":   true,
		"both":      true,
	}
	if !validModes[mode] {
		v.addValidation("Plugin: changelog-generator", false,
			fmt.Sprintf("Invalid mode '%s': must be 'versioned', 'unified', or 'both'", mode), false)
	}

	// Validate format
	format := cfg.GetFormat()
	validFormats := map[string]bool{
		"grouped":        true,
		"keepachangelog": true,
	}
	if !validFormats[format] {
		v.addValidation("Plugin: changelog-generator", false,
			fmt.Sprintf("Invalid format '%s': must be 'grouped' or 'keepachangelog'", format), false)
	}

	// Validate repository config
	if cfg.Repository != nil {
		v.validateRepositoryConfig(cfg.Repository)
	}

	// Validate exclude patterns
	for i, pattern := range cfg.ExcludePatterns {
		if _, err := regexp.Compile(pattern); err != nil {
			v.addValidation("Plugin: changelog-generator", false,
				fmt.Sprintf("Exclude pattern %d: invalid regex: %v", i+1, err), false)
		}
	}

	v.addValidation("Plugin: changelog-generator", true,
		fmt.Sprintf("Mode: %s, Format: %s", mode, format), false)
}

// validateRepositoryConfig validates repository configuration for changelog generator.
func (v *Validator) validateRepositoryConfig(repo *RepositoryConfig) {
	validProviders := map[string]bool{
		"github":    true,
		"gitlab":    true,
		"codeberg":  true,
		"gitea":     true,
		"bitbucket": true,
		"custom":    true,
	}

	if repo.Provider != "" && !validProviders[repo.Provider] {
		v.addValidation("Plugin: changelog-generator", false,
			fmt.Sprintf("Invalid repository provider '%s'", repo.Provider), false)
	}

	// If provider is custom, require host
	if repo.Provider == "custom" && repo.Host == "" {
		v.addValidation("Plugin: changelog-generator", false,
			"Custom provider requires 'host' field", false)
	}
}

// validateReleaseGateConfig validates the release-gate plugin configuration.
func (v *Validator) validateReleaseGateConfig() {
	if v.cfg.Plugins.ReleaseGate == nil || !v.cfg.Plugins.ReleaseGate.Enabled {
		return
	}

	cfg := v.cfg.Plugins.ReleaseGate

	// Check for conflicting branch configurations
	if len(cfg.AllowedBranches) > 0 && len(cfg.BlockedBranches) > 0 {
		v.addValidation("Plugin: release-gate", true,
			"Both allowed and blocked branches configured (blocked takes precedence)", true)
	}

	v.addValidation("Plugin: release-gate", true,
		"Release gate configuration is valid", false)
}

// validateAuditLogConfig validates the audit-log plugin configuration.
func (v *Validator) validateAuditLogConfig() {
	if v.cfg.Plugins.AuditLog == nil || !v.cfg.Plugins.AuditLog.Enabled {
		return
	}

	cfg := v.cfg.Plugins.AuditLog

	// Validate format
	format := cfg.GetFormat()
	if format != "json" && format != "yaml" {
		v.addValidation("Plugin: audit-log", false,
			fmt.Sprintf("Invalid format '%s': must be 'json' or 'yaml'", format), false)
	} else {
		v.addValidation("Plugin: audit-log", true,
			fmt.Sprintf("Audit log format: %s", format), false)
	}
}

// validateWorkspaceConfig validates workspace/multi-module configuration.
func (v *Validator) validateWorkspaceConfig() {
	if v.cfg == nil || v.cfg.Workspace == nil {
		return
	}

	// Validate discovery configuration
	if v.cfg.Workspace.Discovery != nil {
		v.validateDiscoveryConfig()
	}

	// Validate explicit modules
	if len(v.cfg.Workspace.Modules) > 0 {
		v.validateExplicitModules()
	}
}

// validateDiscoveryConfig validates module discovery settings.
func (v *Validator) validateDiscoveryConfig() {
	discovery := v.cfg.Workspace.Discovery

	// Validate max depth
	if discovery.MaxDepth != nil && *discovery.MaxDepth < 0 {
		v.addValidation("Workspace: Discovery", false,
			"max_depth cannot be negative", false)
	}

	// Validate exclude patterns (glob patterns)
	for i, pattern := range discovery.Exclude {
		// Basic validation - check for obviously invalid patterns
		if strings.Contains(pattern, "**/**/**") {
			v.addValidation("Workspace: Discovery", true,
				fmt.Sprintf("Exclude pattern %d: '%s' may be overly broad", i+1, pattern), true)
		}
	}

	v.addValidation("Workspace: Discovery", true,
		fmt.Sprintf("Discovery configured with %d exclude pattern(s)", len(discovery.Exclude)), false)
}

// validateExplicitModules validates explicitly configured modules.
func (v *Validator) validateExplicitModules() {
	modules := v.cfg.Workspace.Modules

	// Check for duplicate module names
	names := make(map[string]bool)
	for i, mod := range modules {
		if names[mod.Name] {
			v.addValidation("Workspace: Modules", false,
				fmt.Sprintf("Module %d: duplicate name '%s'", i+1, mod.Name), false)
		}
		names[mod.Name] = true

		// Check if module path exists
		modPath := mod.Path
		if !filepath.IsAbs(modPath) {
			modPath = filepath.Join(v.rootDir, modPath)
		}

		if _, err := v.fs.Stat(modPath); err != nil {
			if os.IsNotExist(err) {
				v.addValidation("Workspace: Modules", false,
					fmt.Sprintf("Module '%s': path '%s' does not exist", mod.Name, mod.Path), false)
			} else {
				v.addValidation("Workspace: Modules", false,
					fmt.Sprintf("Module '%s': cannot access path '%s': %v", mod.Name, mod.Path, err), false)
			}
		}
	}

	enabledCount := 0
	for _, mod := range modules {
		if mod.IsEnabled() {
			enabledCount++
		}
	}

	v.addValidation("Workspace: Modules", true,
		fmt.Sprintf("Configured with %d module(s) (%d enabled)", len(modules), enabledCount), false)
}

// validateExtensionConfigs validates extension configurations.
func (v *Validator) validateExtensionConfigs() {
	if v.cfg == nil || len(v.cfg.Extensions) == 0 {
		return
	}

	pathErrorCount := 0
	manifestErrorCount := 0

	for i, ext := range v.cfg.Extensions {
		// Check if extension path exists
		extPath := ext.Path
		if !filepath.IsAbs(extPath) {
			extPath = filepath.Join(v.rootDir, extPath)
		}

		if _, err := v.fs.Stat(extPath); err != nil {
			if os.IsNotExist(err) {
				v.addValidation("Extensions", false,
					fmt.Sprintf("Extension %d ('%s'): path '%s' does not exist", i+1, ext.Name, ext.Path), false)
			} else {
				v.addValidation("Extensions", false,
					fmt.Sprintf("Extension %d ('%s'): cannot access path '%s': %v", i+1, ext.Name, ext.Path, err), false)
			}
			pathErrorCount++
			continue
		}

		// Check for extension.yaml manifest file if the extension is enabled
		if ext.Enabled {
			manifestPath := filepath.Join(extPath, "extension.yaml")
			if _, err := v.fs.Stat(manifestPath); err != nil {
				if os.IsNotExist(err) {
					v.addValidation("Extensions", false,
						fmt.Sprintf("Extension %d ('%s'): manifest file 'extension.yaml' not found", i+1, ext.Name), false)
				} else {
					v.addValidation("Extensions", false,
						fmt.Sprintf("Extension %d ('%s'): cannot access manifest: %v", i+1, ext.Name, err), false)
				}
				manifestErrorCount++
			}
		}
	}

	enabledCount := 0
	for _, ext := range v.cfg.Extensions {
		if ext.Enabled {
			enabledCount++
		}
	}

	if pathErrorCount == 0 && manifestErrorCount == 0 {
		v.addValidation("Extensions", true,
			fmt.Sprintf("Configured with %d extension(s) (%d enabled)", len(v.cfg.Extensions), enabledCount), false)
	}
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
