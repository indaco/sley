# Commit Validator Extension

This extension validates that git commits follow the conventional commit format before version bumps. It ensures code quality and consistency in commit messages, especially useful with the `bump auto` command.

## Features

- Validates commits since last tag against conventional commit format
- Configurable allowed commit types
- Optional scope requirement
- Detailed error reporting for invalid commits
- Uses only Python standard library (no external dependencies)

## Installation

Install directly from the sley repository:

```bash
sley extension install --url github.com/indaco/sley/contrib/extensions/commit-validator
```

Or from a local clone:

```bash
sley extension install --path ./contrib/extensions/commit-validator
```

## Usage

Once installed and enabled, the extension runs automatically before version bumps:

```bash
sley bump patch
# Validates all commits since last tag before bumping
```

## Configuration

Add configuration to your `.sley.yaml`:

### Basic Configuration (Default)

Uses standard conventional commit types:

```yaml
extensions:
  - name: commit-validator
    enabled: true
    hooks:
      - pre-bump
```

Default allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

### Advanced Configuration

Customize allowed types and require scope:

```yaml
extensions:
  - name: commit-validator
    enabled: true
    hooks:
      - pre-bump
    config:
      allowed_types:
        - feat
        - fix
        - docs
        - chore
      require_scope: true # Require scope in all commits
```

### Configuration Options

- `allowed_types` (array): List of allowed commit types (default: all conventional commit types)
- `require_scope` (bool): If true, requires scope in commit messages (default: false)

## Conventional Commit Format

Valid commit message formats:

```
type: description
type(scope): description
```

Examples:

```
feat: add user authentication
fix(api): resolve null pointer exception
docs: update installation instructions
chore(deps): upgrade dependencies
```

### Commit Types

Standard conventional commit types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks
- `revert`: Reverting changes

## Examples

### Strict Validation

Only allow feature and fix commits:

```yaml
config:
  allowed_types:
    - feat
    - fix
```

### Require Scopes

Enforce scoped commits:

```yaml
config:
  require_scope: true
```

Valid:

```
feat(auth): add login endpoint
fix(ui): resolve button alignment
```

Invalid:

```
feat: add login endpoint  # Missing scope
```

### Custom Types

Use organization-specific types:

```yaml
config:
  allowed_types:
    - feature
    - bugfix
    - hotfix
    - release
```

## Error Handling

The extension provides detailed error messages for invalid commits:

```
Found 2 invalid commit(s):
  - Add new feature -> must match format 'type: description' or 'type(scope): description'
  - feat(users add endpoint -> must match format 'type: description' or 'type(scope): description'
```

## Validation Behavior

### Commits Checked

The extension validates commits between:

- Last git tag and HEAD (if tags exist)
- All commits (if no tags exist)

### Validation Skipped

No validation occurs when:

- No commits exist in the repository
- Repository is at the same commit as the last tag

### Validation Failure

The extension fails (blocks the bump) if:

- Any commit doesn't match conventional commit format
- Commit type is not in allowed types
- Scope is missing when required

## Hooks Supported

- `pre-bump`: Runs before version bump operations

## Requirements

- Python 3.6 or higher
- Git installed and available in PATH
- Initialized git repository with commits

## Integration with `bump auto`

This extension pairs well with the built-in `commitparser` plugin:

```yaml
# .sley.yaml
plugins:
  commit-parser: true

extensions:
  - name: commit-validator
    enabled: true
    hooks:
      - pre-bump
    config:
      allowed_types:
        - feat
        - fix
        - docs
```

Workflow:

1. Commit validator ensures all commits are properly formatted
2. Commit parser analyzes commits to determine bump type
3. Version is bumped automatically based on commit types

## JSON Input Format

The extension receives the following JSON on stdin:

```json
{
  "hook": "pre-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": null,
  "metadata": null,
  "project_root": "/path/to/project",
  "config": {
    "allowed_types": ["feat", "fix", "docs"],
    "require_scope": false
  }
}
```

## JSON Output Format

Success:

```json
{
  "success": true,
  "message": "All 5 commit(s) follow conventional commit format",
  "data": {
    "commits_checked": 5,
    "invalid_count": 0
  }
}
```

No commits:

```json
{
  "success": true,
  "message": "No commits to validate",
  "data": {
    "commits_checked": 0
  }
}
```

Validation errors:

```json
{
  "success": false,
  "message": "Found 2 invalid commit(s):\n  - Add feature -> must match format 'type: description'...",
  "data": {
    "commits_checked": 5,
    "invalid_count": 2,
    "invalid_commits": [
      {
        "message": "Add feature",
        "error": "must match format 'type: description' or 'type(scope): description'"
      }
    ]
  }
}
```

## Use Cases

- Enforcing commit message standards across teams
- Quality gates for automated version bumps
- Ensuring changelog generation accuracy
- Preparing for automated release processes
- Training teams on conventional commits

## Tips

### Gradual Adoption

Start with validation disabled, then enable:

```yaml
extensions:
  - name: commit-validator
    enabled: false # Start disabled
```

### Pre-commit Hooks

Combine with pre-commit hooks for immediate feedback:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: commit-msg
        name: Conventional Commit
        entry: check-commit-message
        language: system
```

### Training Commits

Use `--help` messages in commits during training:

```
# Good
feat: add user registration

# Bad (but educational)
Add user registration  # Would be rejected by validator
```

## Troubleshooting

### "Failed to retrieve git commits"

Ensure you are in a git repository:

```bash
git status
```

### "No commits to validate"

This is normal when:

- HEAD is at a tagged commit
- No new commits since last version

### Commits not matching

Check commit format carefully:

- Type must be lowercase
- Colon and space required after type/scope
- Scope (if used) must be in parentheses
