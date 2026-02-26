package tagmanager

import (
	"errors"
	"testing"

	"github.com/indaco/sley/internal/semver"
)

func TestTagManagerPlugin_Name(t *testing.T) {
	tm := NewTagManager(nil)
	if got := tm.Name(); got != "tag-manager" {
		t.Errorf("Name() = %q, want %q", got, "tag-manager")
	}
}

func TestTagManagerPlugin_Description(t *testing.T) {
	tm := NewTagManager(nil)
	if got := tm.Description(); got == "" {
		t.Error("Description() should not be empty")
	}
}

func TestTagManagerPlugin_Version(t *testing.T) {
	tm := NewTagManager(nil)
	if got := tm.Version(); got != "v0.1.0" {
		t.Errorf("Version() = %q, want %q", got, "v0.1.0")
	}
}

func TestTagManagerPlugin_FormatTagName(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		version semver.SemVersion
		want    string
	}{
		{
			name:    "default prefix v",
			prefix:  "v",
			version: semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			want:    "v1.2.3",
		},
		{
			name:    "custom prefix release-",
			prefix:  "release-",
			version: semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			want:    "release-2.0.0",
		},
		{
			name:    "empty prefix",
			prefix:  "",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			want:    "1.0.0",
		},
		{
			name:    "with prerelease",
			prefix:  "v",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			want:    "v1.0.0-alpha.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Prefix: tt.prefix}
			tm := NewTagManager(cfg)
			if got := tm.FormatTagName(tt.version); got != tt.want {
				t.Errorf("FormatTagName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagManagerPlugin_TagExists(t *testing.T) {
	tests := []struct {
		name    string
		version semver.SemVersion
		mockFn  func(string) (bool, error)
		want    bool
		wantErr bool
	}{
		{
			name:    "tag exists",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			mockFn: func(name string) (bool, error) {
				return true, nil
			},
			want:    true,
			wantErr: false,
		},
		{
			name:    "tag does not exist",
			version: semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			mockFn: func(name string) (bool, error) {
				return false, nil
			},
			want:    false,
			wantErr: false,
		},
		{
			name:    "error checking tag",
			version: semver.SemVersion{Major: 3, Minor: 0, Patch: 0},
			mockFn: func(name string) (bool, error) {
				return false, errors.New("git error")
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitTagOperations{
				TagExistsFn: tt.mockFn,
			}
			tm := NewTagManagerWithOps(nil, mockOps)

			got, err := tm.TagExists(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("TagExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TagExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagManagerPlugin_ValidateTagAvailable(t *testing.T) {
	tests := []struct {
		name    string
		version semver.SemVersion
		exists  bool
		wantErr bool
	}{
		{
			name:    "tag available",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			exists:  false,
			wantErr: false,
		},
		{
			name:    "tag already exists",
			version: semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			exists:  true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitTagOperations{
				TagExistsFn: func(name string) (bool, error) {
					return tt.exists, nil
				},
			}
			tm := NewTagManagerWithOps(nil, mockOps)

			err := tm.ValidateTagAvailable(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagManagerPlugin_CreateTag(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *Config
		version        semver.SemVersion
		message        string
		tagExists      bool
		createErr      error
		pushErr        error
		wantErr        bool
		wantAnnotated  bool
		wantPushCalled bool
	}{
		{
			name:          "create annotated tag",
			cfg:           &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: true, Push: false},
			version:       semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			message:       "Release 1.0.0",
			tagExists:     false,
			wantErr:       false,
			wantAnnotated: true,
		},
		{
			name:          "create lightweight tag",
			cfg:           &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: false, Push: false},
			version:       semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			message:       "",
			tagExists:     false,
			wantErr:       false,
			wantAnnotated: false,
		},
		{
			name:           "create and push tag",
			cfg:            &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: true, Push: true},
			version:        semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			message:        "Release 1.0.0",
			tagExists:      false,
			wantErr:        false,
			wantAnnotated:  true,
			wantPushCalled: true,
		},
		{
			name:      "tag already exists",
			cfg:       &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: true},
			version:   semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			tagExists: true,
			wantErr:   true,
		},
		{
			name:      "create fails",
			cfg:       &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: true},
			version:   semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			tagExists: false,
			createErr: errors.New("git error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotatedCalled := false
			lightweightCalled := false
			pushCalled := false

			mockOps := &MockGitTagOperations{
				TagExistsFn: func(name string) (bool, error) {
					return tt.tagExists, nil
				},
				CreateAnnotatedTagFn: func(name, msg string) error {
					annotatedCalled = true
					return tt.createErr
				},
				CreateLightweightTagFn: func(name string) error {
					lightweightCalled = true
					return tt.createErr
				},
				PushTagFn: func(name string) error {
					pushCalled = true
					return tt.pushErr
				},
			}

			tm := NewTagManagerWithOps(tt.cfg, mockOps)
			err := tm.CreateTag(tt.version, tt.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantAnnotated && !annotatedCalled {
					t.Error("CreateTag() should have called createAnnotatedTag")
				}
				if !tt.wantAnnotated && !lightweightCalled {
					t.Error("CreateTag() should have called createLightweightTag")
				}
				if tt.wantPushCalled && !pushCalled {
					t.Error("CreateTag() should have called pushTag")
				}
			}
		})
	}
}

func TestTagManagerPlugin_GetLatestTag(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		mockTag string
		mockErr error
		want    semver.SemVersion
		wantErr bool
	}{
		{
			name:    "parse tag with v prefix",
			prefix:  "v",
			mockTag: "v1.2.3",
			want:    semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "parse tag without prefix",
			prefix:  "",
			mockTag: "2.0.0",
			want:    semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name:    "no tags found",
			prefix:  "v",
			mockErr: errors.New("no tags found"),
			wantErr: true,
		},
		{
			name:    "invalid version format",
			prefix:  "v",
			mockTag: "vinvalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitTagOperations{
				GetLatestTagFn: func() (string, error) {
					return tt.mockTag, tt.mockErr
				},
			}

			cfg := &Config{Prefix: tt.prefix}
			tm := NewTagManagerWithOps(cfg, mockOps)

			got, err := tm.GetLatestTag()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagManagerPlugin_IsAutoCreateEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{
			name: "enabled and auto-create",
			cfg:  &Config{Enabled: true, AutoCreate: true},
			want: true,
		},
		{
			name: "enabled but no auto-create",
			cfg:  &Config{Enabled: true, AutoCreate: false},
			want: false,
		},
		{
			name: "disabled",
			cfg:  &Config{Enabled: false, AutoCreate: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTagManager(tt.cfg)
			if got := tm.IsAutoCreateEnabled(); got != tt.want {
				t.Errorf("IsAutoCreateEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled != false {
		t.Errorf("DefaultConfig().Enabled = %v, want false", cfg.Enabled)
	}
	if cfg.AutoCreate != false {
		t.Errorf("DefaultConfig().AutoCreate = %v, want false", cfg.AutoCreate)
	}
	if cfg.Prefix != "v" {
		t.Errorf("DefaultConfig().Prefix = %q, want %q", cfg.Prefix, "v")
	}
	if cfg.Annotate != true {
		t.Errorf("DefaultConfig().Annotate = %v, want true", cfg.Annotate)
	}
	if cfg.Push != false {
		t.Errorf("DefaultConfig().Push = %v, want false", cfg.Push)
	}
	if cfg.Sign != false {
		t.Errorf("DefaultConfig().Sign = %v, want false", cfg.Sign)
	}
	if cfg.SigningKey != "" {
		t.Errorf("DefaultConfig().SigningKey = %q, want empty", cfg.SigningKey)
	}
	if cfg.MessageTemplate != "Release {version}" {
		t.Errorf("DefaultConfig().MessageTemplate = %q, want %q", cfg.MessageTemplate, "Release {version}")
	}
}

func TestTagManagerPlugin_GetConfig(t *testing.T) {
	cfg := &Config{Enabled: true, Prefix: "release-", Push: true}
	tm := NewTagManager(cfg)

	got := tm.GetConfig()
	if got != cfg {
		t.Error("GetConfig() should return the same config passed to NewTagManager")
	}
	if got.Prefix != "release-" {
		t.Errorf("GetConfig().Prefix = %q, want %q", got.Prefix, "release-")
	}
	if got.Push != true {
		t.Errorf("GetConfig().Push = %v, want true", got.Push)
	}
}

func TestTagManagerPlugin_ValidateTagAvailable_Error(t *testing.T) {
	mockOps := &MockGitTagOperations{
		TagExistsFn: func(name string) (bool, error) {
			return false, errors.New("git error")
		},
	}

	tm := NewTagManagerWithOps(nil, mockOps)
	err := tm.ValidateTagAvailable(semver.SemVersion{Major: 1, Minor: 0, Patch: 0})

	if err == nil {
		t.Error("ValidateTagAvailable() should return error when TagExists fails")
	}
}

func TestTagManagerPlugin_CreateTag_PushError(t *testing.T) {
	mockOps := &MockGitTagOperations{
		TagExistsFn: func(name string) (bool, error) {
			return false, nil
		},
		CreateAnnotatedTagFn: func(name, msg string) error {
			return nil
		},
		PushTagFn: func(name string) error {
			return errors.New("push failed")
		},
	}

	cfg := &Config{Enabled: true, AutoCreate: true, Prefix: "v", Annotate: true, Push: true}
	tm := NewTagManagerWithOps(cfg, mockOps)

	err := tm.CreateTag(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, "Release 1.0.0")

	if err == nil {
		t.Error("CreateTag() should return error when push fails")
	}
}

func TestRegister(t *testing.T) {
	// Reset before and after test
	ResetTagManager()
	defer ResetTagManager()

	cfg := &Config{Enabled: true, Prefix: "v"}
	Register(cfg)

	tm := GetTagManagerFn()
	if tm == nil {
		t.Fatal("Register() should register the tag manager")
	}
	if tm.Name() != "tag-manager" {
		t.Errorf("Register() tm.Name() = %q, want %q", tm.Name(), "tag-manager")
	}
}

func TestGetTagManagerFn(t *testing.T) {
	// Reset before and after test
	ResetTagManager()
	defer ResetTagManager()

	// Initially should be nil
	if tm := GetTagManagerFn(); tm != nil {
		t.Error("GetTagManagerFn() should return nil when no manager registered")
	}

	// After registration should return the manager
	Register(&Config{Enabled: true})
	if tm := GetTagManagerFn(); tm == nil {
		t.Error("GetTagManagerFn() should return manager after registration")
	}
}

func TestResetTagManager(t *testing.T) {
	Register(&Config{Enabled: true})

	if tm := GetTagManagerFn(); tm == nil {
		t.Fatal("Expected tag manager to be registered")
	}

	ResetTagManager()

	if tm := GetTagManagerFn(); tm != nil {
		t.Error("ResetTagManager() should clear the registered manager")
	}
}

func TestRegisterTagManager_DoubleRegistration(t *testing.T) {
	// Reset before and after test
	ResetTagManager()
	defer ResetTagManager()

	// Register first manager
	first := NewTagManager(&Config{Enabled: true, Prefix: "v"})
	registerTagManager(first)

	// Attempt to register second manager (should be ignored with warning)
	second := NewTagManager(&Config{Enabled: true, Prefix: "release-"})
	registerTagManager(second)

	// Should still have the first manager
	tm := GetTagManagerFn()
	if tm != first {
		t.Error("Double registration should keep the first manager")
	}
}

func TestTagManagerPlugin_NilConfig(t *testing.T) {
	tm := NewTagManager(nil)

	// Should use defaults
	if got := tm.FormatTagName(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}); got != "v1.0.0" {
		t.Errorf("FormatTagName() with nil config = %q, want %q", got, "v1.0.0")
	}
}

func TestDefaultConfig_TagPrereleases(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TagPrereleases != false {
		t.Errorf("DefaultConfig().TagPrereleases = %v, want false", cfg.TagPrereleases)
	}
}

func TestTagManagerPlugin_ShouldCreateTag(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		version semver.SemVersion
		want    bool
	}{
		{
			name: "stable version with plugin enabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0,
			},
			want: true,
		},
		{
			name: "stable version with plugin disabled",
			cfg:  &Config{Enabled: false, AutoCreate: true, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0,
			},
			want: false,
		},
		{
			name: "stable version with auto-create disabled",
			cfg:  &Config{Enabled: true, AutoCreate: false, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0,
			},
			want: false,
		},
		{
			name: "pre-release with TagPrereleases enabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1",
			},
			want: true,
		},
		{
			name: "pre-release with TagPrereleases disabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: false},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1",
			},
			want: false,
		},
		{
			name: "rc pre-release with TagPrereleases disabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: false},
			version: semver.SemVersion{
				Major: 2, Minor: 0, Patch: 0, PreRelease: "rc.1",
			},
			want: false,
		},
		{
			name: "beta pre-release with TagPrereleases enabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.2",
			},
			want: true,
		},
		{
			name: "stable version after pre-release disabled setting",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: false},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0,
			},
			want: true,
		},
		{
			name: "patch version with TagPrereleases disabled",
			cfg:  &Config{Enabled: true, AutoCreate: true, TagPrereleases: false},
			version: semver.SemVersion{
				Major: 1, Minor: 2, Patch: 5,
			},
			want: true,
		},
		{
			name: "plugin disabled with pre-release",
			cfg:  &Config{Enabled: false, AutoCreate: true, TagPrereleases: true},
			version: semver.SemVersion{
				Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTagManager(tt.cfg)
			if got := tm.ShouldCreateTag(tt.version); got != tt.want {
				t.Errorf("ShouldCreateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagManagerPlugin_ShouldCreateTag_DefaultConfig(t *testing.T) {
	// With default config (AutoCreate: false, TagPrereleases: false)
	// Note: Default config has Enabled: false and AutoCreate: false
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoCreate = true // Enable auto-create to test tagging behavior
	tm := NewTagManager(cfg)

	// Pre-release should NOT be tagged with default config (TagPrereleases: false)
	preRelease := semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"}
	if got := tm.ShouldCreateTag(preRelease); got != false {
		t.Errorf("ShouldCreateTag(preRelease) with default config = %v, want false", got)
	}

	// Stable release should still be tagged
	stable := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}
	if got := tm.ShouldCreateTag(stable); got != true {
		t.Errorf("ShouldCreateTag(stable) with default config = %v, want true", got)
	}
}

func TestTagManagerPlugin_CreateTag_Signed(t *testing.T) {
	t.Run("create signed tag without key", func(t *testing.T) {
		signedCalled := false
		var capturedMessage string
		var capturedKeyID string

		mockOps := &MockGitTagOperations{
			TagExistsFn: func(name string) (bool, error) {
				return false, nil
			},
			CreateSignedTagFn: func(name, msg, keyID string) error {
				signedCalled = true
				capturedMessage = msg
				capturedKeyID = keyID
				return nil
			},
		}

		cfg := &Config{
			Enabled:         true,
			AutoCreate:      true,
			Prefix:          "v",
			Sign:            true,
			MessageTemplate: "Release {version}",
		}
		tm := NewTagManagerWithOps(cfg, mockOps)

		err := tm.CreateTag(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, "")

		if err != nil {
			t.Errorf("CreateTag() error = %v", err)
		}
		if !signedCalled {
			t.Error("CreateTag() should have called createSignedTag")
		}
		if capturedMessage != "Release 1.0.0" {
			t.Errorf("CreateTag() message = %q, want %q", capturedMessage, "Release 1.0.0")
		}
		if capturedKeyID != "" {
			t.Errorf("CreateTag() keyID = %q, want empty", capturedKeyID)
		}
	})

	t.Run("create signed tag with key", func(t *testing.T) {
		signedCalled := false
		var capturedKeyID string

		mockOps := &MockGitTagOperations{
			TagExistsFn: func(name string) (bool, error) {
				return false, nil
			},
			CreateSignedTagFn: func(name, msg, keyID string) error {
				signedCalled = true
				capturedKeyID = keyID
				return nil
			},
		}

		cfg := &Config{
			Enabled:    true,
			AutoCreate: true,
			Prefix:     "v",
			Sign:       true,
			SigningKey: "ABC123DEF456",
		}
		tm := NewTagManagerWithOps(cfg, mockOps)

		err := tm.CreateTag(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, "")

		if err != nil {
			t.Errorf("CreateTag() error = %v", err)
		}
		if !signedCalled {
			t.Error("CreateTag() should have called createSignedTag")
		}
		if capturedKeyID != "ABC123DEF456" {
			t.Errorf("CreateTag() keyID = %q, want %q", capturedKeyID, "ABC123DEF456")
		}
	})

	t.Run("signed tag error", func(t *testing.T) {
		mockOps := &MockGitTagOperations{
			TagExistsFn: func(name string) (bool, error) {
				return false, nil
			},
			CreateSignedTagFn: func(name, msg, keyID string) error {
				return errors.New("gpg signing failed")
			},
		}

		cfg := &Config{
			Enabled:    true,
			AutoCreate: true,
			Prefix:     "v",
			Sign:       true,
		}
		tm := NewTagManagerWithOps(cfg, mockOps)

		err := tm.CreateTag(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, "")

		if err == nil {
			t.Error("CreateTag() should return error when signing fails")
		}
	})
}

func TestTagManagerPlugin_CreateTag_MessageTemplate(t *testing.T) {
	tests := []struct {
		name            string
		template        string
		version         semver.SemVersion
		expectedMessage string
	}{
		{
			name:            "default template",
			template:        "Release {version}",
			version:         semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			expectedMessage: "Release 1.2.3",
		},
		{
			name:            "template with tag",
			template:        "{tag}: Release version {version}",
			version:         semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			expectedMessage: "v2.0.0: Release version 2.0.0",
		},
		{
			name:            "template with prerelease",
			template:        "Release {version} ({prerelease})",
			version:         semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			expectedMessage: "Release 1.0.0-alpha.1 (alpha.1)",
		},
		{
			name:            "empty template uses default",
			template:        "",
			version:         semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			expectedMessage: "Release 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMessage string

			mockOps := &MockGitTagOperations{
				TagExistsFn: func(name string) (bool, error) {
					return false, nil
				},
				CreateAnnotatedTagFn: func(name, msg string) error {
					capturedMessage = msg
					return nil
				},
			}

			cfg := &Config{
				Enabled:         true,
				AutoCreate:      true,
				Prefix:          "v",
				Annotate:        true,
				MessageTemplate: tt.template,
			}
			tm := NewTagManagerWithOps(cfg, mockOps)

			err := tm.CreateTag(tt.version, "")

			if err != nil {
				t.Errorf("CreateTag() error = %v", err)
			}
			if capturedMessage != tt.expectedMessage {
				t.Errorf("CreateTag() message = %q, want %q", capturedMessage, tt.expectedMessage)
			}
		})
	}
}

func TestTagManagerPlugin_CreateTag_ExplicitMessageOverridesTemplate(t *testing.T) {
	var capturedMessage string

	mockOps := &MockGitTagOperations{
		TagExistsFn: func(name string) (bool, error) {
			return false, nil
		},
		CreateAnnotatedTagFn: func(name, msg string) error {
			capturedMessage = msg
			return nil
		},
	}

	cfg := &Config{
		Enabled:         true,
		AutoCreate:      true,
		Prefix:          "v",
		Annotate:        true,
		MessageTemplate: "Template message {version}",
	}
	tm := NewTagManagerWithOps(cfg, mockOps)

	// Explicit message should override template
	explicitMessage := "Custom explicit message"
	err := tm.CreateTag(semver.SemVersion{Major: 1, Minor: 0, Patch: 0}, explicitMessage)

	if err != nil {
		t.Errorf("CreateTag() error = %v", err)
	}
	if capturedMessage != explicitMessage {
		t.Errorf("CreateTag() message = %q, want %q", capturedMessage, explicitMessage)
	}
}

func TestTagManagerPlugin_FormatTagMessage(t *testing.T) {
	tests := []struct {
		name     string
		template string
		version  semver.SemVersion
		prefix   string
		want     string
	}{
		{
			name:     "default template",
			template: "Release {version}",
			version:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			prefix:   "v",
			want:     "Release 1.2.3",
		},
		{
			name:     "complex template",
			template: "{tag}: {version} released",
			version:  semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
			prefix:   "release-",
			want:     "release-2.0.0: 2.0.0 released",
		},
		{
			name:     "empty template uses default",
			template: "",
			version:  semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			prefix:   "v",
			want:     "Release 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Prefix:          tt.prefix,
				MessageTemplate: tt.template,
			}
			tm := NewTagManager(cfg)

			got := tm.FormatTagMessage(tt.version)
			if got != tt.want {
				t.Errorf("FormatTagMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewTagManagerWithOps_NilGitOps(t *testing.T) {
	// When gitOps is nil, it should default to OSGitTagOperations
	tm := NewTagManagerWithOps(nil, nil)

	if tm == nil {
		t.Fatal("NewTagManagerWithOps() returned nil")
	}
	if tm.gitOps == nil {
		t.Error("NewTagManagerWithOps() should set default gitOps when nil")
	}
}
