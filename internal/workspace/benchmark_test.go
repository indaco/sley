package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
)

// setupBenchmarkFS creates a mock filesystem with the specified number of modules.
func setupBenchmarkFS(numModules int) *core.MockFileSystem {
	fs := core.NewMockFileSystem()

	for i := range numModules {
		moduleName := fmt.Sprintf("module-%04d", i)
		versionPath := filepath.Join("/project", moduleName, ".version")
		fs.SetFile(versionPath, []byte("1.0.0"))
		_ = fs.MkdirAll(context.Background(), filepath.Dir(versionPath), 0755)
	}

	return fs
}

// setupBenchmarkFSWithNesting creates a mock filesystem with nested modules.
func setupBenchmarkFSWithNesting(numModules int, nestingLevel int) *core.MockFileSystem {
	fs := core.NewMockFileSystem()

	for i := range numModules {
		// Create nested path
		pathComponents := []string{"/project"}
		for j := range nestingLevel {
			pathComponents = append(pathComponents, fmt.Sprintf("level%d", j))
		}
		pathComponents = append(pathComponents, fmt.Sprintf("module-%04d", i))

		modulePath := filepath.Join(pathComponents...)
		versionPath := filepath.Join(modulePath, ".version")
		fs.SetFile(versionPath, []byte("1.0.0"))
		_ = fs.MkdirAll(context.Background(), filepath.Dir(versionPath), 0755)
	}

	return fs
}

// setupBenchmarkFSWithExcludes creates a filesystem with modules in both included and excluded directories.
func setupBenchmarkFSWithExcludes(numIncluded, numExcluded int) *core.MockFileSystem {
	fs := core.NewMockFileSystem()

	// Add included modules
	for i := range numIncluded {
		moduleName := fmt.Sprintf("module-%04d", i)
		versionPath := filepath.Join("/project", moduleName, ".version")
		fs.SetFile(versionPath, []byte("1.0.0"))
		_ = fs.MkdirAll(context.Background(), filepath.Dir(versionPath), 0755)
	}

	// Add excluded modules
	excludeDirs := []string{"node_modules", ".git", "vendor", "build", "dist"}
	for i := range numExcluded {
		excludeDir := excludeDirs[i%len(excludeDirs)]
		moduleName := fmt.Sprintf("excluded-%04d", i)
		versionPath := filepath.Join("/project", excludeDir, moduleName, ".version")
		fs.SetFile(versionPath, []byte("1.0.0"))
		_ = fs.MkdirAll(context.Background(), filepath.Dir(versionPath), 0755)
	}

	return fs
}

func BenchmarkDetector_10Modules(b *testing.B) {
	fs := setupBenchmarkFS(10)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_50Modules(b *testing.B) {
	fs := setupBenchmarkFS(50)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_100Modules(b *testing.B) {
	fs := setupBenchmarkFS(100)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_200Modules(b *testing.B) {
	fs := setupBenchmarkFS(200)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_NestedModules(b *testing.B) {
	fs := setupBenchmarkFSWithNesting(50, 5)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_WithExcludes(b *testing.B) {
	fs := setupBenchmarkFSWithExcludes(50, 100)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_MaxDepthLimit(b *testing.B) {
	fs := setupBenchmarkFSWithNesting(50, 10)

	maxDepth := 5
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				ModuleMaxDepth: &maxDepth,
			},
		},
	}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_NonRecursive(b *testing.B) {
	fs := setupBenchmarkFS(50)

	recursive := false
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Recursive: &recursive,
			},
		},
	}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetectContext_SingleModule(b *testing.B) {
	fs := setupTestFS(map[string]string{
		"/project/.version": "1.0.0",
	})
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DetectContext(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DetectContext failed: %v", err)
		}
	}
}

func BenchmarkDetectContext_MultiModule(b *testing.B) {
	fs := setupBenchmarkFS(50)
	cfg := &config.Config{}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DetectContext(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DetectContext failed: %v", err)
		}
	}
}

func BenchmarkDetectContext_ExplicitModules(b *testing.B) {
	fs := setupBenchmarkFS(50)

	// Create explicit module configs
	enabled := true
	modules := make([]config.ModuleConfig, 50)
	for i := range 50 {
		moduleName := fmt.Sprintf("module-%04d", i)
		modules[i] = config.ModuleConfig{
			Name:    moduleName,
			Path:    filepath.Join("/project", moduleName, ".version"),
			Enabled: &enabled,
		}
	}

	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{
			Modules: modules,
		},
	}
	detector := NewDetector(fs, cfg)

	for b.Loop() {
		_, err := detector.DetectContext(context.Background(), "/project")
		if err != nil {
			b.Fatalf("DetectContext failed: %v", err)
		}
	}
}

// Real filesystem benchmarks (for comparison)

func BenchmarkDetector_RealFS_10Modules(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping real filesystem benchmark in short mode")
	}

	tmpDir := b.TempDir()

	// Create 10 modules
	for i := range 10 {
		moduleName := fmt.Sprintf("module-%04d", i)
		moduleDir := filepath.Join(tmpDir, moduleName)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		versionPath := filepath.Join(moduleDir, ".version")
		if err := os.WriteFile(versionPath, []byte("1.0.0"), 0644); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}

	cfg := &config.Config{}
	detector := NewDetector(core.NewOSFileSystem(), cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), tmpDir)
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_RealFS_50Modules(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping real filesystem benchmark in short mode")
	}

	tmpDir := b.TempDir()

	// Create 50 modules
	for i := range 50 {
		moduleName := fmt.Sprintf("module-%04d", i)
		moduleDir := filepath.Join(tmpDir, moduleName)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		versionPath := filepath.Join(moduleDir, ".version")
		if err := os.WriteFile(versionPath, []byte("1.0.0"), 0644); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}

	cfg := &config.Config{}
	detector := NewDetector(core.NewOSFileSystem(), cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), tmpDir)
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

func BenchmarkDetector_RealFS_100Modules(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping real filesystem benchmark in short mode")
	}

	tmpDir := b.TempDir()

	// Create 100 modules
	for i := range 100 {
		moduleName := fmt.Sprintf("module-%04d", i)
		moduleDir := filepath.Join(tmpDir, moduleName)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		versionPath := filepath.Join(moduleDir, ".version")
		if err := os.WriteFile(versionPath, []byte("1.0.0"), 0644); err != nil {
			b.Fatalf("Failed to write file: %v", err)
		}
	}

	cfg := &config.Config{}
	detector := NewDetector(core.NewOSFileSystem(), cfg)

	for b.Loop() {
		_, err := detector.DiscoverModules(context.Background(), tmpDir)
		if err != nil {
			b.Fatalf("DiscoverModules failed: %v", err)
		}
	}
}

// =============================================================================
// Executor Benchmarks
// =============================================================================

// benchmarkNoOpOperation is a no-op operation for benchmark purposes.
type benchmarkNoOpOperation struct{}

func (o *benchmarkNoOpOperation) Execute(_ context.Context, _ *Module) error {
	return nil
}

func (o *benchmarkNoOpOperation) Name() string {
	return "benchmark-noop"
}

// setupBenchmarkModules creates a slice of modules for benchmarking.
func setupBenchmarkModules(count int) []*Module {
	modules := make([]*Module, count)
	for i := range count {
		modules[i] = &Module{
			Name:           fmt.Sprintf("module-%04d", i),
			Path:           fmt.Sprintf("/project/module-%04d/.version", i),
			Dir:            fmt.Sprintf("/project/module-%04d", i),
			CurrentVersion: "1.0.0",
		}
	}
	return modules
}

func BenchmarkExecutor_Sequential_10Modules(b *testing.B) {
	modules := setupBenchmarkModules(10)
	executor := NewExecutor(WithParallel(false))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Sequential_50Modules(b *testing.B) {
	modules := setupBenchmarkModules(50)
	executor := NewExecutor(WithParallel(false))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Sequential_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	executor := NewExecutor(WithParallel(false))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Sequential_200Modules(b *testing.B) {
	modules := setupBenchmarkModules(200)
	executor := NewExecutor(WithParallel(false))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Parallel_10Modules(b *testing.B) {
	modules := setupBenchmarkModules(10)
	executor := NewExecutor(WithParallel(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Parallel_50Modules(b *testing.B) {
	modules := setupBenchmarkModules(50)
	executor := NewExecutor(WithParallel(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Parallel_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	executor := NewExecutor(WithParallel(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Parallel_200Modules(b *testing.B) {
	modules := setupBenchmarkModules(200)
	executor := NewExecutor(WithParallel(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

// Executor benchmarks with fail-fast behavior

func BenchmarkExecutor_Sequential_FailFast_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	executor := NewExecutor(WithParallel(false), WithFailFast(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkExecutor_Parallel_FailFast_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	executor := NewExecutor(WithParallel(true), WithFailFast(true))
	operation := &benchmarkNoOpOperation{}

	for b.Loop() {
		_, err := executor.Run(b.Context(), modules, operation)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

// =============================================================================
// Formatter Benchmarks
// =============================================================================

// setupBenchmarkResults creates execution results for benchmarking.
func setupBenchmarkResults(count int) []ExecutionResult {
	results := make([]ExecutionResult, count)
	for i := range count {
		results[i] = ExecutionResult{
			Module: &Module{
				Name:           fmt.Sprintf("module-%04d", i),
				Path:           fmt.Sprintf("/project/module-%04d/.version", i),
				Dir:            fmt.Sprintf("/project/module-%04d", i),
				CurrentVersion: "1.0.0",
			},
			Success:    true,
			OldVersion: "1.0.0",
			NewVersion: "1.0.1",
		}
	}
	return results
}

func BenchmarkTextFormatter_50Results(b *testing.B) {
	results := setupBenchmarkResults(50)
	formatter := NewTextFormatter("Bump Results")

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkTextFormatter_100Results(b *testing.B) {
	results := setupBenchmarkResults(100)
	formatter := NewTextFormatter("Bump Results")

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkJSONFormatter_50Results(b *testing.B) {
	results := setupBenchmarkResults(50)
	formatter := NewJSONFormatter()

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkJSONFormatter_100Results(b *testing.B) {
	results := setupBenchmarkResults(100)
	formatter := NewJSONFormatter()

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkTableFormatter_50Results(b *testing.B) {
	results := setupBenchmarkResults(50)
	formatter := NewTableFormatter("Bump Results")

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkTableFormatter_100Results(b *testing.B) {
	results := setupBenchmarkResults(100)
	formatter := NewTableFormatter("Bump Results")

	for b.Loop() {
		_ = formatter.FormatResults(results)
	}
}

func BenchmarkTextFormatter_ModuleList_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	formatter := NewTextFormatter("")

	for b.Loop() {
		_ = formatter.FormatModuleList(modules)
	}
}

func BenchmarkJSONFormatter_ModuleList_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	formatter := NewJSONFormatter()

	for b.Loop() {
		_ = formatter.FormatModuleList(modules)
	}
}

func BenchmarkTableFormatter_ModuleList_100Modules(b *testing.B) {
	modules := setupBenchmarkModules(100)
	formatter := NewTableFormatter("")

	for b.Loop() {
		_ = formatter.FormatModuleList(modules)
	}
}
