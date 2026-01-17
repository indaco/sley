# Changelog Parser Plugin

The changelog parser plugin parses CHANGELOG.md files to automatically infer version bump types and validate changelog completeness. It supports multiple changelog formats for symmetry with the changelog-generator plugin.

## Plugin Metadata

| Field       | Value                              |
| ----------- | ---------------------------------- |
| Name        | `changelog-parser`                 |
| Type        | `analyzer`                         |
| Description | Infers bump type from CHANGELOG.md |

## Status

Built-in, **disabled by default**

## Features

- **Multiple formats**: Supports `keepachangelog`, `grouped`, `github`, `minimal`, and `auto` detection
- Parses `## [Unreleased]` section to detect change types
- Infers bump type based on changelog subsections
- Validates that changelog has entries before allowing release
- Supports configurable priority over commit-based inference
- Works alongside the commit parser as fallback or primary parser
- Custom section mapping for grouped format

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
    format: "auto" # keepachangelog, grouped, github, minimal, auto
    require-unreleased-section: true
    infer-bump-type: true
    priority: "changelog" # or "commits"
```

### Configuration Options

| Option                       | Type   | Default          | Description                                                      |
| ---------------------------- | ------ | ---------------- | ---------------------------------------------------------------- |
| `enabled`                    | bool   | false            | Enable/disable the plugin                                        |
| `path`                       | string | `CHANGELOG.md`   | Path to the changelog file                                       |
| `format`                     | string | `keepachangelog` | Changelog format: keepachangelog, grouped, github, minimal, auto |
| `require-unreleased-section` | bool   | true             | Enforce presence of Unreleased section                           |
| `infer-bump-type`            | bool   | true             | Enable automatic bump type inference                             |
| `priority`                   | string | `changelog`      | Which parser takes precedence: "changelog" or "commits"          |
| `grouped-section-map`        | map    | (defaults)       | Custom section name to category mapping (grouped format only)    |

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

## Supported Formats

### Keep a Changelog Format (default)

The standard [Keep a Changelog](https://keepachangelog.com) format:

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

**Supported Subsection Headers:**

- `### Added` - New features (minor)
- `### Changed` - Changes in existing functionality (major)
- `### Deprecated` - Soon-to-be removed features (patch)
- `### Removed` - Removed features (major)
- `### Fixed` - Bug fixes (patch)
- `### Security` - Security fixes (patch)

All subsection headers are case-insensitive.

### Grouped Format

The grouped format uses configurable section names:

```markdown
## v1.2.0 - 2024-01-15

### Features

- **scope:** New feature description

### Bug Fixes

- Fixed something
```

**Default Section Mapping:**

| Section Name     | Category | Bump Type |
| ---------------- | -------- | --------- |
| Breaking Changes | Removed  | major     |
| Features         | Added    | minor     |
| Enhancements     | Added    | minor     |
| Bug Fixes        | Fixed    | patch     |
| Fixes            | Fixed    | patch     |
| Performance      | Changed  | major     |
| Refactors        | Changed  | -         |

**Custom Section Mapping:**

```yaml
plugins:
  changelog-parser:
    enabled: true
    format: grouped
    grouped-section-map:
      "New Features": "Added"
      "Bug Fixes": "Fixed"
      "Breaking": "Removed"
```

### GitHub Format

The GitHub release format with "What's Changed" section:

```markdown
## v1.2.0 - 2024-01-15

### Breaking Changes

- Something breaking by @user in #123

### What's Changed

- **scope:** description by @user in #123
```

> **Note:** The GitHub format has limited bump type inference. Only "Breaking Changes" can reliably infer `major`. Entries in "What's Changed" cannot distinguish between features and fixes, resulting in low confidence inference.

### Minimal Format

A condensed format with type prefixes:

```markdown
## v1.2.0

- [Feat] New feature description
- [Fix] Bug fix description
- [Breaking] Breaking change
```

**Type Prefix Mapping:**

| Prefix     | Category | Bump Type |
| ---------- | -------- | --------- |
| [Feat]     | Added    | minor     |
| [Fix]      | Fixed    | patch     |
| [Breaking] | Removed  | major     |
| [Perf]     | Changed  | major     |
| [Revert]   | Removed  | patch     |

### Auto-Detection

Use `format: auto` to let the parser detect the format automatically:

```yaml
plugins:
  changelog-parser:
    enabled: true
    format: auto
```

Detection rules:

- `## [Unreleased]` or `## [1.0.0]` → keepachangelog
- `- [Feat]` or `- [Fix]` entries → minimal
- `### What's Changed` section → github
- `## v1.0.0` without brackets → grouped

## Comparison with Commit Parser

| Feature            | Changelog Parser                        | Commit Parser        |
| ------------------ | --------------------------------------- | -------------------- |
| Input source       | CHANGELOG.md                            | Git commit messages  |
| Format requirement | Multiple (keepachangelog, grouped, etc) | Conventional Commits |
| Manual control     | High                                    | Low                  |
| Automation level   | Semi-automatic                          | Fully automatic      |
| Best for           | Release planning                        | CI/CD workflows      |

## Best Practices

1. **Keep Unreleased Section Updated** - Update as you develop
2. **Use Clear Subsections** - Categorize changes appropriately
3. **Combine with Commits** - Use "commits" priority for CI, "changelog" for manual releases
4. **Validate Before Release** - Enable `require-unreleased-section`

## Troubleshooting

| Issue                       | Solution                                                           |
| --------------------------- | ------------------------------------------------------------------ |
| Plugin not inferring type   | Check `enabled: true`, `infer-bump-type: true`, `priority`         |
| Validation errors           | Verify `## [Unreleased]` exists with `###` subsections             |
| Priority conflicts          | Ensure plugin is in `.sley.yaml` and check `priority` setting      |
| Wrong format detected       | Set explicit `format` instead of `auto`                            |
| GitHub format low inference | Expected - use keepachangelog or grouped for reliable inference    |
| Custom sections not working | Ensure `grouped-section-map` keys match your section names exactly |

## See Also

- [Example Configuration](./examples/changelog-parser.yaml) - Complete changelog-parser setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Commit Parser](./COMMIT_PARSER.md) - Alternative: infer bump from git commits
