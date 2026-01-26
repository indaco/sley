# Docker Tag Sync Extension

Tags Docker images with the new version after successful version bumps.

## Features

- Automatically tags Docker images with the version number
- Configurable version prefix (e.g., `v1.2.3` or `1.2.3`)
- Optional push to registry after tagging
- Support for custom source tags (not just `latest`)
- Optional "also tag as latest" functionality

## Requirements

- Docker CLI installed and accessible
- Source image must exist locally (build before bumping)
- For push: authenticated to the target registry

## Installation

```bash
sley extension install --path ./contrib/extensions/docker-tag-sync
```

## Configuration

Add to your `.sley.yaml`:

```yaml
extensions:
  docker-tag-sync:
    # Required: Docker image name
    image: "myapp"

    # Optional: Source tag to re-tag (default: "latest")
    source-tag: "latest"

    # Optional: Version prefix for the tag (default: "v")
    prefix: "v"

    # Optional: Push to registry after tagging (default: false)
    push: false

    # Optional: Also update the "latest" tag (default: false)
    also-tag-latest: false

    # Optional: Registry (if not included in image name)
    registry: "docker.io/myorg"
```

## Examples

### Basic Usage

Tag local image with version:

```yaml
extensions:
  docker-tag-sync:
    image: "myapp"
```

```bash
# Build your image first
docker build -t myapp:latest .

# Bump version - automatically tags myapp:v1.2.4
sley bump patch
```

### With Registry Push

Tag and push to Docker Hub:

```yaml
extensions:
  docker-tag-sync:
    image: "myorg/myapp"
    push: true
```

### With Custom Registry

Tag and push to private registry:

```yaml
extensions:
  docker-tag-sync:
    image: "myapp"
    registry: "ghcr.io/myorg"
    push: true
    also-tag-latest: true
```

This will:

1. Tag `myapp:latest` as `ghcr.io/myorg/myapp:v1.2.4`
2. Tag `myapp:latest` as `ghcr.io/myorg/myapp:latest`
3. Push both tags to GitHub Container Registry

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
- name: Build Docker image
  run: docker build -t myapp:latest .

- name: Bump version and tag image
  run: sley bump patch

- name: Push to registry
  run: |
    docker push myapp:v${{ steps.version.outputs.new }}
```

Or configure push in `.sley.yaml` for automatic pushing:

```yaml
extensions:
  docker-tag-sync:
    image: "ghcr.io/myorg/myapp"
    push: true
```

## How It Works

1. After a successful version bump, the extension receives the new version
2. It looks for the source image (default: `image:latest`)
3. Tags the source image with the new version tag
4. Optionally pushes to the registry

## Error Handling

The extension will fail (and report the error) if:

- The source image doesn't exist locally
- Docker tagging fails
- Push fails (if enabled)

The version bump itself will still succeed - only the Docker tagging step fails.

## Troubleshooting

### "Source image not found"

Build your Docker image before running the version bump:

```bash
docker build -t myapp:latest .
sley bump patch
```

### "Failed to push"

Ensure you're authenticated to the registry:

```bash
docker login ghcr.io
# or
docker login docker.io
```

### Using a different source tag

If your CI builds with a different tag (e.g., commit SHA):

```yaml
extensions:
  docker-tag-sync:
    image: "myapp"
    source-tag: "sha-abc123"
```

## See Also

- [Extensions Overview](https://sley.indaco.dev/extensions/) - How extensions work
- [tag-manager plugin](https://sley.indaco.dev/plugins/tag-manager.html) - Git tag automation (built-in)
