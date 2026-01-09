package clix

import (
	"testing"

	"github.com/indaco/sley/internal/workspace"
)

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

func TestFilterModulesByNames(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "api", Path: "services/api/.version"},
		{Name: "web", Path: "apps/web/.version"},
		{Name: "shared", Path: "packages/shared/.version"},
		{Name: "auth", Path: "services/auth/.version"},
	}

	tests := []struct {
		name      string
		modules   []*workspace.Module
		names     []string
		wantCount int
		wantNames []string
	}{
		{
			name:      "filter multiple modules",
			modules:   modules,
			names:     []string{"api", "web"},
			wantCount: 2,
			wantNames: []string{"api", "web"},
		},
		{
			name:      "filter single module",
			modules:   modules,
			names:     []string{"shared"},
			wantCount: 1,
			wantNames: []string{"shared"},
		},
		{
			name:      "filter with comma-separated value",
			modules:   modules,
			names:     []string{"api,auth"},
			wantCount: 2,
			wantNames: []string{"api", "auth"},
		},
		{
			name:      "filter with mixed format",
			modules:   modules,
			names:     []string{"api", "web,shared"},
			wantCount: 3,
			wantNames: []string{"api", "web", "shared"},
		},
		{
			name:      "filter non-existent module",
			modules:   modules,
			names:     []string{"nonexistent"},
			wantCount: 0,
		},
		{
			name:      "empty names returns all modules",
			modules:   modules,
			names:     []string{},
			wantCount: 4,
		},
		{
			name:      "nil modules",
			modules:   nil,
			names:     []string{"api"},
			wantCount: 0,
		},
		{
			name:      "filter with spaces in comma-separated",
			modules:   modules,
			names:     []string{"api, web , shared"},
			wantCount: 3,
			wantNames: []string{"api", "web", "shared"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterModulesByNames(tt.modules, tt.names)
			if len(got) != tt.wantCount {
				t.Errorf("filterModulesByNames() returned %d modules, want %d", len(got), tt.wantCount)
			}

			if len(tt.wantNames) > 0 {
				gotNames := make(map[string]bool)
				for _, mod := range got {
					gotNames[mod.Name] = true
				}
				for _, wantName := range tt.wantNames {
					if !gotNames[wantName] {
						t.Errorf("filterModulesByNames() missing expected module %q", wantName)
					}
				}
			}
		})
	}
}

func TestFilterModulesByPattern(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "api", Path: "services/api/.version"},
		{Name: "auth", Path: "services/auth/.version"},
		{Name: "web", Path: "apps/web/.version"},
		{Name: "shared", Path: "packages/shared/.version"},
	}

	tests := []struct {
		name      string
		modules   []*workspace.Module
		pattern   string
		wantCount int
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "match by directory pattern",
			modules:   modules,
			pattern:   "services/*",
			wantCount: 2,
			wantNames: []string{"api", "auth"},
		},
		{
			name:      "match by module name",
			modules:   modules,
			pattern:   "api",
			wantCount: 1,
			wantNames: []string{"api"},
		},
		{
			name:      "match by full path",
			modules:   modules,
			pattern:   "apps/web/.version",
			wantCount: 1,
			wantNames: []string{"web"},
		},
		{
			name:      "no matches",
			modules:   modules,
			pattern:   "nonexistent/*",
			wantCount: 0,
		},
		{
			name:      "empty pattern returns all",
			modules:   modules,
			pattern:   "",
			wantCount: 4,
		},
		{
			name:      "wildcard name pattern",
			modules:   modules,
			pattern:   "a*",
			wantCount: 2,
			wantNames: []string{"api", "auth"},
		},
		{
			name:    "invalid pattern",
			modules: modules,
			pattern: "[invalid",
			wantErr: true,
		},
		{
			name:      "nil modules",
			modules:   nil,
			pattern:   "api",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterModulesByPattern(tt.modules, tt.pattern)

			if tt.wantErr {
				if err == nil {
					t.Error("filterModulesByPattern() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("filterModulesByPattern() unexpected error: %v", err)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("filterModulesByPattern() returned %d modules, want %d", len(got), tt.wantCount)
			}

			if len(tt.wantNames) > 0 {
				gotNames := make(map[string]bool)
				for _, mod := range got {
					gotNames[mod.Name] = true
				}
				for _, wantName := range tt.wantNames {
					if !gotNames[wantName] {
						t.Errorf("filterModulesByPattern() missing expected module %q", wantName)
					}
				}
			}
		})
	}
}
