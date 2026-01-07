package auditlog

import (
	"testing"

	"github.com/indaco/sley/internal/config"
)

func TestRegister(t *testing.T) {
	// Reset before test
	ResetAuditLog()

	cfg := &config.AuditLogConfig{
		Enabled: true,
		Path:    "test.json",
		Format:  "json",
	}

	Register(cfg)

	al := GetAuditLogFn()
	if al == nil {
		t.Fatal("expected audit log to be registered")
	}

	plugin, ok := al.(*AuditLogPlugin)
	if !ok {
		t.Fatal("expected AuditLogPlugin type")
	}

	if !plugin.IsEnabled() {
		t.Error("expected plugin to be enabled")
	}

	if plugin.GetConfig().GetPath() != "test.json" {
		t.Errorf("expected path 'test.json', got %q", plugin.GetConfig().GetPath())
	}

	// Cleanup
	ResetAuditLog()
}

func TestRegisterMultiple(t *testing.T) {
	// Reset before test
	ResetAuditLog()

	cfg1 := &config.AuditLogConfig{
		Enabled: true,
		Path:    "first.json",
	}

	Register(cfg1)
	first := GetAuditLogFn()

	// Try to register another
	cfg2 := &config.AuditLogConfig{
		Enabled: true,
		Path:    "second.json",
	}

	Register(cfg2)
	second := GetAuditLogFn()

	// Should still be the first one
	if first != second {
		t.Error("expected second registration to be ignored")
	}

	plugin := first.(*AuditLogPlugin)
	if plugin.GetConfig().GetPath() != "first.json" {
		t.Error("expected first registration to remain")
	}

	// Cleanup
	ResetAuditLog()
}

func TestFromConfigStruct(t *testing.T) {
	tests := []struct {
		name     string
		input    *config.AuditLogConfig
		wantPath string
	}{
		{
			name:     "nil config uses defaults",
			input:    nil,
			wantPath: ".version-history.json",
		},
		{
			name: "custom config",
			input: &config.AuditLogConfig{
				Enabled:          true,
				Path:             "custom.json",
				Format:           "yaml",
				IncludeAuthor:    true,
				IncludeTimestamp: false,
				IncludeCommitSHA: true,
				IncludeBranch:    false,
			},
			wantPath: "custom.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := FromConfigStruct(tt.input)
			if cfg.GetPath() != tt.wantPath {
				t.Errorf("expected path %q, got %q", tt.wantPath, cfg.GetPath())
			}
		})
	}
}
