package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// validateYAMLSyntax checks if the config file is valid YAML.
func (v *Validator) validateYAMLSyntax(ctx context.Context) {
	if v.configPath == "" {
		// No config file, use defaults
		v.addValidation("YAML Syntax", true, "No .sley.yaml file found, using defaults", false)
		return
	}

	// Check if file exists
	if _, err := v.fs.Stat(ctx, v.configPath); err != nil {
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
func (v *Validator) validatePluginConfigs(ctx context.Context) {
	if v.cfg == nil || v.cfg.Plugins == nil {
		v.addValidation("Plugin Configuration", true, "No plugin configuration found (using defaults)", false)
		return
	}

	// Validate tag-manager plugin
	v.validateTagManagerConfig()

	// Validate version-validator plugin
	v.validateVersionValidatorConfig()

	// Validate dependency-check plugin
	v.validateDependencyCheckConfig(ctx)

	// Validate changelog-parser plugin
	v.validateChangelogParserConfig(ctx)

	// Validate changelog-generator plugin
	v.validateChangelogGeneratorConfig()

	// Validate release-gate plugin
	v.validateReleaseGateConfig()

	// Validate audit-log plugin
	v.validateAuditLogConfig(ctx)
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
		if containsInvalidTagChars(prefix) {
			v.addValidation("Plugin: tag-manager", false,
				fmt.Sprintf("Invalid prefix '%s': contains whitespace or path separators", prefix), false)
		} else {
			v.addValidation("Plugin: tag-manager", true,
				fmt.Sprintf("Tag prefix '%s' is valid", prefix), false)
		}
	}
}

// containsInvalidTagChars checks if a string contains invalid tag characters.
func containsInvalidTagChars(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '/' || r == '\\' {
			return true
		}
	}
	return false
}

// validVersionValidatorRuleTypes defines the set of valid rule types for version-validator.
var validVersionValidatorRuleTypes = map[string]bool{
	"pre-release-format":         true,
	"major-version-max":          true,
	"minor-version-max":          true,
	"patch-version-max":          true,
	"require-pre-release-for-0x": true,
	"branch-constraint":          true,
	"no-major-bump":              true,
	"no-minor-bump":              true,
	"no-patch-bump":              true,
	"max-prerelease-iterations":  true,
	"require-even-minor":         true,
}

// validateVersionValidatorConfig validates the version-validator plugin configuration.
func (v *Validator) validateVersionValidatorConfig() {
	if v.cfg.Plugins.VersionValidator == nil || !v.cfg.Plugins.VersionValidator.Enabled {
		return
	}

	cfg := v.cfg.Plugins.VersionValidator

	if len(cfg.Rules) == 0 {
		v.addValidation("Plugin: version-validator", true,
			"Version validator enabled but no rules configured", true)
		return
	}

	for i, rule := range cfg.Rules {
		v.validateVersionValidatorRule(i, rule)
	}

	v.addValidation("Plugin: version-validator", true,
		fmt.Sprintf("Configured with %d validation rule(s)", len(cfg.Rules)), false)
}

// validateVersionValidatorRule validates a single version-validator rule.
func (v *Validator) validateVersionValidatorRule(idx int, rule ValidationRule) {
	ruleNum := idx + 1

	if !validVersionValidatorRuleTypes[rule.Type] {
		v.addValidation("Plugin: version-validator", false,
			fmt.Sprintf("Rule %d: unknown rule type '%s'", ruleNum, rule.Type), false)
		return
	}

	switch rule.Type {
	case "pre-release-format":
		v.validatePreReleaseFormatRule(ruleNum, rule)
	case "branch-constraint":
		v.validateBranchConstraintRule(ruleNum, rule)
	case "max-prerelease-iterations":
		v.validateValueRequiredRule(ruleNum, rule)
	case "major-version-max", "minor-version-max", "patch-version-max":
		v.validateValueRequiredRule(ruleNum, rule)
	}
}

// validatePreReleaseFormatRule validates a pre-release-format rule.
func (v *Validator) validatePreReleaseFormatRule(ruleNum int, rule ValidationRule) {
	if rule.Pattern == "" {
		return
	}
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		v.addValidation("Plugin: version-validator", false,
			fmt.Sprintf("Rule %d: invalid regex pattern: %v", ruleNum, err), false)
	}
}

// validateBranchConstraintRule validates a branch-constraint rule.
func (v *Validator) validateBranchConstraintRule(ruleNum int, rule ValidationRule) {
	if rule.Branch == "" {
		v.addValidation("Plugin: version-validator", false,
			fmt.Sprintf("Rule %d: branch-constraint requires 'branch' field", ruleNum), false)
	}
	if len(rule.Allowed) == 0 {
		v.addValidation("Plugin: version-validator", false,
			fmt.Sprintf("Rule %d: branch-constraint requires 'allowed' field", ruleNum), false)
	}
}

// validateValueRequiredRule validates rules that require a positive value.
func (v *Validator) validateValueRequiredRule(ruleNum int, rule ValidationRule) {
	if rule.Value <= 0 {
		v.addValidation("Plugin: version-validator", true,
			fmt.Sprintf("Rule %d: %s has no value set (rule will be skipped)", ruleNum, rule.Type), true)
	}
}

// validateDependencyCheckConfig validates the dependency-check plugin configuration.
func (v *Validator) validateDependencyCheckConfig(ctx context.Context) {
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

		if _, err := v.fs.Stat(ctx, filePath); err != nil {
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
func (v *Validator) validateChangelogParserConfig(ctx context.Context) {
	if v.cfg.Plugins.ChangelogParser == nil || !v.cfg.Plugins.ChangelogParser.Enabled {
		return
	}

	cfg := v.cfg.Plugins.ChangelogParser

	// Validate changelog file path
	changelogPath := cfg.GetPath()
	if !filepath.IsAbs(changelogPath) {
		changelogPath = filepath.Join(v.rootDir, changelogPath)
	}

	if _, err := v.fs.Stat(ctx, changelogPath); err != nil {
		if os.IsNotExist(err) {
			v.addValidation("Plugin: changelog-parser", false,
				fmt.Sprintf("Changelog file '%s' does not exist", cfg.GetPath()), false)
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
		return
	}

	// Validate format setting
	validFormats := map[string]bool{
		"keepachangelog": true,
		"grouped":        true,
		"github":         true,
		"minimal":        true,
		"auto":           true,
		"":               true,
	}
	if !validFormats[cfg.Format] {
		v.addValidation("Plugin: changelog-parser", false,
			fmt.Sprintf("Invalid format '%s': must be 'keepachangelog', 'grouped', 'github', 'minimal', or 'auto'", cfg.Format), false)
		return
	}

	v.addValidation("Plugin: changelog-parser", true,
		fmt.Sprintf("Changelog file '%s' is accessible (format: %s)", cfg.GetPath(), cfg.GetFormat()), false)
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
		"github":         true,
		"minimal":        true,
	}
	if !validFormats[format] {
		v.addValidation("Plugin: changelog-generator", false,
			fmt.Sprintf("Invalid format '%s': must be 'grouped', 'keepachangelog', 'github', or 'minimal'", format), false)
	}

	// Validate merge-after
	mergeAfter := cfg.GetMergeAfter()
	validMergeAfter := map[string]bool{
		"immediate": true,
		"manual":    true,
		"prompt":    true,
	}
	if !validMergeAfter[mergeAfter] {
		v.addValidation("Plugin: changelog-generator", false,
			fmt.Sprintf("Invalid merge-after '%s': must be 'immediate', 'manual', or 'prompt'", mergeAfter), false)
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
func (v *Validator) validateAuditLogConfig(_ context.Context) {
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
