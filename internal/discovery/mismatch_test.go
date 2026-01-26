package discovery

import (
	"testing"
)

func TestDetectMismatches(t *testing.T) {
	tests := []struct {
		name      string
		result    *Result
		wantCount int
	}{
		{
			name:      "nil result",
			result:    nil,
			wantCount: 0,
		},
		{
			name: "no mismatches",
			result: &Result{
				Modules: []Module{
					{RelPath: ".version", Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{RelPath: "package.json", Version: "1.0.0"},
				},
			},
			wantCount: 0,
		},
		{
			name: "module mismatch",
			result: &Result{
				Modules: []Module{
					{RelPath: ".version", Version: "1.0.0"},
					{RelPath: "other/.version", Version: "2.0.0"},
				},
			},
			wantCount: 1,
		},
		{
			name: "manifest mismatch",
			result: &Result{
				Modules: []Module{
					{RelPath: ".version", Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{RelPath: "package.json", Version: "2.0.0"},
				},
			},
			wantCount: 1,
		},
		{
			name: "multiple mismatches",
			result: &Result{
				Modules: []Module{
					{RelPath: ".version", Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{RelPath: "package.json", Version: "2.0.0"},
					{RelPath: "Cargo.toml", Version: "3.0.0"},
				},
			},
			wantCount: 2,
		},
		{
			name: "empty version ignored",
			result: &Result{
				Modules: []Module{
					{RelPath: ".version", Version: "1.0.0"},
					{RelPath: "other/.version", Version: ""},
				},
			},
			wantCount: 0,
		},
		{
			name: "no primary version",
			result: &Result{
				Modules: []Module{
					{RelPath: "other/.version", Version: ""},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mismatches := DetectMismatches(tt.result)
			if len(mismatches) != tt.wantCount {
				t.Errorf("DetectMismatches() returned %d mismatches, want %d", len(mismatches), tt.wantCount)
			}
		})
	}
}

func TestDetectMismatchesWithCustomBase(t *testing.T) {
	result := &Result{
		Modules: []Module{
			{RelPath: ".version", Version: "1.0.0"},
		},
		Manifests: []ManifestSource{
			{RelPath: "package.json", Version: "2.0.0"},
		},
	}

	// With custom base that matches manifest
	mismatches := DetectMismatchesWithCustomBase(result, "2.0.0")
	if len(mismatches) != 1 {
		t.Errorf("expected 1 mismatch (module), got %d", len(mismatches))
	}

	if mismatches[0].Source != ".version" {
		t.Errorf("mismatch source = %q, want %q", mismatches[0].Source, ".version")
	}

	// Nil result
	nilMismatches := DetectMismatchesWithCustomBase(nil, "1.0.0")
	if nilMismatches != nil {
		t.Errorf("expected nil for nil result, got %v", nilMismatches)
	}

	// Empty base version
	emptyMismatches := DetectMismatchesWithCustomBase(result, "")
	if emptyMismatches != nil {
		t.Errorf("expected nil for empty base version, got %v", emptyMismatches)
	}
}

func TestGetUniqueVersions(t *testing.T) {
	tests := []struct {
		name      string
		result    *Result
		wantCount int
		wantFirst string
	}{
		{
			name:      "nil result",
			result:    nil,
			wantCount: 0,
		},
		{
			name: "single version",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
				},
			},
			wantCount: 1,
			wantFirst: "1.0.0",
		},
		{
			name: "multiple same version",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
					{Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{Version: "1.0.0"},
				},
			},
			wantCount: 1,
			wantFirst: "1.0.0",
		},
		{
			name: "multiple different versions",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
					{Version: "2.0.0"},
				},
			},
			wantCount: 2,
			wantFirst: "1.0.0", // Sorted alphabetically
		},
		{
			name: "empty versions ignored",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
					{Version: ""},
				},
			},
			wantCount: 1,
			wantFirst: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := GetUniqueVersions(tt.result)
			if len(versions) != tt.wantCount {
				t.Errorf("GetUniqueVersions() returned %d versions, want %d", len(versions), tt.wantCount)
			}
			if tt.wantCount > 0 && versions[0] != tt.wantFirst {
				t.Errorf("first version = %q, want %q", versions[0], tt.wantFirst)
			}
		})
	}
}

func TestIsVersionConsistent(t *testing.T) {
	tests := []struct {
		name   string
		result *Result
		want   bool
	}{
		{
			name:   "nil result",
			result: nil,
			want:   true,
		},
		{
			name: "single version",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
				},
			},
			want: true,
		},
		{
			name: "consistent versions",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{Version: "1.0.0"},
				},
			},
			want: true,
		},
		{
			name: "inconsistent versions",
			result: &Result{
				Modules: []Module{
					{Version: "1.0.0"},
				},
				Manifests: []ManifestSource{
					{Version: "2.0.0"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsVersionConsistent(tt.result); got != tt.want {
				t.Errorf("IsVersionConsistent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVersionSummary(t *testing.T) {
	result := &Result{
		Modules: []Module{
			{RelPath: "a/.version", Version: "1.0.0"},
			{RelPath: "b/.version", Version: "1.0.0"},
			{RelPath: "c/.version", Version: "2.0.0"},
		},
		Manifests: []ManifestSource{
			{RelPath: "package.json", Version: "1.0.0"},
		},
	}

	summaries := GetVersionSummary(result)

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	// First should be 1.0.0 with count 3 (sorted by count descending)
	if summaries[0].Version != "1.0.0" {
		t.Errorf("first version = %q, want %q", summaries[0].Version, "1.0.0")
	}
	if summaries[0].Count != 3 {
		t.Errorf("first count = %d, want 3", summaries[0].Count)
	}

	// Second should be 2.0.0 with count 1
	if summaries[1].Version != "2.0.0" {
		t.Errorf("second version = %q, want %q", summaries[1].Version, "2.0.0")
	}
	if summaries[1].Count != 1 {
		t.Errorf("second count = %d, want 1", summaries[1].Count)
	}
}

func TestGetVersionSummary_NilResult(t *testing.T) {
	summaries := GetVersionSummary(nil)
	if summaries != nil {
		t.Errorf("expected nil for nil result, got %v", summaries)
	}
}

func TestMismatchSorting(t *testing.T) {
	result := &Result{
		Modules: []Module{
			{RelPath: ".version", Version: "1.0.0"},
		},
		Manifests: []ManifestSource{
			{RelPath: "z.json", Version: "2.0.0"},
			{RelPath: "a.json", Version: "2.0.0"},
			{RelPath: "m.json", Version: "2.0.0"},
		},
	}

	mismatches := DetectMismatches(result)

	// Should be sorted alphabetically
	if len(mismatches) != 3 {
		t.Fatalf("expected 3 mismatches, got %d", len(mismatches))
	}

	expected := []string{"a.json", "m.json", "z.json"}
	for i, m := range mismatches {
		if m.Source != expected[i] {
			t.Errorf("mismatch[%d].Source = %q, want %q", i, m.Source, expected[i])
		}
	}
}
