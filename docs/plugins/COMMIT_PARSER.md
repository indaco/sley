# Commit Parser Plugin

The commit parser plugin analyzes git commit messages following the Conventional Commits specification and automatically determines the appropriate version bump type.

## Plugin Metadata

| Field       | Value                                          |
| ----------- | ---------------------------------------------- |
| Name        | `commit-parser`                                |
| Version     | v0.1.0                                         |
| Type        | `commit-parser`                                |
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
2. Parses commit messages for conventional commit types:
   - `feat:` or `feat!:` -> minor bump (major if breaking)
   - `fix:` or `fix!:` -> patch bump (major if breaking)
   - `BREAKING CHANGE:` in commit body -> major bump
3. Returns the highest-priority bump type found

## Configuration

Enable/disable in `.sley.yaml`:

```yaml
plugins:
  commit-parser: true # Enabled by default
```

To disable:

```yaml
plugins:
  commit-parser: false
```

## Usage

### With `bump auto`

The plugin integrates with the `bump auto` command:

```bash
# Automatic bump based on conventional commits
sley bump auto

# Manual override with --label
sley bump auto --label minor

# Disable plugin inference
sley bump auto --no-infer
```

### Example Workflow

```bash
# Make commits following conventional format
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login timeout"
git commit -m "feat!: redesign API endpoints"

# Plugin analyzes commits and determines major bump
sley bump auto
# Output: Inferred bump type: major
# Version bumped from 1.2.3 to 2.0.0
```

## Conventional Commit Format

### Valid Message Formats

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
fix(auth)!: change token format
```

### Breaking Changes in Body

You can also indicate breaking changes in the commit body:

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

Other conventional commit types (`docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`) do not trigger version bumps by default.

## Bump Priority

When multiple commits exist, the highest-priority bump wins:

1. **major** (breaking changes)
2. **minor** (features)
3. **patch** (fixes)

Example:

```bash
git commit -m "fix: correct typo"        # patch
git commit -m "feat: add search"         # minor
git commit -m "fix: handle edge case"    # patch

sley bump auto
# Result: minor (highest priority)
```

## Disabling Inference

When you need manual control over bump types:

### Via Configuration

```yaml
# .sley.yaml
plugins:
  commit-parser: false
```

### Via Command Flags

```bash
# Always bumps patch (ignores commits)
sley bump auto --no-infer

# Manual override (commits still analyzed but overridden)
sley bump auto --label minor
```

## Integration with Other Plugins

The commit parser works alongside other plugins:

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

Execution flow:

1. `commit-parser` analyzes commits -> determines bump type
2. `version-validator` validates the bump is allowed
3. Version file updated
4. `tag-manager` creates git tag

## Best Practices

1. **Consistent commit messages**: Enforce conventional commits in your team with tools like `commitlint`
2. **Meaningful scopes**: Use scopes to categorize changes (`feat(api):`, `fix(ui):`)
3. **Clear breaking change indicators**: Use `!` suffix for breaking changes
4. **Detailed bodies**: Include context in commit body for major changes

## Troubleshooting

### Plugin Not Detecting Commits

Ensure commits follow the conventional format exactly:

```bash
# Correct
feat: add feature

# Incorrect (missing colon)
feat add feature

# Incorrect (extra space)
feat : add feature
```

### Wrong Bump Type Inferred

Check your commit history:

```bash
git log --oneline -10
```

Verify commits use correct prefixes:

- `feat` for features (minor)
- `fix` for bug fixes (patch)
- `!` suffix for breaking changes (major)

### No Bump Type Found

If no conventional commits are found, `bump auto` defaults to patch:

```bash
sley bump auto
# Output: No conventional commits found, defaulting to patch
```

## See Also

- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs from commits
- [Changelog Parser](./CHANGELOG_PARSER.md) - Alternative: infer bump from CHANGELOG.md
