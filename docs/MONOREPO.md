# Monorepo / Multi-Module Support

Manage multiple `.version` files across a monorepo or multi-module project.

> [!NOTE]
> **Terminology**: This documentation uses "monorepo" and "multi-module" interchangeably. A **monorepo** is a repository containing multiple projects/services. A **module** in sley refers to any directory with its own `.version` file. The `workspace` term appears only in configuration contexts (`.sley.yaml`).

## Overview

When multiple services, packages, or modules exist in a single repository, each with its own `.version` file, `sley` can detect and operate on all of them automatically.

**Key features:**

- Automatic discovery of `.version` files in subdirectories
- Interactive TUI for selecting which modules to operate on
- Non-interactive flags for CI/CD pipelines
- Parallel execution for faster operations
- Multiple output formats (text, JSON, table)

## Quick Start

```bash
# Bump all modules
sley bump patch --all

# Show versions for all modules
sley show --all

# Bump specific module by name
sley bump patch --module api

# Interactive selection (without flags)
sley bump patch
```

### List Discovered Modules

```bash
sley modules list
# api     ./services/api/.version    1.2.3
# web     ./apps/web/.version        2.0.0
# shared  ./packages/shared/.version 0.5.1
```

## Detection Hierarchy

```
1. --path flag provided       -> Single-module mode (explicit path)
2. SLEY_PATH env set          -> Single-module mode (env path)
3. .version in current dir    -> Single-module mode (current dir)
4. Multiple .version found    -> Multi-module mode (discovery)
5. No .version file found     -> Error
```

## Module Discovery

### Automatic Discovery

`sley` recursively searches for `.version` files:

```
my-monorepo/
  services/
    api/.version        # Discovered as "api"
    auth/.version       # Discovered as "auth"
  packages/
    shared/.version     # Discovered as "shared"
```

The module name is derived from the parent directory name.

### Exclude Patterns

Create a `.sleyignore` file to exclude directories:

```
# .sleyignore
vendor/
node_modules/
testdata/
**/fixtures/
```

Default excluded patterns: `node_modules`, `.git`, `vendor`, `tmp`, `build`, `dist`, `.cache`, `__pycache__`

## Configuration

Configure discovery and modules in `.sley.yaml`:

```yaml
workspace:
  discovery:
    enabled: true
    recursive: true
    max_depth: 10
    exclude:
      - "testdata"
      - "examples"

  # Optional: explicit module definitions (overrides auto-discovery)
  modules:
    - name: api
      path: ./services/api/.version
      enabled: true
    - name: legacy
      path: ./legacy/.version
      enabled: false # Skip this module
```

### Config Inheritance

Module-specific `.sley.yaml` files can override workspace settings:

```yaml
# services/api/.sley.yaml
path: VERSION
plugins:
  commit-parser: false
```

## Interactive Mode

Without `--all` or `--module`, you get an interactive prompt:

```
Found 3 modules with .version files:
  - api (1.2.3)
  - web (2.0.0)
  - shared (0.5.1)

? How would you like to proceed?
  > Apply to all modules
    Select specific modules...
    Cancel
```

Use `--yes` to auto-select all modules: `sley bump patch --yes`

## Non-Interactive Mode (CI/CD)

```bash
# Operate on all modules
sley bump patch --all

# Operate on specific module by name
sley bump patch --module api

# Operate on multiple modules by name
sley bump patch --modules api,web,shared

# Operate on modules matching a pattern
sley bump patch --pattern "services/*"

# Execution control
sley bump patch --all --parallel          # Run in parallel
sley bump patch --all --continue-on-error # Don't stop on failures
sley bump patch --all --quiet             # Summary only
```

## Output Formats

```bash
sley show --all                  # Text (default)
sley show --all --format json    # JSON array
sley show --all --format table   # ASCII table
```

## CI/CD Integration

`sley` auto-detects CI environments (GitHub Actions, GitLab CI, CircleCI, etc.) and disables interactive prompts.

### GitHub Actions Example

```yaml
- name: Bump versions
  run: |
    if [ "${{ inputs.modules }}" = "all" ]; then
      sley bump ${{ inputs.bump_type }} --all
    else
      sley bump ${{ inputs.bump_type }} --modules ${{ inputs.modules }}
    fi
```

## Command Reference

| Flag                  | Short | Description                              |
| --------------------- | ----- | ---------------------------------------- |
| `--all`               | `-a`  | Operate on all discovered modules        |
| `--module`            | `-m`  | Operate on specific module by name       |
| `--modules`           |       | Operate on multiple modules (comma-sep)  |
| `--pattern`           |       | Operate on modules matching glob pattern |
| `--yes`               | `-y`  | Auto-select all without prompting        |
| `--non-interactive`   |       | Disable interactive prompts              |
| `--parallel`          |       | Execute operations in parallel           |
| `--fail-fast`         |       | Stop on first error (default)            |
| `--continue-on-error` |       | Continue even if some modules fail       |
| `--quiet`             | `-q`  | Suppress per-module output               |
| `--format`            |       | Output format: text, json, table         |

### Module Commands

```bash
sley modules list              # List all modules
sley modules list --verbose    # Detailed output
sley modules list --format json
sley modules discover          # Test discovery settings
```

## Troubleshooting

| Issue                        | Solution                                           |
| ---------------------------- | -------------------------------------------------- |
| No modules found             | Ensure `.version` files exist in subdirectories    |
| Module not detected          | Check `.sleyignore` and exclude patterns           |
| Interactive mode not working | Use `--all` or `--module` flags in CI/CD           |
| Permission denied            | Ensure `.version` files are writable (`chmod 644`) |

## See Also

- [README.md](../README.md) - Main documentation
- [Plugin System](PLUGINS.md) - Built-in plugins
