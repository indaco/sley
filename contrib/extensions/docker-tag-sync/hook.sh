#!/usr/bin/env bash
#
# Docker Tag Sync Extension for sley
# Tags and optionally pushes Docker images with the new version after bumps
#
# Configuration options (via .sley.yaml):
#   image: Image name (required, e.g., "myapp" or "registry.io/org/myapp")
#   source-tag: Source tag to re-tag (default: "latest")
#   prefix: Version prefix for the tag (default: "v")
#   push: Push the new tag to registry (default: false)
#   also-tag-latest: Also update the "latest" tag (default: false)
#   registry: Registry to push to (optional, uses image registry if specified in image name)
#
# Input JSON format (from stdin):
# {
#   "version": "1.2.3",
#   "previous_version": "1.2.2",
#   "bump_type": "patch",
#   "project_root": "/path/to/project",
#   "config": {
#     "image": "myapp",
#     "source-tag": "latest",
#     "prefix": "v",
#     "push": false
#   }
# }
#
# Output JSON format (to stdout):
# {
#   "success": true,
#   "message": "Tagged myapp:v1.2.3",
#   "data": {}
# }

set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Parse JSON fields using jq (or fallback to grep/sed for basic parsing)
if command -v jq &>/dev/null; then
    VERSION=$(echo "$INPUT" | jq -r '.version // empty')
    PROJECT_ROOT=$(echo "$INPUT" | jq -r '.project_root // empty')
    IMAGE=$(echo "$INPUT" | jq -r '.config.image // empty')
    SOURCE_TAG=$(echo "$INPUT" | jq -r '.config["source-tag"] // "latest"')
    PREFIX=$(echo "$INPUT" | jq -r '.config.prefix // "v"')
    PUSH=$(echo "$INPUT" | jq -r '.config.push // false')
    ALSO_TAG_LATEST=$(echo "$INPUT" | jq -r '.config["also-tag-latest"] // false')
    REGISTRY=$(echo "$INPUT" | jq -r '.config.registry // empty')
else
    # Fallback: basic parsing without jq
    VERSION=$(echo "$INPUT" | grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*: *"\([^"]*\)".*/\1/')
    # shellcheck disable=SC2034  # PROJECT_ROOT parsed for consistency, not used by this extension
    PROJECT_ROOT=$(echo "$INPUT" | grep -o '"project_root"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*: *"\([^"]*\)".*/\1/')
    IMAGE=""
    SOURCE_TAG="latest"
    PREFIX="v"
    PUSH="false"
    ALSO_TAG_LATEST="false"
    REGISTRY=""
fi

# Helper function to output JSON result
output_result() {
    local success="$1"
    local message="$2"
    echo "{\"success\": $success, \"message\": \"$message\", \"data\": {}}"
}

# Validate required fields
if [ -z "$VERSION" ]; then
    output_result "false" "Missing required field: version"
    exit 1
fi

if [ -z "$IMAGE" ]; then
    output_result "false" "Missing required config: image. Set 'image' in extension config."
    exit 1
fi

# Build full image names
FULL_IMAGE="$IMAGE"
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE="$REGISTRY/$IMAGE"
fi

SOURCE_IMAGE="$FULL_IMAGE:$SOURCE_TAG"
TARGET_TAG="${PREFIX}${VERSION}"
TARGET_IMAGE="$FULL_IMAGE:$TARGET_TAG"

# Check if source image exists
if ! docker image inspect "$SOURCE_IMAGE" &>/dev/null; then
    output_result "false" "Source image not found: $SOURCE_IMAGE. Build the image first."
    exit 1
fi

# Tag the image
if ! docker tag "$SOURCE_IMAGE" "$TARGET_IMAGE" 2>&1; then
    output_result "false" "Failed to tag image: $SOURCE_IMAGE -> $TARGET_IMAGE"
    exit 1
fi

RESULT_MESSAGE="Tagged $TARGET_IMAGE"

# Also tag as latest if requested
if [ "$ALSO_TAG_LATEST" = "true" ]; then
    LATEST_IMAGE="$FULL_IMAGE:latest"
    if docker tag "$SOURCE_IMAGE" "$LATEST_IMAGE" 2>&1; then
        RESULT_MESSAGE="$RESULT_MESSAGE and $LATEST_IMAGE"
    fi
fi

# Push if requested
if [ "$PUSH" = "true" ]; then
    if ! docker push "$TARGET_IMAGE" 2>&1; then
        output_result "false" "Tagged but failed to push: $TARGET_IMAGE"
        exit 1
    fi
    RESULT_MESSAGE="$RESULT_MESSAGE (pushed)"

    if [ "$ALSO_TAG_LATEST" = "true" ]; then
        docker push "$LATEST_IMAGE" 2>&1 || true
    fi
fi

output_result "true" "$RESULT_MESSAGE"
exit 0
