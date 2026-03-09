package changelogparser

import "testing"

func TestRegisterAndGet(t *testing.T) {
	ResetChangelogParser()
	defer ResetChangelogParser()

	cfg := &Config{Enabled: true, Path: "CHANGELOG.md"}
	Register(cfg)

	cp := GetChangelogParserFn()
	if cp == nil {
		t.Fatal("expected changelog parser to be registered")
	}

	plugin, ok := cp.(*ChangelogParserPlugin)
	if !ok {
		t.Fatal("expected ChangelogParserPlugin type")
	}

	if !plugin.IsEnabled() {
		t.Error("expected plugin to be enabled")
	}
}

func TestGetReturnsNilWhenNotRegistered(t *testing.T) {
	ResetChangelogParser()
	defer ResetChangelogParser()

	cp := GetChangelogParserFn()
	if cp != nil {
		t.Error("expected nil when not registered")
	}
}

func TestReset(t *testing.T) {
	ResetChangelogParser()
	defer ResetChangelogParser()

	Register(&Config{Enabled: true})
	ResetChangelogParser()

	cp := GetChangelogParserFn()
	if cp != nil {
		t.Error("expected nil after reset")
	}
}
