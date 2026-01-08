# Changelog Generator Plugin

The changelog generator plugin automatically generates changelog entries from conventional commits after version bumps. It supports multiple git hosting providers (GitHub, GitLab, Codeberg, Bitbucket, etc.) and can output to versioned files, a unified CHANGELOG.md, or both.

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

## How It Works

1. After a successful version bump, retrieves commits since the previous version
2. Parses commits using conventional commit format
3. Groups commits by type using configurable patterns
4. Generates markdown content with links to commits, PRs, and version comparisons
5. Writes to versioned file (`.changes/vX.Y.Z.md`), unified CHANGELOG.md, or both

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned" # "versioned", "unified", or "both"
    format: "grouped" # "grouped" or "keepachangelog"
    changes-dir: ".changes" # Directory for versioned files
    changelog-path: "CHANGELOG.md" # Path for unified changelog
    # Custom header template (optional)
    # header-template: ".changes/header.md"
    repository:
      auto-detect: true # Auto-detect from git remote
    groups:
      - pattern: "^feat"
        label: "Enhancements"
      - pattern: "^fix"
        label: "Fixes"
      - pattern: "^docs?"
        label: "Documentation"
    exclude-patterns:
      - "^Merge"
      - "^WIP"
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

### Repository Configuration

The plugin supports multiple git hosting providers:

```yaml
repository:
  provider: "github" # github, gitlab, codeberg, gitea, bitbucket, custom
  host: "github.com" # Git server hostname
  owner: "myorg" # Repository owner/organization
  repo: "myproject" # Repository name
  auto-detect: true # Auto-detect from git remote (recommended)
```

When `auto-detect: true`, the plugin parses the git remote URL and automatically detects:

- GitHub (github.com)
- GitLab (gitlab.com)
- Codeberg (codeberg.org)
- Gitea (gitea.io)
- Bitbucket (bitbucket.org)
- SourceHut (sr.ht)
- Custom/self-hosted instances

### Format Configuration

The plugin supports two changelog formats:

#### Format: `grouped` (Default)

The default format groups commits by their configured labels and supports custom icons:

```yaml
format: "grouped"
```

Example output:

```markdown
## v1.2.0 - 2026-01-04

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))

### Fixes

- **parser:** Handle edge case ([ghi789](https://github.com/owner/repo/commit/ghi789))

### New Contributors

- [@newdev](https://github.com/newdev) made their first contribution in [#42](https://github.com/owner/repo/pull/42)

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Contributors

- [@alice](https://github.com/alice)
- [@newdev](https://github.com/newdev)
```

**Features**:

- Custom group labels via `groups` configuration
- Optional icons via `group-icons` or `groups[].icon`
- New Contributors section (detects first-time contributors)
- Full Changelog compare links between versions
- Commit and PR/MR links

#### Format: `keepachangelog`

Follows the [Keep a Changelog](https://keepachangelog.com) specification with standard sections:

```yaml
format: "keepachangelog"
```

Example output:

```markdown
## [1.2.0] - 2026-01-04

### Added

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))

### Fixed

- **parser:** Handle edge case ([ghi789](https://github.com/owner/repo/commit/ghi789))

### New Contributors

- [@newdev](https://github.com/newdev) made their first contribution in [#42](https://github.com/owner/repo/pull/42)

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Contributors

- [@alice](https://github.com/alice)
- [@newdev](https://github.com/newdev)
```

**Features**:

- Standard sections: Added, Changed, Deprecated, Removed, Fixed, Security, Breaking Changes
- Version header with brackets (no "v" prefix)
- New Contributors section (detects first-time contributors)
- Full Changelog compare links
- Commit and PR/MR links
- Custom group configuration is ignored

**Commit type mapping**:

| Conventional Commit Type               | Keep a Changelog Section |
| -------------------------------------- | ------------------------ |
| `feat`                                 | Added                    |
| `fix`                                  | Fixed                    |
| `refactor`, `perf`, `style`            | Changed                  |
| `revert`                               | Removed                  |
| `docs`, `test`, `chore`, `ci`, `build` | (skipped)                |
| Any type with `!` or `BREAKING CHANGE` | Breaking Changes         |

### Groups Configuration

**Note**: Group configuration only applies when using the `grouped` format. The `keepachangelog` format uses fixed standard sections and ignores custom groups.

There are three ways to configure commit groups (for `grouped` format):

#### Option 1: Use Default Icons (Simplest)

Enable `use-default-icons` to automatically apply predefined icons to all groups and contributors:

```yaml
plugins:
  changelog-generator:
    enabled: true
    use-default-icons: true
    contributors:
      enabled: true
```

This applies the following default icons:

| Group         | Icon |
| ------------- | ---- |
| Enhancements  | 🚀   |
| Fixes         | 🩹   |
| Refactors     | 💅   |
| Documentation | 📖   |
| Performance   | ⚡   |
| Styling       | 🎨   |
| Tests         | ✅   |
| Chores        | 🏡   |
| CI            | 🤖   |
| Build         | 📦   |
| Reverts       | ◀️   |

The default contributor icon is ❤️.

You can override specific icons while using defaults for the rest:

```yaml
plugins:
  changelog-generator:
    enabled: true
    use-default-icons: true
    group-icons:
      Enhancements: "✨" # Override just this one
    contributors:
      enabled: true
      icon: "⭐" # Override contributor icon
```

#### Option 2: Add Icons to Defaults

Use `group-icons` to manually add icons while keeping default patterns and labels:

```yaml
group-icons:
  Enhancements: "🚀"
  Fixes: "🩹"
  Refactors: "💅"
  Documentation: "📖"
  Performance: "⚡"
  Styling: "🎨"
  Tests: "✅"
  Chores: "🏡"
  CI: "🤖"
  Build: "📦"
  Reverts: "◀️"
```

Keys must match default labels exactly. You can specify only the icons you want.

**Note**: Consider using `use-default-icons: true` instead for a simpler configuration with the same result.

#### Option 3: Full Custom Groups

Use `groups` for complete control over patterns, labels, and order:

```yaml
groups:
  - pattern: "^feat"
    label: "New Features"
    icon: "🚀"
  - pattern: "^fix"
    label: "Bug Fixes"
    icon: "🐛"
  - pattern: "^docs?"
    label: "Documentation"
```

The `pattern` field uses Go regex syntax and matches against the commit type.
Order is derived from array position. When `groups` is specified, `group-icons` is ignored.

### Default Groups

If no groups are specified, the plugin uses these defaults:

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

### Exclude Patterns

Filter out unwanted commits using regex patterns:

```yaml
exclude-patterns:
  - "^Merge" # Merge commits
  - "^WIP" # Work in progress
  - "^wip" # Case variant
  - "^fixup!" # Fixup commits
  - "^squash!" # Squash commits
```

### Non-Conventional Commits

By default, commits that don't follow the conventional commit format (e.g., `Update README` instead of `docs: update README`) are skipped, and a warning is printed:

```
Warning: 2 non-conventional commit(s) skipped:
  - abc123: Update README
  - def456: Bump version
Tip: Use conventional commit format (type: description) or set 'include-non-conventional: true' in config.
```

To include these commits in an "Other Changes" section instead of skipping them:

```yaml
include-non-conventional: true
```

This adds a section at the end of the changelog:

```markdown
### Other Changes

- Update README ([abc123](https://github.com/owner/repo/commit/abc123))
- Bump version ([def456](https://github.com/owner/repo/commit/def456))
```

## Merging Versioned Files

If you've been using versioned mode and want to create a unified CHANGELOG.md, use the built-in merge command:

```bash
sley changelog merge
```

This command combines all versioned changelog files from the `.changes` directory into a single CHANGELOG.md file, sorted by version (newest first).

### Options

| Flag                | Default        | Description                          |
| ------------------- | -------------- | ------------------------------------ |
| `--changes-dir`     | `.changes`     | Directory containing versioned files |
| `--output`          | `CHANGELOG.md` | Output path for unified changelog    |
| `--header-template` | (built-in)     | Path to custom header template file  |

### Examples

Merge with default settings:

```bash
sley changelog merge
```

Merge with custom paths:

```bash
sley changelog merge --changes-dir .changes --output CHANGELOG.md
```

Merge with custom header:

```bash
sley changelog merge --header-template .changes/header.md
```

### Behavior

1. Reads all `.changes/v*.md` files
2. Sorts by version (newest first)
3. Prepends default or custom header
4. Writes to CHANGELOG.md

The command respects configuration from `.sley.yaml` but flags take precedence:

```yaml
plugins:
  changelog-generator:
    enabled: true
    changes-dir: ".changes"
    changelog-path: "CHANGELOG.md"
    header-template: ".changes/header.md"
```

## Alternative: Changie Integration

For teams that prefer [changie](https://changie.dev/), sley's versioned output is fully compatible with changie's merge workflow. Changie is a popular changelog management tool that provides additional features like interactive entry creation and advanced templating.

### Why Use Changie?

- Interactive changelog entry creation
- Advanced templating with custom formats
- Project-specific changelog workflows
- Built-in validation and linting
- Team collaboration features

### Setup

Install changie:

```bash
# macOS
brew install changie

# Or use go install
go install github.com/miniscruff/changie@latest
```

Initialize changie in your project:

```bash
changie init
```

### Configuration

Configure changie to work with sley's versioned files. Edit `.changie.yaml`:

```yaml
changesDir: .changes
outputPath: CHANGELOG.md
headerPath: .changes/header.md

# Configure to read sley's versioned files
kinds:
  - label: Enhancements
  - label: Fixes
  - label: Documentation
  - label: Refactors
  - label: Performance
  - label: Tests
  - label: Chores

# Custom format matching sley's output
versionFormat: '## {{.Version}} - {{.Time.Format "2006-01-02"}}'
kindFormat: "### {{.Kind}}"
changeFormat: "- {{.Body}}"
```

### Workflow

1. Configure sley to use versioned mode:

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
    changes-dir: ".changes"
```

2. Bump version with sley (generates `.changes/vX.Y.Z.md`):

```bash
sley bump patch
```

3. Merge changelog with changie:

```bash
changie merge
```

This creates or updates CHANGELOG.md with all versioned entries.

### When to Use Each Tool

Use **sley's built-in merge**:

- Quick one-command changelog management
- Prefer minimal tooling

Use **changie**:

- Complex projects with custom changelog formats
- Interactive changelog workflows
- Advanced templating requirements
- Team collaboration on changelog entries

## Output Modes

### Versioned Mode (Default)

Creates individual files for each version:

```
.changes/
  v1.0.0.md
  v1.1.0.md
  v1.2.0.md
```

Example `.changes/v1.2.0.md`:

```markdown
## v1.2.0 - 2026-01-03

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))
- **config:** Support multiple git providers ([def456](https://github.com/owner/repo/commit/def456)) ([#42](https://github.com/owner/repo/pull/42))

### Fixes

- **parser:** Handle edge case in commit parsing ([ghi789](https://github.com/owner/repo/commit/ghi789))

### New Contributors

- [@bob](https://github.com/bob) made their first contribution in [#42](https://github.com/owner/repo/pull/42)

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Contributors

- [@alice](https://github.com/alice)
- [@bob](https://github.com/bob)
```

### Unified Mode

Appends to a single CHANGELOG.md file, with new versions at the top:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## v1.2.0 - 2026-01-03

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))

### New Contributors

- [@newdev](https://github.com/newdev) made their first contribution in [#42](https://github.com/owner/repo/pull/42)

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Contributors

- [@alice](https://github.com/alice)
- [@newdev](https://github.com/newdev)

## v1.1.0 - 2025-12-15

...
```

### Both Mode

Writes to both versioned files and the unified changelog.

## Usage

Once enabled, the plugin works automatically with all bump commands.

### Basic Usage

```bash
sley bump patch
# Output: Version bumped from 1.2.3 to 1.2.4
# Creates: .changes/v1.2.4.md
```

### With Auto Bump

```bash
sley bump auto
# 1. Analyzes commits to determine bump type
# 2. Bumps version
# 3. Generates changelog entry
```

## Provider-Specific URLs

The plugin generates correct URLs for each provider:

### GitHub/Gitea/Codeberg

- Compare: `https://github.com/owner/repo/compare/v1.0.0...v1.1.0`
- Commit: `https://github.com/owner/repo/commit/abc123`
- PR: `https://github.com/owner/repo/pull/42`

### GitLab

- Compare: `https://gitlab.com/owner/repo/-/compare/v1.0.0...v1.1.0`
- Commit: `https://gitlab.com/owner/repo/-/commit/abc123`
- MR: `https://gitlab.com/owner/repo/-/merge_requests/42`

### Bitbucket

- Compare: `https://bitbucket.org/owner/repo/branches/compare/v1.1.0%0Dv1.0.0`
- Commit: `https://bitbucket.org/owner/repo/commits/abc123`
- PR: `https://bitbucket.org/owner/repo/pull-requests/42`

### SourceHut

- Compare: `https://git.sr.ht/owner/repo/log/v1.0.0..v1.1.0`
- Commit: `https://git.sr.ht/owner/repo/commit/abc123`

## Common Configurations

### GitHub Repository

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
    repository:
      auto-detect: true
```

### With Default Icons

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
    use-default-icons: true
    repository:
      auto-detect: true
    contributors:
      enabled: true
```

### GitLab Self-Hosted

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "unified"
    repository:
      provider: "gitlab"
      host: "gitlab.mycompany.com"
      owner: "team"
      repo: "project"
```

### Full Configuration

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "both"
    changes-dir: ".changes"
    changelog-path: "CHANGELOG.md"
    repository:
      auto-detect: true
    groups:
      - pattern: "^feat"
        label: "New Features"
        icon: ""
      - pattern: "^fix"
        label: "Bug Fixes"
        icon: ""
      - pattern: "^docs?"
        label: "Documentation"
      - pattern: "^refactor"
        label: "Code Refactoring"
      - pattern: "^perf"
        label: "Performance Improvements"
      - pattern: "^test"
        label: "Tests"
      - pattern: "^chore|^ci|^build"
        label: "Maintenance"
    exclude-patterns:
      - "^Merge"
      - "^WIP"
      - "^wip"
    contributors:
      enabled: true
```

### Contributors Configuration

The contributors section lists all unique contributors for a version. Additionally, the plugin can detect and highlight first-time contributors in a separate "New Contributors" section.

```yaml
contributors:
  enabled: true
  format: "- [@{{.Username}}](https://{{.Host}}/{{.Username}})" # default
  icon: "" # optional icon before "Contributors" header
  show-new-contributors: true # enable "New Contributors" section (default: true)
  new-contributors-format: "" # custom format for new contributor entries
  new-contributors-icon: "" # optional icon before "New Contributors" header
```

#### Configuration Options

| Option                    | Type   | Default | Description                             |
| ------------------------- | ------ | ------- | --------------------------------------- |
| `enabled`                 | bool   | true    | Enable/disable contributors section     |
| `format`                  | string | (link)  | Go template for contributor formatting  |
| `icon`                    | string | ""      | Icon before "Contributors" header       |
| `show-new-contributors`   | bool   | true    | Enable "New Contributors" section       |
| `new-contributors-format` | string | (auto)  | Go template for new contributor entries |
| `new-contributors-icon`   | string | ""      | Icon before "New Contributors" header   |

When `use-default-icons: true` is set, the default icons are:

- Contributors: ❤️
- New Contributors: 🎉

#### New Contributors Section

The plugin automatically detects first-time contributors by checking if a username has any commits before the previous version tag. New contributors are displayed with a link to their first PR:

```markdown
### 🎉 New Contributors

- [@newdev](https://github.com/newdev) made their first contribution in [#42](https://github.com/owner/repo/pull/42)
- [@another](https://github.com/another) made their first contribution in abc123
```

If no PR number is found in the commit message, the commit hash is shown instead.

#### New Contributors Format Template Variables

| Variable          | Description                                      |
| ----------------- | ------------------------------------------------ |
| `{{.Name}}`       | Full name from git (e.g., "Alice Smith")         |
| `{{.Username}}`   | Username extracted from email or derived         |
| `{{.Email}}`      | Email address                                    |
| `{{.Host}}`       | Git host for URL generation (e.g., "github.com") |
| `{{.PRNumber}}`   | PR number extracted from commit message          |
| `{{.CommitHash}}` | Short commit hash of first contribution          |

#### Contributors Format Template Variables

| Variable        | Description                                      |
| --------------- | ------------------------------------------------ |
| `{{.Name}}`     | Full name from git (e.g., "Alice Smith")         |
| `{{.Username}}` | Username extracted from email or derived         |
| `{{.Email}}`    | Email address                                    |
| `{{.Host}}`     | Git host for URL generation (e.g., "github.com") |

#### Format Examples

Username only (default):

```yaml
format: "- [@{{.Username}}](https://{{.Host}}/{{.Username}})"
# Output: - [@alice](https://github.com/alice)
```

Full name with username link:

```yaml
format: "- {{.Name}} ([@{{.Username}}](https://{{.Host}}/{{.Username}}))"
# Output: - Alice Smith ([@alice](https://github.com/alice))
```

Simple username without link:

```yaml
format: "- @{{.Username}}"
# Output: - @alice
```

#### Disabling New Contributors

To disable the new contributors section:

```yaml
contributors:
  enabled: true
  show-new-contributors: false
```

## Integration with Other Plugins

### With Tag Manager

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

Execution flow:

1. Version file updated
2. Changelog generated
3. Tag created and pushed

### With Commit Parser

```yaml
plugins:
  commit-parser: true
  changelog-generator:
    enabled: true
```

Workflow:

```bash
sley bump auto
# 1. commit-parser analyzes commits -> determines bump type
# 2. Version bumped
# 3. changelog-generator creates entry with same commits
```

### Full Release Workflow

```yaml
plugins:
  commit-parser: true
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
  changelog-generator:
    enabled: true
    mode: "both"
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

## Custom Header Template

Create a custom header file for changelogs. The header template is used when:

- Writing to unified CHANGELOG.md (mode: "unified" or "both")
- Merging versioned files with `sley changelog merge`

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned" # or "unified" or "both"
    header-template: ".changes/header.md"
```

Example `.changes/header.md`:

```markdown
# Changelog

All notable changes to MyProject are documented here.

This project adheres to [Semantic Versioning](https://semver.org/). The changelog is generated by [sley](https://github.com/indaco/sley).
```

## Best Practices

1. **Use versioned mode for larger projects**: Individual files are easier to review in PRs
2. **Enable auto-detect**: Let the plugin determine repository info from git remote
3. **Customize groups for your workflow**: Match your commit types to meaningful categories
4. **Exclude noise commits**: Filter merge commits and WIP entries
5. **Combine with tag-manager**: Create a complete release workflow

## Troubleshooting

### Changelog Not Generated

1. Verify plugin is enabled:

   ```yaml
   plugins:
     changelog-generator:
       enabled: true
   ```

2. Check there are commits since last version:
   ```bash
   git log v1.0.0..HEAD --oneline
   ```

### Links Not Working

1. Verify repository configuration:

   ```yaml
   repository:
     auto-detect: true
   ```

2. Check git remote is configured:
   ```bash
   git remote -v
   ```

### Wrong Grouping

1. Verify commit message format follows conventional commits:

   ```
   feat(scope): description
   fix: description
   ```

2. Check group patterns match your commit types

### Contributors Missing

Ensure contributors section is enabled:

```yaml
contributors:
  enabled: true
```

## Acknowledgments

This plugin took inspiration from:

- [changie](https://changie.dev/) - Automated changelog tool for preparing releases
- [git-cliff](https://git-cliff.org/) - Highly customizable changelog generator

## See Also

- [Example Configuration](./examples/changelog-generator.yaml) - Complete changelog-generator setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Commit Parser](./COMMIT_PARSER.md) - Automatic bump type detection
- [Tag Manager](./TAG_MANAGER.md) - Git tag automation
