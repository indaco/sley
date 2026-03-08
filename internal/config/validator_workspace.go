package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validateWorkspaceConfig validates workspace/multi-module configuration.
func (v *Validator) validateWorkspaceConfig(ctx context.Context) {
	if v.cfg == nil || v.cfg.Workspace == nil {
		return
	}

	// Validate discovery configuration
	if v.cfg.Workspace.Discovery != nil {
		v.validateDiscoveryConfig()
	}

	// Validate explicit modules
	if len(v.cfg.Workspace.Modules) > 0 {
		v.validateExplicitModules(ctx)
	}
}

// validateDiscoveryConfig validates module discovery settings.
func (v *Validator) validateDiscoveryConfig() {
	discovery := v.cfg.Workspace.Discovery

	// Validate max depth
	if discovery.ModuleMaxDepth != nil && *discovery.ModuleMaxDepth < 0 {
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
func (v *Validator) validateExplicitModules(ctx context.Context) {
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

		if _, err := v.fs.Stat(ctx, modPath); err != nil {
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
