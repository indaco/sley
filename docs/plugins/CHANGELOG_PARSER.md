# Changelog Parser Plugin

The changelog parser plugin parses CHANGELOG.md files in [Keep a Changelog](https://keepachangelog.com) format to automatically infer version bump types and validate changelog completeness.

## Plugin Metadata

| Field       | Value                              |
| ----------- | ---------------------------------- |
| Name        | `changelog-parser`                 |
| Version     | v0.1.0                             |
| Type        | `changelog-parser`                 |
| Description | Infers bump type from CHANGELOG.md |

## Status

Built-in, **disabled by default**

## Features

- Parses `## [Unreleased]` section to detect change types
- Infers bump type based on changelog subsections
- Validates that changelog has entries before allowing release
- Supports configurable priority over commit-based inference
- Works alongside the commit parser as fallback or primary parser

## Bump Type Inference Rules

| Changelog Section | Bump Type | Rationale                      |
| ----------------- | --------- | ------------------------------ |
| Removed           | major     | Breaking change (removal)      |
| Changed           | major     | Breaking change (modification) |
| Added             | minor     | New feature                    |
| Fixed             | patch     | Bug fix                        |
| Security          | patch     | Security fix                   |
| Deprecated        | patch     | Deprecation notice             |

Priority: major (Removed/Changed) > minor (Added) > patch (Fixed/Security/Deprecated)

## Configuration

Enable and configure in `.sley.yaml`. See [changelog-parser.yaml](./examples/changelog-parser.yaml) for complete examples.

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

| Option                       | Type   | Default        | Description                                             |
| ---------------------------- | ------ | -------------- | ------------------------------------------------------- |
| `enabled`                    | bool   | false          | Enable/disable the plugin                               |
| `path`                       | string | `CHANGELOG.md` | Path to the changelog file                              |
| `require-unreleased-section` | bool   | true           | Enforce presence of Unreleased section                  |
| `infer-bump-type`            | bool   | true           | Enable automatic bump type inference                    |
| `priority`                   | string | `changelog`    | Which parser takes precedence: "changelog" or "commits" |

## Usage with `bump auto`

```bash
# Edit your CHANGELOG.md
cat CHANGELOG.md
# ## [Unreleased]
# ### Added
# - New feature X
# ### Fixed
# - Bug fix Y

# Run bump auto - detects "minor" from "Added" section
sley bump auto
# Inferred from changelog: minor
# Bumped version from 1.2.3 to 1.3.0
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

```yaml
plugins:
  changelog-parser:
    enabled: true
  tag-manager:
    enabled: true
    auto-create: true
```

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

## Comparison with Commit Parser

| Feature            | Changelog Parser | Commit Parser        |
| ------------------ | ---------------- | -------------------- |
| Input source       | CHANGELOG.md     | Git commit messages  |
| Format requirement | Keep a Changelog | Conventional Commits |
| Manual control     | High             | Low                  |
| Automation level   | Semi-automatic   | Fully automatic      |
| Best for           | Release planning | CI/CD workflows      |

## Best Practices

1. **Keep Unreleased Section Updated** - Update as you develop
2. **Use Clear Subsections** - Categorize changes appropriately
3. **Combine with Commits** - Use "commits" priority for CI, "changelog" for manual releases
4. **Validate Before Release** - Enable `require-unreleased-section`

## Troubleshooting

| Issue                     | Solution                                                      |
| ------------------------- | ------------------------------------------------------------- |
| Plugin not inferring type | Check `enabled: true`, `infer-bump-type: true`, `priority`    |
| Validation errors         | Verify `## [Unreleased]` exists with `###` subsections        |
| Priority conflicts        | Ensure plugin is in `.sley.yaml` and check `priority` setting |

## See Also

- [Example Configuration](./examples/changelog-parser.yaml) - Complete changelog-parser setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Commit Parser](./COMMIT_PARSER.md) - Alternative: infer bump from git commits
