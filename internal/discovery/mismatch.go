package discovery

import (
	"sort"
)

// DetectMismatches analyzes discovery results and identifies version inconsistencies.
// It uses the primary version as the expected version and flags any sources that differ.
func DetectMismatches(result *Result) []Mismatch {
	if result == nil {
		return nil
	}

	// Get the expected version (primary source of truth)
	expectedVersion := result.PrimaryVersion()
	if expectedVersion == "" {
		// No primary version, can't detect mismatches
		return nil
	}

	var mismatches []Mismatch

	// Check all modules
	for _, m := range result.Modules {
		if m.Version != "" && m.Version != expectedVersion {
			mismatches = append(mismatches, Mismatch{
				Source:          m.RelPath,
				ExpectedVersion: expectedVersion,
				ActualVersion:   m.Version,
			})
		}
	}

	// Check all manifests
	for _, m := range result.Manifests {
		if m.Version != "" && m.Version != expectedVersion {
			mismatches = append(mismatches, Mismatch{
				Source:          m.RelPath,
				ExpectedVersion: expectedVersion,
				ActualVersion:   m.Version,
			})
		}
	}

	// Sort mismatches by source path for consistent output
	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Source < mismatches[j].Source
	})

	return mismatches
}

// DetectMismatchesWithCustomBase checks for mismatches using a specified base version.
func DetectMismatchesWithCustomBase(result *Result, baseVersion string) []Mismatch {
	if result == nil || baseVersion == "" {
		return nil
	}

	var mismatches []Mismatch

	// Check all modules
	for _, m := range result.Modules {
		if m.Version != "" && m.Version != baseVersion {
			mismatches = append(mismatches, Mismatch{
				Source:          m.RelPath,
				ExpectedVersion: baseVersion,
				ActualVersion:   m.Version,
			})
		}
	}

	// Check all manifests
	for _, m := range result.Manifests {
		if m.Version != "" && m.Version != baseVersion {
			mismatches = append(mismatches, Mismatch{
				Source:          m.RelPath,
				ExpectedVersion: baseVersion,
				ActualVersion:   m.Version,
			})
		}
	}

	// Sort mismatches by source path
	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Source < mismatches[j].Source
	})

	return mismatches
}

// GetUniqueVersions returns a list of unique versions found in the discovery result.
func GetUniqueVersions(result *Result) []string {
	if result == nil {
		return nil
	}

	versionSet := make(map[string]struct{})

	// Collect versions from modules
	for _, m := range result.Modules {
		if m.Version != "" {
			versionSet[m.Version] = struct{}{}
		}
	}

	// Collect versions from manifests
	for _, m := range result.Manifests {
		if m.Version != "" {
			versionSet[m.Version] = struct{}{}
		}
	}

	// Convert to slice
	versions := make([]string, 0, len(versionSet))
	for v := range versionSet {
		versions = append(versions, v)
	}

	// Sort for consistent output
	sort.Strings(versions)

	return versions
}

// IsVersionConsistent returns true if all discovered sources have the same version.
func IsVersionConsistent(result *Result) bool {
	if result == nil {
		return true
	}

	versions := GetUniqueVersions(result)
	return len(versions) <= 1
}

// VersionSummary provides a summary of version distribution.
type VersionSummary struct {
	// Version is the version string.
	Version string

	// Count is how many sources have this version.
	Count int

	// Sources lists the paths with this version.
	Sources []string
}

// GetVersionSummary returns a summary of version distribution across sources.
func GetVersionSummary(result *Result) []VersionSummary {
	if result == nil {
		return nil
	}

	versionMap := make(map[string][]string)

	// Collect versions from modules
	for _, m := range result.Modules {
		if m.Version != "" {
			versionMap[m.Version] = append(versionMap[m.Version], m.RelPath)
		}
	}

	// Collect versions from manifests
	for _, m := range result.Manifests {
		if m.Version != "" {
			versionMap[m.Version] = append(versionMap[m.Version], m.RelPath)
		}
	}

	// Convert to slice
	summaries := make([]VersionSummary, 0, len(versionMap))
	for v, sources := range versionMap {
		sort.Strings(sources)
		summaries = append(summaries, VersionSummary{
			Version: v,
			Count:   len(sources),
			Sources: sources,
		})
	}

	// Sort by count (descending), then by version
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Count != summaries[j].Count {
			return summaries[i].Count > summaries[j].Count
		}
		return summaries[i].Version < summaries[j].Version
	})

	return summaries
}
