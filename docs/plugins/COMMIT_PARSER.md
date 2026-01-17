# Commit Parser Plugin

The commit parser plugin analyzes git commit messages following the Conventional Commits specification and automatically determines the appropriate version bump type.

## Plugin Metadata

| Field       | Value                                          |
| ----------- | ---------------------------------------------- |
| Name        | `commit-parser`                                |
| Type        | `analyzer`                                     |
| Description | Parses conventional commits to infer bump type |

## Status

Built-in, **enabled by default**

## Features

- Parses conventional commit messages automatically
- Determines bump type based on commit prefixes (`feat`, `fix`, etc.)
- Detects breaking changes via `!` suffix or `BREAKING CHANGE:` in body
- Integrates with `bump auto` command
- Supports scoped commits (`feat(api):`, `fix(auth):`)

## How It Works

1. Retrieves commits since the last git tag (or HEAD~10 if no tags exist)
2. Parses commit messages for conventional commit types
3. Returns the highest-priority bump type found

## Configuration

```yaml
plugins:
  commit-parser: true  # Enabled by default

  # To disable:
  commit-parser: false
```

## Usage

```bash
# Automatic bump based on conventional commits
sley bump auto

# Manual override
sley bump auto --label minor

# Disable plugin inference
sley bump auto --no-infer
```

### Example Workflow

```bash
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login timeout"
git commit -m "feat!: redesign API endpoints"

sley bump auto
# => Inferred bump type: major
# => Version bumped from 1.2.3 to 2.0.0
```

## Conventional Commit Format

```
type: description
type(scope): description
type!: description          # Breaking change
type(scope)!: description   # Breaking change with scope
```

### Examples

```
feat: add user dashboard
fix(api): handle null response
docs: update installation guide
feat!: redesign authentication flow
```

### Breaking Changes in Body

```
feat: update authentication flow

BREAKING CHANGE: Token format has changed from JWT to opaque tokens.
Users must re-authenticate after upgrading.
```

## Commit Type to Bump Mapping

| Type                     | Bump  | Description             |
| ------------------------ | ----- | ----------------------- |
| `feat`                   | minor | New feature             |
| `fix`                    | patch | Bug fix                 |
| `feat!` or `fix!`        | major | Breaking change         |
| Any + `BREAKING CHANGE:` | major | Breaking change in body |

Other types (`docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`) do not trigger version bumps.

## Bump Priority

When multiple commits exist, the highest-priority bump wins:

1. **major** (breaking changes)
2. **minor** (features)
3. **patch** (fixes)

```bash
git commit -m "fix: correct typo"        # patch
git commit -m "feat: add search"         # minor
sley bump auto                           # Result: minor
```

## Integration with Other Plugins

### With Changelog Parser

When both plugins are enabled, use the `priority` setting in changelog-parser to control precedence:

```yaml
plugins:
  commit-parser: true
  changelog-parser:
    enabled: true
    format: auto # keepachangelog, grouped, github, minimal, auto
    priority: "commits" # commit-parser takes precedence
```

- **priority: "commits"**: Commit parser runs first (recommended for CI/CD)
- **priority: "changelog"**: Changelog parser runs first, commit parser is fallback

### With Other Plugins

```yaml
plugins:
  commit-parser: true
  version-validator:
    enabled: true
    rules:
      - type: "branch-constraint"
        branch: "main"
        allowed: ["minor", "patch"]
  tag-manager:
    enabled: true
    prefix: "v"
```

Flow: commit-parser analyzes commits -> version-validator validates -> version updated -> tag-manager creates tag

## Best Practices

1. **Consistent commit messages** - Enforce with tools like `commitlint`
2. **Meaningful scopes** - Use scopes to categorize (`feat(api):`, `fix(ui):`)
3. **Clear breaking change indicators** - Use `!` suffix
4. **Detailed bodies** - Include context for major changes

## Troubleshooting

| Issue                    | Solution                                                 |
| ------------------------ | -------------------------------------------------------- |
| Plugin not detecting     | Ensure format: `type: description` (colon required)      |
| Wrong bump type inferred | Check `git log --oneline -10` for correct prefixes       |
| No bump type found       | `bump auto` defaults to patch if no conventional commits |

## Comparison with Changelog Parser

| Feature            | Commit Parser        | Changelog Parser                        |
| ------------------ | -------------------- | --------------------------------------- |
| Input source       | Git commit messages  | CHANGELOG.md                            |
| Format requirement | Conventional Commits | Multiple (keepachangelog, grouped, etc) |
| Manual control     | Low                  | High                                    |
| Automation level   | Fully automatic      | Semi-automatic                          |
| Best for           | CI/CD workflows      | Release planning                        |

## See Also

- [Example Configuration](./examples/commit-parser.yaml) - Commit-parser setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs from commits
- [Changelog Parser](./CHANGELOG_PARSER.md) - Alternative: infer bump from CHANGELOG.md (supports multiple formats)
