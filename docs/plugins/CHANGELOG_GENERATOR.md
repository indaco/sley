# Changelog Generator Plugin

The changelog generator plugin automatically generates changelog entries from conventional commits after version bumps. It supports multiple git hosting providers (GitHub, GitLab, Codeberg, Bitbucket, etc.) and can output to versioned files, a unified CHANGELOG.md, or both.

## Plugin Metadata

| Field       | Value                                |
| ----------- | ------------------------------------ |
| Name        | `changelog-generator`                |
| Version     | v0.1.0                               |
| Type        | `changelog-generator`                |
| Description | Generates changelog from git commits |

## Status

Built-in, **disabled by default**

## Features

- Automatic changelog generation from conventional commits
- Multiple changelog formats: grouped (default) or Keep a Changelog
- Multiple output modes: versioned files, unified CHANGELOG.md, or both
- Commit grouping by type (feat, fix, docs, etc.) with customizable labels
- GitHub, GitLab, Codeberg, Bitbucket, and custom git hosting support
- Full Changelog compare links between versions
- Commit and PR/MR links
- New Contributors section (first-time contributors detection)
- Contributors section
- Configurable exclude patterns for filtering commits
- Optional icons/emojis per commit group (grouped format only)

## Why Another Changelog Generator?

Excellent standalone tools like [changie](https://changie.dev/) and [git-cliff](https://git-cliff.org/) already exist. The changelog-generator plugin exists because sley aims to be a **unified versioning tool** - one tool that handles `.version` files across any language or stack, with changelog generation as part of that workflow.

- **One tool, one workflow**: `sley bump patch` handles version update, changelog, and tag in sequence
- **Shared configuration**: Everything lives in `.sley.yaml`
- **Plugin coordination**: Works with `commit-parser` and `tag-manager` in a defined execution order

For teams using versioned output mode, the generated `.changes/vX.Y.Z.md` files remain compatible with changie's merge workflow.

This plugin isn't trying to match the flexibility of dedicated changelog tools - it's providing a good-enough solution for projects that want everything in one place.

## How It Works

1. After a successful version bump, retrieves commits since the previous version
2. Parses commits using conventional commit format
3. Groups commits by type using configurable patterns
4. Generates markdown content with links to commits, PRs, and version comparisons
5. Writes to versioned file (`.changes/vX.Y.Z.md`), unified CHANGELOG.md, or both

## PR Links in Changelog

The changelog is generated from **commit messages**I. For PR numbers to appear in the changelog, they must be present in the commit message itself (format: `#123` or `(#123)`).

**Option 1: Use Squash and Merge (Recommended)**

GitHub's squash merge automatically appends `(#123)` to the commit message, which the parser will detect.

**Option 2: Include PR Numbers Manually**

Add the PR number to your commit messages:

```
feat(api): add new endpoint (#123)
fix: resolve timeout issue (#456)
```

**Option 3: Rebase and Merge**

Include the PR number in your original commit messages before merging.

## Usage

Once enabled, the plugin works automatically with all bump commands:

```bash
sley bump patch
# => Version bumped from 1.2.3 to 1.2.4
# => Creates: .changes/v1.2.4.md

sley bump auto
# Analyzes commits, bumps version, generates changelog
```

## Configuration

Enable and configure in `.sley.yaml`. See [changelog-generator.yaml](./examples/changelog-generator.yaml) for a complete example.

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
    format: "grouped"
    repository:
      auto-detect: true
    contributors:
      enabled: true
```

### Configuration Options

| Option                     | Type   | Default          | Description                                                          |
| -------------------------- | ------ | ---------------- | -------------------------------------------------------------------- |
| `enabled`                  | bool   | false            | Enable/disable the plugin                                            |
| `mode`                     | string | `"versioned"`    | Output mode: versioned, unified, or both                             |
| `format`                   | string | `"grouped"`      | Changelog format: "grouped" or "keepachangelog"                      |
| `changes-dir`              | string | `".changes"`     | Directory for versioned changelog files                              |
| `changelog-path`           | string | `"CHANGELOG.md"` | Path to unified changelog file                                       |
| `header-template`          | string | (built-in)       | Path to custom header template                                       |
| `repository`               | object | auto-detect      | Git repository configuration for links                               |
| `groups`                   | array  | (defaults)       | Full custom commit grouping rules (ignored in keepachangelog format) |
| `use-default-icons`        | bool   | false            | Enable predefined icons for all groups and contributors              |
| `group-icons`              | map    | (none)           | Add icons to default groups by label (ignored in keepachangelog)     |
| `exclude-patterns`         | array  | (defaults)       | Regex patterns for commits to exclude                                |
| `include-non-conventional` | bool   | false            | Include non-conventional commits in "Other Changes"                  |
| `contributors`             | object | enabled          | Contributors section configuration                                   |

### Output Modes

The `mode` option controls where changelog entries are written:

- **versioned** (default): Creates `.changes/v{version}.md` files for each version
- **unified**: Appends to a single CHANGELOG.md file (newest first)
- **both**: Writes to both versioned files and unified changelog

Example versioned output (`.changes/v1.2.0.md`):

```markdown
## v1.2.0 - 2026-01-03

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))

### Fixes

- **parser:** Handle edge case ([ghi789](https://github.com/owner/repo/commit/ghi789))

### New Contributors

- [@bob](https://github.com/bob) made their first contribution in [#42](https://github.com/owner/repo/pull/42)

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Contributors

- [@alice](https://github.com/alice)
- [@bob](https://github.com/bob)
```

### Repository Configuration

The plugin auto-detects repository info from git remote (recommended). Supported providers: GitHub, GitLab, Codeberg, Gitea, Bitbucket, SourceHut, and custom/self-hosted instances.

```yaml
repository:
  auto-detect: true # recommended
  # Or specify manually:
  # provider: "gitlab"
  # host: "gitlab.mycompany.com"
  # owner: "team"
  # repo: "project"
```

### Format Configuration

#### Format: `grouped` (Default)

Groups commits by configured labels with optional icons. Supports custom group labels via `groups` configuration.

#### Format: `keepachangelog`

Follows the [Keep a Changelog](https://keepachangelog.com) specification with standard sections. Custom group configuration is ignored.

| Conventional Commit Type               | Keep a Changelog Section |
| -------------------------------------- | ------------------------ |
| `feat`                                 | Added                    |
| `fix`                                  | Fixed                    |
| `refactor`, `perf`, `style`            | Changed                  |
| `revert`                               | Removed                  |
| `docs`, `test`, `chore`, `ci`, `build` | (skipped)                |
| Any type with `!` or `BREAKING CHANGE` | Breaking Changes         |

### Groups Configuration

Group configuration only applies to the `grouped` format. Three configuration approaches:

**Option 1: Default Icons (Simplest)**

```yaml
use-default-icons: true
```

Applies predefined icons: Enhancements (🚀), Fixes (🩹), Refactors (💅), Documentation (📖), Performance (⚡), Styling (🎨), Tests (✅), Chores (🏡), CI (🤖), Build (📦), Reverts (◀️). Contributor icon: ❤️.

**Option 2: Add Icons to Defaults**

```yaml
group-icons:
  Enhancements: "🚀"
  Fixes: "🩹"
```

Keys must match default labels exactly.

**Option 3: Full Custom Groups**

```yaml
groups:
  - pattern: "^feat"
    label: "New Features"
    icon: "🚀"
  - pattern: "^fix"
    label: "Bug Fixes"
    icon: "🐛"
```

The `pattern` field uses Go regex syntax. Order is derived from array position.

#### Default Groups

| Pattern     | Label         |
| ----------- | ------------- |
| `^feat`     | Enhancements  |
| `^fix`      | Fixes         |
| `^refactor` | Refactors     |
| `^docs?`    | Documentation |
| `^perf`     | Performance   |
| `^style`    | Styling       |
| `^test`     | Tests         |
| `^chore`    | Chores        |
| `^ci`       | CI            |
| `^build`    | Build         |
| `^revert`   | Reverts       |

### Contributors Configuration

Lists unique contributors per version and detects first-time contributors.

| Option                    | Type   | Default | Description                             |
| ------------------------- | ------ | ------- | --------------------------------------- |
| `enabled`                 | bool   | true    | Enable/disable contributors section     |
| `format`                  | string | (link)  | Go template for contributor formatting  |
| `icon`                    | string | ""      | Icon before "Contributors" header       |
| `show-new-contributors`   | bool   | true    | Enable "New Contributors" section       |
| `new-contributors-format` | string | (auto)  | Go template for new contributor entries |
| `new-contributors-icon`   | string | ""      | Icon before "New Contributors" header   |

Template variables: `{{.Name}}`, `{{.Username}}`, `{{.Email}}`, `{{.Host}}`. New contributors also have `{{.PRNumber}}` and `{{.CommitHash}}`.

### Exclude Patterns

Filter unwanted commits using regex patterns:

```yaml
exclude-patterns:
  - "^Merge"
  - "^WIP"
  - "^fixup!"
  - "^squash!"
```

### Non-Conventional Commits

By default, non-conventional commits are skipped with a warning. Set `include-non-conventional: true` to include them in an "Other Changes" section.

### Custom Header Template

For unified changelogs, specify a custom header file:

```yaml
header-template: ".changes/header.md"
```

## Working with Changelogs

### Merging Versioned Files

Combine versioned files into a unified CHANGELOG.md:

```bash
sley changelog merge
sley changelog merge --changes-dir .changes --output CHANGELOG.md
sley changelog merge --header-template .changes/header.md
```

| Flag                | Default        | Description                          |
| ------------------- | -------------- | ------------------------------------ |
| `--changes-dir`     | `.changes`     | Directory containing versioned files |
| `--output`          | `CHANGELOG.md` | Output path for unified changelog    |
| `--header-template` | (built-in)     | Path to custom header template file  |

### Changie Integration

For teams preferring [changie](https://changie.dev/), sley's versioned output is compatible with changie's merge workflow:

1. Configure sley with `mode: "versioned"`
2. Bump version with sley (generates `.changes/vX.Y.Z.md`)
3. Merge changelog with `changie merge`

Use sley's built-in merge for minimal tooling; use changie for advanced templating and team collaboration.

## Provider-Specific URLs

The plugin generates correct URLs for each provider:

| Provider              | Compare URL Pattern                 | PR/MR Term |
| --------------------- | ----------------------------------- | ---------- |
| GitHub/Gitea/Codeberg | `/compare/v1.0.0...v1.1.0`          | PR         |
| GitLab                | `/-/compare/v1.0.0...v1.1.0`        | MR         |
| Bitbucket             | `/branches/compare/v1.1.0%0Dv1.0.0` | PR         |
| SourceHut             | `/log/v1.0.0..v1.1.0`               | -          |

## Integration with Other Plugins

### With Tag Manager

```yaml
plugins:
  changelog-generator:
    enabled: true
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

Flow: Version updated -> Changelog generated -> Tag created and pushed.

### With Commit Parser

```yaml
plugins:
  commit-parser: true
  changelog-generator:
    enabled: true
```

Run `sley bump auto` to analyze commits, determine bump type, and generate changelog.

## Examples

For complete configuration examples, see:

- [changelog-generator.yaml](./examples/changelog-generator.yaml) - Full plugin configuration with all options
- [full-config.yaml](./examples/full-config.yaml) - All plugins working together

### Quick Start Configurations

**Minimal (GitHub with auto-detect):**

```yaml
plugins:
  changelog-generator:
    enabled: true
```

**With icons:**

```yaml
plugins:
  changelog-generator:
    enabled: true
    use-default-icons: true
```

**GitLab self-hosted:**

```yaml
plugins:
  changelog-generator:
    enabled: true
    repository:
      provider: "gitlab"
      host: "gitlab.mycompany.com"
      owner: "team"
      repo: "project"
```

## Best Practices

1. **Use versioned mode for larger projects** - Individual files are easier to review in PRs
2. **Enable auto-detect** - Let the plugin determine repository info from git remote
3. **Customize groups for your workflow** - Match commit types to meaningful categories
4. **Exclude noise commits** - Filter merge commits and WIP entries
5. **Combine with tag-manager** - Create a complete release workflow

## Troubleshooting

| Issue                   | Solution                                                      |
| ----------------------- | ------------------------------------------------------------- |
| Changelog not generated | Verify `enabled: true` and commits exist since last version   |
| Links not working       | Check `repository.auto-detect: true` and `git remote -v`      |
| Wrong grouping          | Verify conventional commit format: `type(scope): description` |
| Contributors missing    | Ensure `contributors.enabled: true`                           |

## Acknowledgments

Inspired by [changie](https://changie.dev/) and [git-cliff](https://git-cliff.org/).

## See Also

- [Example Configuration](./examples/changelog-generator.yaml) - Complete changelog-generator setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Commit Parser](./COMMIT_PARSER.md) - Automatic bump type detection
- [Tag Manager](./TAG_MANAGER.md) - Git tag automation
