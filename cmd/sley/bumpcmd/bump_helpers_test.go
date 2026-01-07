package bumpcmd

import (
	"testing"

	"github.com/indaco/sley/internal/semver"
)

/* ------------------------------------------------------------------------- */
/* MOCK IMPLEMENTATIONS FOR HELPER TESTS                                     */
/* ------------------------------------------------------------------------- */

// mockTagManager implements tagmanager.TagManager for testing
type mockTagManager struct {
	validateErr error
	createErr   error
}

func (m *mockTagManager) Name() string                                    { return "mock-tag-manager" }
func (m *mockTagManager) Description() string                             { return "mock tag manager" }
func (m *mockTagManager) Version() string                                 { return "1.0.0" }
func (m *mockTagManager) ValidateTagAvailable(v semver.SemVersion) error  { return m.validateErr }
func (m *mockTagManager) CreateTag(v semver.SemVersion, msg string) error { return m.createErr }
func (m *mockTagManager) FormatTagName(v semver.SemVersion) string        { return "v" + v.String() }
func (m *mockTagManager) TagExists(v semver.SemVersion) (bool, error)     { return false, nil }
func (m *mockTagManager) PushTag(v semver.SemVersion) error               { return nil }
func (m *mockTagManager) DeleteTag(v semver.SemVersion) error             { return nil }
func (m *mockTagManager) GetLatestTag() (semver.SemVersion, error)        { return semver.SemVersion{}, nil }
func (m *mockTagManager) ListTags() ([]string, error)                     { return nil, nil }

// mockVersionValidator implements versionvalidator.VersionValidator for testing
type mockVersionValidator struct {
	validateErr error
}

func (m *mockVersionValidator) Name() string        { return "mock-version-validator" }
func (m *mockVersionValidator) Description() string { return "mock version validator" }
func (m *mockVersionValidator) Version() string     { return "1.0.0" }
func (m *mockVersionValidator) Validate(newV, prevV semver.SemVersion, bumpType string) error {
	return m.validateErr
}
func (m *mockVersionValidator) ValidateSet(v semver.SemVersion) error { return nil }

// mockReleaseGate implements releasegate.ReleaseGate for testing
type mockReleaseGate struct {
	validateErr error
}

func (m *mockReleaseGate) Name() string        { return "mock-release-gate" }
func (m *mockReleaseGate) Description() string { return "mock release gate" }
func (m *mockReleaseGate) Version() string     { return "1.0.0" }
func (m *mockReleaseGate) ValidateRelease(newV, prevV semver.SemVersion, bumpType string) error {
	return m.validateErr
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTION TESTS                                                     */
/* ------------------------------------------------------------------------- */

func TestCalculateNewBuild(t *testing.T) {
	tests := []struct {
		name         string
		meta         string
		preserveMeta bool
		currentBuild string
		expected     string
	}{
		{"new meta overrides", "ci.123", false, "old.456", "ci.123"},
		{"new meta with preserve", "ci.123", true, "old.456", "ci.123"},
		{"preserve existing", "", true, "old.456", "old.456"},
		{"clear when not preserving", "", false, "old.456", ""},
		{"empty when no current", "", true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewBuild(tt.meta, tt.preserveMeta, tt.currentBuild)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractVersionPointers(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name          string
		version       semver.SemVersion
		expectedPre   *string
		expectedBuild *string
	}{
		{
			name:          "both populated",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			expectedPre:   strPtr("alpha.1"),
			expectedBuild: strPtr("ci.99"),
		},
		{
			name:          "only prerelease",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.2"},
			expectedPre:   strPtr("beta.2"),
			expectedBuild: nil,
		},
		{
			name:          "only build",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "build.42"},
			expectedPre:   nil,
			expectedBuild: strPtr("build.42"),
		},
		{
			name:          "both empty",
			version:       semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			expectedPre:   nil,
			expectedBuild: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pre, build := extractVersionPointers(tt.version)
			assertStringPtr(t, "prerelease", tt.expectedPre, pre)
			assertStringPtr(t, "build", tt.expectedBuild, build)
		})
	}
}

func assertStringPtr(t *testing.T, name string, expected, actual *string) {
	t.Helper()
	if expected == nil && actual != nil {
		t.Errorf("expected %s pointer to be nil, got %q", name, *actual)
	}
	if expected != nil && actual == nil {
		t.Errorf("expected %s pointer to be %q, got nil", name, *expected)
	}
	if expected != nil && actual != nil && *expected != *actual {
		t.Errorf("expected %s %q, got %q", name, *expected, *actual)
	}
}

func TestPromotePreRelease(t *testing.T) {
	tests := []struct {
		name         string
		current      semver.SemVersion
		preserveMeta bool
		expected     semver.SemVersion
	}{
		{
			name:         "promote without preserving meta",
			current:      semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			preserveMeta: false,
			expected:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:         "promote with preserving meta",
			current:      semver.SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "ci.99"},
			preserveMeta: true,
			expected:     semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "ci.99"},
		},
		{
			name:         "promote without meta",
			current:      semver.SemVersion{Major: 2, Minor: 0, Patch: 0, PreRelease: "rc.1"},
			preserveMeta: true,
			expected:     semver.SemVersion{Major: 2, Minor: 0, Patch: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := promotePreRelease(tt.current, tt.preserveMeta)
			if result.String() != tt.expected.String() {
				t.Errorf("expected %q, got %q", tt.expected.String(), result.String())
			}
		})
	}
}

func TestSetBuildMetadata(t *testing.T) {
	tests := []struct {
		name     string
		current  semver.SemVersion
		next     semver.SemVersion
		meta     string
		preserve bool
		expected string
	}{
		{
			name:     "set new meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "new",
			preserve: false,
			expected: "new",
		},
		{
			name:     "preserve meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "",
			preserve: true,
			expected: "old",
		},
		{
			name:     "clear meta",
			current:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "old"},
			next:     semver.SemVersion{Major: 1, Minor: 2, Patch: 4},
			meta:     "",
			preserve: false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setBuildMetadata(tt.current, tt.next, tt.meta, tt.preserve)
			if result.Build != tt.expected {
				t.Errorf("expected build %q, got %q", tt.expected, result.Build)
			}
		})
	}
}

func TestModuleInfoFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantNil  bool
		wantDir  string
		wantName string
	}{
		{
			name:    "current directory version file",
			path:    ".version",
			wantNil: true,
		},
		{
			name:    "dot path version file",
			path:    "./.version",
			wantNil: true,
		},
		{
			name:     "subdirectory version file",
			path:     "packages/app/.version",
			wantNil:  false,
			wantDir:  "packages/app",
			wantName: "app",
		},
		{
			name:     "absolute path version file",
			path:     "/project/packages/lib/.version",
			wantNil:  false,
			wantDir:  "/project/packages/lib",
			wantName: "lib",
		},
		{
			name:     "nested module path",
			path:     "apps/frontend/web/.version",
			wantNil:  false,
			wantDir:  "apps/frontend/web",
			wantName: "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := moduleInfoFromPath(tt.path)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected non-nil result, got nil")
				return
			}

			if result.Dir != tt.wantDir {
				t.Errorf("expected Dir=%q, got %q", tt.wantDir, result.Dir)
			}
			if result.Name != tt.wantName {
				t.Errorf("expected Name=%q, got %q", tt.wantName, result.Name)
			}
		})
	}
}
