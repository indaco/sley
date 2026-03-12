package workspace

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

// mockOperation is a test helper that implements Operation.
type mockOperation struct {
	name     string
	execFunc func(ctx context.Context, mod *Module) error
}

func (m *mockOperation) Execute(ctx context.Context, mod *Module) error {
	if m.execFunc != nil {
		return m.execFunc(ctx, mod)
	}
	return nil
}

func (m *mockOperation) Name() string {
	return m.name
}

func TestNewExecutor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opts     []ExecutorOption
		wantPar  bool
		wantFail bool
	}{
		{
			name:     "default executor",
			opts:     nil,
			wantPar:  false,
			wantFail: false,
		},
		{
			name:     "parallel executor",
			opts:     []ExecutorOption{WithParallel(true)},
			wantPar:  true,
			wantFail: false,
		},
		{
			name:     "fail-fast executor",
			opts:     []ExecutorOption{WithFailFast(true)},
			wantPar:  false,
			wantFail: true,
		},
		{
			name:     "parallel and fail-fast",
			opts:     []ExecutorOption{WithParallel(true), WithFailFast(true)},
			wantPar:  true,
			wantFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := NewExecutor(tt.opts...)
			if e.parallel != tt.wantPar {
				t.Errorf("NewExecutor() parallel = %v, want %v", e.parallel, tt.wantPar)
			}
			if e.failFast != tt.wantFail {
				t.Errorf("NewExecutor() failFast = %v, want %v", e.failFast, tt.wantFail)
			}
		})
	}
}

func TestExecutor_Run_EmptyModules(t *testing.T) {
	t.Parallel()
	e := NewExecutor()
	op := &mockOperation{name: "test"}

	_, err := e.Run(context.Background(), nil, op)
	if err == nil {
		t.Error("Run() with nil modules should return error")
	}

	_, err = e.Run(context.Background(), []*Module{}, op)
	if err == nil {
		t.Error("Run() with empty modules should return error")
	}
}

func TestExecutor_Run_NilOperation(t *testing.T) {
	t.Parallel()
	e := NewExecutor()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
	}

	_, err := e.Run(context.Background(), modules, nil)
	if err == nil {
		t.Error("Run() with nil operation should return error")
	}
}

func TestExecutor_Run_Success(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version", CurrentVersion: "1.0.0"},
		{Name: "module-b", Path: "/path/to/module-b/.version", CurrentVersion: "2.0.0"},
	}

	executed := make([]string, 0)
	op := &mockOperation{
		name: "test",
		execFunc: func(ctx context.Context, mod *Module) error {
			executed = append(executed, mod.Name)
			return nil
		},
	}

	e := NewExecutor()
	results, err := e.Run(context.Background(), modules, op)

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if len(results) != len(modules) {
		t.Errorf("Run() returned %d results, want %d", len(results), len(modules))
	}

	for _, result := range results {
		if !result.Success {
			t.Errorf("Result for %s should be successful", result.Module.Name)
		}
		if result.Error != nil {
			t.Errorf("Result for %s should not have error", result.Module.Name)
		}
	}

	if len(executed) != len(modules) {
		t.Errorf("Operation executed %d times, want %d", len(executed), len(modules))
	}
}

func TestExecutor_Run_PartialFailure(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	expectedErr := errors.New("operation failed")
	op := &mockOperation{
		name: "test",
		execFunc: func(ctx context.Context, mod *Module) error {
			if mod.Name == "module-b" {
				return expectedErr
			}
			return nil
		},
	}

	e := NewExecutor()
	results, err := e.Run(context.Background(), modules, op)

	// Without fail-fast, all operations should complete
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if len(results) != len(modules) {
		t.Errorf("Run() returned %d results, want %d", len(results), len(modules))
	}

	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successful results, got %d", successCount)
	}

	if errorCount != 1 {
		t.Errorf("Expected 1 failed result, got %d", errorCount)
	}
}

func TestExecutor_Run_FailFast(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	expectedErr := errors.New("operation failed")
	executed := make([]string, 0)
	op := &mockOperation{
		name: "test",
		execFunc: func(ctx context.Context, mod *Module) error {
			executed = append(executed, mod.Name)
			if mod.Name == "module-b" {
				return expectedErr
			}
			return nil
		},
	}

	e := NewExecutor(WithFailFast(true))
	results, err := e.Run(context.Background(), modules, op)

	// With fail-fast, execution should stop on first error
	if err == nil {
		t.Error("Run() with fail-fast should return error")
	}

	// Should have results for modules up to and including the failure
	if len(results) < 1 {
		t.Errorf("Run() returned %d results, expected at least 1", len(results))
	}
}

func TestExecutor_Run_ContextCancellation(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	op := &mockOperation{
		name: "test",
		execFunc: func(ctx context.Context, mod *Module) error {
			if mod.Name == "module-a" {
				cancel() // Cancel after first module
			}
			// Check if context was cancelled instead of sleeping
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}

	e := NewExecutor()
	results, err := e.Run(ctx, modules, op)

	// Should get context cancellation error
	if err == nil {
		t.Error("Run() should return context cancellation error")
	}

	// Should have at least one result
	if len(results) == 0 {
		t.Error("Run() should have at least one result before cancellation")
	}
}

func TestHasErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		results []ExecutionResult
		want    bool
	}{
		{
			name: "no errors",
			results: []ExecutionResult{
				{Success: true},
				{Success: true},
			},
			want: false,
		},
		{
			name: "has errors",
			results: []ExecutionResult{
				{Success: true},
				{Success: false, Error: errors.New("error")},
			},
			want: true,
		},
		{
			name:    "empty results",
			results: []ExecutionResult{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HasErrors(tt.results)
			if got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuccessCount(t *testing.T) {
	t.Parallel()
	results := []ExecutionResult{
		{Success: true},
		{Success: false},
		{Success: true},
	}

	got := SuccessCount(results)
	want := 2

	if got != want {
		t.Errorf("SuccessCount() = %d, want %d", got, want)
	}
}

func TestErrorCount(t *testing.T) {
	t.Parallel()
	results := []ExecutionResult{
		{Success: true},
		{Success: false},
		{Success: true},
	}

	got := ErrorCount(results)
	want := 1

	if got != want {
		t.Errorf("ErrorCount() = %d, want %d", got, want)
	}
}

func TestTotalDuration(t *testing.T) {
	t.Parallel()
	results := []ExecutionResult{
		{Duration: 100 * time.Millisecond},
		{Duration: 200 * time.Millisecond},
		{Duration: 150 * time.Millisecond},
	}

	got := TotalDuration(results)
	want := 450 * time.Millisecond

	if got != want {
		t.Errorf("TotalDuration() = %v, want %v", got, want)
	}
}

func TestExecutor_RunParallel_Success(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version", CurrentVersion: "1.0.0"},
		{Name: "module-b", Path: "/path/to/module-b/.version", CurrentVersion: "2.0.0"},
		{Name: "module-c", Path: "/path/to/module-c/.version", CurrentVersion: "3.0.0"},
	}

	var mu sync.Mutex
	executed := make(map[string]bool)
	op := &mockOperation{
		name: "test-parallel",
		execFunc: func(ctx context.Context, mod *Module) error {
			mu.Lock()
			executed[mod.Name] = true
			mu.Unlock()
			return nil
		},
	}

	e := NewExecutor(WithParallel(true))
	results, err := e.Run(context.Background(), modules, op)

	if err != nil {
		t.Errorf("RunParallel() unexpected error: %v", err)
	}

	if len(results) != len(modules) {
		t.Errorf("RunParallel() returned %d results, want %d", len(results), len(modules))
	}

	// All results should be successful
	for _, result := range results {
		if !result.Success {
			t.Errorf("Result for %s should be successful", result.Module.Name)
		}
		if result.Error != nil {
			t.Errorf("Result for %s should not have error", result.Module.Name)
		}
	}

	// Verify all modules were executed
	if len(executed) != len(modules) {
		t.Errorf("Expected %d modules to be executed, got %d", len(modules), len(executed))
	}
}

func TestExecutor_RunParallel_PartialFailure(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
		{Name: "module-d", Path: "/path/to/module-d/.version"},
	}

	expectedErr := errors.New("operation failed")
	op := &mockOperation{
		name: "test-parallel-fail",
		execFunc: func(ctx context.Context, mod *Module) error {
			if mod.Name == "module-b" || mod.Name == "module-d" {
				return expectedErr
			}
			return nil
		},
	}

	e := NewExecutor(WithParallel(true))
	results, err := e.Run(context.Background(), modules, op)

	// Without fail-fast, all operations should complete even with failures
	if err != nil {
		t.Errorf("RunParallel() without fail-fast should not return error, got: %v", err)
	}

	if len(results) != len(modules) {
		t.Errorf("RunParallel() returned %d results, want %d", len(results), len(modules))
	}

	// Count successes and failures
	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successful results, got %d", successCount)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 failed results, got %d", errorCount)
	}
}

func TestExecutor_RunParallel_FailFast(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
		{Name: "module-d", Path: "/path/to/module-d/.version"},
	}

	expectedErr := errors.New("operation failed")
	var mu sync.Mutex
	executed := make(map[string]bool)

	op := &mockOperation{
		name: "test-parallel-failfast",
		execFunc: func(ctx context.Context, mod *Module) error {
			mu.Lock()
			executed[mod.Name] = true
			mu.Unlock()

			// module-b fails quickly
			if mod.Name == "module-b" {
				return expectedErr
			}

			// Other modules take longer - should be cancelled if fail-fast works
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
	}

	e := NewExecutor(WithParallel(true), WithFailFast(true))
	results, err := e.Run(context.Background(), modules, op)

	// With fail-fast, should get an error
	if err == nil {
		t.Error("RunParallel() with fail-fast should return error on failure")
	}

	if len(results) != len(modules) {
		t.Errorf("RunParallel() returned %d results, want %d", len(results), len(modules))
	}

	// Should have at least one error result
	hasError := false
	for _, result := range results {
		if !result.Success {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("Expected at least one error result with fail-fast")
	}
}

func TestExecutor_RunParallel_ContextCancellation(t *testing.T) {
	t.Parallel()
	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
		{Name: "module-c", Path: "/path/to/module-c/.version"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	op := &mockOperation{
		name: "test-parallel-cancel",
		execFunc: func(ctx context.Context, mod *Module) error {
			// Simulate long-running operation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		},
	}

	e := NewExecutor(WithParallel(true))
	results, _ := e.Run(ctx, modules, op)

	// All results should be present (though some may have context errors)
	if len(results) != len(modules) {
		t.Errorf("RunParallel() returned %d results, want %d", len(results), len(modules))
	}
}

func TestExecutor_VersionTracking(t *testing.T) {
	t.Parallel()
	module := &Module{
		Name:           "test-module",
		Path:           "/path/to/.version",
		CurrentVersion: "1.0.0",
	}

	op := &mockOperation{
		name: "version-update",
		execFunc: func(ctx context.Context, mod *Module) error {
			// Simulate version update
			mod.CurrentVersion = "1.1.0"
			return nil
		},
	}

	e := NewExecutor()
	results, err := e.Run(context.Background(), []*Module{module}, op)

	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.OldVersion != "1.0.0" {
		t.Errorf("OldVersion = %q, want %q", result.OldVersion, "1.0.0")
	}
	if result.NewVersion != "1.1.0" {
		t.Errorf("NewVersion = %q, want %q", result.NewVersion, "1.1.0")
	}
}

func TestWithParallel(t *testing.T) {
	t.Parallel()
	e := NewExecutor(WithParallel(true))
	if !e.parallel {
		t.Error("WithParallel(true) should set parallel to true")
	}

	e = NewExecutor(WithParallel(false))
	if e.parallel {
		t.Error("WithParallel(false) should set parallel to false")
	}
}

func TestWithFailFast(t *testing.T) {
	t.Parallel()
	e := NewExecutor(WithFailFast(true))
	if !e.failFast {
		t.Error("WithFailFast(true) should set failFast to true")
	}

	e = NewExecutor(WithFailFast(false))
	if e.failFast {
		t.Error("WithFailFast(false) should set failFast to false")
	}
}

// Race Detector Stress Tests
// These tests are specifically designed to catch race conditions when run with -race flag.

func TestExecutor_RaceDetector_ParallelModuleAccess(t *testing.T) {
	t.Parallel(
	// Test concurrent access to module state
	)

	const numModules = 50
	modules := make([]*Module, numModules)
	for i := range numModules {
		modules[i] = &Module{
			Name:           "module-" + string(rune('a'+i%26)),
			Path:           "/path/to/module/.version",
			CurrentVersion: "1.0.0",
		}
	}

	var counter int64
	var mu sync.Mutex

	op := &mockOperation{
		name: "race-test",
		execFunc: func(ctx context.Context, mod *Module) error {
			mu.Lock()
			counter++
			mu.Unlock()
			// Simulate version update
			mod.CurrentVersion = "1.1.0"
			return nil
		},
	}

	e := NewExecutor(WithParallel(true))
	results, err := e.Run(context.Background(), modules, op)

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	mu.Lock()
	if counter != int64(numModules) {
		t.Errorf("Expected %d operations, got %d", numModules, counter)
	}
	mu.Unlock()

	if len(results) != numModules {
		t.Errorf("Expected %d results, got %d", numModules, len(results))
	}
}

func TestExecutor_RaceDetector_ParallelFailFastCancellation(t *testing.T) {
	t.Parallel(
	// Test that fail-fast correctly cancels concurrent operations without races
	)

	const numModules = 20
	modules := make([]*Module, numModules)
	for i := range numModules {
		modules[i] = &Module{
			Name:           "module-" + string(rune('a'+i%26)),
			Path:           "/path/to/module/.version",
			CurrentVersion: "1.0.0",
		}
	}

	var started int64
	var mu sync.Mutex

	op := &mockOperation{
		name: "fail-fast-race-test",
		execFunc: func(ctx context.Context, mod *Module) error {
			mu.Lock()
			started++
			count := started
			mu.Unlock()

			// First few operations succeed quickly
			if count <= 3 {
				return nil
			}

			// Module at position 5 fails
			if count == 5 {
				return errors.New("intentional failure")
			}

			// Other modules wait and respect context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
	}

	e := NewExecutor(WithParallel(true), WithFailFast(true))
	_, err := e.Run(context.Background(), modules, op)

	// Should have an error from fail-fast
	if err == nil {
		t.Error("Expected error from fail-fast, got nil")
	}
}

func TestExecutor_RaceDetector_ResultsSliceAccess(t *testing.T) {
	t.Parallel(
	// Test that results slice is properly synchronized
	)

	const numModules = 30
	modules := make([]*Module, numModules)
	for i := range numModules {
		modules[i] = &Module{
			Name:           "module-" + string(rune('a'+i%26)),
			Path:           "/path/to/module/.version",
			CurrentVersion: "1.0.0",
		}
	}

	op := &mockOperation{
		name: "results-race-test",
		execFunc: func(ctx context.Context, mod *Module) error {
			// Yield to increase chance of interleaving for race detection
			runtime.Gosched()
			mod.CurrentVersion = "2.0.0"
			return nil
		},
	}

	e := NewExecutor(WithParallel(true))
	results, err := e.Run(context.Background(), modules, op)

	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	// Verify all results are properly set
	if len(results) != numModules {
		t.Errorf("Expected %d results, got %d", numModules, len(results))
	}

	for i, result := range results {
		if result.Module == nil {
			t.Errorf("Result %d has nil module", i)
		}
	}
}

func TestExecutor_RaceDetector_ConcurrentContextCancellation(t *testing.T) {
	t.Parallel(
	// Test concurrent context cancellation handling
	)

	const numModules = 25
	modules := make([]*Module, numModules)
	for i := range numModules {
		modules[i] = &Module{
			Name:           "module-" + string(rune('a'+i%26)),
			Path:           "/path/to/module/.version",
			CurrentVersion: "1.0.0",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	var mu sync.Mutex
	completed := 0

	op := &mockOperation{
		name: "cancel-race-test",
		execFunc: func(ctx context.Context, mod *Module) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(20 * time.Millisecond):
				mu.Lock()
				completed++
				mu.Unlock()
				return nil
			}
		},
	}

	e := NewExecutor(WithParallel(true))
	results, _ := e.Run(ctx, modules, op)

	// All results should be present regardless of completion status
	if len(results) != numModules {
		t.Errorf("Expected %d results, got %d", numModules, len(results))
	}
}

func TestExecutor_RaceDetector_MixedSuccessFailure(t *testing.T) {
	t.Parallel(
	// Test mixed success/failure results with concurrent access
	)

	const numModules = 40
	modules := make([]*Module, numModules)
	for i := range numModules {
		modules[i] = &Module{
			Name:           "module-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i%10)),
			Path:           "/path/to/module/.version",
			CurrentVersion: "1.0.0",
		}
	}

	op := &mockOperation{
		name: "mixed-race-test",
		execFunc: func(ctx context.Context, mod *Module) error {
			// Fail every 5th module
			if mod.Name[len(mod.Name)-1]%5 == 0 {
				return errors.New("planned failure")
			}
			mod.CurrentVersion = "2.0.0"
			return nil
		},
	}

	e := NewExecutor(WithParallel(true))
	results, err := e.Run(context.Background(), modules, op)

	if err != nil {
		t.Errorf("Run() without fail-fast should not return error: %v", err)
	}

	if len(results) != numModules {
		t.Errorf("Expected %d results, got %d", numModules, len(results))
	}

	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	// Should have some failures
	if errorCount == 0 {
		t.Error("Expected at least one failure")
	}
	// Should have some successes
	if successCount == 0 {
		t.Error("Expected at least one success")
	}
}
