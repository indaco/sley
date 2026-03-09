package versionvalidator

import "testing"

func TestRegisterAndGet(t *testing.T) {
	Unregister()
	defer Unregister()

	cfg := &Config{Enabled: true}
	Register(cfg)

	vv := GetVersionValidatorFn()
	if vv == nil {
		t.Fatal("expected version validator to be registered")
	}

	plugin, ok := vv.(*VersionValidatorPlugin)
	if !ok {
		t.Fatal("expected VersionValidatorPlugin type")
	}

	if !plugin.IsEnabled() {
		t.Error("expected plugin to be enabled")
	}
}

func TestGetReturnsNilWhenNotRegistered(t *testing.T) {
	Unregister()
	defer Unregister()

	vv := GetVersionValidatorFn()
	if vv != nil {
		t.Error("expected nil when not registered")
	}
}

func TestRegistryUnregister(t *testing.T) {
	Unregister()
	defer Unregister()

	Register(&Config{Enabled: true})
	Unregister()

	vv := GetVersionValidatorFn()
	if vv != nil {
		t.Error("expected nil after unregister")
	}
}
