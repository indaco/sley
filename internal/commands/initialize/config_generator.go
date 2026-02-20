package initialize

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
)

// GenerateConfigWithComments creates a Config struct and returns it as YAML with helpful comments.
// The path parameter specifies the .version file path to use in the config.
// The selectedPlugins parameter should contain the names of plugins to enable.
func GenerateConfigWithComments(path string, selectedPlugins []string) ([]byte, error) {
	cfg := &config.Config{
		Path: path,
	}

	// Create plugins config based on selections
	pluginsCfg := &config.PluginConfig{}

	for _, name := range selectedPlugins {
		switch name {
		case "commit-parser":
			pluginsCfg.CommitParser = true
		case "tag-manager":
			pluginsCfg.TagManager = &config.TagManagerConfig{
				Enabled: true,
			}
		case "version-validator":
			pluginsCfg.VersionValidator = &config.VersionValidatorConfig{
				Enabled: true,
			}
		case "dependency-check":
			pluginsCfg.DependencyCheck = &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
			}
		case "changelog-parser":
			pluginsCfg.ChangelogParser = &config.ChangelogParserConfig{
				Enabled: true,
			}
		case "changelog-generator":
			pluginsCfg.ChangelogGenerator = &config.ChangelogGeneratorConfig{
				Enabled: true,
			}
		case "release-gate":
			pluginsCfg.ReleaseGate = &config.ReleaseGateConfig{
				Enabled: true,
			}
		case "audit-log":
			pluginsCfg.AuditLog = &config.AuditLogConfig{
				Enabled: true,
			}
		}
	}

	cfg.Plugins = pluginsCfg

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add helpful header comments
	var buf bytes.Buffer
	buf.WriteString("# sley configuration file\n")
	buf.WriteString("# Documentation: https://github.com/indaco/sley\n")
	buf.WriteString("\n")

	// Add plugin-specific comments
	if len(selectedPlugins) > 0 {
		buf.WriteString("# Enabled plugins:\n")
		for _, name := range selectedPlugins {
			fmt.Fprintf(&buf, "#   - %s\n", name)
		}
		buf.WriteString("\n")
	}

	buf.Write(data)

	// Add inline comments for plugin configuration sections
	result := addPluginComments(buf.Bytes(), selectedPlugins)

	return result, nil
}

// GenerateConfigWithDiscovery creates a Config with dependency-check configuration
// based on discovered sync candidates.
func GenerateConfigWithDiscovery(path string, selectedPlugins []string, syncCandidates []discovery.SyncCandidate) ([]byte, error) {
	cfg := &config.Config{
		Path: path,
	}

	// Create plugins config based on selections
	pluginsCfg := &config.PluginConfig{}

	for _, name := range selectedPlugins {
		switch name {
		case "commit-parser":
			pluginsCfg.CommitParser = true
		case "tag-manager":
			pluginsCfg.TagManager = &config.TagManagerConfig{
				Enabled: true,
			}
		case "version-validator":
			pluginsCfg.VersionValidator = &config.VersionValidatorConfig{
				Enabled: true,
			}
		case "dependency-check":
			// Configure dependency-check with discovered files
			depCheck := &config.DependencyCheckConfig{
				Enabled:  true,
				AutoSync: true,
			}
			if len(syncCandidates) > 0 {
				depCheck.Files = make([]config.DependencyFileConfig, len(syncCandidates))
				for i, c := range syncCandidates {
					depCheck.Files[i] = config.DependencyFileConfig{
						Path:    c.Path,
						Format:  c.Format.String(),
						Field:   c.Field,
						Pattern: c.Pattern,
					}
				}
			}
			pluginsCfg.DependencyCheck = depCheck
		case "changelog-parser":
			pluginsCfg.ChangelogParser = &config.ChangelogParserConfig{
				Enabled: true,
			}
		case "changelog-generator":
			pluginsCfg.ChangelogGenerator = &config.ChangelogGeneratorConfig{
				Enabled: true,
			}
		case "release-gate":
			pluginsCfg.ReleaseGate = &config.ReleaseGateConfig{
				Enabled: true,
			}
		case "audit-log":
			pluginsCfg.AuditLog = &config.AuditLogConfig{
				Enabled: true,
			}
		}
	}

	cfg.Plugins = pluginsCfg

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add helpful header comments
	var buf bytes.Buffer
	buf.WriteString("# sley configuration file\n")
	buf.WriteString("# Documentation: https://github.com/indaco/sley\n")
	buf.WriteString("\n")

	// Add plugin-specific comments
	if len(selectedPlugins) > 0 {
		buf.WriteString("# Enabled plugins:\n")
		for _, name := range selectedPlugins {
			fmt.Fprintf(&buf, "#   - %s\n", name)
		}
		buf.WriteString("\n")
	}

	buf.Write(data)

	// Add inline comments for plugin configuration sections
	result := addPluginComments(buf.Bytes(), selectedPlugins)

	return result, nil
}

// addPluginComments adds helpful inline comments to the YAML configuration.
func addPluginComments(yamlData []byte, selectedPlugins []string) []byte {
	var buf bytes.Buffer
	lines := bytes.Split(yamlData, []byte("\n"))

	pluginComments := map[string]string{
		"commit-parser":       "# Analyzes conventional commits to suggest version bumps",
		"tag-manager":         "# Automatically creates git tags after version changes",
		"version-validator":   "# Enforces versioning policies and constraints",
		"dependency-check":    "# Syncs version to package.json and other dependency files",
		"changelog-parser":    "# Infers version bump from CHANGELOG.md entries",
		"changelog-generator": "# Generates changelogs from git commits",
		"release-gate":        "# Pre-bump validation checks (worktree, CI, branches)",
		"audit-log":           "# Records version history with timestamps and metadata",
	}

	for i, line := range lines {
		// Add comment before plugin sections
		for _, plugin := range selectedPlugins {
			if bytes.Contains(line, []byte(plugin+":")) {
				if comment, ok := pluginComments[plugin]; ok {
					buf.WriteString("  " + comment + "\n")
				}
			}
		}

		buf.Write(line)

		// Add newline except for last line if it's empty
		if i < len(lines)-1 || len(line) > 0 {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes()
}
