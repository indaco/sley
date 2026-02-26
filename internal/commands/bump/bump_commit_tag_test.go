package bump

import (
	"fmt"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/semver"
)

/* ------------------------------------------------------------------------- */
/* COMMIT AND TAG AFTER BUMP TESTS                                           */
/* ------------------------------------------------------------------------- */

// newTestTagManagerPlugin creates a TagManagerPlugin with mock ops for testing.
func newTestTagManagerPlugin(cfg *tagmanager.Config, gitOps *tagmanager.MockGitTagOperations, commitOps *tagmanager.MockGitCommitOperations) *tagmanager.TagManagerPlugin {
	return tagmanager.NewTagManagerWithOps(cfg, gitOps, commitOps)
}

func TestCommitAndTagAfterBump_AutoCreateDisabled(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: false,
		Prefix:     "v",
	}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", "")
	if err != nil {
		t.Errorf("expected nil error for disabled auto-create, got %v", err)
	}
}

func TestCommitAndTagAfterBump_Success(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	var stagedFiles []string
	var commitMsg string
	var createdTag string

	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn: func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error {
			createdTag = name
			return nil
		},
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{}, nil },
		StageFilesFn: func(files ...string) error {
			stagedFiles = append(stagedFiles, files...)
			return nil
		},
		CommitFn: func(message string) error {
			commitMsg = message
			return nil
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(stagedFiles) != 1 || stagedFiles[0] != ".version" {
		t.Errorf("expected staged files [.version], got %v", stagedFiles)
	}
	if commitMsg != "chore(release): v1.2.3" {
		t.Errorf("expected commit message %q, got %q", "chore(release): v1.2.3", commitMsg)
	}
	if createdTag != "v1.2.3" {
		t.Errorf("expected tag v1.2.3, got %q", createdTag)
	}
}

func TestCommitAndTagAfterBump_WithModifiedFiles(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	var stagedFiles []string

	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn:          func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error { return nil },
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) {
			return []string{"CHANGELOG.md", "package.json"}, nil
		},
		StageFilesFn: func(files ...string) error {
			stagedFiles = append(stagedFiles, files...)
			return nil
		},
		CommitFn: func(message string) error { return nil },
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	// Should stage bumpedPath + auto-detected modified files
	if len(stagedFiles) != 3 {
		t.Errorf("expected 3 staged files, got %d: %v", len(stagedFiles), stagedFiles)
	}
}

func TestCommitAndTagAfterBump_CommitFails(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
	}, &tagmanager.MockGitTagOperations{}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{}, nil },
		StageFilesFn:       func(files ...string) error { return nil },
		CommitFn:           func(message string) error { return fmt.Errorf("commit failed") },
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to commit release changes") {
		t.Errorf("expected commit error, got: %v", err)
	}
}

func TestCommitAndTagAfterBump_TagCreationFails(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn:          func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error { return fmt.Errorf("tag creation failed") },
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{}, nil },
		StageFilesFn:       func(files ...string) error { return nil },
		CommitFn:           func(message string) error { return nil },
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create tag") {
		t.Errorf("expected tag creation error, got: %v", err)
	}
}

func TestCommitAndTagAfterBump_WithoutBumpedPath(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn:          func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error { return nil },
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{".version"}, nil },
		StageFilesFn:       func(files ...string) error { return nil },
		CommitFn:           func(message string) error { return nil },
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	// Call with empty bumpedPath (via createTagAfterBump)
	err := createTagAfterBump(registry, version, "patch")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestCommitAndTagAfterBump_WithPush(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	var pushedTag string

	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:    true,
		AutoCreate: true,
		Prefix:     "v",
		Annotate:   true,
		Push:       true,
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn:          func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error { return nil },
		PushTagFn: func(name string) error {
			pushedTag = name
			return nil
		},
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{}, nil },
		StageFilesFn:       func(files ...string) error { return nil },
		CommitFn:           func(message string) error { return nil },
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if pushedTag != "v1.2.3" {
		t.Errorf("expected pushed tag v1.2.3, got %q", pushedTag)
	}
}

func TestCommitAndTagAfterBump_CustomCommitMessageTemplate(t *testing.T) {
	version := semver.SemVersion{Major: 1, Minor: 2, Patch: 3}
	var commitMsg string

	plugin := newTestTagManagerPlugin(&tagmanager.Config{
		Enabled:               true,
		AutoCreate:            true,
		Prefix:                "v",
		Annotate:              true,
		CommitMessageTemplate: "release: bump to {version}",
	}, &tagmanager.MockGitTagOperations{
		TagExistsFn:          func(name string) (bool, error) { return false, nil },
		CreateAnnotatedTagFn: func(name, message string) error { return nil },
	}, &tagmanager.MockGitCommitOperations{
		GetModifiedFilesFn: func() ([]string, error) { return []string{}, nil },
		StageFilesFn:       func(files ...string) error { return nil },
		CommitFn: func(message string) error {
			commitMsg = message
			return nil
		},
	})

	registry := plugins.NewPluginRegistry()
	if err := registry.RegisterTagManager(plugin); err != nil {
		t.Fatalf("failed to register tag manager: %v", err)
	}

	err := commitAndTagAfterBump(registry, version, "patch", ".version")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if commitMsg != "release: bump to 1.2.3" {
		t.Errorf("expected commit message %q, got %q", "release: bump to 1.2.3", commitMsg)
	}
}
