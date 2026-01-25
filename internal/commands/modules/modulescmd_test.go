package modules

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestRun(t *testing.T) {
	cmd := Run()

	if cmd == nil {
		t.Fatal("Run() returned nil")
	}

	if cmd.Name != "modules" {
		t.Errorf("command name = %q, want %q", cmd.Name, "modules")
	}

	expectedAliases := []string{"mods"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("command has %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	} else {
		for i, alias := range cmd.Aliases {
			if alias != expectedAliases[i] {
				t.Errorf("alias[%d] = %q, want %q", i, alias, expectedAliases[i])
			}
		}
	}

	expectedUsage := "Manage and discover modules in workspace"
	if cmd.Usage != expectedUsage {
		t.Errorf("command usage = %q, want %q", cmd.Usage, expectedUsage)
	}
}

func TestRun_HasSubcommands(t *testing.T) {
	cmd := Run()

	if len(cmd.Commands) == 0 {
		t.Fatal("command has no subcommands")
	}

	expectedSubcommands := map[string]bool{
		"list":     true,
		"discover": true,
	}

	foundSubcommands := make(map[string]bool)
	for _, subcmd := range cmd.Commands {
		foundSubcommands[subcmd.Name] = true
	}

	for expectedName := range expectedSubcommands {
		if !foundSubcommands[expectedName] {
			t.Errorf("missing expected subcommand %q", expectedName)
		}
	}
}

func TestRun_ListSubcommand(t *testing.T) {
	cmd := Run()

	var listCmd *cli.Command
	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "list" {
			listCmd = subcmd
			break
		}
	}

	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	expectedAliases := []string{"ls"}
	if len(listCmd.Aliases) != len(expectedAliases) {
		t.Errorf("list command has %d aliases, want %d", len(listCmd.Aliases), len(expectedAliases))
	} else {
		for i, alias := range listCmd.Aliases {
			if alias != expectedAliases[i] {
				t.Errorf("list alias[%d] = %q, want %q", i, alias, expectedAliases[i])
			}
		}
	}

	expectedUsage := "List all discovered modules in workspace"
	if listCmd.Usage != expectedUsage {
		t.Errorf("list command usage = %q, want %q", listCmd.Usage, expectedUsage)
	}

	if listCmd.Action == nil {
		t.Error("list command has no action")
	}
}

func TestRun_DiscoverSubcommand(t *testing.T) {
	cmd := Run()

	var discoverCmd *cli.Command
	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "discover" {
			discoverCmd = subcmd
			break
		}
	}

	if discoverCmd == nil {
		t.Fatal("discover subcommand not found")
	}

	expectedUsage := "Test module discovery without running operations"
	if discoverCmd.Usage != expectedUsage {
		t.Errorf("discover command usage = %q, want %q", discoverCmd.Usage, expectedUsage)
	}

	if discoverCmd.Action == nil {
		t.Error("discover command has no action")
	}
}

func TestRun_ListFlags(t *testing.T) {
	cmd := Run()

	var listCmd *cli.Command
	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "list" {
			listCmd = subcmd
			break
		}
	}

	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	expectedFlags := map[string]bool{
		"verbose": true,
		"format":  true,
	}

	foundFlags := make(map[string]bool)
	for _, flag := range listCmd.Flags {
		for _, name := range flag.Names() {
			foundFlags[name] = true
		}
	}

	for expectedFlag := range expectedFlags {
		if !foundFlags[expectedFlag] {
			t.Errorf("list command missing expected flag %q", expectedFlag)
		}
	}
}

func TestRun_DiscoverFlags(t *testing.T) {
	cmd := Run()

	var discoverCmd *cli.Command
	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "discover" {
			discoverCmd = subcmd
			break
		}
	}

	if discoverCmd == nil {
		t.Fatal("discover subcommand not found")
	}

	expectedFlags := map[string]bool{
		"dry-run": true,
	}

	foundFlags := make(map[string]bool)
	for _, flag := range discoverCmd.Flags {
		for _, name := range flag.Names() {
			foundFlags[name] = true
		}
	}

	for expectedFlag := range expectedFlags {
		if !foundFlags[expectedFlag] {
			t.Errorf("discover command missing expected flag %q", expectedFlag)
		}
	}
}
