package extension

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func createExtensionDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}
}

func writeConfigFile(t *testing.T, path string) {
	t.Helper()
	content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func checkCLIOutput(t *testing.T, output, extensionName string, deleted bool) {
	t.Helper()
	var expected string
	if deleted {
		expected = fmt.Sprintf("Extension %q and its directory uninstalled successfully.", extensionName)
	} else {
		expected = fmt.Sprintf("Extension %q uninstalled, but its directory is preserved.", extensionName)
	}
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

func checkExtensionDirDeleted(t *testing.T, dir string, expectDeleted bool) {
	t.Helper()
	_, err := os.Stat(dir)
	if expectDeleted {
		if !os.IsNotExist(err) {
			t.Errorf("expected extension directory to be deleted, but it still exists")
		}
	} else {
		if err != nil {
			t.Errorf("expected extension directory to exist, got: %v", err)
		}
	}
}

func checkExtensionRemovedFromConfig(t *testing.T, configPath, extensionName string) {
	t.Helper()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// The extension entry should no longer appear in the file.
	if strings.Contains(string(data), "name: "+extensionName) {
		t.Errorf("expected extension %q to be removed from config, but it is still present.\nConfig content:\n%s", extensionName, string(data))
	}
}
