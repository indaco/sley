# Changelog Parser Plugin

The `changelogparser` plugin parses CHANGELOG.md files in [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format to automatically infer version bump types and validate changelog completeness.

## Plugin Metadata

| Field       | Value                              |
| ----------- | ---------------------------------- |
| Name        | `changelog-parser`                 |
| Version     | v0.1.0                             |
| Type        | `changelog-parser`                 |
| Description | Infers bump type from CHANGELOG.md |

## Features

- Parses `## [Unreleased]` section to detect change types
- Infers bump type based on changelog subsections
- Validates that changelog has entries before allowing release
- Supports configurable priority over commit-based inference
- Works alongside the commit parser as fallback or primary parser

## Bump Type Inference Rules

The plugin analyzes subsections in the Unreleased section and determines the bump type based on the following priority:

| Changelog Section | Bump Type | Rationale                      |
| ----------------- | --------- | ------------------------------ |
| Removed           | major     | Breaking change (removal)      |
| Changed           | major     | Breaking change (modification) |
| Added             | minor     | New feature                    |
| Fixed             | patch     | Bug fix                        |
| Security          | patch     | Security fix                   |
| Deprecated        | patch     | Deprecation notice             |

### Priority Order

1. **Major**: Removed or Changed sections take highest priority
2. **Minor**: Added section if no major changes
3. **Patch**: Fixed, Security, or Deprecated if no major/minor changes

## Configuration

Add the following to your `.sley.yaml` file:

```yaml
plugins:
  changelog-parser:
    enabled: true
    path: "CHANGELOG.md"
    require-unreleased-section: true
    infer-bump-type: true
    priority: "changelog" # or "commits"
```

### Configuration Options

- `enabled` (bool): Controls whether the plugin is active (default: false)
- `path` (string): Path to the changelog file (default: "CHANGELOG.md")
- `require-unreleased-section` (bool): Enforce presence of Unreleased section (default: true)
- `infer-bump-type` (bool): Enable automatic bump type inference (default: true)
- `priority` (string): Which parser takes precedence - "changelog" or "commits" (default: "changelog")

## Usage with `bump auto`

When enabled with `priority: "changelog"`, the plugin will:

1. Parse the CHANGELOG.md file
2. Extract the `## [Unreleased]` section
3. Analyze subsections (Added, Changed, Fixed, etc.)
4. Infer the appropriate bump type (major, minor, or patch)
5. Fall back to commit parser if changelog has no entries

### Example Workflow

```bash
# 1. Edit your CHANGELOG.md
$ cat CHANGELOG.md
# Changelog

## [Unreleased]

### Added
- New feature X

### Fixed
- Bug fix Y

# 2. Run bump auto - it will detect "minor" from "Added" section
$ sley bump auto
Inferred from changelog: minor
Bumped version from 1.2.3 to 1.3.0

# 3. The changelog takes precedence over commits
# Even if commits suggest "patch", "Added" section forces "minor"
```

## Integration with Other Plugins

### With Commit Parser

When both plugins are enabled:

- **priority: "changelog"**: Changelog parser runs first, commit parser is fallback
- **priority: "commits"**: Commit parser runs first, changelog parser is ignored

```yaml
plugins:
  changelog-parser:
    enabled: true
    priority: "changelog"
  commit-parser: true
```

### With Tag Manager

Combine with tag manager for automatic git tag creation:

```yaml
plugins:
  changelog-parser:
    enabled: true
  tag-manager:
    enabled: true
    auto-create: true
```

### With Version Validator

Ensure changelog entries exist before bumping:

```yaml
plugins:
  changelog-parser:
    enabled: true
    require-unreleased-section: true
  version-validator:
    enabled: true
```

## Validation

The plugin can validate changelog completeness before version bumps:

- Checks that `## [Unreleased]` section exists
- Ensures the section has at least one entry
- Fails fast if validation is enabled and requirements aren't met

## Keep a Changelog Format

The plugin expects standard Keep a Changelog format:

```markdown
# Changelog

## [Unreleased]

### Added

- New feature description

### Changed

- Modified behavior description

### Fixed

- Bug fix description

### Security

- Security fix description

### Deprecated

- Deprecation notice

### Removed

- Removed feature description

## [1.2.3] - 2024-01-15

### Added

- Previous feature
```

### Supported Subsection Headers

- `### Added` - New features
- `### Changed` - Changes in existing functionality
- `### Deprecated` - Soon-to-be removed features
- `### Removed` - Removed features
- `### Fixed` - Bug fixes
- `### Security` - Security fixes

All subsection headers are case-insensitive.

## See Also

- [Example Configuration](./examples/changelog-parser.yaml) - Complete changelog-parser setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Example CHANGELOG.md](./examples/CHANGELOG-example.md) - Sample changelog file format

## Testing

The plugin includes comprehensive test coverage (99.0%):

- Parser tests for various changelog formats
- Inference logic tests for all bump types
- Integration tests with bump auto command
- Error handling and edge case tests

Run tests:

```bash
go test ./internal/plugins/changelogparser/...
```

## Implementation Details

### File Structure

```
internal/plugins/changelogparser/
├── parser.go          # Changelog file parsing logic
├── parser_test.go     # Parser unit tests
├── plugin.go          # Plugin implementation
├── plugin_test.go     # Plugin unit tests
└── registry.go        # Singleton registry pattern
```

### Key Components

1. **changelogFileParser**: Parses CHANGELOG.md files
2. **UnreleasedSection**: Represents parsed Unreleased section
3. **ChangelogParserPlugin**: Plugin implementation with configuration
4. **Registry**: Singleton pattern for plugin registration

### Function Variables for Testability

The implementation uses function variables (like `openFileFn`) to enable dependency injection during testing, following the patterns established in other plugins.

## Comparison with Commit Parser

| Feature                | Changelog Parser | Commit Parser        |
| ---------------------- | ---------------- | -------------------- |
| Input source           | CHANGELOG.md     | Git commit messages  |
| Format requirement     | Keep a Changelog | Conventional Commits |
| Explicit documentation | Yes              | No                   |
| Manual control         | High             | Low                  |
| Automation level       | Semi-automatic   | Fully automatic      |
| Best for               | Release planning | CI/CD workflows      |

## Best Practices

1. **Keep Unreleased Section Updated**: Update the Unreleased section as you develop
2. **Use Clear Subsections**: Categorize changes appropriately (Added, Fixed, etc.)
3. **Combine with Commits**: Use "commits" priority for automated CI, "changelog" for manual releases
4. **Validate Before Release**: Enable `require-unreleased-section` to prevent empty releases
5. **Document Breaking Changes**: Use "Changed" or "Removed" for breaking changes

## Troubleshooting

### Plugin Not Inferring Bump Type

- Check that `enabled: true` and `infer-bump-type: true`
- Verify `priority: "changelog"` if you want it to override commits
- Ensure Unreleased section has properly formatted subsections

### Validation Errors

- Verify `## [Unreleased]` section exists
- Check that subsections use `###` headers
- Ensure entries start with `- ` (dash and space)

### Priority Conflicts

- If commits override changelog, check `priority` setting
- Ensure plugin is registered in `.sley.yaml`
- Verify no conflicts with custom inference logic
