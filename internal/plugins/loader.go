package plugins

import (
	"fmt"
	"os"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/changelogparser"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
)

// RegisterBuiltinPlugins registers all builtin plugins with the provided registry.
func RegisterBuiltinPlugins(cfg *config.Config, registry *PluginRegistry) {
	if cfg == nil || cfg.Plugins == nil {
		return
	}

	registerCommitParser(cfg.Plugins, registry)
	registerTagManager(cfg.Plugins, registry)
	registerVersionValidator(cfg.Plugins, registry)
	registerDependencyCheck(cfg.Plugins, registry)
	registerChangelogParser(cfg.Plugins, registry)
	registerChangelogGenerator(cfg.Plugins, registry)
	registerReleaseGate(cfg.Plugins, registry)
	registerAuditLog(cfg.Plugins, registry)
}

func registerCommitParser(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.CommitParser {
		plugin := commitparser.NewCommitParser()
		if err := registry.RegisterCommitParser(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerTagManager(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.TagManager != nil && plugins.TagManager.Enabled {
		tmCfg := &tagmanager.Config{
			Enabled:    true,
			AutoCreate: plugins.TagManager.GetAutoCreate(),
			Prefix:     plugins.TagManager.GetPrefix(),
			Annotate:   plugins.TagManager.GetAnnotate(),
			Push:       plugins.TagManager.Push,
		}
		plugin := tagmanager.NewTagManager(tmCfg)
		if err := registry.RegisterTagManager(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerVersionValidator(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.VersionValidator != nil && plugins.VersionValidator.Enabled {
		vvCfg := &versionvalidator.Config{
			Enabled: true,
			Rules:   convertValidationRules(plugins.VersionValidator.Rules),
		}
		plugin := versionvalidator.NewVersionValidator(vvCfg)
		if err := registry.RegisterVersionValidator(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerDependencyCheck(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.DependencyCheck != nil && plugins.DependencyCheck.Enabled {
		dcCfg := convertDependencyCheckConfig(plugins.DependencyCheck)
		plugin := dependencycheck.NewDependencyChecker(dcCfg)
		if err := registry.RegisterDependencyChecker(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerChangelogParser(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.ChangelogParser != nil && plugins.ChangelogParser.Enabled {
		clCfg := convertChangelogParserConfig(plugins.ChangelogParser)
		plugin := changelogparser.NewChangelogParser(clCfg)
		if err := registry.RegisterChangelogParser(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerChangelogGenerator(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.ChangelogGenerator != nil && plugins.ChangelogGenerator.Enabled {
		internalCfg := changeloggenerator.FromConfigStruct(plugins.ChangelogGenerator)
		plugin, err := changeloggenerator.NewChangelogGenerator(internalCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create changelog generator: %v\n", err)
			return
		}
		if err := registry.RegisterChangelogGenerator(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerReleaseGate(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.ReleaseGate != nil && plugins.ReleaseGate.Enabled {
		rgCfg := convertReleaseGateConfig(plugins.ReleaseGate)
		plugin := releasegate.NewReleaseGate(rgCfg)
		if err := registry.RegisterReleaseGate(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

func registerAuditLog(plugins *config.PluginConfig, registry *PluginRegistry) {
	if plugins.AuditLog != nil && plugins.AuditLog.Enabled {
		internalCfg := auditlog.FromConfigStruct(plugins.AuditLog)
		plugin := auditlog.NewAuditLog(internalCfg)
		if err := registry.RegisterAuditLog(plugin); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

// convertValidationRules converts config rules to versionvalidator rules.
func convertValidationRules(configRules []config.ValidationRule) []versionvalidator.Rule {
	rules := make([]versionvalidator.Rule, len(configRules))
	for i, r := range configRules {
		rules[i] = versionvalidator.Rule{
			Type:    versionvalidator.RuleType(r.Type),
			Pattern: r.Pattern,
			Value:   r.Value,
			Enabled: r.Enabled,
			Branch:  r.Branch,
			Allowed: r.Allowed,
		}
	}
	return rules
}

// convertDependencyCheckConfig converts config to dependencycheck config.
func convertDependencyCheckConfig(cfg *config.DependencyCheckConfig) *dependencycheck.Config {
	files := make([]dependencycheck.FileConfig, len(cfg.Files))
	for i, f := range cfg.Files {
		files[i] = dependencycheck.FileConfig{
			Path:    f.Path,
			Field:   f.Field,
			Format:  f.Format,
			Pattern: f.Pattern,
		}
	}
	return &dependencycheck.Config{
		Enabled:  cfg.Enabled,
		AutoSync: cfg.AutoSync,
		Files:    files,
	}
}

// convertChangelogParserConfig converts config to changelogparser config.
func convertChangelogParserConfig(cfg *config.ChangelogParserConfig) *changelogparser.Config {
	return &changelogparser.Config{
		Enabled:                  cfg.Enabled,
		Path:                     cfg.Path,
		RequireUnreleasedSection: cfg.RequireUnreleasedSection,
		InferBumpType:            cfg.InferBumpType,
		Priority:                 cfg.Priority,
	}
}

// convertReleaseGateConfig converts config to releasegate config.
func convertReleaseGateConfig(cfg *config.ReleaseGateConfig) *releasegate.Config {
	return &releasegate.Config{
		Enabled:              cfg.Enabled,
		RequireCleanWorktree: cfg.RequireCleanWorktree,
		RequireCIPass:        cfg.RequireCIPass,
		BlockedOnWIPCommits:  cfg.BlockedOnWIPCommits,
		AllowedBranches:      cfg.AllowedBranches,
		BlockedBranches:      cfg.BlockedBranches,
	}
}
