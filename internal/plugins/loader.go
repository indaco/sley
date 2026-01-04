package plugins

import (
	"fmt"
	"os"

	"github.com/indaco/verso/internal/config"
	"github.com/indaco/verso/internal/plugins/auditlog"
	"github.com/indaco/verso/internal/plugins/changeloggenerator"
	"github.com/indaco/verso/internal/plugins/changelogparser"
	"github.com/indaco/verso/internal/plugins/commitparser"
	"github.com/indaco/verso/internal/plugins/dependencycheck"
	"github.com/indaco/verso/internal/plugins/releasegate"
	"github.com/indaco/verso/internal/plugins/tagmanager"
	"github.com/indaco/verso/internal/plugins/versionvalidator"
)

func RegisterBuiltinPlugins(cfg *config.Config) {
	if cfg == nil || cfg.Plugins == nil {
		return
	}

	registerCommitParser(cfg.Plugins)
	registerTagManager(cfg.Plugins)
	registerVersionValidator(cfg.Plugins)
	registerDependencyCheck(cfg.Plugins)
	registerChangelogParser(cfg.Plugins)
	registerChangelogGenerator(cfg.Plugins)
	registerReleaseGate(cfg.Plugins)
	registerAuditLog(cfg.Plugins)
}

func registerCommitParser(plugins *config.PluginConfig) {
	if plugins.CommitParser {
		commitparser.Register()
	}
}

func registerTagManager(plugins *config.PluginConfig) {
	if plugins.TagManager != nil && plugins.TagManager.Enabled {
		tmCfg := &tagmanager.Config{
			Enabled:    true,
			AutoCreate: plugins.TagManager.GetAutoCreate(),
			Prefix:     plugins.TagManager.GetPrefix(),
			Annotate:   plugins.TagManager.GetAnnotate(),
			Push:       plugins.TagManager.Push,
		}
		tagmanager.Register(tmCfg)
	}
}

func registerVersionValidator(plugins *config.PluginConfig) {
	if plugins.VersionValidator != nil && plugins.VersionValidator.Enabled {
		vvCfg := &versionvalidator.Config{
			Enabled: true,
			Rules:   convertValidationRules(plugins.VersionValidator.Rules),
		}
		versionvalidator.Register(vvCfg)
	}
}

func registerDependencyCheck(plugins *config.PluginConfig) {
	if plugins.DependencyCheck != nil && plugins.DependencyCheck.Enabled {
		dcCfg := convertDependencyCheckConfig(plugins.DependencyCheck)
		dependencycheck.Register(dcCfg)
	}
}

func registerChangelogParser(plugins *config.PluginConfig) {
	if plugins.ChangelogParser != nil && plugins.ChangelogParser.Enabled {
		clCfg := convertChangelogParserConfig(plugins.ChangelogParser)
		changelogparser.Register(clCfg)
	}
}

func registerChangelogGenerator(plugins *config.PluginConfig) {
	if plugins.ChangelogGenerator != nil && plugins.ChangelogGenerator.Enabled {
		if err := changeloggenerator.Register(plugins.ChangelogGenerator); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to register changelog generator: %v\n", err)
		}
	}
}

func registerReleaseGate(plugins *config.PluginConfig) {
	if plugins.ReleaseGate != nil && plugins.ReleaseGate.Enabled {
		rgCfg := convertReleaseGateConfig(plugins.ReleaseGate)
		releasegate.Register(rgCfg)
	}
}

func registerAuditLog(plugins *config.PluginConfig) {
	if plugins.AuditLog != nil && plugins.AuditLog.Enabled {
		auditlog.Register(plugins.AuditLog)
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
