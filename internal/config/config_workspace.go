package config

// DiscoveryConfig configures automatic module discovery behavior.
type DiscoveryConfig struct {
	// Enabled controls whether auto-discovery is active (default: true).
	Enabled *bool `yaml:"enabled,omitempty"`

	// Recursive enables searching subdirectories (default: true).
	Recursive *bool `yaml:"recursive,omitempty"`

	// ModuleMaxDepth limits directory traversal depth for module discovery (default: 10).
	ModuleMaxDepth *int `yaml:"module_max_depth,omitempty"`

	// ManifestMaxDepth limits directory traversal depth for manifest discovery (default: 3).
	// This is separate from ModuleMaxDepth to allow different depths for modules vs manifests.
	ManifestMaxDepth *int `yaml:"manifest_max_depth,omitempty"`

	// Exclude lists paths/patterns to skip during discovery.
	Exclude []string `yaml:"exclude,omitempty"`
}

// ModuleConfig defines an explicitly configured module.
type ModuleConfig struct {
	// Name is the module identifier.
	Name string `yaml:"name"`

	// Path is the path to the module's .version file.
	Path string `yaml:"path"`

	// Enabled controls whether this module is active (default: true).
	Enabled *bool `yaml:"enabled,omitempty"`
}

// IsEnabled returns true if the module is enabled.
// Modules are enabled by default if the Enabled field is nil.
func (m *ModuleConfig) IsEnabled() bool {
	if m.Enabled == nil {
		return true
	}
	return *m.Enabled
}

// WorkspaceConfig configures multi-module/monorepo behavior.
type WorkspaceConfig struct {
	// Discovery configures automatic module discovery.
	Discovery *DiscoveryConfig `yaml:"discovery,omitempty"`

	// Modules explicitly defines modules (overrides discovery if non-empty).
	Modules []ModuleConfig `yaml:"modules,omitempty"`
}

// DefaultExcludePatterns returns the default patterns to exclude during module discovery.
var DefaultExcludePatterns = []string{
	"node_modules",
	".git",
	"vendor",
	"tmp",
	"build",
	"dist",
	".cache",
	"__pycache__",
}

// DiscoveryDefaults returns a DiscoveryConfig with default values.
func DiscoveryDefaults() *DiscoveryConfig {
	enabled := true
	recursive := true
	maxDepth := 10
	manifestMaxDepth := 3

	return &DiscoveryConfig{
		Enabled:          &enabled,
		Recursive:        &recursive,
		ModuleMaxDepth:   &maxDepth,
		ManifestMaxDepth: &manifestMaxDepth,
		Exclude:          DefaultExcludePatterns,
	}
}

// GetDiscoveryConfig returns the discovery configuration with defaults applied.
// If workspace or discovery is not configured, returns default discovery settings.
func (c *Config) GetDiscoveryConfig() *DiscoveryConfig {
	if c.Workspace == nil || c.Workspace.Discovery == nil {
		return DiscoveryDefaults()
	}

	cfg := c.Workspace.Discovery
	defaults := DiscoveryDefaults()

	// Apply defaults for nil pointer fields
	result := &DiscoveryConfig{
		Enabled:          cfg.Enabled,
		Recursive:        cfg.Recursive,
		ModuleMaxDepth:   cfg.ModuleMaxDepth,
		ManifestMaxDepth: cfg.ManifestMaxDepth,
		Exclude:          cfg.Exclude,
	}

	if result.Enabled == nil {
		result.Enabled = defaults.Enabled
	}
	if result.Recursive == nil {
		result.Recursive = defaults.Recursive
	}
	if result.ModuleMaxDepth == nil {
		result.ModuleMaxDepth = defaults.ModuleMaxDepth
	}
	if result.ManifestMaxDepth == nil {
		result.ManifestMaxDepth = defaults.ManifestMaxDepth
	}

	return result
}

// GetExcludePatterns returns the merged list of default and configured exclude patterns.
// Configured patterns are appended to defaults, allowing for extension.
func (c *Config) GetExcludePatterns() []string {
	discovery := c.GetDiscoveryConfig()

	// Start with defaults
	patterns := make([]string, len(DefaultExcludePatterns))
	copy(patterns, DefaultExcludePatterns)

	// Add configured patterns if they differ from defaults
	if c.Workspace != nil && c.Workspace.Discovery != nil && len(c.Workspace.Discovery.Exclude) > 0 {
		// Use a map to avoid duplicates
		seen := make(map[string]bool)
		for _, p := range DefaultExcludePatterns {
			seen[p] = true
		}

		for _, p := range c.Workspace.Discovery.Exclude {
			if !seen[p] {
				patterns = append(patterns, p)
				seen[p] = true
			}
		}
	} else if discovery.Exclude != nil {
		// If using defaults, return them directly
		return discovery.Exclude
	}

	return patterns
}

// HasExplicitModules returns true if modules are explicitly defined in the workspace configuration.
func (c *Config) HasExplicitModules() bool {
	return c.Workspace != nil && len(c.Workspace.Modules) > 0
}

// IsModuleEnabled checks if a specific module is enabled by name.
// Returns false if the module is not found or workspace is not configured.
func (c *Config) IsModuleEnabled(name string) bool {
	if !c.HasExplicitModules() {
		return false
	}

	for _, module := range c.Workspace.Modules {
		if module.Name == name {
			return module.IsEnabled()
		}
	}

	return false
}
