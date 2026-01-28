package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR URL VALIDATION                                    */
/* ------------------------------------------------------------------------- */

func TestIsURL(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "HTTPS GitHub URL",
			str:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "HTTP URL",
			str:  "http://example.com/path",
			want: true,
		},
		{
			name: "GitHub without protocol",
			str:  "github.com/user/repo",
			want: true,
		},
		{
			name: "GitLab without protocol",
			str:  "gitlab.com/org/project",
			want: true,
		},
		{
			name: "local path",
			str:  "./local/extension",
			want: false,
		},
		{
			name: "absolute local path",
			str:  "/home/user/extension",
			want: false,
		},
		{
			name: "relative path",
			str:  "../extensions/my-ext",
			want: false,
		},
		{
			name: "empty string",
			str:  "",
			want: false,
		},
		{
			name: "just domain",
			str:  "github.com",
			want: false,
		},
		{
			name: "domain with only one path segment",
			str:  "github.com/user",
			want: false,
		},
		{
			name: "whitespace",
			str:  "   ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.str)
			if got != tt.want {
				t.Errorf("IsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsURL_EdgeCases tests IsURL with various edge cases
func TestIsURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "HTTPS with port",
			str:  "https://github.com:443/user/repo",
			want: true,
		},
		{
			name: "with .git extension",
			str:  "github.com/user/repo.git",
			want: true,
		},
		{
			name: "with trailing spaces",
			str:  "  github.com/user/repo  ",
			want: true,
		},
		{
			name: "filename only",
			str:  "extension.yaml",
			want: false,
		},
		{
			name: "current directory",
			str:  ".",
			want: false,
		},
		{
			name: "parent directory",
			str:  "..",
			want: false,
		},
		{
			name: "Windows path",
			str:  "C:\\Users\\ext",
			want: false,
		},
		{
			name: "GitHub with four segments",
			str:  "github.com/org/team/repo",
			want: true,
		},
		{
			name: "GitLab with many segments",
			str:  "gitlab.com/group/subgroup/project",
			want: true,
		},
		{
			name: "bitbucket (not explicitly supported but should match pattern)",
			str:  "bitbucket.org/user/repo",
			want: false, // Not github or gitlab
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.str)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.str, got, tt.want)
			}
		})
	}
}
