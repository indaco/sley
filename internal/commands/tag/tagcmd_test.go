package tag

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

// mockGitTagOps is a mock implementation of core.GitTagOperations for testing.
type mockGitTagOps struct {
	tagExists          func(name string) (bool, error)
	listTags           func(pattern string) ([]string, error)
	pushTag            func(name string) error
	deleteTag          func(name string) error
	deleteRemoteTag    func(name string) error
	createAnnotatedTag func(name, message string) error
	createLightweight  func(name string) error
	createSignedTag    func(name, message, keyID string) error
	getLatestTag       func() (string, error)
}

func (m *mockGitTagOps) TagExists(name string) (bool, error) {
	if m.tagExists != nil {
		return m.tagExists(name)
	}
	return false, nil
}

func (m *mockGitTagOps) ListTags(pattern string) ([]string, error) {
	if m.listTags != nil {
		return m.listTags(pattern)
	}
	return []string{}, nil
}

func (m *mockGitTagOps) PushTag(name string) error {
	if m.pushTag != nil {
		return m.pushTag(name)
	}
	return nil
}

func (m *mockGitTagOps) DeleteTag(name string) error {
	if m.deleteTag != nil {
		return m.deleteTag(name)
	}
	return nil
}

func (m *mockGitTagOps) DeleteRemoteTag(name string) error {
	if m.deleteRemoteTag != nil {
		return m.deleteRemoteTag(name)
	}
	return nil
}

func (m *mockGitTagOps) CreateAnnotatedTag(name, message string) error {
	if m.createAnnotatedTag != nil {
		return m.createAnnotatedTag(name, message)
	}
	return nil
}

func (m *mockGitTagOps) CreateLightweightTag(name string) error {
	if m.createLightweight != nil {
		return m.createLightweight(name)
	}
	return nil
}

func (m *mockGitTagOps) CreateSignedTag(name, message, keyID string) error {
	if m.createSignedTag != nil {
		return m.createSignedTag(name, message, keyID)
	}
	return nil
}

func (m *mockGitTagOps) GetLatestTag() (string, error) {
	if m.getLatestTag != nil {
		return m.getLatestTag()
	}
	return "", nil
}

func TestGetVersionPath(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		pathFlag string
		want     string
	}{
		{
			name:     "default path",
			cfg:      nil,
			pathFlag: "",
			want:     ".version",
		},
		{
			name:     "config path",
			cfg:      &config.Config{Path: "custom/.version"},
			pathFlag: "",
			want:     "custom/.version",
		},
		{
			name:     "empty config path uses default",
			cfg:      &config.Config{Path: ""},
			pathFlag: "",
			want:     ".version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path"},
				},
			}

			got := getVersionPath(cmd, tt.cfg)
			if got != tt.want {
				t.Errorf("getVersionPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTagPrefix(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: "v",
		},
		{
			name: "nil plugins",
			cfg:  &config.Config{},
			want: "v",
		},
		{
			name: "nil tag manager",
			cfg:  &config.Config{Plugins: &config.PluginConfig{}},
			want: "v",
		},
		{
			name: "default prefix",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{},
				},
			},
			want: "v",
		},
		{
			name: "custom prefix",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{
						Prefix: "release-",
					},
				},
			},
			want: "release-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTagPrefix(tt.cfg)
			if got != tt.want {
				t.Errorf("getTagPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTagManagerConfig(t *testing.T) {
	annotateTrue := true
	annotateFalse := false
	autoCreateTrue := true
	autoCreateFalse := false

	tests := []struct {
		name       string
		cfg        *config.Config
		wantPrefix string
		wantSign   bool
	}{
		{
			name:       "nil config",
			cfg:        nil,
			wantPrefix: "v",
			wantSign:   false,
		},
		{
			name:       "nil plugins",
			cfg:        &config.Config{},
			wantPrefix: "v",
			wantSign:   false,
		},
		{
			name: "custom config",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{
						Enabled:    true,
						Prefix:     "ver-",
						Sign:       true,
						SigningKey: "ABC123",
						Annotate:   &annotateTrue,
						AutoCreate: &autoCreateTrue,
					},
				},
			},
			wantPrefix: "ver-",
			wantSign:   true,
		},
		{
			name: "false booleans",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{
						Enabled:    false,
						Annotate:   &annotateFalse,
						AutoCreate: &autoCreateFalse,
					},
				},
			},
			wantPrefix: "v",
			wantSign:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTagManagerConfig(tt.cfg)
			if got.Prefix != tt.wantPrefix {
				t.Errorf("buildTagManagerConfig().Prefix = %v, want %v", got.Prefix, tt.wantPrefix)
			}
			if got.Sign != tt.wantSign {
				t.Errorf("buildTagManagerConfig().Sign = %v, want %v", got.Sign, tt.wantSign)
			}
		})
	}
}

func TestSortTagsBySemver(t *testing.T) {
	tests := []struct {
		name   string
		tags   []string
		prefix string
		want   []string
	}{
		{
			name:   "basic sort",
			tags:   []string{"v1.0.0", "v2.0.0", "v1.5.0"},
			prefix: "v",
			want:   []string{"v2.0.0", "v1.5.0", "v1.0.0"},
		},
		{
			name:   "with patch versions",
			tags:   []string{"v1.0.1", "v1.0.0", "v1.0.10", "v1.0.2"},
			prefix: "v",
			want:   []string{"v1.0.10", "v1.0.2", "v1.0.1", "v1.0.0"},
		},
		{
			name:   "pre-releases after stable",
			tags:   []string{"v1.0.0-alpha.1", "v1.0.0", "v1.0.0-rc.1"},
			prefix: "v",
			want:   []string{"v1.0.0", "v1.0.0-rc.1", "v1.0.0-alpha.1"},
		},
		{
			name:   "custom prefix",
			tags:   []string{"release-1.0.0", "release-2.0.0"},
			prefix: "release-",
			want:   []string{"release-2.0.0", "release-1.0.0"},
		},
		{
			name:   "empty tags",
			tags:   []string{},
			prefix: "v",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortTagsBySemver(tt.tags, tt.prefix)
			if len(tt.tags) != len(tt.want) {
				t.Fatalf("sortTagsBySemver() len = %v, want %v", len(tt.tags), len(tt.want))
			}
			for i := range tt.tags {
				if tt.tags[i] != tt.want[i] {
					t.Errorf("sortTagsBySemver()[%d] = %v, want %v", i, tt.tags[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseVersionFromTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		prefix  string
		want    semver.SemVersion
		isValid bool
	}{
		{
			name:    "valid tag",
			tag:     "v1.2.3",
			prefix:  "v",
			want:    semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			isValid: true,
		},
		{
			name:    "with pre-release",
			tag:     "v1.0.0-beta.1",
			prefix:  "v",
			want:    semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "beta.1"},
			isValid: true,
		},
		{
			name:    "invalid tag",
			tag:     "invalid",
			prefix:  "v",
			want:    semver.SemVersion{},
			isValid: false,
		},
		{
			name:    "custom prefix",
			tag:     "release-2.0.0",
			prefix:  "release-",
			want:    semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVersionFromTag(tt.tag, tt.prefix)
			if tt.isValid {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch || got.PreRelease != tt.want.PreRelease {
					t.Errorf("parseVersionFromTag() = %v, want %v", got, tt.want)
				}
			} else {
				if got.Major != 0 || got.Minor != 0 || got.Patch != 0 {
					t.Errorf("parseVersionFromTag() = %v, want zero version", got)
				}
			}
		})
	}
}

func TestRunCreateCmd_MissingVersionFile(t *testing.T) {
	mockOps := &mockGitTagOps{}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{
		Path: "/nonexistent/path/.version",
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path"},
			&cli.BoolFlag{Name: "push"},
			&cli.StringFlag{Name: "message"},
		},
	}

	err := tc.runCreateCmd(context.Background(), cmd, cfg)
	if err == nil {
		t.Error("runCreateCmd() expected error for missing version file")
	}
}

func TestRunCreateCmd_TagAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{
		Path: versionFile,
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Value: versionFile},
			&cli.BoolFlag{Name: "push"},
			&cli.StringFlag{Name: "message"},
		},
	}

	err := tc.runCreateCmd(context.Background(), cmd, cfg)
	if err == nil {
		t.Error("runCreateCmd() expected error for existing tag")
	}
	if err != nil && err.Error() != "tag v1.0.0 already exists" {
		t.Errorf("runCreateCmd() unexpected error: %v", err)
	}
}

func TestRunCreateCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	var createdTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
		createAnnotatedTag: func(name, message string) error {
			createdTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{
		Path: versionFile,
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Value: versionFile},
			&cli.BoolFlag{Name: "push"},
			&cli.StringFlag{Name: "message"},
		},
	}

	err := tc.runCreateCmd(context.Background(), cmd, cfg)
	if err != nil {
		t.Errorf("runCreateCmd() unexpected error: %v", err)
	}
	if createdTag != "v1.0.0" {
		t.Errorf("runCreateCmd() created tag = %v, want v1.0.0", createdTag)
	}
}

func TestRunListCmd(t *testing.T) {
	mockOps := &mockGitTagOps{
		listTags: func(pattern string) ([]string, error) {
			return []string{"v1.0.0", "v2.0.0", "v1.5.0"}, nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Prefix: "v",
			},
		},
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "limit"},
		},
	}

	err := tc.runListCmd(context.Background(), cmd, cfg)
	if err != nil {
		t.Errorf("runListCmd() unexpected error: %v", err)
	}
}

func TestRunPushCmd_TagExists(t *testing.T) {
	var pushedTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		pushTag: func(name string) error {
			pushedTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push", "v1.0.0"})
	if err != nil {
		t.Errorf("runPushCmd() unexpected error: %v", err)
	}
	if pushedTag != "v1.0.0" {
		t.Errorf("runPushCmd() pushed tag = %v, want v1.0.0", pushedTag)
	}
}

func TestRunPushCmd_TagNotExists(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push", "v1.0.0"})
	if err == nil {
		t.Error("runPushCmd() expected error for non-existing tag")
	}
}

func TestRunDeleteCmd_TagExists(t *testing.T) {
	var deletedTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		deleteTag: func(name string) error {
			deletedTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0"})
	if err != nil {
		t.Errorf("runDeleteCmd() unexpected error: %v", err)
	}
	if deletedTag != "v1.0.0" {
		t.Errorf("runDeleteCmd() deleted tag = %v, want v1.0.0", deletedTag)
	}
}

func TestRunDeleteCmd_TagNotExists(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0"})
	if err == nil {
		t.Error("runDeleteCmd() expected error for non-existing tag")
	}
}

func TestRunDeleteCmd_WithRemote(t *testing.T) {
	var deletedLocal, deletedRemote string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		deleteTag: func(name string) error {
			deletedLocal = name
			return nil
		},
		deleteRemoteTag: func(name string) error {
			deletedRemote = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0", "--remote"})
	if err != nil {
		t.Errorf("runDeleteCmd() unexpected error: %v", err)
	}
	if deletedLocal != "v1.0.0" {
		t.Errorf("runDeleteCmd() deleted local tag = %v, want v1.0.0", deletedLocal)
	}
	if deletedRemote != "v1.0.0" {
		t.Errorf("runDeleteCmd() deleted remote tag = %v, want v1.0.0", deletedRemote)
	}
}

func TestIsTagManagerEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: false,
		},
		{
			name: "nil plugins",
			cfg:  &config.Config{},
			want: false,
		},
		{
			name: "nil tag manager",
			cfg:  &config.Config{Plugins: &config.PluginConfig{}},
			want: false,
		},
		{
			name: "disabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{Enabled: false},
				},
			},
			want: false,
		},
		{
			name: "enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{Enabled: true},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTagManagerEnabled(tt.cfg)
			if got != tt.want {
				t.Errorf("isTagManagerEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireTagManagerEnabled(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name:    "nil plugins",
			cfg:     &config.Config{},
			wantErr: true,
		},
		{
			name: "disabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "enabled",
			cfg: &config.Config{
				Plugins: &config.PluginConfig{
					TagManager: &config.TagManagerConfig{Enabled: true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := requireTagManagerEnabled(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("requireTagManagerEnabled() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "tag-manager plugin is not enabled") {
					t.Errorf("requireTagManagerEnabled() error message = %q, want it to contain 'tag-manager plugin is not enabled'", err.Error())
				}
			}
		})
	}
}

func TestTagCommand_BeforeHookBlocksSubcommands(t *testing.T) {
	// Verify that running any tag subcommand via Run() fails when plugin is disabled.
	cfg := &config.Config{
		Path: ".version",
		// No tag-manager plugin configured.
	}

	app := &cli.Command{
		Name:      "sley",
		Writer:    io.Discard,
		ErrWriter: io.Discard,
		Commands:  []*cli.Command{Run(cfg)},
	}

	subcommands := [][]string{
		{"sley", "tag", "list"},
		{"sley", "tag", "create"},
		{"sley", "tag", "push", "v1.0.0"},
		{"sley", "tag", "delete", "v1.0.0"},
	}

	for _, args := range subcommands {
		t.Run(strings.Join(args[1:], " "), func(t *testing.T) {
			err := app.Run(context.Background(), args)
			if err == nil {
				t.Errorf("expected error for %v when tag-manager is not enabled", args)
			}
			if err != nil && !strings.Contains(err.Error(), "tag-manager plugin is not enabled") {
				t.Errorf("unexpected error for %v: %v", args, err)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	cfg := &config.Config{
		Path: ".version",
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled: true,
				Prefix:  "v",
			},
		},
	}

	cmd := Run(cfg)

	if cmd.Name != "tag" {
		t.Errorf("Run().Name = %v, want %v", cmd.Name, "tag")
	}

	if len(cmd.Commands) != 4 {
		t.Errorf("Run().Commands len = %v, want 4", len(cmd.Commands))
	}

	expectedSubcommands := map[string]bool{
		"create": false,
		"list":   false,
		"push":   false,
		"delete": false,
	}

	for _, subcmd := range cmd.Commands {
		if _, ok := expectedSubcommands[subcmd.Name]; ok {
			expectedSubcommands[subcmd.Name] = true
		}
	}

	for name, found := range expectedSubcommands {
		if !found {
			t.Errorf("Run() missing subcommand %v", name)
		}
	}
}

func TestCreateTag_Signed(t *testing.T) {
	var signedTag, signedMessage, signedKeyID string
	mockOps := &mockGitTagOps{
		createSignedTag: func(name, message, keyID string) error {
			signedTag = name
			signedMessage = message
			signedKeyID = keyID
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &tagmanager.Config{
		Sign:       true,
		SigningKey: "ABC123",
	}

	err := tc.createTag("v1.0.0", "Release 1.0.0", cfg)
	if err != nil {
		t.Errorf("createTag() unexpected error: %v", err)
	}
	if signedTag != "v1.0.0" {
		t.Errorf("createTag() signed tag = %v, want v1.0.0", signedTag)
	}
	if signedMessage != "Release 1.0.0" {
		t.Errorf("createTag() signed message = %v, want Release 1.0.0", signedMessage)
	}
	if signedKeyID != "ABC123" {
		t.Errorf("createTag() signed keyID = %v, want ABC123", signedKeyID)
	}
}

func TestCreateTag_Lightweight(t *testing.T) {
	var lightweightTag string
	mockOps := &mockGitTagOps{
		createLightweight: func(name string) error {
			lightweightTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &tagmanager.Config{
		Sign:     false,
		Annotate: false,
	}

	err := tc.createTag("v1.0.0", "ignored message", cfg)
	if err != nil {
		t.Errorf("createTag() unexpected error: %v", err)
	}
	if lightweightTag != "v1.0.0" {
		t.Errorf("createTag() lightweight tag = %v, want v1.0.0", lightweightTag)
	}
}

func TestCreateTag_SignedError(t *testing.T) {
	mockOps := &mockGitTagOps{
		createSignedTag: func(name, message, keyID string) error {
			return fmt.Errorf("gpg signing failed")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &tagmanager.Config{Sign: true}

	err := tc.createTag("v1.0.0", "Release", cfg)
	if err == nil {
		t.Error("createTag() expected error for signed tag failure")
	}
}

func TestCreateTag_AnnotatedError(t *testing.T) {
	mockOps := &mockGitTagOps{
		createAnnotatedTag: func(name, message string) error {
			return fmt.Errorf("annotated tag failed")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &tagmanager.Config{Sign: false, Annotate: true}

	err := tc.createTag("v1.0.0", "Release", cfg)
	if err == nil {
		t.Error("createTag() expected error for annotated tag failure")
	}
}

func TestCreateTag_LightweightError(t *testing.T) {
	mockOps := &mockGitTagOps{
		createLightweight: func(name string) error {
			return fmt.Errorf("lightweight tag failed")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &tagmanager.Config{Sign: false, Annotate: false}

	err := tc.createTag("v1.0.0", "ignored", cfg)
	if err == nil {
		t.Error("createTag() expected error for lightweight tag failure")
	}
}

func TestRunListCmd_Empty(t *testing.T) {
	mockOps := &mockGitTagOps{
		listTags: func(pattern string) ([]string, error) {
			return []string{}, nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "limit"},
		},
	}

	err := tc.runListCmd(context.Background(), cmd, cfg)
	if err != nil {
		t.Errorf("runListCmd() unexpected error: %v", err)
	}
}

func TestRunListCmd_Error(t *testing.T) {
	mockOps := &mockGitTagOps{
		listTags: func(pattern string) ([]string, error) {
			return nil, fmt.Errorf("git error")
		},
	}
	tc := NewTagCommand(mockOps)

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "limit"},
		},
	}

	err := tc.runListCmd(context.Background(), cmd, nil)
	if err == nil {
		t.Error("runListCmd() expected error")
	}
}

func TestRunListCmd_WithLimit(t *testing.T) {
	mockOps := &mockGitTagOps{
		listTags: func(pattern string) ([]string, error) {
			return []string{"v3.0.0", "v2.0.0", "v1.0.0"}, nil
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "list",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "limit", Aliases: []string{"n"}},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runListCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "list", "--limit", "2"})
	if err != nil {
		t.Errorf("runListCmd() unexpected error: %v", err)
	}
}

func TestRunPushCmd_NoArg(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("2.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	var pushedTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		pushTag: func(name string) error {
			pushedTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: versionFile}

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, cfg)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push"})
	if err != nil {
		t.Errorf("runPushCmd() unexpected error: %v", err)
	}
	if pushedTag != "v2.0.0" {
		t.Errorf("runPushCmd() pushed tag = %v, want v2.0.0", pushedTag)
	}
}

func TestRunPushCmd_VersionReadError(t *testing.T) {
	mockOps := &mockGitTagOps{}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: "/nonexistent/.version"}

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, cfg)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push"})
	if err == nil {
		t.Error("runPushCmd() expected error for missing version file")
	}
}

func TestRunPushCmd_TagExistsError(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, fmt.Errorf("git error")
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push", "v1.0.0"})
	if err == nil {
		t.Error("runPushCmd() expected error for tagExists failure")
	}
}

func TestRunPushCmd_PushError(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		pushTag: func(name string) error {
			return fmt.Errorf("push failed")
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "push",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runPushCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "push", "v1.0.0"})
	if err == nil {
		t.Error("runPushCmd() expected error for push failure")
	}
}

func TestRunCreateCmd_WithPush(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	var createdTag, pushedTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
		createAnnotatedTag: func(name, message string) error {
			createdTag = name
			return nil
		},
		pushTag: func(name string) error {
			pushedTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: versionFile}

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "create",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path"},
					&cli.BoolFlag{Name: "push"},
					&cli.StringFlag{Name: "message"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runCreateCmd(ctx, cmd, cfg)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "create", "--push"})
	if err != nil {
		t.Errorf("runCreateCmd() unexpected error: %v", err)
	}
	if createdTag != "v1.0.0" {
		t.Errorf("runCreateCmd() created tag = %v, want v1.0.0", createdTag)
	}
	if pushedTag != "v1.0.0" {
		t.Errorf("runCreateCmd() pushed tag = %v, want v1.0.0", pushedTag)
	}
}

func TestRunCreateCmd_PushError(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
		createAnnotatedTag: func(name, message string) error {
			return nil
		},
		pushTag: func(name string) error {
			return fmt.Errorf("push failed")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: versionFile}

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "create",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path"},
					&cli.BoolFlag{Name: "push"},
					&cli.StringFlag{Name: "message"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runCreateCmd(ctx, cmd, cfg)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "create", "--push"})
	if err == nil {
		t.Error("runCreateCmd() expected error for push failure")
	}
}

func TestRunCreateCmd_TagExistsError(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, fmt.Errorf("git error")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: versionFile}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path"},
			&cli.BoolFlag{Name: "push"},
			&cli.StringFlag{Name: "message"},
		},
	}

	err := tc.runCreateCmd(context.Background(), cmd, cfg)
	if err == nil {
		t.Error("runCreateCmd() expected error for tagExists failure")
	}
}

func TestRunCreateCmd_CreateTagError(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
		createAnnotatedTag: func(name, message string) error {
			return fmt.Errorf("create tag failed")
		},
	}
	tc := NewTagCommand(mockOps)

	cfg := &config.Config{Path: versionFile}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path"},
			&cli.BoolFlag{Name: "push"},
			&cli.StringFlag{Name: "message"},
		},
	}

	err := tc.runCreateCmd(context.Background(), cmd, cfg)
	if err == nil {
		t.Error("runCreateCmd() expected error for createTag failure")
	}
}

func TestRunDeleteCmd_MissingArg(t *testing.T) {
	mockOps := &mockGitTagOps{}
	tc := NewTagCommand(mockOps)

	var capturedErr error
	app := &cli.Command{
		Name:      "test",
		Writer:    io.Discard,
		ErrWriter: io.Discard,
		ExitErrHandler: func(_ context.Context, _ *cli.Command, err error) {
			capturedErr = err
		},
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	_ = app.Run(context.Background(), []string{"test", "delete"})
	// cli.Exit returns an ExitCoder error captured by ExitErrHandler
	if capturedErr == nil {
		t.Error("runDeleteCmd() expected error for missing argument")
	}
}

func TestRunDeleteCmd_TagExistsError(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, fmt.Errorf("git error")
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0"})
	if err == nil {
		t.Error("runDeleteCmd() expected error for tagExists failure")
	}
}

func TestRunDeleteCmd_DeleteLocalError(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		deleteTag: func(name string) error {
			return fmt.Errorf("delete failed")
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0"})
	if err == nil {
		t.Error("runDeleteCmd() expected error for delete local failure")
	}
}

func TestRunDeleteCmd_DeleteRemoteError(t *testing.T) {
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		deleteTag: func(name string) error {
			return nil
		},
		deleteRemoteTag: func(name string) error {
			return fmt.Errorf("remote delete failed")
		},
	}
	tc := NewTagCommand(mockOps)

	app := &cli.Command{
		Name: "test",
		Commands: []*cli.Command{
			{
				Name: "delete",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "remote"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return tc.runDeleteCmd(ctx, cmd, nil)
				},
			},
		},
	}

	err := app.Run(context.Background(), []string{"test", "delete", "v1.0.0", "--remote"})
	if err == nil {
		t.Error("runDeleteCmd() expected error for delete remote failure")
	}
}

func TestRunCommand_HasMultiModuleFlags(t *testing.T) {
	cfg := &config.Config{Path: ".version"}
	cmd := Run(cfg)

	// The tag command should now include multi-module flags
	if len(cmd.Flags) == 0 {
		t.Fatal("Run() expected tag command to have flags from MultiModuleFlags()")
	}

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

	for _, f := range cmd.Flags {
		for _, name := range f.Names() {
			if _, ok := expectedFlags[name]; ok {
				expectedFlags[name] = true
			}
		}
	}

	for name, found := range expectedFlags {
		if !found {
			t.Errorf("Run() missing expected multi-module flag %q", name)
		}
	}
}

func TestResolveVersionPath_SingleModule(t *testing.T) {
	tmpDir := t.TempDir()
	versionFile := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// When cfg.Path is explicitly set (not ".version"), resolveVersionPath
	// should return SingleModuleMode via getSingleModuleFromFlags.
	cfg := &config.Config{Path: versionFile}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path"},
		},
	}

	path, err := resolveVersionPath(context.Background(), cmd, cfg)
	if err != nil {
		t.Fatalf("resolveVersionPath() unexpected error: %v", err)
	}
	if path != versionFile {
		t.Errorf("resolveVersionPath() = %v, want %v", path, versionFile)
	}
}

func TestResolveVersionPath_MultiModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "2.3.4")
	testutils.WriteTempVersionFile(t, moduleB, "2.3.4")

	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	// Build a CLI app that has the multi-module flags (from tag parent command)
	// plus the global --path flag, and an action that calls resolveVersionPath.
	var resolvedPath string
	var resolveErr error

	app := &cli.Command{
		Name: "sley",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: cfg.Path,
			},
			&cli.BoolFlag{Name: "strict"},
		},
		Commands: []*cli.Command{
			{
				Name:  "tag",
				Flags: Run(cfg).Flags, // inherit the multi-module flags
				Commands: []*cli.Command{
					{
						Name: "resolve",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							resolvedPath, resolveErr = resolveVersionPath(ctx, cmd, cfg)
							return resolveErr
						},
					},
				},
			},
		},
	}

	// Change to tmpDir so workspace detection discovers modules
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	err = app.Run(context.Background(), []string{"sley", "tag", "resolve", "--all"})
	if err != nil {
		t.Fatalf("resolveVersionPath() in multi-module mode returned error: %v", err)
	}

	// The resolved path should be a module's .version file, not the root
	if resolvedPath == ".version" {
		t.Error("resolveVersionPath() returned default '.version' instead of a module path")
	}

	// It should point to an actual file
	if _, err := os.Stat(resolvedPath); err != nil {
		t.Errorf("resolveVersionPath() returned path that doesn't exist: %v", resolvedPath)
	}

	// The version should be readable
	v, err := semver.ReadVersion(resolvedPath)
	if err != nil {
		t.Fatalf("failed to read version from resolved path %s: %v", resolvedPath, err)
	}
	if v.String() != "2.3.4" {
		t.Errorf("version from resolved path = %v, want 2.3.4", v.String())
	}
}

func TestCLI_TagCreate_MultiModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "3.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "3.0.0")

	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled: true,
			},
		},
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	var createdTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return false, nil
		},
		createAnnotatedTag: func(name, message string) error {
			createdTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	appCli := &cli.Command{
		Name: "sley",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: cfg.Path,
			},
			&cli.BoolFlag{Name: "strict"},
		},
		Commands: []*cli.Command{
			func() *cli.Command {
				cmd := Run(cfg)
				// Replace the create subcommand with one that uses our mock
				for i, sub := range cmd.Commands {
					if sub.Name == "create" {
						cmd.Commands[i] = tc.createCmd(cfg)
					}
				}
				return cmd
			}(),
		},
	}

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "tag", "create", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if createdTag != "v3.0.0" {
		t.Errorf("tag create in multi-module mode created tag = %v, want v3.0.0", createdTag)
	}

	// Output should mention which module the version was sourced from
	if !strings.Contains(output, "Using version from module") {
		t.Errorf("expected output to mention source module, got: %q", output)
	}
}

func TestCLI_TagPush_MultiModule_NoArg(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "4.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "4.0.0")

	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled: true,
			},
		},
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	var pushedTag string
	mockOps := &mockGitTagOps{
		tagExists: func(name string) (bool, error) {
			return true, nil
		},
		pushTag: func(name string) error {
			pushedTag = name
			return nil
		},
	}
	tc := NewTagCommand(mockOps)

	appCli := &cli.Command{
		Name: "sley",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: cfg.Path,
			},
			&cli.BoolFlag{Name: "strict"},
		},
		Commands: []*cli.Command{
			func() *cli.Command {
				cmd := Run(cfg)
				// Replace the push subcommand with one that uses our mock
				for i, sub := range cmd.Commands {
					if sub.Name == "push" {
						cmd.Commands[i] = tc.pushCmd(cfg)
					}
				}
				return cmd
			}(),
		},
	}

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "tag", "push", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if pushedTag != "v4.0.0" {
		t.Errorf("tag push in multi-module mode pushed tag = %v, want v4.0.0", pushedTag)
	}

	// Output should mention which module the version was sourced from
	if !strings.Contains(output, "Using version from module") {
		t.Errorf("expected output to mention source module, got: %q", output)
	}
}
