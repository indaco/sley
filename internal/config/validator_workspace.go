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

// validateExtensionConfigs validates extension configurations.
func (v *Validator) validateExtensionConfigs(ctx context.Context) {
	if v.cfg == nil || len(v.cfg.Extensions) == 0 {
		return
	}

	pathErrorCount := 0
	manifestErrorCount := 0

	for i, ext := range v.cfg.Extensions {
		extPath := v.resolveExtensionPath(ext.Path)

		if !v.validateExtensionPath(ctx, i, ext, extPath) {
			pathErrorCount++
			continue
		}

		if ext.Enabled && !v.validateExtensionManifest(ctx, i, ext, extPath) {
			manifestErrorCount++
		}
	}

	if pathErrorCount == 0 && manifestErrorCount == 0 {
		enabledCount := v.countEnabledExtensions()
		v.addValidation("Extensions", true,
			fmt.Sprintf("Configured with %d extension(s) (%d enabled)", len(v.cfg.Extensions), enabledCount), false)
	}
}

// resolveExtensionPath returns the absolute path for an extension.
func (v *Validator) resolveExtensionPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(v.rootDir, path)
}

// validateExtensionPath checks if an extension path exists and is accessible.
func (v *Validator) validateExtensionPath(ctx context.Context, index int, ext ExtensionConfig, extPath string) bool {
	_, err := v.fs.Stat(ctx, extPath)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		v.addValidation("Extensions", false,
			fmt.Sprintf("Extension %d ('%s'): path '%s' does not exist", index+1, ext.Name, ext.Path), false)
	} else {
		v.addValidation("Extensions", false,
			fmt.Sprintf("Extension %d ('%s'): cannot access path '%s': %v", index+1, ext.Name, ext.Path, err), false)
	}
	return false
}

// validateExtensionManifest checks if an enabled extension has a valid manifest file.
func (v *Validator) validateExtensionManifest(ctx context.Context, index int, ext ExtensionConfig, extPath string) bool {
	manifestPath := filepath.Join(extPath, "extension.yaml")
	_, err := v.fs.Stat(ctx, manifestPath)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		v.addValidation("Extensions", false,
			fmt.Sprintf("Extension %d ('%s'): manifest file 'extension.yaml' not found", index+1, ext.Name), false)
	} else {
		v.addValidation("Extensions", false,
			fmt.Sprintf("Extension %d ('%s'): cannot access manifest: %v", index+1, ext.Name, err), false)
	}
	return false
}

// countEnabledExtensions returns the number of enabled extensions.
func (v *Validator) countEnabledExtensions() int {
	count := 0
	for _, ext := range v.cfg.Extensions {
		if ext.Enabled {
			count++
		}
	}
	return count
}
