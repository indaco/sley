package dependencycheck

import "testing"

func TestRegisterAndGet(t *testing.T) {
	ResetDependencyChecker()
	defer ResetDependencyChecker()

	cfg := &Config{Enabled: true, AutoSync: true}
	Register(cfg)

	dc := GetDependencyCheckerFn()
	if dc == nil {
		t.Fatal("expected dependency checker to be registered")
	}

	plugin, ok := dc.(*DependencyCheckerPlugin)
	if !ok {
		t.Fatal("expected DependencyCheckerPlugin type")
	}

	if !plugin.IsEnabled() {
		t.Error("expected plugin to be enabled")
	}
}

func TestGetReturnsNilWhenNotRegistered(t *testing.T) {
	ResetDependencyChecker()
	defer ResetDependencyChecker()

	dc := GetDependencyCheckerFn()
	if dc != nil {
		t.Error("expected nil when not registered")
	}
}

func TestReset(t *testing.T) {
	ResetDependencyChecker()
	defer ResetDependencyChecker()

	Register(&Config{Enabled: true})
	ResetDependencyChecker()

	dc := GetDependencyCheckerFn()
	if dc != nil {
		t.Error("expected nil after reset")
	}
}

func TestUnregister(t *testing.T) {
	ResetDependencyChecker()
	defer ResetDependencyChecker()

	Register(&Config{Enabled: true})
	Unregister()

	dc := GetDependencyCheckerFn()
	if dc != nil {
		t.Error("expected nil after unregister")
	}
}
