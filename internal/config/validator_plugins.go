package config

import (
	"context"
	"fmt"
	"regexp"
)

// validateYAMLSyntax checks if the config file is valid YAML.
func (v *Validator) validateYAMLSyntax(ctx context.Context) {
	if v.configPath == "" {
		v.addValidation("YAML Syntax", true, "No .sley.yaml file found, using defaults", false)
		return
	}

	if !v.validateFileExists(ctx, "YAML Syntax", "Config file", v.configPath) {
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

	v.validateTagManagerConfig()
	v.validateVersionValidatorConfig()
	v.validateDependencyCheckConfig(ctx)
	v.validateChangelogParserConfig(ctx)
	v.validateChangelogGeneratorConfig()
	v.validateReleaseGateConfig()
	v.validateAuditLogConfig()
}

// validateTagManagerConfig validates the tag-manager plugin configuration.
func (v *Validator) validateTagManagerConfig() {
	if v.cfg.Plugins.TagManager == nil || !v.cfg.Plugins.TagManager.Enabled {
		return
	}

	prefix := v.cfg.Plugins.TagManager.GetPrefix()
	if prefix != "" {
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
		label := fmt.Sprintf("File %d", i+1)

		v.validateFileExists(ctx, "Plugin: dependency-check", label, file.Path)

		if !validFormats[file.Format] {
			v.addValidation("Plugin: dependency-check", false,
				fmt.Sprintf("File %d: unknown format '%s'", i+1, file.Format), false)
		}

		if file.Format == "regex" && file.Pattern != "" {
			v.validateRegex("Plugin: dependency-check", fmt.Sprintf("File %d", i+1), file.Pattern)
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

	if !v.validateFileExists(ctx, "Plugin: changelog-parser", fmt.Sprintf("Changelog file '%s'", cfg.GetPath()), cfg.GetPath()) {
		return
	}

	// Validate priority setting
	if cfg.Priority != "" {
		validPriorities := map[string]bool{
			"changelog": true,
			"commits":   true,
		}
		if !v.validateEnum("Plugin: changelog-parser", "priority", cfg.Priority, validPriorities) {
			return
		}
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
	if !v.validateEnum("Plugin: changelog-parser", "format", cfg.Format, validFormats) {
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
	validModes := map[string]bool{
		"versioned": true,
		"unified":   true,
		"both":      true,
	}
	v.validateEnum("Plugin: changelog-generator", "mode", cfg.GetMode(), validModes)

	// Validate format
	validFormats := map[string]bool{
		"grouped":        true,
		"keepachangelog": true,
		"github":         true,
		"minimal":        true,
	}
	v.validateEnum("Plugin: changelog-generator", "format", cfg.GetFormat(), validFormats)

	// Validate merge-after
	validMergeAfter := map[string]bool{
		"immediate": true,
		"manual":    true,
		"prompt":    true,
	}
	v.validateEnum("Plugin: changelog-generator", "merge-after", cfg.GetMergeAfter(), validMergeAfter)

	// Validate repository config
	if cfg.Repository != nil {
		v.validateRepositoryConfig(cfg.Repository)
	}

	// Validate exclude patterns
	for i, pattern := range cfg.ExcludePatterns {
		v.validateRegex("Plugin: changelog-generator", fmt.Sprintf("Exclude pattern %d", i+1), pattern)
	}

	v.addValidation("Plugin: changelog-generator", true,
		fmt.Sprintf("Mode: %s, Format: %s", cfg.GetMode(), cfg.GetFormat()), false)
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

	if repo.Provider != "" {
		v.validateEnum("Plugin: changelog-generator", "repository provider", repo.Provider, validProviders)
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

	validFormats := map[string]bool{
		"json": true,
		"yaml": true,
	}

	format := cfg.GetFormat()
	if v.validateEnum("Plugin: audit-log", "format", format, validFormats) {
		v.addValidation("Plugin: audit-log", true,
			fmt.Sprintf("Audit log format: %s", format), false)
	}
}
