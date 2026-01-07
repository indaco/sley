package plugins

import (
	"sync"
	"testing"

	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/changelogparser"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
	"github.com/indaco/sley/internal/semver"
)

// Mock implementations for testing
type mockCommitParser struct{ name string }

func (m *mockCommitParser) Name() string                   { return m.name }
func (m *mockCommitParser) Parse([]string) (string, error) { return "patch", nil }

type mockTagManager struct{ name string }

func (m *mockTagManager) Name() string        { return m.name }
func (m *mockTagManager) Description() string { return "mock" }
func (m *mockTagManager) Version() string     { return "v1.0.0" }
func (m *mockTagManager) CreateTag(version semver.SemVersion, message string) error {
	return nil
}
func (m *mockTagManager) TagExists(version semver.SemVersion) (bool, error) { return false, nil }
func (m *mockTagManager) GetLatestTag() (semver.SemVersion, error) {
	return semver.SemVersion{}, nil
}
func (m *mockTagManager) ValidateTagAvailable(version semver.SemVersion) error {
	return nil
}
func (m *mockTagManager) FormatTagName(version semver.SemVersion) string { return "v1.0.0" }
func (m *mockTagManager) PushTag(version semver.SemVersion) error        { return nil }
func (m *mockTagManager) DeleteTag(version semver.SemVersion) error      { return nil }
func (m *mockTagManager) ListTags() ([]string, error)                    { return nil, nil }

func TestPluginRegistry_CommitParser(t *testing.T) {
	t.Run("register and get", func(t *testing.T) {
		registry := NewPluginRegistry()
		parser := &mockCommitParser{name: "test-parser"}

		if err := registry.RegisterCommitParser(parser); err != nil {
			t.Fatalf("RegisterCommitParser failed: %v", err)
		}

		retrieved := registry.GetCommitParser()
		if retrieved == nil {
			t.Fatal("GetCommitParser returned nil")
		}
		if retrieved.Name() != "test-parser" {
			t.Errorf("expected name 'test-parser', got %q", retrieved.Name())
		}
	})

	t.Run("duplicate registration returns error", func(t *testing.T) {
		registry := NewPluginRegistry()
		parser1 := &mockCommitParser{name: "parser1"}
		parser2 := &mockCommitParser{name: "parser2"}

		if err := registry.RegisterCommitParser(parser1); err != nil {
			t.Fatalf("first registration failed: %v", err)
		}

		err := registry.RegisterCommitParser(parser2)
		if err == nil {
			t.Fatal("expected error on duplicate registration, got nil")
		}
	})

	t.Run("get returns nil when not registered", func(t *testing.T) {
		registry := NewPluginRegistry()
		if parser := registry.GetCommitParser(); parser != nil {
			t.Errorf("expected nil, got %v", parser)
		}
	})
}

func TestPluginRegistry_TagManager(t *testing.T) {
	t.Run("register and get", func(t *testing.T) {
		registry := NewPluginRegistry()
		tm := &mockTagManager{name: "test-tm"}

		if err := registry.RegisterTagManager(tm); err != nil {
			t.Fatalf("RegisterTagManager failed: %v", err)
		}

		retrieved := registry.GetTagManager()
		if retrieved == nil {
			t.Fatal("GetTagManager returned nil")
		}
		if retrieved.Name() != "test-tm" {
			t.Errorf("expected name 'test-tm', got %q", retrieved.Name())
		}
	})

	t.Run("duplicate registration returns error", func(t *testing.T) {
		registry := NewPluginRegistry()
		tm1 := &mockTagManager{name: "tm1"}
		tm2 := &mockTagManager{name: "tm2"}

		if err := registry.RegisterTagManager(tm1); err != nil {
			t.Fatalf("first registration failed: %v", err)
		}

		err := registry.RegisterTagManager(tm2)
		if err == nil {
			t.Fatal("expected error on duplicate registration, got nil")
		}
	})
}

func TestPluginRegistry_Reset(t *testing.T) {
	registry := NewPluginRegistry()

	// Register multiple plugins
	if err := registry.RegisterCommitParser(&mockCommitParser{name: "test"}); err != nil {
		t.Fatalf("RegisterCommitParser failed: %v", err)
	}
	if err := registry.RegisterTagManager(&mockTagManager{name: "test"}); err != nil {
		t.Fatalf("RegisterTagManager failed: %v", err)
	}

	// Verify they're registered
	if registry.GetCommitParser() == nil {
		t.Fatal("expected commit parser to be registered")
	}
	if registry.GetTagManager() == nil {
		t.Fatal("expected tag manager to be registered")
	}

	// Reset the registry
	registry.Reset()

	// Verify they're cleared
	if registry.GetCommitParser() != nil {
		t.Error("expected commit parser to be nil after reset")
	}
	if registry.GetTagManager() != nil {
		t.Error("expected tag manager to be nil after reset")
	}
}

func TestPluginRegistry_ThreadSafety(t *testing.T) {
	registry := NewPluginRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Only first registration should succeed
			_ = registry.RegisterCommitParser(&mockCommitParser{name: "parser"})
		}(i)
	}

	wg.Wait()

	// Verify exactly one is registered
	if parser := registry.GetCommitParser(); parser == nil {
		t.Fatal("expected a parser to be registered")
	}
}

func TestPluginRegistry_AllPluginTypes(t *testing.T) {
	registry := NewPluginRegistry()

	// Test all plugin types can be registered
	tests := []struct {
		name     string
		register func() error
		get      func() any
	}{
		{
			name: "CommitParser",
			register: func() error {
				return registry.RegisterCommitParser(&mockCommitParser{name: "test"})
			},
			get: func() any { return registry.GetCommitParser() },
		},
		{
			name: "TagManager",
			register: func() error {
				return registry.RegisterTagManager(&mockTagManager{name: "test"})
			},
			get: func() any { return registry.GetTagManager() },
		},
		{
			name: "VersionValidator",
			register: func() error {
				return registry.RegisterVersionValidator(
					versionvalidator.NewVersionValidator(versionvalidator.DefaultConfig()),
				)
			},
			get: func() any { return registry.GetVersionValidator() },
		},
		{
			name: "DependencyChecker",
			register: func() error {
				return registry.RegisterDependencyChecker(
					dependencycheck.NewDependencyChecker(dependencycheck.DefaultConfig()),
				)
			},
			get: func() any { return registry.GetDependencyChecker() },
		},
		{
			name: "ChangelogParser",
			register: func() error {
				return registry.RegisterChangelogParser(
					changelogparser.NewChangelogParser(changelogparser.DefaultConfig()),
				)
			},
			get: func() any { return registry.GetChangelogParser() },
		},
		{
			name: "ChangelogGenerator",
			register: func() error {
				cfg := changeloggenerator.DefaultConfig()
				gen, err := changeloggenerator.NewChangelogGenerator(cfg)
				if err != nil {
					return err
				}
				return registry.RegisterChangelogGenerator(gen)
			},
			get: func() any { return registry.GetChangelogGenerator() },
		},
		{
			name: "ReleaseGate",
			register: func() error {
				return registry.RegisterReleaseGate(
					releasegate.NewReleaseGate(releasegate.DefaultConfig()),
				)
			},
			get: func() any { return registry.GetReleaseGate() },
		},
		{
			name: "AuditLog",
			register: func() error {
				return registry.RegisterAuditLog(
					auditlog.NewAuditLog(auditlog.DefaultConfig()),
				)
			},
			get: func() any { return registry.GetAuditLog() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.register(); err != nil {
				t.Fatalf("registration failed: %v", err)
			}

			if plugin := tt.get(); plugin == nil {
				t.Error("expected plugin to be registered")
			}
		})
	}
}
