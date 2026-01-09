# Extension System

Extensions are executable scripts that run at specific hook points during version management operations. They receive JSON input on stdin and return JSON output on stdout.

## When to Use Extensions vs Plugins

### Use Extensions When

- **Custom to your workflow**: Organization-specific automation or processes
- **Requires external tools**: Need to call AWS CLI, curl, custom scripts, etc.
- **Prototyping new features**: Testing ideas before proposing as built-in plugins
- **Language-specific needs**: Python/Node.js/Ruby tooling integration
- **Examples**: Custom notification systems, deployment triggers, proprietary tool integration

### Use Plugins When

- **Performance matters**: Plugins execute in <1ms with native Go performance
- **Feature is widely applicable**: Common versioning needs across many projects
- **Deep integration needed**: Requires tight coupling with bump logic or validation
- **Built-in reliability required**: No external dependencies or installation steps
- **Examples**: Git tagging, conventional commit parsing, version validation, file syncing

> [!NOTE]
> Most users will only need plugins (see [PLUGINS.md](PLUGINS.md)). Extensions are for advanced customization when built-in plugins don't meet your specific needs.

## Hook Points

| Hook          | When                    | Use Cases                                      |
| ------------- | ----------------------- | ---------------------------------------------- |
| `pre-bump`    | Before version bump     | Validate preconditions, run linters/tests      |
| `post-bump`   | After successful bump   | Update files, create tags, send notifications  |
| `pre-release` | Before pre-release      | Validate pre-release labels, check readiness   |
| `validate`    | Custom validation       | Enforce policies, validate version format      |

### Input Context

All hooks receive JSON on stdin:

```json
{
  "hook": "post-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": "alpha",
  "metadata": "build123",
  "project_root": "/path/to/project",
  "module_dir": "./services/api",
  "module_name": "api"
}
```

Note: `module_dir` and `module_name` are included in monorepo contexts.

## Creating an Extension

### Directory Structure

```
my-extension/
  extension.yaml    # Manifest (required)
  hook.sh           # Entry point script
  README.md         # Documentation (recommended)
```

### Extension Manifest

```yaml
# extension.yaml
name: my-extension
version: 1.0.0
description: Brief description
author: Your Name
repository: https://github.com/username/my-extension
entry: hook.sh
hooks:
  - pre-bump
  - post-bump
```

### Hook Script (Shell)

```bash
#!/bin/sh

# Read JSON input
read -r input

# Parse fields
version=$(echo "$input" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
bump_type=$(echo "$input" | grep -o '"bump_type":"[^"]*"' | cut -d'"' -f4)

# Custom logic here
echo "Processing version $version" >&2

# Return success
echo '{"success": true, "message": "Extension executed"}'
exit 0
```

See `contrib/extensions/` for Python and Node.js examples.

### JSON Output Format

```json
{
  "success": true,
  "message": "Optional status message",
  "data": { "key": "Optional data to return" }
}
```

On failure, return `success: false`:

```json
{
  "success": false,
  "message": "Extension failed: reason"
}
```

## Installing Extensions

```bash
# From local path
sley extension install /path/to/my-extension
```

This copies the extension to `~/.sley-extensions/my-extension/` and adds it to `.sley.yaml`.

### Configuration

```yaml
# .sley.yaml
extensions:
  - name: changelog-generator
    path: /Users/username/.sley-extensions/changelog-generator
    enabled: true
```

## Managing Extensions

```bash
sley extension list              # List installed extensions
sley extension remove NAME       # Remove an extension
```

Disable by setting `enabled: false` in `.sley.yaml`.

## Running Extensions

Extensions run automatically during bump commands:

```bash
sley bump patch
# Pre-bump extensions run
# Version bumped
# Post-bump extensions run
```

Skip extensions with `--skip-hooks`:

```bash
sley bump patch --skip-hooks
```

## Security Considerations

- Extensions run with your user permissions - only install from trusted sources
- 30-second execution timeout (prevents hanging)
- 1MB output limit (prevents memory exhaustion)
- Each extension runs as separate process (isolation)

## Best Practices

1. Keep extensions focused on one task
2. Validate input before processing
3. Handle errors gracefully with meaningful messages
4. Test with different inputs
5. Use stderr for debug output

## Example Extensions

See `contrib/extensions/` for reference implementations:

- **commit-validator** - Validates conventional commit format
- **docker-tag-sync** - Syncs version to Docker image tags
- **git-tagger** - Creates git tags on version bumps
- **package-sync** - Syncs version across package manifests
- **version-policy** - Enforces custom versioning policies

## Troubleshooting

| Issue                 | Solution                                                  |
| --------------------- | --------------------------------------------------------- |
| Extension not found   | Check `sley extension list` and `.sley.yaml`              |
| Not executing         | Verify `enabled: true`, script executable, proper shebang |
| Permission denied     | Run `chmod +x hook.sh`                                    |
| Timeout errors        | Optimize script or split into smaller tasks               |
| Invalid JSON output   | Test manually with `echo '{"hook":"post-bump"}' \| ./hook.sh` |

## Error Propagation

If an extension fails (`success: false`), the bump operation is aborted and subsequent extensions are not executed.

## See Also

- [Plugin System](PLUGINS.md) - Built-in plugins (compiled into CLI)
- [Example Extensions](../contrib/extensions/) - Reference implementations
