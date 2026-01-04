# Extension System

The verso extension system allows you to extend the functionality of the CLI tool by writing custom scripts that execute at specific hook points during version management operations.

## Overview

Extensions are executable scripts (shell scripts, Python, Node.js, etc.) that:

1. Receive JSON input on stdin
2. Perform custom operations
3. Return JSON output on stdout

Extensions are installed locally and configured in `.verso.yaml`.

## Hook Points

The extension system supports the following hook points:

### `pre-bump`

Called **before** a version bump operation is applied.

**Use cases:**

- Validate preconditions before bumping
- Run linters or tests
- Check for uncommitted changes

**Input context:**

```json
{
  "hook": "pre-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "project_root": "/path/to/project"
}
```

### `post-bump`

Called **after** a version bump operation completes successfully.

**Use cases:**

- Update CHANGELOG.md
- Create git tags
- Send notifications
- Update documentation

**Input context:**

```json
{
  "hook": "post-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": "alpha",
  "metadata": "build123",
  "project_root": "/path/to/project"
}
```

### `pre-release`

Called before applying pre-release changes.

**Use cases:**

- Validate pre-release labels
- Check release readiness

**Input context:**

```json
{
  "hook": "pre-release",
  "version": "1.2.3-alpha.1",
  "project_root": "/path/to/project"
}
```

### `validate`

Called to perform custom validation on version changes.

**Use cases:**

- Enforce versioning policies
- Validate version format
- Check compatibility requirements

**Input context:**

```json
{
  "hook": "validate",
  "version": "1.2.3",
  "project_root": "/path/to/project"
}
```

## Creating an Extension

### Directory Structure

```
my-extension/
├── extension.yaml    # Extension manifest (required)
├── hook.sh          # Entry point script (can be any executable)
└── README.md        # Documentation (recommended)
```

### Extension Manifest

The `extension.yaml` file defines the extension metadata:

```yaml
name: my-extension
version: 1.0.0
description: Brief description of what the extension does
author: Your Name
repository: https://github.com/username/my-extension
entry: hook.sh
hooks:
  - pre-bump
  - post-bump
```

**Required fields:**

- `name`: Unique extension identifier
- `version`: Extension version (SemVer format)
- `description`: Brief description
- `author`: Extension author
- `repository`: Source repository URL
- `entry`: Path to executable script (relative to extension directory)

**Optional fields:**

- `hooks`: List of hook points this extension supports

### Writing the Hook Script

Extensions receive JSON input on stdin and must return JSON on stdout.

**Shell script example:**

```bash
#!/bin/sh

# Read JSON input from stdin
read -r input

# Parse JSON to extract fields
version=$(echo "$input" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
bump_type=$(echo "$input" | grep -o '"bump_type":"[^"]*"' | cut -d'"' -f4)

# Perform your custom logic here
echo "Processing version $version (bump type: $bump_type)" >&2

# Return success
echo '{"success": true, "message": "Extension executed successfully"}'
exit 0
```

**Python example:**

```python
#!/usr/bin/env python3
import sys
import json

# Read JSON input from stdin
input_data = json.load(sys.stdin)

version = input_data.get('version')
bump_type = input_data.get('bump_type')

# Perform your custom logic here
print(f"Processing version {version} (bump type: {bump_type})", file=sys.stderr)

# Return success
output = {
    "success": True,
    "message": "Extension executed successfully"
}
print(json.dumps(output))
sys.exit(0)
```

**Node.js example:**

```javascript
#!/usr/bin/env node

const readline = require("readline");

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false,
});

rl.on("line", (line) => {
  const input = JSON.parse(line);
  const { version, bump_type } = input;

  // Perform your custom logic here
  console.error(`Processing version ${version} (bump type: ${bump_type})`);

  // Return success
  const output = {
    success: true,
    message: "Extension executed successfully",
  };
  console.log(JSON.stringify(output));
  process.exit(0);
});
```

### JSON Output Format

Extensions must return JSON with the following structure:

```json
{
  "success": true,
  "message": "Optional status message",
  "data": {
    "key": "Optional data to return"
  }
}
```

**Fields:**

- `success` (required): Boolean indicating success or failure
- `message` (optional): Human-readable status message
- `data` (optional): Additional data to return

**Error handling:**

If your extension fails, return `success: false`:

```json
{
  "success": false,
  "message": "Extension failed: reason for failure"
}
```

## Installing Extensions

### From Local Path

```bash
verso extension install /path/to/my-extension
```

This will:

1. Copy the extension to `~/.verso-extensions/my-extension/`
2. Add the extension to `.verso.yaml`
3. Enable the extension by default

### Configuration

Extensions are configured in `.verso.yaml`:

```yaml
path: .version

extensions:
  - name: changelog-generator
    path: /Users/username/.verso-extensions/changelog-generator
    enabled: true

  - name: git-tagger
    path: /Users/username/.verso-extensions/git-tagger
    enabled: true
```

**Extension configuration fields:**

- `name`: Extension name (from manifest)
- `path`: Full path to installed extension
- `enabled`: Whether the extension is active (default: true)

## Managing Extensions

### List Installed Extensions

```bash
verso extension list
```

Output:

```
Installed Extensions:

changelog-generator (1.0.0)
  Description: Automatically updates CHANGELOG.md on version bumps
  Location: /Users/username/.verso-extensions/changelog-generator
  Status: enabled
```

### Remove an Extension

```bash
verso extension remove changelog-generator
```

This will:

1. Remove the extension from `.verso.yaml`
2. Optionally delete the extension files

### Disable an Extension

Edit `.verso.yaml` and set `enabled: false`:

```yaml
extensions:
  - name: changelog-generator
    path: /Users/username/.verso-extensions/changelog-generator
    enabled: false # Disabled
```

## Running Extensions

Extensions run automatically when you execute bump commands:

```bash
verso bump patch
# Pre-bump extensions run
# Version is bumped
# Post-bump extensions run
```

### Skip Extension Hooks

To skip extension execution, use the `--skip-hooks` flag:

```bash
verso bump patch --skip-hooks
```

Note: This also skips pre-release hooks.

## Security Considerations

1. **Trust**: Extensions are executed as shell scripts with your user permissions. Only install extensions from trusted sources.

2. **Timeouts**: Extensions have a 30-second execution timeout by default to prevent hanging.

3. **Output Limits**: Extension output is limited to 1MB to prevent memory exhaustion.

4. **Validation**: The CLI validates extension manifests and JSON output format.

5. **Isolation**: Each extension runs as a separate process with no access to other extensions.

## Best Practices

1. **Keep extensions focused**: Each extension should do one thing well
2. **Validate input**: Check for required fields before processing
3. **Handle errors gracefully**: Return proper error messages
4. **Document your extension**: Include a README with usage examples
5. **Test thoroughly**: Test your extension with different inputs
6. **Use semantic versioning**: Version your extensions properly
7. **Provide feedback**: Write meaningful messages to stderr for debugging

## Example Extensions

See the `contrib/extensions/` directory for reference implementations:

- **commit-validator**: Validates commit messages follow conventional commit format
- **docker-tag-sync**: Syncs version to Docker image tags
- **git-tagger**: Creates git tags on version bumps
- **package-sync**: Syncs version across package manifest files
- **version-policy**: Enforces custom versioning policies

## Troubleshooting

### Extension not found

Ensure the extension is installed and listed in `.verso.yaml`:

```bash
verso extension list
```

### Extension not executing

1. Check if the extension is enabled in `.verso.yaml`
2. Verify the script is executable: `chmod +x hook.sh`
3. Ensure the script has a proper shebang line: `#!/bin/sh`
4. Check hook points match: extension must support the hook being triggered

### Permission denied

Make the script executable:

```bash
chmod +x /path/to/extension/hook.sh
```

### Timeout errors

If your extension takes longer than 30 seconds, it will be terminated. Optimize your script or split operations into smaller tasks.

### Invalid JSON output

Ensure your script outputs valid JSON on stdout:

```bash
# Test manually
echo '{"hook":"post-bump","version":"1.0.0","project_root":"."}' | ./hook.sh
```

The output should be valid JSON that can be parsed.

## Advanced Topics

### Multiple Extensions

Multiple extensions can be configured for the same hook point. They execute in the order they appear in `.verso.yaml`.

### Extension Data

Extensions can return custom data in the `data` field for logging or debugging:

```json
{
  "success": true,
  "message": "Changelog updated",
  "data": {
    "entries_added": 1,
    "file": "CHANGELOG.md"
  }
}
```

### Error Propagation

If an extension fails (returns `success: false`), the entire bump operation is aborted and subsequent extensions are not executed.

## API Reference

### Input JSON Schema

```typescript
interface HookInput {
  hook: "pre-bump" | "post-bump" | "pre-release" | "validate";
  version: string;
  previous_version?: string;
  bump_type?: "patch" | "minor" | "major";
  prerelease?: string;
  metadata?: string;
  project_root: string;
}
```

### Output JSON Schema

```typescript
interface HookOutput {
  success: boolean;
  message?: string;
  data?: Record<string, any>;
}
```
