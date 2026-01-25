package initialize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectContext_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	ctx := DetectProjectContext()

	if ctx.IsGitRepo {
		t.Error("expected IsGitRepo to be false in empty directory")
	}
	if ctx.HasPackageJSON {
		t.Error("expected HasPackageJSON to be false")
	}
	if ctx.HasGoMod {
		t.Error("expected HasGoMod to be false")
	}
	if ctx.HasCargoToml {
		t.Error("expected HasCargoToml to be false")
	}
	if ctx.HasPyprojectToml {
		t.Error("expected HasPyprojectToml to be false")
	}
}

func TestDetectProjectContext_GitRepository(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	ctx := DetectProjectContext()

	if !ctx.IsGitRepo {
		t.Error("expected IsGitRepo to be true")
	}
}

func TestDetectProjectContext_GitSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git in root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to subdirectory
	t.Chdir(subDir)

	ctx := DetectProjectContext()

	if !ctx.IsGitRepo {
		t.Error("expected IsGitRepo to be true in subdirectory of git repo")
	}
}

func TestDetectProjectContext_PackageJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	ctx := DetectProjectContext()

	if !ctx.HasPackageJSON {
		t.Error("expected HasPackageJSON to be true")
	}
}

func TestDetectProjectContext_GoMod(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	ctx := DetectProjectContext()

	if !ctx.HasGoMod {
		t.Error("expected HasGoMod to be true")
	}
}

func TestDetectProjectContext_MultipleMarkers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create package.json
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	ctx := DetectProjectContext()

	if !ctx.IsGitRepo {
		t.Error("expected IsGitRepo to be true")
	}
	if !ctx.HasPackageJSON {
		t.Error("expected HasPackageJSON to be true")
	}
	if !ctx.HasGoMod {
		t.Error("expected HasGoMod to be true")
	}
}

func TestProjectContext_SuggestedPlugins(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *ProjectContext
		expected []string
	}{
		{
			name:     "empty project",
			ctx:      &ProjectContext{},
			expected: []string{},
		},
		{
			name: "git repo only",
			ctx: &ProjectContext{
				IsGitRepo: true,
			},
			expected: []string{"commit-parser", "tag-manager"},
		},
		{
			name: "git repo with package.json",
			ctx: &ProjectContext{
				IsGitRepo:      true,
				HasPackageJSON: true,
			},
			expected: []string{"commit-parser", "tag-manager", "dependency-check"},
		},
		{
			name: "go project with git",
			ctx: &ProjectContext{
				IsGitRepo: true,
				HasGoMod:  true,
			},
			expected: []string{"commit-parser", "tag-manager", "dependency-check"},
		},
		{
			name: "rust project",
			ctx: &ProjectContext{
				IsGitRepo:    true,
				HasCargoToml: true,
			},
			expected: []string{"commit-parser", "tag-manager", "dependency-check"},
		},
		{
			name: "python project",
			ctx: &ProjectContext{
				HasPyprojectToml: true,
			},
			expected: []string{"dependency-check"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.SuggestedPlugins()

			if len(got) != len(tt.expected) {
				t.Errorf("expected %d suggestions, got %d: %v", len(tt.expected), len(got), got)
				return
			}

			for i, exp := range tt.expected {
				if got[i] != exp {
					t.Errorf("suggestion[%d]: expected %q, got %q", i, exp, got[i])
				}
			}
		})
	}
}

func TestProjectContext_FormatDetectionSummary(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *ProjectContext
		contains []string
		empty    bool
	}{
		{
			name:  "empty project",
			ctx:   &ProjectContext{},
			empty: true,
		},
		{
			name: "git repo",
			ctx: &ProjectContext{
				IsGitRepo: true,
			},
			contains: []string{"Detected:", "Git repository"},
		},
		{
			name: "node project",
			ctx: &ProjectContext{
				HasPackageJSON: true,
			},
			contains: []string{"Detected:", "package.json", "Node.js project"},
		},
		{
			name: "go project",
			ctx: &ProjectContext{
				HasGoMod: true,
			},
			contains: []string{"Detected:", "go.mod", "Go project"},
		},
		{
			name: "rust project",
			ctx: &ProjectContext{
				HasCargoToml: true,
			},
			contains: []string{"Detected:", "Cargo.toml", "Rust project"},
		},
		{
			name: "python project",
			ctx: &ProjectContext{
				HasPyprojectToml: true,
			},
			contains: []string{"Detected:", "pyproject.toml", "Python project"},
		},
		{
			name: "multiple markers",
			ctx: &ProjectContext{
				IsGitRepo:      true,
				HasPackageJSON: true,
				HasGoMod:       true,
			},
			contains: []string{"Detected:", "Git repository", "package.json", "go.mod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.ctx.FormatDetectionSummary()

			if tt.empty {
				if summary != "" {
					t.Errorf("expected empty summary, got: %q", summary)
				}
				return
			}

			for _, expected := range tt.contains {
				if summary == "" || !contains(summary, expected) {
					t.Errorf("expected summary to contain %q, got: %q", expected, summary)
				}
			}
		})
	}
}

func TestProjectContext_HasAnyDetection(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *ProjectContext
		expected bool
	}{
		{
			name:     "empty",
			ctx:      &ProjectContext{},
			expected: false,
		},
		{
			name: "has git",
			ctx: &ProjectContext{
				IsGitRepo: true,
			},
			expected: true,
		},
		{
			name: "has package.json",
			ctx: &ProjectContext{
				HasPackageJSON: true,
			},
			expected: true,
		},
		{
			name: "has go.mod",
			ctx: &ProjectContext{
				HasGoMod: true,
			},
			expected: true,
		},
		{
			name: "has cargo.toml",
			ctx: &ProjectContext{
				HasCargoToml: true,
			},
			expected: true,
		},
		{
			name: "has pyproject.toml",
			ctx: &ProjectContext{
				HasPyprojectToml: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.HasAnyDetection()
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
