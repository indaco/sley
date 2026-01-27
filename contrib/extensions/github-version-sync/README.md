# GitHub Version Sync Extension

Syncs the `.version` file to the latest release from a GitHub repository before bumping.

## Features

- Fetches latest release from any public GitHub repository
- Configurable tag prefix stripping (e.g., `v1.2.3` -> `1.2.3`)
- Optional GitHub token for private repos or higher rate limits
- Support for GitHub Enterprise via custom API URL
- Works with `jq`/`jaq` for robust JSON parsing, falls back to basic shell parsing

## Requirements

- `curl` (standard on most systems)
- `jq` or `jaq` (recommended, but not required)

## Installation

Install directly from the sley repository:

```bash
sley extension install --url github.com/indaco/sley/contrib/extensions/github-version-sync
```

Or from a local clone:

```bash
sley extension install --path ./contrib/extensions/github-version-sync
```

## Configuration

Add to your `.sley.yaml`. Extension settings are defined under the `config` key:

```yaml
extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      # Required: GitHub repository in "owner/repo" format
      repo: owner/repo

      # Optional: Prefix to strip from tag (default: "v")
      strip-prefix: v

      # Optional: GitHub token for private repos or rate limits
      github-token: ""

      # Optional: GitHub API URL for Enterprise (default: "https://api.github.com")
      api-url: https://api.github.com
```

### Configuration Options

All configuration options are specified under the `config` key and passed to the hook script as JSON:

- **`repo`** (required): GitHub repository in "owner/repo" format
- **`strip-prefix`** (optional): Prefix to strip from tag names (default: `"v"`)
- **`github-token`** (optional): GitHub token for private repos or higher rate limits
- **`api-url`** (optional): GitHub API URL for Enterprise instances (default: `"https://api.github.com"`)

When using this extension, set `commit-parser: false` in your config. The extension sets the definitive version - no additional bump should be applied.

```yaml
plugins:
  commit-parser: false # Let the extension control the version
```

## Examples

### Basic Usage

Sync to a public repository:

```yaml
plugins:
  commit-parser: false
  tag-manager:
    enabled: true

extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: indaco/sley
```

```bash
# Fetches latest sley release (e.g., v0.9.0)
# Updates .version to 0.9.0
sley bump auto
```

### With GitHub Token

For private repositories or to avoid rate limits:

```yaml
extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: myorg/private-repo
      github-token: "${GITHUB_TOKEN}" # Use environment variable
```

### GitHub Enterprise

```yaml
extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: myorg/myrepo
      api-url: https://github.mycompany.com/api/v3
      github-token: "${GHE_TOKEN}"
```

### Custom Tag Format

If the upstream repo uses tags without `v` prefix:

```yaml
extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: some/repo
      strip-prefix: "" # Don't strip any prefix
```

### Documentation Site Workflow

Complete example for a docs site tracking a library:

```yaml
# .sley.yaml
path: .version

plugins:
  commit-parser: false
  tag-manager:
    enabled: true
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json

extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: indaco/sley
```

```bash
# Sync docs version to latest sley release
sley bump auto
# 1. Fetches latest sley release (v0.9.0)
# 2. Updates .version to 0.9.0
# 3. Syncs package.json version
# 4. Creates git tag v0.9.0
```

## How It Works

1. Pre-bump hook receives the current project state
2. Extension fetches the latest release from the configured GitHub repo
3. Extracts the tag name and strips the configured prefix
4. Writes the version to the `.version` file
5. sley detects the change and uses this version (no additional bump applied)

## Error Handling

The extension will fail (and block the bump) if:

- The GitHub API request fails
- The repository is not found or has no releases
- Rate limit is exceeded (hint: use `github-token`)
- The `.version` file cannot be written

## Troubleshooting

### "GitHub API rate limit exceeded"

GitHub limits unauthenticated requests to 60/hour. Set a token in your extension config:

```yaml
extensions:
  - name: github-version-sync
    path: .sley-extensions/github-version-sync
    enabled: true
    config:
      repo: owner/repo
      github-token: "${GITHUB_TOKEN}"
```

Create a token at: https://github.com/settings/tokens

### "Repository not found or no releases"

- Check the repo name is correct (`owner/repo` format)
- Ensure the repository has at least one release (not just tags)
- For private repos, ensure the token has `repo` scope

### "Version not changing"

Ensure `commit-parser: false` is set. Otherwise, sley may recalculate and override the version.

### Using with jq/jaq vs without

The extension works best with `jq` or `jaq` (a Rust alternative) installed for robust JSON parsing:

```bash
# jq
brew install jq        # macOS
apt-get install jq     # Ubuntu/Debian
apk add jq             # Alpine

# jaq (Rust alternative)
brew install jaq       # macOS
cargo install jaq      # via Cargo
```

Without `jq` or `jaq`, basic shell parsing is used (works for simple cases).

## Use Cases

- Documentation sites tracking a library's version
- Projects that need to stay in sync with an upstream dependency
- Monorepos where one package tracks another's release

## See Also

- [Extensions Overview](https://sley.indaco.dev/extensions/) - How extensions work
- [dependency-check plugin](https://sley.indaco.dev/plugins/dependency-check) - Sync version to package.json
- [tag-manager plugin](https://sley.indaco.dev/plugins/tag-manager) - Git tag automation
