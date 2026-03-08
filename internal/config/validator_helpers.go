package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// validateEnum checks that a value is in the allowed set. Returns true if valid.
func (v *Validator) validateEnum(category, field, value string, allowed map[string]bool) bool {
	if !allowed[value] {
		keys := make([]string, 0, len(allowed))
		for k := range allowed {
			if k != "" {
				keys = append(keys, "'"+k+"'")
			}
		}
		v.addValidation(category, false,
			fmt.Sprintf("Invalid %s '%s': must be one of %s", field, value, joinWithOr(keys)), false)
		return false
	}
	return true
}

// validateFileExists checks that a file exists at the given path.
// label is used in the validation message (e.g., "File 1", "Changelog file").
// Returns true if the file exists.
func (v *Validator) validateFileExists(ctx context.Context, category, label, rawPath string) bool {
	absPath := v.resolvePath(rawPath)

	if _, err := v.fs.Stat(ctx, absPath); err != nil {
		if os.IsNotExist(err) {
			v.addValidation(category, false,
				fmt.Sprintf("%s: '%s' does not exist", label, rawPath), false)
		} else {
			v.addValidation(category, false,
				fmt.Sprintf("%s: cannot access '%s': %v", label, rawPath, err), false)
		}
		return false
	}
	return true
}

// validateRegex checks that a pattern is a valid regular expression.
// Returns true if valid.
func (v *Validator) validateRegex(category, label, pattern string) bool {
	if _, err := regexp.Compile(pattern); err != nil {
		v.addValidation(category, false,
			fmt.Sprintf("%s: invalid regex: %v", label, err), false)
		return false
	}
	return true
}

// resolvePath resolves a path relative to the validator's root directory.
// Absolute paths are returned as-is.
func (v *Validator) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(v.rootDir, path)
}

// joinWithOr joins strings with commas and "or" before the last element.
func joinWithOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		var b strings.Builder
		for i, item := range items {
			if i == len(items)-1 {
				b.WriteString("or ")
				b.WriteString(item)
			} else {
				b.WriteString(item)
				b.WriteString(", ")
			}
		}
		return b.String()
	}
}
