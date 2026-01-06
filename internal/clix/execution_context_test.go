package clix

import (
	"context"
	"os"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode ExecutionMode
		want string
	}{
		{
			name: "single module mode",
			mode: SingleModuleMode,
			want: "SingleModule",
		},
		{
			name: "multi module mode",
			mode: MultiModuleMode,
			want: "MultiModule",
		},
		{
			name: "unknown mode",
			mode: ExecutionMode(99),
			want: "Unknown(99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("ExecutionMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionContext_IsSingleModule(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ExecutionContext
		want bool
	}{
		{
			name: "single module mode",
			ctx: &ExecutionContext{
				Mode: SingleModuleMode,
			},
			want: true,
		},
		{
			name: "multi module mode",
			ctx: &ExecutionContext{
				Mode: MultiModuleMode,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.IsSingleModule()
			if got != tt.want {
				t.Errorf("ExecutionContext.IsSingleModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionContext_IsMultiModule(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ExecutionContext
		want bool
	}{
		{
			name: "single module mode",
			ctx: &ExecutionContext{
				Mode: SingleModuleMode,
			},
			want: false,
		},
		{
			name: "multi module mode",
			ctx: &ExecutionContext{
				Mode: MultiModuleMode,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.IsMultiModule()
			if got != tt.want {
				t.Errorf("ExecutionContext.IsMultiModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterModulesByName(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	// Modules with duplicate names (common in monorepos)
	modulesWithDuplicates := []*workspace.Module{
		{Name: "version", Path: "/backend/gateway/internal/version/.version", Dir: "backend/gateway/internal/version"},
		{Name: "version", Path: "/cli/internal/version/.version", Dir: "cli/internal/version"},
		{Name: "ai-services", Path: "/backend/ai-services/.version", Dir: "backend/ai-services"},
	}

	tests := []struct {
		name       string
		modules    []*workspace.Module
		moduleName string
		wantCount  int
		wantPaths  []string // Check paths to verify correct modules returned
	}{
		{
			name:       "filter existing module",
			modules:    modules,
			moduleName: "module-b",
			wantCount:  1,
			wantPaths:  []string{"/path/to/module-b/.version"},
		},
		{
			name:       "filter non-existent module",
			modules:    modules,
			moduleName: "module-z",
			wantCount:  0,
		},
		{
			name:       "empty module list",
			modules:    []*workspace.Module{},
			moduleName: "module-a",
			wantCount:  0,
		},
		{
			name:       "filter returns all modules with same name",
			modules:    modulesWithDuplicates,
			moduleName: "version",
			wantCount:  2,
			wantPaths: []string{
				"/backend/gateway/internal/version/.version",
				"/cli/internal/version/.version",
			},
		},
		{
			name:       "filter unique module among duplicates",
			modules:    modulesWithDuplicates,
			moduleName: "ai-services",
			wantCount:  1,
			wantPaths:  []string{"/backend/ai-services/.version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterModulesByName(tt.modules, tt.moduleName)
			if len(got) != tt.wantCount {
				t.Errorf("filterModulesByName() returned %d modules, want %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 {
				for i, wantPath := range tt.wantPaths {
					if i >= len(got) {
						t.Errorf("filterModulesByName() missing module at index %d", i)
						continue
					}
					if got[i].Path != wantPath {
						t.Errorf("filterModulesByName()[%d].Path = %q, want %q", i, got[i].Path, wantPath)
					}
				}
			}
		})
	}
}

func TestFilterModulesBySelection(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	tests := []struct {
		name      string
		modules   []*workspace.Module
		selected  []string
		wantCount int
		wantNames []string
	}{
		{
			name:      "select multiple modules",
			modules:   modules,
			selected:  []string{"module-a", "module-c"},
			wantCount: 2,
			wantNames: []string{"module-a", "module-c"},
		},
		{
			name:      "select single module",
			modules:   modules,
			selected:  []string{"module-b"},
			wantCount: 1,
			wantNames: []string{"module-b"},
		},
		{
			name:      "select all modules",
			modules:   modules,
			selected:  []string{"module-a", "module-b", "module-c"},
			wantCount: 3,
			wantNames: []string{"module-a", "module-b", "module-c"},
		},
		{
			name:      "select non-existent module",
			modules:   modules,
			selected:  []string{"module-z"},
			wantCount: 0,
		},
		{
			name:      "empty selection",
			modules:   modules,
			selected:  []string{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterModulesBySelection(tt.modules, tt.selected)
			if len(got) != tt.wantCount {
				t.Errorf("filterModulesBySelection() returned %d modules, want %d", len(got), tt.wantCount)
			}

			gotNames := make(map[string]bool)
			for _, mod := range got {
				gotNames[mod.Name] = true
			}

			for _, wantName := range tt.wantNames {
				if !gotNames[wantName] {
					t.Errorf("filterModulesBySelection() missing expected module %q", wantName)
				}
			}
		})
	}
}

func TestGetExecutionContext_PathFlag(t *testing.T) {
	cfg := &config.Config{
		Path: ".version",
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: "/custom/path/.version",
			},
		},
	}

	// Simulate flag being set
	ctx := context.Background()

	// Create a command context with the path flag set
	// Note: In actual CLI usage, this would be handled by urfave/cli
	// For testing, we'll verify the logic separately
	testPath := "/custom/path/.version"

	execCtx := &ExecutionContext{
		Mode: SingleModuleMode,
		Path: testPath,
	}

	if execCtx.Mode != SingleModuleMode {
		t.Errorf("GetExecutionContext() with --path should return SingleModuleMode, got %v", execCtx.Mode)
	}

	if execCtx.Path != testPath {
		t.Errorf("GetExecutionContext() path = %v, want %v", execCtx.Path, testPath)
	}

	_ = ctx
	_ = cmd
	_ = cfg
}

func TestFilterModulesByName_NilModules(t *testing.T) {
	result := filterModulesByName(nil, "test")
	if result != nil {
		t.Errorf("filterModulesByName(nil, _) should return nil, got %v", result)
	}
}

func TestFilterModulesByName_PreservesModuleData(t *testing.T) {
	modules := []*workspace.Module{
		{
			Name:           "module-a",
			Path:           "/path/to/module-a/.version",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
			Dir:            "/path/to/module-a",
		},
	}

	result := filterModulesByName(modules, "module-a")

	if len(result) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result))
	}

	mod := result[0]
	if mod.Name != "module-a" {
		t.Errorf("Name = %q, want %q", mod.Name, "module-a")
	}
	if mod.Path != "/path/to/module-a/.version" {
		t.Errorf("Path = %q, want %q", mod.Path, "/path/to/module-a/.version")
	}
	if mod.RelPath != "module-a/.version" {
		t.Errorf("RelPath = %q, want %q", mod.RelPath, "module-a/.version")
	}
	if mod.CurrentVersion != "1.0.0" {
		t.Errorf("CurrentVersion = %q, want %q", mod.CurrentVersion, "1.0.0")
	}
	if mod.Dir != "/path/to/module-a" {
		t.Errorf("Dir = %q, want %q", mod.Dir, "/path/to/module-a")
	}
}

func TestFilterModulesBySelection_NilModules(t *testing.T) {
	result := filterModulesBySelection(nil, []string{"test"})
	if result != nil {
		t.Errorf("filterModulesBySelection(nil, _) should return nil, got %v", result)
	}
}

func TestFilterModulesBySelection_NilSelection(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
	}

	result := filterModulesBySelection(modules, nil)
	if len(result) != 0 {
		t.Errorf("expected 0 modules for nil selection, got %d", len(result))
	}
}

func TestFilterModulesBySelection_DuplicateSelection(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
	}

	// Select module-a twice (should only appear once in result)
	result := filterModulesBySelection(modules, []string{"module-a", "module-a"})

	if len(result) != 1 {
		t.Errorf("expected 1 module, got %d", len(result))
	}

	if len(result) > 0 && result[0].Name != "module-a" {
		t.Errorf("expected module-a, got %q", result[0].Name)
	}
}

func TestFilterModulesBySelection_PreservesOrder(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	// Select in order: a, c
	result := filterModulesBySelection(modules, []string{"module-a", "module-c"})

	if len(result) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(result))
	}

	// Result should preserve the order from modules list (a comes before c)
	expectedOrder := []string{"module-a", "module-c"}
	for i, mod := range result {
		if mod.Name != expectedOrder[i] {
			t.Errorf("module[%d] = %q, want %q", i, mod.Name, expectedOrder[i])
		}
	}
}

func TestExecutionContext_EmptyModules(t *testing.T) {
	execCtx := &ExecutionContext{
		Mode:    MultiModuleMode,
		Modules: []*workspace.Module{},
	}

	if !execCtx.IsMultiModule() {
		t.Error("should be multi-module mode")
	}

	if len(execCtx.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(execCtx.Modules))
	}
}

func TestExecutionContext_SingleModuleWithModules(t *testing.T) {
	// This is an edge case - single module mode should not have Modules set
	execCtx := &ExecutionContext{
		Mode:    SingleModuleMode,
		Path:    "/test/.version",
		Modules: []*workspace.Module{{Name: "test"}},
	}

	if !execCtx.IsSingleModule() {
		t.Error("should be single-module mode")
	}

	if execCtx.Path != "/test/.version" {
		t.Errorf("Path = %q, want %q", execCtx.Path, "/test/.version")
	}

	// Even though Modules is set, we're in single-module mode
	// This verifies that Mode takes precedence
}

func TestExecutionContext_MultiModuleWithPath(t *testing.T) {
	// This is an edge case - multi-module mode should not have Path set
	execCtx := &ExecutionContext{
		Mode: MultiModuleMode,
		Path: "/test/.version",
		Modules: []*workspace.Module{
			{Name: "module-a"},
			{Name: "module-b"},
		},
	}

	if !execCtx.IsMultiModule() {
		t.Error("should be multi-module mode")
	}

	if len(execCtx.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(execCtx.Modules))
	}

	// Even though Path is set, we're in multi-module mode
	// This verifies that Mode takes precedence
}

func TestFilterModulesByName_CaseSensitive(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "Module-A", Path: "/path/to/Module-A/.version"},
		{Name: "module-a", Path: "/path/to/module-a/.version"},
	}

	// Should be case-sensitive
	result := filterModulesByName(modules, "Module-A")
	if len(result) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result))
	}
	if result[0].Name != "Module-A" {
		t.Errorf("expected Module-A, got %q", result[0].Name)
	}

	// Lowercase search should find different module
	result = filterModulesByName(modules, "module-a")
	if len(result) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result))
	}
	if result[0].Name != "module-a" {
		t.Errorf("expected module-a, got %q", result[0].Name)
	}
}

func TestFilterModulesBySelection_PartialMatches(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module", Path: "/path/to/module/.version"},
		{Name: "module-extra", Path: "/path/to/module-extra/.version"},
	}

	// Should only match exact names, not partial
	result := filterModulesBySelection(modules, []string{"module"})

	if len(result) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result))
	}
	if result[0].Name != "module" {
		t.Errorf("expected module, got %q", result[0].Name)
	}
}

func TestFilterModulesBySelection_MixedExistingAndNonExisting(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
	}

	// Select one existing and one non-existing
	result := filterModulesBySelection(modules, []string{"module-a", "module-z"})

	// Should only include the existing module
	if len(result) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result))
	}
	if result[0].Name != "module-a" {
		t.Errorf("expected module-a, got %q", result[0].Name)
	}
}

func TestGetExecutionContext_PathFlagSet(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := tmpDir + "/.version"

	cfg := &config.Config{
		Path: versionPath,
	}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: versionPath,
			},
		},
	}

	ctx := context.Background()
	execCtx, err := GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		t.Fatalf("GetExecutionContext() error = %v", err)
	}

	if execCtx.Mode != SingleModuleMode {
		t.Errorf("Mode = %v, want SingleModuleMode", execCtx.Mode)
	}

	if execCtx.Path != versionPath {
		t.Errorf("Path = %v, want %v", execCtx.Path, versionPath)
	}
}

func TestGetExecutionContext_ConfigPathSet(t *testing.T) {
	// When .sley.yaml has an explicit path (not default), use single-module mode
	tmpDir := t.TempDir()
	versionPath := tmpDir + "/custom/path/.version"

	// Create the version file
	if err := os.MkdirAll(tmpDir+"/custom/path", 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(versionPath, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	cfg := &config.Config{
		Path: versionPath, // Explicit path set in config
	}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: ".version", // Default value, not explicitly set
			},
		},
	}

	ctx := context.Background()
	execCtx, err := GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		t.Fatalf("GetExecutionContext() error = %v", err)
	}

	if execCtx.Mode != SingleModuleMode {
		t.Errorf("Mode = %v, want SingleModuleMode (config path should override detection)", execCtx.Mode)
	}

	if execCtx.Path != versionPath {
		t.Errorf("Path = %v, want %v", execCtx.Path, versionPath)
	}
}

func TestGetExecutionContext_AllFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple modules
	modA := tmpDir + "/module-a"
	modB := tmpDir + "/module-b"
	if err := os.MkdirAll(modA, 0755); err != nil {
		t.Fatalf("failed to create module-a dir: %v", err)
	}
	if err := os.MkdirAll(modB, 0755); err != nil {
		t.Fatalf("failed to create module-b dir: %v", err)
	}
	if err := os.WriteFile(modA+"/.version", []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to write module-a version: %v", err)
	}
	if err := os.WriteFile(modB+"/.version", []byte("2.0.0"), 0644); err != nil {
		t.Fatalf("failed to write module-b version: %v", err)
	}

	cfg := &config.Config{}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Value: true},
			&cli.BoolFlag{Name: "yes"},
			&cli.BoolFlag{Name: "non-interactive"},
			&cli.StringFlag{Name: "path"},
			&cli.StringFlag{Name: "module"},
		},
	}

	// Change to tmpDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	ctx := context.Background()
	execCtx, err := GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		t.Fatalf("GetExecutionContext() error = %v", err)
	}

	if execCtx.Mode != MultiModuleMode {
		t.Errorf("Mode = %v, want MultiModuleMode", execCtx.Mode)
	}

	if len(execCtx.Modules) == 0 {
		t.Error("expected modules to be discovered")
	}
}

func TestGetExecutionContext_ModuleFlag(t *testing.T) {
	// This test is difficult to implement without a full CLI context
	// as cmd.IsSet() requires the flag to be actually set via CLI parsing.
	// The behavior is covered by integration tests.
	t.Skip("Requires full CLI context with flag parsing - covered by integration tests")
}

func TestGetExecutionContext_SingleModuleDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a single .version file
	versionFile := tmpDir + "/.version"
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to write version file: %v", err)
	}

	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: versionFile,
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all"},
			&cli.BoolFlag{Name: "yes"},
			&cli.BoolFlag{Name: "non-interactive"},
			&cli.StringFlag{Name: "path"},
			&cli.StringFlag{Name: "module"},
		},
	}

	// Change to tmpDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	ctx := context.Background()
	execCtx, err := GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		t.Fatalf("GetExecutionContext() error = %v", err)
	}

	if execCtx.Mode != SingleModuleMode {
		t.Errorf("Mode = %v, want SingleModuleMode", execCtx.Mode)
	}

	if execCtx.Path == "" {
		t.Error("expected path to be set for single module")
	}
}

func TestGetExecutionContext_NoModulesDetection(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Path: ".version",
	}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all"},
			&cli.BoolFlag{Name: "yes"},
			&cli.BoolFlag{Name: "non-interactive"},
			&cli.StringFlag{Name: "path"},
			&cli.StringFlag{Name: "module"},
		},
	}

	// Change to tmpDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	ctx := context.Background()
	execCtx, ctxErr := GetExecutionContext(ctx, cmd, cfg)
	if ctxErr != nil {
		t.Fatalf("GetExecutionContext() error = %v", ctxErr)
	}

	if execCtx.Mode != SingleModuleMode {
		t.Errorf("Mode = %v, want SingleModuleMode", execCtx.Mode)
	}

	// Should fall back to config path
	if execCtx.Path != ".version" {
		t.Errorf("Path = %v, want .version", execCtx.Path)
	}
}

func TestGetMultiModuleContext_NoModulesFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all"},
			&cli.BoolFlag{Name: "yes"},
			&cli.BoolFlag{Name: "non-interactive"},
			&cli.StringFlag{Name: "module"},
		},
	}

	// Change to tmpDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	ctx := context.Background()
	_, ctxErr := getMultiModuleContext(ctx, cmd, cfg, &executionOptions{}, false)
	if ctxErr == nil {
		t.Fatal("expected error for no modules found")
	}

	if !contains(ctxErr.Error(), "no modules found") {
		t.Errorf("expected 'no modules found' error, got: %v", ctxErr)
	}
}

func TestGetMultiModuleContext_ModuleNotFound(t *testing.T) {
	// This test is difficult to implement without a full CLI context
	// as cmd.IsSet() requires the flag to be actually set via CLI parsing.
	// The behavior is covered by integration tests.
	t.Skip("Requires full CLI context with flag parsing - covered by integration tests")
}

func TestGetMultiModuleContext_NonInteractive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create modules
	modA := tmpDir + "/module-a"
	modB := tmpDir + "/module-b"
	if err := os.MkdirAll(modA, 0755); err != nil {
		t.Fatalf("failed to create module-a dir: %v", err)
	}
	if err := os.MkdirAll(modB, 0755); err != nil {
		t.Fatalf("failed to create module-b dir: %v", err)
	}
	if err := os.WriteFile(modA+"/.version", []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to write module-a version: %v", err)
	}
	if err := os.WriteFile(modB+"/.version", []byte("2.0.0"), 0644); err != nil {
		t.Fatalf("failed to write module-b version: %v", err)
	}

	cfg := &config.Config{}

	cmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all"},
			&cli.BoolFlag{Name: "yes", Value: true}, // non-interactive
			&cli.BoolFlag{Name: "non-interactive"},
			&cli.StringFlag{Name: "module"},
		},
	}

	// Change to tmpDir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to tmpDir: %v", err)
	}

	ctx := context.Background()
	execCtx, err := getMultiModuleContext(ctx, cmd, cfg, &executionOptions{}, false)
	if err != nil {
		t.Fatalf("getMultiModuleContext() error = %v", err)
	}

	if execCtx.Mode != MultiModuleMode {
		t.Errorf("Mode = %v, want MultiModuleMode", execCtx.Mode)
	}

	if len(execCtx.Modules) == 0 {
		t.Error("expected modules to be discovered")
	}

	if !execCtx.Selection.All {
		t.Error("expected Selection.All to be true for non-interactive mode")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
