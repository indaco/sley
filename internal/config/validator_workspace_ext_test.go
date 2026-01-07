package config

import (
	"context"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/core"
)

/* ------------------------------------------------------------------------- */
/* WORKSPACE VALIDATION TESTS                                                */
/* ------------------------------------------------------------------------- */

func TestValidator_ValidateWorkspaceConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		setupFS   func(context.Context, *core.MockFileSystem)
		wantError bool
	}{
		{
			name: "valid explicit modules",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{
							Name: "module1",
							Path: "module1/.version",
						},
						{
							Name: "module2",
							Path: "module2/.version",
						},
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "module1/.version", []byte("1.0.0"), 0644)
				_ = fs.WriteFile(ctx, "module2/.version", []byte("2.0.0"), 0644)
			},
			wantError: false,
		},
		{
			name: "module path does not exist",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{
							Name: "missing",
							Path: "missing/.version",
						},
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: true,
		},
		{
			name: "duplicate module names",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{
							Name: "duplicate",
							Path: "module1/.version",
						},
						{
							Name: "duplicate",
							Path: "module2/.version",
						},
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "module1/.version", []byte("1.0.0"), 0644)
				_ = fs.WriteFile(ctx, "module2/.version", []byte("2.0.0"), 0644)
			},
			wantError: true,
		},
		{
			name: "valid discovery config",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Discovery: &DiscoveryConfig{
						Exclude: []string{"node_modules", "vendor"},
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: false,
		},
		{
			name: "negative max depth",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Discovery: &DiscoveryConfig{
						MaxDepth: func() *int { v := -1; return &v }(),
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			tt.setupFS(ctx, fs)

			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if (r.Category == "Workspace: Modules" || r.Category == "Workspace: Discovery") &&
					!r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("workspace validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateWorkspaceOverbreadExcludePattern(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Workspace: &WorkspaceConfig{
			Discovery: &DiscoveryConfig{
				Exclude: []string{"**/**/**"},
			},
		},
	}

	fs := core.NewMockFileSystem()
	validator := NewValidator(fs, config, "", ".")

	results, err := validator.Validate(ctx)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should have a warning about overly broad pattern
	hasWarning := false
	for _, r := range results {
		if r.Category == "Workspace: Discovery" && r.Warning && strings.Contains(r.Message, "overly broad") {
			hasWarning = true
			break
		}
	}

	if !hasWarning {
		t.Error("expected warning for overly broad exclude pattern")
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION VALIDATION TESTS                                                */
/* ------------------------------------------------------------------------- */

func TestValidator_ValidateExtensionConfigs(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		setupFS   func(context.Context, *core.MockFileSystem)
		wantError bool
	}{
		{
			name: "valid extension with manifest",
			config: &Config{
				Extensions: []ExtensionConfig{
					{
						Name:    "test-extension",
						Path:    "extensions/test",
						Enabled: true,
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				// Create the directory explicitly
				_ = fs.MkdirAll(ctx, "extensions/test", 0755)
				_ = fs.WriteFile(ctx, "extensions/test/extension.yaml", []byte("name: test"), 0644)
			},
			wantError: false,
		},
		{
			name: "extension path does not exist",
			config: &Config{
				Extensions: []ExtensionConfig{
					{
						Name:    "missing-extension",
						Path:    "extensions/missing",
						Enabled: false,
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: true,
		},
		{
			name: "enabled extension missing manifest",
			config: &Config{
				Extensions: []ExtensionConfig{
					{
						Name:    "no-manifest",
						Path:    "extensions/no-manifest",
						Enabled: true,
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				// Create directory with a file but no manifest
				_ = fs.MkdirAll(ctx, "extensions/no-manifest", 0755)
				_ = fs.WriteFile(ctx, "extensions/no-manifest/script.sh", []byte("#!/bin/bash"), 0755)
			},
			wantError: true,
		},
		{
			name: "disabled extension missing manifest - should not error",
			config: &Config{
				Extensions: []ExtensionConfig{
					{
						Name:    "disabled-ext",
						Path:    "extensions/disabled",
						Enabled: false,
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				// Create directory with a file
				_ = fs.MkdirAll(ctx, "extensions/disabled", 0755)
				_ = fs.WriteFile(ctx, "extensions/disabled/script.sh", []byte("#!/bin/bash"), 0755)
			},
			wantError: false,
		},
		{
			name: "multiple extensions mixed states",
			config: &Config{
				Extensions: []ExtensionConfig{
					{
						Name:    "valid-ext",
						Path:    "extensions/valid",
						Enabled: true,
					},
					{
						Name:    "disabled-ext",
						Path:    "extensions/disabled",
						Enabled: false,
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.MkdirAll(ctx, "extensions/valid", 0755)
				_ = fs.WriteFile(ctx, "extensions/valid/extension.yaml", []byte("name: valid"), 0644)
				_ = fs.MkdirAll(ctx, "extensions/disabled", 0755)
				_ = fs.WriteFile(ctx, "extensions/disabled/script.sh", []byte("#!/bin/bash"), 0755)
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			tt.setupFS(ctx, fs)

			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Extensions" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("extension validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}
