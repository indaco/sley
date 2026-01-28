package semver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/testutils"
)

// mockGitTagReader implements GitTagReader for testing.
type mockGitTagReader struct {
	tag string
	err error
}

func (m *mockGitTagReader) DescribeTags(ctx context.Context) (string, error) {
	return m.tag, m.err
}

func TestSemVersion_String_WithBuildOnly(t *testing.T) {
	v := SemVersion{
		Major: 1,
		Minor: 0,
		Patch: 0,
		Build: "exp.sha.5114f85",
	}

	got := v.String()
	want := "1.0.0+exp.sha.5114f85"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTIONS FOR COMPARISON                                           */
/* ------------------------------------------------------------------------- */

func Test_compareInt(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{0, 0, 0},
		{1, 1, 0},
		{1, 2, -1},
		{2, 1, 1},
		{0, 1, -1},
		{1, 0, 1},
		{-1, 1, -1},
		{1, -1, 1},
		{100, 99, 1},
		{99, 100, -1},
	}

	for _, tt := range tests {
		got := compareInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func Test_parseNumericIdentifier(t *testing.T) {
	tests := []struct {
		input   string
		wantNum int
		wantOk  bool
	}{
		// Valid numeric identifiers
		{"0", 0, true},
		{"1", 1, true},
		{"10", 10, true},
		{"123", 123, true},

		// Invalid: empty string
		{"", 0, false},

		// Invalid: leading zeros (per semver spec)
		{"00", 0, false},
		{"01", 0, false},
		{"007", 0, false},

		// Invalid: non-numeric
		{"a", 0, false},
		{"1a", 0, false},
		{"a1", 0, false},
		{"alpha", 0, false},
		{"1.0", 0, false},
		{"-1", 0, false},
	}

	for _, tt := range tests {
		num, ok := parseNumericIdentifier(tt.input)
		if ok != tt.wantOk {
			t.Errorf("parseNumericIdentifier(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
		}
		if ok && num != tt.wantNum {
			t.Errorf("parseNumericIdentifier(%q) = %d, want %d", tt.input, num, tt.wantNum)
		}
	}
}

func Test_compareIdentifier(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		// Both numeric: compare numerically
		{"1", "2", -1},
		{"2", "1", 1},
		{"1", "1", 0},
		{"2", "10", -1}, // numeric: 2 < 10
		{"10", "2", 1},  // numeric: 10 > 2
		{"10", "10", 0},
		{"99", "100", -1},

		// Numeric vs alphanumeric: numeric has lower precedence
		{"1", "alpha", -1},
		{"alpha", "1", 1},
		{"0", "a", -1},
		{"999", "a", -1},

		// Both alphanumeric: compare lexicographically
		{"alpha", "alpha", 0},
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},
		{"rc", "rc", 0},
		{"a", "b", -1},
		{"z", "a", 1},
	}

	for _, tt := range tests {
		got := compareIdentifier(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareIdentifier(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func Test_comparePreRelease(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		// Simple alphanumeric
		{"alpha", "alpha", 0},
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},

		// Numeric identifiers compared numerically
		{"rc.1", "rc.2", -1},
		{"rc.2", "rc.1", 1},
		{"rc.2", "rc.10", -1}, // 2 < 10 numerically (not lexicographically!)
		{"rc.10", "rc.2", 1},
		{"rc.10", "rc.10", 0},

		// Mixed: alpha.1 vs alpha.2
		{"alpha.1", "alpha.2", -1},
		{"alpha.2", "alpha.1", 1},

		// Different prefixes
		{"alpha.1", "beta.1", -1},
		{"beta.1", "alpha.1", 1},
		{"alpha", "alpha.1", -1}, // shorter has lower precedence
		{"alpha.1", "alpha", 1},

		// Complex pre-release identifiers
		{"1.0.0", "1.0.1", -1},
		{"alpha.1.beta", "alpha.1.beta", 0},
		{"alpha.1.beta", "alpha.1.gamma", -1},

		// Numeric vs alphanumeric identifier
		{"alpha.1", "alpha.beta", -1}, // 1 (numeric) < beta (alphanumeric)
		{"alpha.beta", "alpha.1", 1},

		// Edge cases
		{"0", "0", 0},
		{"0", "1", -1},
		{"1", "0", 1},
	}

	for _, tt := range tests {
		got := comparePreRelease(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("comparePreRelease(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION COMPARISON                                                        */
/* ------------------------------------------------------------------------- */

func TestSemVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    SemVersion
		b    SemVersion
		want int
	}{
		{
			name: "equal versions",
			a:    SemVersion{Major: 1, Minor: 2, Patch: 3},
			b:    SemVersion{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "major less",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0},
			b:    SemVersion{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "major greater",
			a:    SemVersion{Major: 2, Minor: 0, Patch: 0},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0},
			want: 1,
		},
		{
			name: "minor less",
			a:    SemVersion{Major: 1, Minor: 2, Patch: 0},
			b:    SemVersion{Major: 1, Minor: 3, Patch: 0},
			want: -1,
		},
		{
			name: "minor greater",
			a:    SemVersion{Major: 1, Minor: 3, Patch: 0},
			b:    SemVersion{Major: 1, Minor: 2, Patch: 0},
			want: 1,
		},
		{
			name: "patch less",
			a:    SemVersion{Major: 1, Minor: 2, Patch: 3},
			b:    SemVersion{Major: 1, Minor: 2, Patch: 4},
			want: -1,
		},
		{
			name: "patch greater",
			a:    SemVersion{Major: 1, Minor: 2, Patch: 4},
			b:    SemVersion{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			name: "0.10.0 greater than 0.9.1 (semantic, not lexicographic)",
			a:    SemVersion{Major: 0, Minor: 10, Patch: 0},
			b:    SemVersion{Major: 0, Minor: 9, Patch: 1},
			want: 1,
		},
		{
			name: "0.10.0 greater than 0.2.0 (semantic, not lexicographic)",
			a:    SemVersion{Major: 0, Minor: 10, Patch: 0},
			b:    SemVersion{Major: 0, Minor: 2, Patch: 0},
			want: 1,
		},
		{
			name: "pre-release less than release",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "release greater than pre-release",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"},
			want: 1,
		},
		{
			name: "alpha less than beta pre-release",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "beta"},
			want: -1,
		},
		{
			name: "equal pre-release",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.1"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.1"},
			want: 0,
		},
		{
			name: "rc.2 less than rc.10 (numeric comparison)",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.2"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.10"},
			want: -1,
		},
		{
			name: "rc.10 greater than rc.2 (numeric comparison)",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.10"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.2"},
			want: 1,
		},
		{
			name: "alpha.1 less than alpha.beta (numeric < alphanumeric)",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.beta"},
			want: -1,
		},
		{
			name: "alpha less than alpha.1 (shorter has lower precedence)",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			want: -1,
		},
		{
			name: "build metadata ignored",
			a:    SemVersion{Major: 1, Minor: 0, Patch: 0, Build: "build.1"},
			b:    SemVersion{Major: 1, Minor: 0, Patch: 0, Build: "build.2"},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Compare(tt.b)
			if got != tt.want {
				t.Errorf("(%s).Compare(%s) = %d, want %d",
					tt.a.String(), tt.b.String(), got, tt.want)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION FILE INITIALIZATION                                               */
/* ------------------------------------------------------------------------- */

func TestInitializeVersionFileWithFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("file already exists and is valid", func(t *testing.T) {
		path := testutils.WriteTempVersionFile(t, tmpDir, "2.3.4")

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}

	})

	t.Run("file already exists and is invalid", func(t *testing.T) {
		path := testutils.WriteTempVersionFile(t, tmpDir, "not-a-version")

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error from feedback function: %v", err)
		}
		if created {
			t.Errorf("expected created=false for existing file, got true")
		}

		// Now test the actual parse failure
		_, err = ReadVersion(path)
		if err == nil {
			t.Fatal("expected error from ReadVersion, got nil")
		}
		if !strings.Contains(err.Error(), "invalid version format") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("file does not exist, fallback to git tag", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, ".version")

		// Use mock git tag reader that returns a valid tag
		mockGit := &mockGitTagReader{tag: "v1.2.3\n", err: nil}
		mgr := NewVersionManager(core.NewOSFileSystem(), mockGit)
		restore := SetDefaultManager(mgr)
		defer restore()

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}

		// Verify content
		data, _ := os.ReadFile(path)
		got := strings.TrimSpace(string(data))
		if got != "1.2.3" {
			t.Errorf("expected 1.2.3, got %q", got)
		}
	})

	t.Run("file does not exist, fallback to default 0.1.0", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, ".version")

		// Use mock git tag reader that returns an invalid tag
		mockGit := &mockGitTagReader{tag: "invalid-tag\n", err: nil}
		mgr := NewVersionManager(core.NewOSFileSystem(), mockGit)
		restore := SetDefaultManager(mgr)
		defer restore()

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}

		// Verify content
		data, _ := os.ReadFile(path)
		got := strings.TrimSpace(string(data))
		if got != "0.1.0" {
			t.Errorf("expected 0.1.0, got %q", got)
		}
	})
}

/* ------------------------------------------------------------------------- */
/* VERSION PARSING                                                           */
/* ------------------------------------------------------------------------- */

func TestParseAndString(t *testing.T) {
	tests := []struct {
		raw      string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
		{"  1.2.3-beta.1 ", "1.2.3-beta.1"},
	}

	for _, tt := range tests {
		v, err := ParseVersion(tt.raw)
		if err != nil {
			t.Errorf("ParseVersion(%q) failed: %v", tt.raw, err)
			continue
		}
		if v.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, v.String())
		}
	}
}

func TestParseVersion_ValidWithVPrefix(t *testing.T) {
	v, err := ParseVersion("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("unexpected version parsed: %+v", v)
	}
}

func TestParseVersion_ErrorCases(t *testing.T) {
	t.Run("invalid format (missing patch)", func(t *testing.T) {
		_, err := ParseVersion("1.2")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric major", func(t *testing.T) {
		_, err := ParseVersion("a.2.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric minor", func(t *testing.T) {
		_, err := ParseVersion("1.b.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric patch", func(t *testing.T) {
		_, err := ParseVersion("1.2.c")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("version string too long (ReDoS prevention)", func(t *testing.T) {
		// Create a version string that exceeds maxVersionLength
		longVersion := "1.0.0-" + strings.Repeat("a", maxVersionLength)
		_, err := ParseVersion(longVersion)
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion for too-long version, got %v", err)
		}
		if err != nil && !strings.Contains(err.Error(), "exceeds maximum length") {
			t.Errorf("expected error message about maximum length, got %v", err)
		}
	})
}

func TestParseVersion_InvalidFormat(t *testing.T) {
	invalidVersions := []string{
		"",
		"1",
		"1.2",
		"abc.def.ghi",
		"1.2.3.4", // too many parts
	}

	for _, raw := range invalidVersions {
		_, err := ParseVersion(raw)
		if err == nil {
			t.Errorf("expected error for invalid version %q, got nil", raw)
		}
	}
}

func TestParseVersion_NumberConversionErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{"a.2.3", "invalid major version"},
		{"1.b.3", "invalid minor version"},
		{"1.2.c", "invalid patch version"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseVersion(tt.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %v", tt.expectedError, err)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION UPDATES                                                           */
/* ------------------------------------------------------------------------- */

func TestUpdateVersion_Scenarios(t *testing.T) {
	tmpDir := os.TempDir()
	tests := []struct {
		name        string
		initial     string
		level       string
		pre         string
		meta        string
		preserve    bool
		expected    string
		expectErr   bool
		expectedErr string
	}{
		{"patch bump", "1.2.3", "patch", "", "", false, "1.2.4", false, ""},
		{"minor bump", "1.2.3", "minor", "", "", false, "1.3.0", false, ""},
		{"major bump", "1.2.3", "major", "", "", false, "2.0.0", false, ""},
		{"with pre-release", "1.2.3", "patch", "alpha.1", "", false, "1.2.4-alpha.1", false, ""},
		{"with metadata", "1.2.3", "patch", "", "ci.123", false, "1.2.4+ci.123", false, ""},
		{"with pre + metadata", "1.2.3", "patch", "rc.1", "ci.456", false, "1.2.4-rc.1+ci.456", false, ""},
		{"preserve metadata", "1.2.3+build.789", "patch", "", "", true, "1.2.4+build.789", false, ""},
		{"clear metadata", "1.2.3+build.789", "patch", "", "", false, "1.2.4", false, ""},
		{"preserve metadata but override", "1.2.3+build.789", "patch", "", "custom.1", true, "1.2.4+custom.1", false, ""},
		{"invalid bump level", "1.2.3", "banana", "", "", false, "", true, "invalid bump type"},
		{"invalid initial version", "not-a-version", "patch", "", "", false, "", true, "invalid version format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			defer os.Remove(path)

			err := UpdateVersion(path, tt.level, tt.pre, tt.meta, tt.preserve)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error to contain %q, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := strings.TrimSpace(testutils.ReadFile(t, path))
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION READ/WRITE                                                        */
/* ------------------------------------------------------------------------- */

func TestReadVersion_FileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.version")

	_, err := ReadVersion(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	if !os.IsNotExist(err) {
		t.Errorf("expected file-not-found error, got %v", err)
	}
}

func TestSetPreRelease(t *testing.T) {
	tmpDir := os.TempDir()
	path := testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	defer os.Remove(path)

	version, _ := ReadVersion(path)
	version.PreRelease = "rc.1"
	if err := SaveVersion(path, version); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(testutils.ReadFile(t, path))
	want := "1.2.3-rc.1"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

/* ------------------------------------------------------------------------- */
/* PRE-RELEASE INCREMENTS                                                    */
/* ------------------------------------------------------------------------- */

func TestIncrementPreRelease(t *testing.T) {
	cases := []struct {
		current string
		base    string
		want    string
	}{
		// Dot separator (e.g., rc.1)
		{"alpha", "alpha", "alpha.1"},
		{"alpha.", "alpha", "alpha.1"},
		{"alpha.1", "alpha", "alpha.2"},
		{"alpha.9", "alpha", "alpha.10"},
		{"rc.1", "rc", "rc.2"},
		{"rc.99", "rc", "rc.100"},

		// Dash separator (e.g., rc-1)
		{"rc-1", "rc", "rc-2"},
		{"rc-9", "rc", "rc-10"},
		{"beta-5", "beta", "beta-6"},

		// No separator (e.g., rc1)
		{"rc1", "rc", "rc2"},
		{"rc9", "rc", "rc10"},
		{"beta5", "beta", "beta6"},

		// Label switch (different base)
		{"beta", "alpha", "alpha.1"},
		{"beta.3", "alpha", "alpha.1"},
		{"rc1", "beta", "beta.1"},
		{"rc-2", "beta", "beta.1"},

		// Empty current
		{"", "rc", "rc.1"},
	}

	for _, c := range cases {
		got := IncrementPreRelease(c.current, c.base)
		if got != c.want {
			t.Errorf("incrementPreRelease(%q, %q) = %q, want %q", c.current, c.base, got, c.want)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION FILE INITIALIZATION WITH MOCKS                                    */
/* ------------------------------------------------------------------------- */

func TestInitialize_NewFile_WithValidGitTag(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Use mock git tag reader that returns a valid tag
	mockGit := &mockGitTagReader{tag: "v1.2.3\n", err: nil}
	mgr := NewVersionManager(core.NewOSFileSystem(), mockGit)
	restore := SetDefaultManager(mgr)
	defer restore()

	err := mgr.Initialize(context.Background(), versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	want := "1.2.3"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestInitialize_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Use mock git tag reader - should not be called since file exists
	mockGit := &mockGitTagReader{tag: "v9.9.9\n", err: nil}
	mgr := NewVersionManager(core.NewOSFileSystem(), mockGit)
	restore := SetDefaultManager(mgr)
	defer restore()

	err = mgr.Initialize(context.Background(), versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	if got != "1.2.3" {
		t.Errorf("expected file content to remain '1.2.3', got %q", got)
	}
}

func TestSaveVersion_MkdirAllFails(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file where the directory is expected
	conflictPath := filepath.Join(tmpDir, "conflict")
	if err := os.WriteFile(conflictPath, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}

	versionFile := filepath.Join(conflictPath, ".version") // invalid: parent is a file

	err := SaveVersion(versionFile, SemVersion{1, 2, 3, "", ""})
	if err == nil {
		t.Fatal("expected error due to mkdir on a file, got nil")
	}

	if !strings.Contains(err.Error(), "not a directory") && !strings.Contains(err.Error(), "is a file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInitialize_InvalidGitTagFormat(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Use mock git tag reader that returns an invalid tag
	mockGit := &mockGitTagReader{tag: "invalid-tag\n", err: nil}
	mgr := NewVersionManager(core.NewOSFileSystem(), mockGit)
	restore := SetDefaultManager(mgr)
	defer restore()

	err := mgr.Initialize(context.Background(), versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	want := "0.1.0"

	if got != want {
		t.Errorf("expected fallback version %q, got %q", want, got)
	}
}

func TestInitializeVersionFileWithFeedback_InitializationFails(t *testing.T) {
	tmp := t.TempDir()
	noWrite := filepath.Join(tmp, "nowrite")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join(noWrite, ".version")

	created, err := InitializeVersionFileWithFeedback(versionPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if created {
		t.Errorf("expected created to be false, got true")
	}
}

func TestReadVersion_InvalidContent(t *testing.T) {
	// Test that ReadVersion correctly reports an error for invalid version content.
	// This uses the new dependency injection pattern with MockFileSystem.
	mockFS := core.NewMockFileSystem()
	mockFS.SetFile("/test/.version", []byte("not-a-version\n"))

	mgr := NewVersionManager(mockFS, nil)
	restore := SetDefaultManager(mgr)
	defer restore()

	_, err := ReadVersion("/test/.version")
	if err == nil {
		t.Fatal("expected error from ReadVersion, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP NEXT                                                                 */
/* ------------------------------------------------------------------------- */

func TestBumpNext(t *testing.T) {
	tests := []struct {
		name     string
		current  SemVersion
		expected SemVersion
	}{
		{
			name: "promote alpha pre-release",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1",
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name: "promote rc pre-release",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "rc.1",
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name: "default patch bump",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 4,
			},
		},
		{
			name: "promote 0.x alpha to final",
			current: SemVersion{
				Major: 0, Minor: 9, Patch: 0, PreRelease: "alpha.1",
			},
			expected: SemVersion{
				Major: 0, Minor: 9, Patch: 0,
			},
		},
		{
			name: "optional heuristic bump from 0.9.0 to 0.10.0",
			current: SemVersion{
				Major: 0, Minor: 9, Patch: 0,
			},
			expected: SemVersion{
				Major: 0, Minor: 10, Patch: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpNext(tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.String(), got.String())
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP BY LABEL                                                             */
/* ------------------------------------------------------------------------- */

func TestBumpByLabel(t *testing.T) {
	tests := []struct {
		name     string
		current  SemVersion
		label    string
		expected string
		wantErr  bool
	}{
		{"patch bump", SemVersion{1, 2, 3, "", ""}, "patch", "1.2.4", false},
		{"minor bump", SemVersion{1, 2, 3, "", ""}, "minor", "1.3.0", false},
		{"major bump", SemVersion{1, 2, 3, "", ""}, "major", "2.0.0", false},
		{"invalid label", SemVersion{1, 2, 3, "", ""}, "foobar", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpByLabel(tt.current, tt.label)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got.String())
			}
		})
	}
}
