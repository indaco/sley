package extensionmgr

import (
	"testing"
)

func TestReplaceYAMLSection(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		key         string
		replacement string
		wantResult  string
		wantFound   bool
	}{
		{
			name:        "replace empty section",
			content:     "path: .version\nextensions: []\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: foo\n    path: bar\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "replace section with existing entries",
			content:     "path: .version\nextensions:\n  - name: old\n    path: old/path\n    enabled: true\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: old\n    path: old/path\n    enabled: true\n  - name: new\n    path: new/path\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: old\n    path: old/path\n    enabled: true\n  - name: new\n    path: new/path\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "key not found",
			content:     "path: .version\nplugins:\n  commit-parser: true\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nplugins:\n  commit-parser: true\n",
			wantFound:   false,
		},
		{
			name:        "preserves surrounding content and comments",
			content:     "# Header comment\npath: .version # inline comment\n\n# Extensions section\nextensions: []\n\n# Hooks section\npre-release-hooks:\n  - changelog:\n      command: git-chglog\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: test\n    path: test/path\n    enabled: true",
			wantResult:  "# Header comment\npath: .version # inline comment\n\n# Extensions section\nextensions:\n  - name: test\n    path: test/path\n    enabled: true\n\n# Hooks section\npre-release-hooks:\n  - changelog:\n      command: git-chglog\n",
			wantFound:   true,
		},
		{
			name:        "section at end of file without trailing newline",
			content:     "path: .version\nextensions: []",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: foo\n    path: bar\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "does not match indented key",
			content:     "path: .version\nplugins:\n  extensions: something\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo",
			wantResult:  "path: .version\nplugins:\n  extensions: something\n",
			wantFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotFound := replaceYAMLSection(tt.content, tt.key, tt.replacement)
			if gotFound != tt.wantFound {
				t.Errorf("replaceYAMLSection() found = %v, want %v", gotFound, tt.wantFound)
			}
			if gotResult != tt.wantResult {
				t.Errorf("replaceYAMLSection() result mismatch.\ngot:\n%s\nwant:\n%s", gotResult, tt.wantResult)
			}
		})
	}
}
