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

// --- Monorepo Detection Tests ---

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func mkdirTest(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}
}

func sliceContainsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func TestDetectMonorepo_GoWork(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "go.work", `go 1.21

use (
	./cobra
	./kong
	./urfave
)
`)
	mkdirTest(t, "cobra")
	mkdirTest(t, "kong")
	mkdirTest(t, "urfave")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if info.Type != "go-work" {
		t.Errorf("expected type go-work, got %s", info.Type)
	}
	if len(info.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %d: %v", len(info.Modules), info.Modules)
	}
	for _, name := range []string{"cobra", "kong", "urfave"} {
		if !sliceContainsStr(info.Modules, name) {
			t.Errorf("expected module %q in %v", name, info.Modules)
		}
	}
}

func TestDetectMonorepo_GoWork_SingleLine(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "go.work", "go 1.21\n\nuse ./single\n")
	mkdirTest(t, "single")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if len(info.Modules) != 1 || info.Modules[0] != "single" {
		t.Errorf("expected [single], got %v", info.Modules)
	}
}

func TestDetectMonorepo_GoWork_ExternalPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "go.work", "go 1.21\n\nuse (\n\t./local\n\t../external\n)\n")
	mkdirTest(t, "local")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if len(info.Modules) != 1 {
		t.Errorf("expected 1 module (external excluded), got %d: %v", len(info.Modules), info.Modules)
	}
	if info.Modules[0] != "local" {
		t.Errorf("expected module 'local', got %q", info.Modules[0])
	}
}

func TestDetectMonorepo_PnpmWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "pnpm-workspace.yaml", "packages:\n  - \"packages/*\"\n")
	mkdirTest(t, "packages/foo")
	mkdirTest(t, "packages/bar")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if info.Type != "pnpm" {
		t.Errorf("expected type pnpm, got %s", info.Type)
	}
	if len(info.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d: %v", len(info.Modules), info.Modules)
	}
}

func TestDetectMonorepo_NpmWorkspaces(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "package.json", `{"name":"root","workspaces":["packages/*"]}`)
	mkdirTest(t, "packages/app")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if info.Type != "npm" {
		t.Errorf("expected type npm, got %s", info.Type)
	}
	if len(info.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d: %v", len(info.Modules), info.Modules)
	}
}

func TestDetectMonorepo_CargoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	writeTestFile(t, "Cargo.toml", `[package]
name = "root"

[workspace]
members = ["crate-a", "crate-b"]
`)
	mkdirTest(t, "crate-a")
	mkdirTest(t, "crate-b")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if info.Type != "cargo" {
		t.Errorf("expected type cargo, got %s", info.Type)
	}
}

func TestDetectMonorepo_None(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil for no markers, got %+v", info)
	}
}

func TestDetectMonorepo_Priority(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Both go.work and pnpm-workspace.yaml present — go.work should win
	writeTestFile(t, "go.work", "go 1.21\nuse ./mod1\n")
	mkdirTest(t, "mod1")
	writeTestFile(t, "pnpm-workspace.yaml", "packages:\n  - \"packages/*\"\n")

	info, err := DetectMonorepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MonorepoInfo, got nil")
	}
	if info.Type != "go-work" {
		t.Errorf("expected go-work to take priority, got %s", info.Type)
	}
}

func TestCreateMonorepoVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mkdirTest(t, "mod-new")
	mkdirTest(t, "mod-existing")
	writeTestFile(t, "mod-existing/.version", "1.5.0\n")

	info := &MonorepoInfo{
		Type:    "go-work",
		Modules: []string{"mod-new", "mod-existing"},
	}
	createMonorepoVersionFiles(info)

	// mod-new should have .version with 0.0.0
	data, err := os.ReadFile("mod-new/.version")
	if err != nil {
		t.Fatalf("expected mod-new/.version to exist: %v", err)
	}
	if string(data) != "0.0.0\n" {
		t.Errorf("expected '0.0.0\\n', got %q", string(data))
	}

	// mod-existing should NOT be overwritten
	data, err = os.ReadFile("mod-existing/.version")
	if err != nil {
		t.Fatalf("expected mod-existing/.version to exist: %v", err)
	}
	if string(data) != "1.5.0\n" {
		t.Errorf("expected '1.5.0\\n' (unchanged), got %q", string(data))
	}
}
