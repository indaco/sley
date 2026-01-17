package changelogparser

import (
	"slices"
	"testing"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		format      string
		wantErr     bool
		errContains string
		wantFormat  string
	}{
		{format: "keepachangelog", wantFormat: "keepachangelog"},
		{format: "", wantFormat: "keepachangelog"},
		{format: "grouped", wantFormat: "grouped"},
		{format: "github", wantFormat: "github"},
		{format: "minimal", wantFormat: "minimal"},
		{format: "auto", wantFormat: "auto"},
		{format: "invalid", wantErr: true, errContains: "unknown changelog format"},
		{format: "xml", wantErr: true, errContains: "unknown changelog format"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			parser, err := NewParser(tt.format, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if parser.Format() != tt.wantFormat {
				t.Errorf("Format() = %q, want %q", parser.Format(), tt.wantFormat)
			}
		})
	}
}

func TestValidFormats(t *testing.T) {
	formats := ValidFormats()

	expected := []string{"keepachangelog", "grouped", "github", "minimal", "auto"}
	if len(formats) != len(expected) {
		t.Errorf("ValidFormats() returned %d formats, want %d", len(formats), len(expected))
	}

	for _, exp := range expected {
		found := slices.Contains(formats, exp)
		if !found {
			t.Errorf("ValidFormats() missing %q", exp)
		}
	}
}

func TestParsedSection_Fields(t *testing.T) {
	ps := &ParsedSection{
		Version:            "1.2.3",
		Date:               "2024-01-15",
		HasEntries:         true,
		InferredBumpType:   "minor",
		BumpTypeConfidence: "high",
		Entries: []ParsedEntry{
			{Category: "Added", Description: "Feature"},
		},
	}

	if ps.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", ps.Version, "1.2.3")
	}
	if ps.Date != "2024-01-15" {
		t.Errorf("Date = %q, want %q", ps.Date, "2024-01-15")
	}
	if !ps.HasEntries {
		t.Error("HasEntries should be true")
	}
	if ps.InferredBumpType != "minor" {
		t.Errorf("InferredBumpType = %q, want %q", ps.InferredBumpType, "minor")
	}
	if ps.BumpTypeConfidence != "high" {
		t.Errorf("BumpTypeConfidence = %q, want %q", ps.BumpTypeConfidence, "high")
	}
	if len(ps.Entries) != 1 {
		t.Errorf("Entries length = %d, want 1", len(ps.Entries))
	}
}

func TestParsedEntry_Fields(t *testing.T) {
	pe := ParsedEntry{
		Category:        "Added",
		OriginalSection: "Features",
		Description:     "New authentication",
		Scope:           "auth",
		IsBreaking:      true,
		CommitType:      "feat",
	}

	if pe.Category != "Added" {
		t.Errorf("Category = %q, want %q", pe.Category, "Added")
	}
	if pe.OriginalSection != "Features" {
		t.Errorf("OriginalSection = %q, want %q", pe.OriginalSection, "Features")
	}
	if pe.Description != "New authentication" {
		t.Errorf("Description = %q, want %q", pe.Description, "New authentication")
	}
	if pe.Scope != "auth" {
		t.Errorf("Scope = %q, want %q", pe.Scope, "auth")
	}
	if !pe.IsBreaking {
		t.Error("IsBreaking should be true")
	}
	if pe.CommitType != "feat" {
		t.Errorf("CommitType = %q, want %q", pe.CommitType, "feat")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
