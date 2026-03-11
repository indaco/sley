package auditlog

import "github.com/indaco/sley/internal/config"

// Config holds configuration for the audit log plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// Path is the path to the audit log file.
	Path string

	// Format specifies the output format: json or yaml.
	Format string

	// IncludeAuthor includes git author in log entries.
	IncludeAuthor bool

	// IncludeTimestamp includes ISO 8601 timestamp in log entries.
	IncludeTimestamp bool

	// IncludeCommitSHA includes current commit SHA in log entries.
	IncludeCommitSHA bool

	// IncludeBranch includes current branch name in log entries.
	IncludeBranch bool
}

// DefaultConfig returns the default audit log configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:          false,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    true,
		IncludeTimestamp: true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}
}

// GetPath returns the path with default ".version-history.json".
func (c *Config) GetPath() string {
	if c.Path == "" {
		return ".version-history.json"
	}
	return c.Path
}

// GetFormat returns the format with default "json".
func (c *Config) GetFormat() string {
	if c.Format == "" {
		return "json"
	}
	return c.Format
}

// FromConfigStruct converts the config package struct to internal config.
func FromConfigStruct(cfg *config.AuditLogConfig) *Config {
	if cfg == nil {
		return DefaultConfig()
	}

	return &Config{
		Enabled:          cfg.Enabled,
		Path:             cfg.GetPath(),
		Format:           cfg.GetFormat(),
		IncludeAuthor:    cfg.IncludeAuthor,
		IncludeTimestamp: cfg.IncludeTimestamp,
		IncludeCommitSHA: cfg.IncludeCommitSHA,
		IncludeBranch:    cfg.IncludeBranch,
	}
}
