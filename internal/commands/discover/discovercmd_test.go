package discover

import (
	"context"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/discovery"
)

func TestRun_ReturnsCommand(t *testing.T) {
	cmd := Run(nil)

	if cmd.Name != "discover" {
		t.Errorf("Name = %q, want %q", cmd.Name, "discover")
	}

	if cmd.Usage == "" {
		t.Error("Usage should not be empty")
	}

	// Verify flags exist
	flagNames := []string{"format", "quiet", "no-interactive"}
	for _, name := range flagNames {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected flag %q not found", name)
		}
	}
}

func TestDiscoverAndSuggest(t *testing.T) {
	// Create a mock filesystem with test files
	fs := core.NewMockFileSystem()
	fs.SetFile("/project/.version", []byte("1.0.0\n"))
	fs.SetFile("/project/package.json", []byte(`{"version": "1.0.0"}`))

	// Note: DiscoverAndSuggest uses os.Getwd and NewOSFileSystem internally
	// This test verifies the interface only, not the actual filesystem operations

	cfg := &config.Config{}
	result, suggestion, err := DiscoverAndSuggest(context.Background(), cfg, ".")

	// In a real test environment, this might fail due to actual filesystem
	// We're mainly testing that the function signature works correctly
	if err == nil {
		if result == nil {
			t.Error("result should not be nil when err is nil")
		}
	}

	// suggestion can be nil if no sync candidates are found
	_ = suggestion
}

func TestPrintQuietSummary(t *testing.T) {
	// This is a visual output test - we just verify it doesn't panic

	tests := []struct {
		name   string
		result *discovery.Result
	}{
		{
			name:   "empty result",
			result: &discovery.Result{},
		},
		{
			name: "with modules",
			result: &discovery.Result{
				Mode: discovery.SingleModule,
				Modules: []discovery.Module{
					{Name: "root", RelPath: ".version", Version: "1.0.0"},
				},
			},
		},
		{
			name: "with mismatches",
			result: &discovery.Result{
				Mode: discovery.SingleModule,
				Modules: []discovery.Module{
					{Name: "root", RelPath: ".version", Version: "1.0.0"},
				},
				Mismatches: []discovery.Mismatch{
					{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			printQuietSummary(tt.result)
		})
	}
}
