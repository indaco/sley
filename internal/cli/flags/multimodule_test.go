package flags

import (
	"slices"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestMultiModuleFlags(t *testing.T) {
	flags := MultiModuleFlags()

	if len(flags) == 0 {
		t.Fatal("MultiModuleFlags() should return flags")
	}

	// Check that all expected flags are present
	expectedFlags := map[string]bool{
		"all":               false,
		"module":            false,
		"modules":           false,
		"pattern":           false,
		"yes":               false,
		"non-interactive":   false,
		"parallel":          false,
		"fail-fast":         false,
		"continue-on-error": false,
		"quiet":             false,
		"format":            false,
	}

	for _, flag := range flags {
		names := flag.Names()
		if len(names) > 0 {
			expectedFlags[names[0]] = true
		}
	}

	for name, found := range expectedFlags {
		if !found {
			t.Errorf("Expected flag %q not found in MultiModuleFlags()", name)
		}
	}
}

func TestMultiModuleFlags_FlagTypes(t *testing.T) {
	flags := MultiModuleFlags()

	// Verify some specific flag configurations
	flagMap := make(map[string]cli.Flag)
	for _, f := range flags {
		names := f.Names()
		if len(names) > 0 {
			flagMap[names[0]] = f
		}
	}

	// Check --all is a BoolFlag
	if _, ok := flagMap["all"].(*cli.BoolFlag); !ok {
		t.Error("Expected 'all' to be a BoolFlag")
	}

	// Check --module is a StringFlag
	if _, ok := flagMap["module"].(*cli.StringFlag); !ok {
		t.Error("Expected 'module' to be a StringFlag")
	}

	// Check --modules is a StringSliceFlag
	if _, ok := flagMap["modules"].(*cli.StringSliceFlag); !ok {
		t.Error("Expected 'modules' to be a StringSliceFlag")
	}

	// Check --format has default value "text"
	if f, ok := flagMap["format"].(*cli.StringFlag); ok {
		if f.Value != "text" {
			t.Errorf("Expected 'format' default value to be 'text', got %q", f.Value)
		}
	}

	// Check --fail-fast has default value true
	if f, ok := flagMap["fail-fast"].(*cli.BoolFlag); ok {
		if f.Value != true {
			t.Error("Expected 'fail-fast' default value to be true")
		}
	}
}

func TestMultiModuleFlags_Aliases(t *testing.T) {
	flags := MultiModuleFlags()

	aliasMap := make(map[string][]string)
	for _, f := range flags {
		names := f.Names()
		if len(names) > 1 {
			aliasMap[names[0]] = names[1:]
		}
	}

	// Check expected aliases
	expectedAliases := map[string]string{
		"all":   "a",
		"yes":   "y",
		"quiet": "q",
	}

	for flag, expectedAlias := range expectedAliases {
		aliases, exists := aliasMap[flag]
		if !exists {
			t.Errorf("Expected flag %q to have aliases", flag)
			continue
		}

		found := slices.Contains(aliases, expectedAlias)
		if !found {
			t.Errorf("Expected flag %q to have alias %q, got %v", flag, expectedAlias, aliases)
		}
	}
}
