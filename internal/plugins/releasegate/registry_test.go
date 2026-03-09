package releasegate

import "testing"

func TestRegisterAndGet(t *testing.T) {
	Unregister()
	defer Unregister()

	cfg := &Config{Enabled: true}
	Register(cfg)

	rg := GetReleaseGateFn()
	if rg == nil {
		t.Fatal("expected release gate to be registered")
	}

	plugin, ok := rg.(*ReleaseGatePlugin)
	if !ok {
		t.Fatal("expected ReleaseGatePlugin type")
	}

	if !plugin.IsEnabled() {
		t.Error("expected plugin to be enabled")
	}
}

func TestGetReturnsNilWhenNotRegistered(t *testing.T) {
	Unregister()
	defer Unregister()

	rg := GetReleaseGateFn()
	if rg != nil {
		t.Error("expected nil when not registered")
	}
}

func TestRegistryUnregister(t *testing.T) {
	Unregister()
	defer Unregister()

	Register(&Config{Enabled: true})
	Unregister()

	rg := GetReleaseGateFn()
	if rg != nil {
		t.Error("expected nil after unregister")
	}
}
