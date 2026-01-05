# Package Sync Extension

This extension automatically synchronizes version numbers to package.json and other JSON files after version bumps using sley.

## Features

- Updates package.json version field automatically
- Supports multiple JSON files in a single configuration
- Supports nested JSON paths (e.g., "metadata.version")
- Preserves original file formatting and indentation
- No external dependencies (uses only Node.js standard library)
- Monorepo/multi-module support: automatically detects module directory

## Installation

**From local path:**

```bash
sley extension install --path ./contrib/extensions/package-sync
```

**From URL (after cloning the repo):**

```bash
sley extension install --url https://github.com/indaco/sley
# Then copy from contrib/extensions/package-sync
```

## Usage

Once installed and enabled, the extension will automatically run on every version bump:

```bash
sley bump patch
# Updates package.json version field
```

## Configuration

Add configuration to your `.sley.yaml`:

### Basic Configuration (Default)

By default, updates `package.json` version field:

```yaml
extensions:
  - name: package-sync
    enabled: true
    hooks:
      - post-bump
```

### Advanced Configuration

Update multiple files and custom JSON paths:

```yaml
extensions:
  - name: package-sync
    enabled: true
    hooks:
      - post-bump
    config:
      files:
        - path: package.json
          json_paths:
            - version
        - path: package-lock.json
          json_paths:
            - version
            - packages..version # Root package version in lockfile
        - path: manifest.json
          json_paths:
            - version
            - metadata.appVersion
```

### Configuration Options

- `files`: Array of file configurations
  - `path`: Relative path to JSON file from project root
  - `json_paths`: Array of dot-separated JSON paths to update (default: `["version"]`)

## Examples

### Basic Usage

Update package.json only:

```bash
sley bump minor
# package.json version: 1.2.3 -> 1.3.0
```

### Multiple Files

Update package.json and package-lock.json:

```yaml
config:
  files:
    - path: package.json
      json_paths:
        - version
    - path: package-lock.json
      json_paths:
        - version
```

### Nested Paths

Update version in nested JSON structures:

```yaml
config:
  files:
    - path: app-manifest.json
      json_paths:
        - version
        - metadata.version
        - app.version
```

Given `app-manifest.json`:

```json
{
  "version": "1.2.3",
  "metadata": {
    "version": "1.2.3"
  },
  "app": {
    "version": "1.2.3"
  }
}
```

All three paths will be updated to the new version.

### Short-Form Configuration

For simple cases, you can use a string instead of an object:

```yaml
config:
  files:
    - package.json
    - manifest.json
# Both files will have their "version" field updated
```

## Hooks Supported

- `post-bump`: Runs after version is bumped

## Requirements

- Node.js 12 or higher
- Write permissions in the project directory
- Valid JSON files at specified paths

## Error Handling

The extension will fail if:

- Node.js is not installed
- Specified file does not exist
- File contains invalid JSON
- File is not writable
- JSON path does not exist in file

Partial failures are reported: if updating one file fails, other files will still be processed.

## JSON Input Format

The extension receives the following JSON on stdin:

```json
{
  "hook": "post-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": null,
  "metadata": null,
  "project_root": "/path/to/project",
  "module_dir": "/path/to/project/packages/app",
  "module_name": "app",
  "config": {
    "files": [
      {
        "path": "package.json",
        "json_paths": ["version"]
      }
    ]
  }
}
```

### Field Descriptions

| Field          | Description                                                                                                                            |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `project_root` | The root directory of the project (where `.sley.yaml` is located)                                                                      |
| `module_dir`   | The directory containing the `.version` file being bumped (monorepo support). Falls back to `project_root` for single-module projects. |
| `module_name`  | The name of the module being bumped (derived from directory name). Only present in monorepo mode.                                      |

## JSON Output Format

Success:

```json
{
  "success": true,
  "message": "Updated package.json: version: 1.2.2 -> 1.2.3",
  "data": {
    "files_processed": 1,
    "version": "1.2.3"
  }
}
```

Multiple files:

```json
{
  "success": true,
  "message": "Updated package.json: version: 1.2.2 -> 1.2.3; Updated manifest.json: version: 1.2.2 -> 1.2.3",
  "data": {
    "files_processed": 2,
    "version": "1.2.3"
  }
}
```

Error:

```json
{
  "success": false,
  "message": "File not found: package.json",
  "data": {}
}
```

## Formatting Preservation

The extension automatically detects and preserves the indentation style of each JSON file:

- Detects spaces vs tabs
- Preserves indentation size (2 spaces, 4 spaces, etc.)
- Adds trailing newline for consistency
- Maintains key ordering

## Monorepo Support

When working with monorepos, sley automatically provides the module directory context to extensions. The `package-sync` extension uses this to locate `package.json` relative to each module's `.version` file.

### Example Monorepo Structure

```
/project
├── .sley.yaml
├── packages/
│   ├── app/
│   │   ├── .version          # app version
│   │   └── package.json      # automatically updated
│   └── lib/
│       ├── .version          # lib version
│       └── package.json      # automatically updated
```

When you bump `packages/app/.version`, the extension automatically updates `packages/app/package.json`.

### Monorepo Configuration

For monorepos, the default configuration works automatically:

```yaml
extensions:
  - name: package-sync
    enabled: true
    hooks:
      - post-bump
```

If you need to update files relative to the project root instead of the module directory, use the `relative_to` option:

```yaml
extensions:
  - name: package-sync
    enabled: true
    hooks:
      - post-bump
    config:
      files:
        - path: package.json
          json_paths:
            - version
        - path: root-manifest.json
          relative_to: project # Always relative to project root
          json_paths:
            - version
```

## Use Cases

- JavaScript/TypeScript projects with package.json
- Monorepos with multiple package files
- Projects with custom manifest files
- Applications with version embedded in configuration files
