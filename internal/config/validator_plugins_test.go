package config

import (
	"context"
	"testing"

	"github.com/indaco/sley/internal/core"
)

func TestValidator_ValidateYAMLSyntax(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		setupFS    func(context.Context, *core.MockFileSystem)
		wantPass   bool
		wantMsg    string
	}{
		{
			name:       "no config file",
			configPath: "",
			setupFS:    func(ctx context.Context, fs *core.MockFileSystem) {},
			wantPass:   true,
			wantMsg:    "No .sley.yaml file found, using defaults",
		},
		{
			name:       "valid config file",
			configPath: ".sley.yaml",
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, ".sley.yaml", []byte("path: .version\n"), 0644)
			},
			wantPass: true,
			wantMsg:  "Configuration file is valid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			tt.setupFS(ctx, fs)

			cfg := &Config{Path: ".version"}
			validator := NewValidator(fs, cfg, tt.configPath, ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Find YAML Syntax validation result
			var found bool
			for _, r := range results {
				if r.Category == "YAML Syntax" {
					found = true
					if r.Passed != tt.wantPass {
						t.Errorf("YAML Syntax validation passed = %v, want %v", r.Passed, tt.wantPass)
					}
					if r.Message != tt.wantMsg {
						t.Errorf("YAML Syntax validation message = %q, want %q", r.Message, tt.wantMsg)
					}
				}
			}

			if !found {
				t.Error("YAML Syntax validation result not found")
			}
		})
	}
}

func TestValidator_ValidateTagManagerConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantPass  bool
		wantError bool
	}{
		{
			name: "valid tag manager config",
			config: &Config{
				Plugins: &PluginConfig{
					TagManager: &TagManagerConfig{
						Enabled: true,
						Prefix:  "v",
					},
				},
			},
			wantPass:  true,
			wantError: false,
		},
		{
			name: "invalid prefix with whitespace",
			config: &Config{
				Plugins: &PluginConfig{
					TagManager: &TagManagerConfig{
						Enabled: true,
						Prefix:  "v ",
					},
				},
			},
			wantPass:  false,
			wantError: true,
		},
		{
			name: "invalid prefix with slash",
			config: &Config{
				Plugins: &PluginConfig{
					TagManager: &TagManagerConfig{
						Enabled: true,
						Prefix:  "v/",
					},
				},
			},
			wantPass:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Find tag-manager validation result
			var found bool
			for _, r := range results {
				if r.Category == "Plugin: tag-manager" {
					found = true
					if tt.wantError && r.Passed {
						t.Errorf("Expected tag-manager validation to fail, but it passed")
					}
					if !tt.wantError && !r.Passed {
						t.Errorf("Expected tag-manager validation to pass, but it failed: %s", r.Message)
					}
				}
			}

			if tt.config.Plugins.TagManager != nil && tt.config.Plugins.TagManager.Enabled && !found {
				t.Error("tag-manager validation result not found")
			}
		})
	}
}

func TestValidator_ValidateVersionValidatorConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid rules",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "pre-release-format",
								Pattern: `^(alpha|beta|rc)\.\d+$`,
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid rule type",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "unknown-rule-type",
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid regex pattern",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "pre-release-format",
								Pattern: "[invalid(regex",
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "branch-constraint missing branch",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "branch-constraint",
								Allowed: []string{"patch", "minor"},
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Plugin: version-validator" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("version-validator validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateDependencyCheckConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		setupFS   func(context.Context, *core.MockFileSystem)
		wantError bool
	}{
		{
			name: "valid file exists",
			config: &Config{
				Plugins: &PluginConfig{
					DependencyCheck: &DependencyCheckConfig{
						Enabled: true,
						Files: []DependencyFileConfig{
							{
								Path:   "package.json",
								Format: "json",
								Field:  "version",
							},
						},
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "package.json", []byte(`{"version": "1.0.0"}`), 0644)
			},
			wantError: false,
		},
		{
			name: "file does not exist",
			config: &Config{
				Plugins: &PluginConfig{
					DependencyCheck: &DependencyCheckConfig{
						Enabled: true,
						Files: []DependencyFileConfig{
							{
								Path:   "missing.json",
								Format: "json",
								Field:  "version",
							},
						},
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: true,
		},
		{
			name: "invalid format",
			config: &Config{
				Plugins: &PluginConfig{
					DependencyCheck: &DependencyCheckConfig{
						Enabled: true,
						Files: []DependencyFileConfig{
							{
								Path:   "file.txt",
								Format: "unknown",
								Field:  "version",
							},
						},
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "file.txt", []byte("content"), 0644)
			},
			wantError: true,
		},
		{
			name: "regex format with invalid pattern",
			config: &Config{
				Plugins: &PluginConfig{
					DependencyCheck: &DependencyCheckConfig{
						Enabled: true,
						Files: []DependencyFileConfig{
							{
								Path:    "file.txt",
								Format:  "regex",
								Pattern: "[invalid(regex",
							},
						},
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "file.txt", []byte("content"), 0644)
			},
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
				if r.Category == "Plugin: dependency-check" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("dependency-check validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateChangelogGeneratorConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Mode:    "versioned",
						Format:  "grouped",
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid mode",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Mode:    "invalid",
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid format",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Format:  "invalid",
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid repository provider",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Repository: &RepositoryConfig{
							Provider: "invalid",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "custom provider without host",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Repository: &RepositoryConfig{
							Provider: "custom",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid exclude pattern",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled:         true,
						ExcludePatterns: []string{"[invalid(regex"},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Plugin: changelog-generator" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("changelog-generator validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateAuditLogConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid json format",
			config: &Config{
				Plugins: &PluginConfig{
					AuditLog: &AuditLogConfig{
						Enabled: true,
						Format:  "json",
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid yaml format",
			config: &Config{
				Plugins: &PluginConfig{
					AuditLog: &AuditLogConfig{
						Enabled: true,
						Format:  "yaml",
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid format",
			config: &Config{
				Plugins: &PluginConfig{
					AuditLog: &AuditLogConfig{
						Enabled: true,
						Format:  "xml",
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Plugin: audit-log" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("audit-log validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateChangelogParserConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		setupFS   func(context.Context, *core.MockFileSystem)
		wantError bool
	}{
		{
			name: "valid changelog file",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogParser: &ChangelogParserConfig{
						Enabled:  true,
						Path:     "CHANGELOG.md",
						Priority: "changelog",
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "CHANGELOG.md", []byte("# Changelog\n"), 0644)
			},
			wantError: false,
		},
		{
			name: "changelog file does not exist",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogParser: &ChangelogParserConfig{
						Enabled: true,
						Path:    "MISSING.md",
					},
				},
			},
			setupFS:   func(ctx context.Context, fs *core.MockFileSystem) {},
			wantError: true,
		},
		{
			name: "invalid priority value",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogParser: &ChangelogParserConfig{
						Enabled:  true,
						Path:     "CHANGELOG.md",
						Priority: "invalid-priority",
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "CHANGELOG.md", []byte("# Changelog\n"), 0644)
			},
			wantError: true,
		},
		{
			name: "default changelog path",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogParser: &ChangelogParserConfig{
						Enabled: true,
						Path:    "",
					},
				},
			},
			setupFS: func(ctx context.Context, fs *core.MockFileSystem) {
				_ = fs.WriteFile(ctx, "CHANGELOG.md", []byte("# Changelog\n"), 0644)
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
				if r.Category == "Plugin: changelog-parser" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("changelog-parser validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateReleaseGateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantWarning bool
	}{
		{
			name: "no conflicting branches",
			config: &Config{
				Plugins: &PluginConfig{
					ReleaseGate: &ReleaseGateConfig{
						Enabled:         true,
						AllowedBranches: []string{"main", "develop"},
					},
				},
			},
			wantWarning: false,
		},
		{
			name: "conflicting allowed and blocked branches",
			config: &Config{
				Plugins: &PluginConfig{
					ReleaseGate: &ReleaseGateConfig{
						Enabled:         true,
						AllowedBranches: []string{"main"},
						BlockedBranches: []string{"develop"},
					},
				},
			},
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasWarning := false
			for _, r := range results {
				if r.Category == "Plugin: release-gate" && r.Warning {
					hasWarning = true
					break
				}
			}

			if hasWarning != tt.wantWarning {
				t.Errorf("release-gate validation warning = %v, want %v", hasWarning, tt.wantWarning)
			}
		})
	}
}

func TestValidator_ValidateCommitParserPlugin(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "commit parser enabled",
			config: &Config{
				Plugins: &PluginConfig{
					CommitParser: true,
				},
			},
		},
		{
			name: "commit parser disabled",
			config: &Config{
				Plugins: &PluginConfig{
					CommitParser: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			_, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			// Commit parser doesn't have specific validation, this just ensures it doesn't break
		})
	}
}

func TestValidator_ValidateVersionValidatorBranchConstraint(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid branch constraint",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "branch-constraint",
								Branch:  "main",
								Allowed: []string{"major", "minor"},
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "branch constraint missing allowed",
			config: &Config{
				Plugins: &PluginConfig{
					VersionValidator: &VersionValidatorConfig{
						Enabled: true,
						Rules: []ValidationRule{
							{
								Type:    "branch-constraint",
								Branch:  "main",
								Enabled: true,
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Plugin: version-validator" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("version-validator validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateChangelogGeneratorCustomProvider(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "custom provider with host",
			config: &Config{
				Plugins: &PluginConfig{
					ChangelogGenerator: &ChangelogGeneratorConfig{
						Enabled: true,
						Repository: &RepositoryConfig{
							Provider: "custom",
							Host:     "git.example.com",
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := core.NewMockFileSystem()
			validator := NewValidator(fs, tt.config, "", ".")

			results, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			hasError := false
			for _, r := range results {
				if r.Category == "Plugin: changelog-generator" && !r.Passed && !r.Warning {
					hasError = true
					break
				}
			}

			if hasError != tt.wantError {
				t.Errorf("changelog-generator validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}
