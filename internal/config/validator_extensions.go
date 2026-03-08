package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// validateExtensionConfigs validates extension configurations.
func (v *Validator) validateExtensionConfigs(ctx context.Context) {
	if v.cfg == nil || len(v.cfg.Extensions) == 0 {
		return
	}

	pathErrorCount := 0
	manifestErrorCount := 0

	for i, ext := range v.cfg.Extensions {
		extPath := v.resolvePath(ext.Path)

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
