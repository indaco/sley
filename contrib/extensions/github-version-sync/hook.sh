#!/usr/bin/env bash
#
# GitHub Version Sync Extension for sley
# Syncs the .version file to the latest release from a GitHub repository
#
# Configuration options (via .sley.yaml):
#   repo: GitHub repository in "owner/repo" format (required)
#   strip-prefix: Prefix to strip from tag (default: "v")
#   github-token: GitHub token for private repos or higher rate limits (optional)
#   api-url: GitHub API base URL for GitHub Enterprise (default: "https://api.github.com")
#
# Input JSON format (from stdin):
# {
#   "hook": "pre-bump",
#   "version": "1.2.3",
#   "previous_version": "1.2.2",
#   "bump_type": "patch",
#   "project_root": "/path/to/project",
#   "config": {
#     "repo": "owner/repo",
#     "strip-prefix": "v",
#     "github-token": "",
#     "api-url": "https://api.github.com"
#   }
# }
#
# Output JSON format (to stdout):
# {
#   "success": true,
#   "message": "Synced to release v1.2.3",
#   "data": {"version": "1.2.3", "tag": "v1.2.3"}
# }

set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Detect JSON processor: jq or jaq (Rust alternative)
JQ_CMD=""
if command -v jq &>/dev/null; then
	JQ_CMD="jq"
elif command -v jaq &>/dev/null; then
	JQ_CMD="jaq"
fi

# Parse JSON fields using jq/jaq (or fallback to grep/sed for basic parsing)
if [ -n "$JQ_CMD" ]; then
	PROJECT_ROOT=$(echo "$INPUT" | "$JQ_CMD" -r '.project_root // empty')
	REPO=$(echo "$INPUT" | "$JQ_CMD" -r '.config.repo // empty')
	STRIP_PREFIX=$(echo "$INPUT" | "$JQ_CMD" -r '.config["strip-prefix"] // "v"')
	GITHUB_TOKEN=$(echo "$INPUT" | "$JQ_CMD" -r '.config["github-token"] // empty')
	API_URL=$(echo "$INPUT" | "$JQ_CMD" -r '.config["api-url"] // "https://api.github.com"')
else
	# Fallback: basic parsing without jq/jaq
	PROJECT_ROOT=$(echo "$INPUT" | grep -o '"project_root"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*: *"\([^"]*\)".*/\1/')
	REPO=$(echo "$INPUT" | grep -o '"repo"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*: *"\([^"]*\)".*/\1/')
	STRIP_PREFIX="v"
	GITHUB_TOKEN=""
	API_URL="https://api.github.com"
fi

# Helper function to output JSON result
output_result() {
	local success="$1"
	local message="$2"
	local version="${3:-}"
	local tag="${4:-}"

	if [ -n "$version" ] && [ -n "$tag" ]; then
		echo "{\"success\": $success, \"message\": \"$message\", \"data\": {\"version\": \"$version\", \"tag\": \"$tag\"}}"
	else
		echo "{\"success\": $success, \"message\": \"$message\", \"data\": {}}"
	fi
}

# Validate required fields
if [ -z "$PROJECT_ROOT" ]; then
	output_result "false" "Missing required field: project_root"
	exit 1
fi

if [ -z "$REPO" ]; then
	output_result "false" "Missing required config: repo. Set 'repo' in extension config (e.g., 'owner/repo')."
	exit 1
fi

# Build API URL
RELEASES_URL="${API_URL}/repos/${REPO}/releases/latest"

# Prepare curl options
CURL_OPTS=(-s -f)
if [ -n "$GITHUB_TOKEN" ]; then
	CURL_OPTS+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi
CURL_OPTS+=(-H "Accept: application/vnd.github.v3+json")

# Fetch latest release
echo "Fetching latest release from ${REPO}..." >&2

if ! RESPONSE=$(curl "${CURL_OPTS[@]}" "$RELEASES_URL" 2>&1); then
	# Check for common errors
	if echo "$RESPONSE" | grep -q "rate limit"; then
		output_result "false" "GitHub API rate limit exceeded. Consider setting 'github-token' in config."
	elif echo "$RESPONSE" | grep -q "Not Found"; then
		output_result "false" "Repository not found or no releases: ${REPO}"
	else
		output_result "false" "Failed to fetch release from GitHub: ${RESPONSE}"
	fi
	exit 1
fi

# Extract tag name
if [ -n "$JQ_CMD" ]; then
	TAG_NAME=$(echo "$RESPONSE" | "$JQ_CMD" -r '.tag_name // empty')
else
	TAG_NAME=$(echo "$RESPONSE" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/.*: *"\([^"]*\)".*/\1/')
fi

if [ -z "$TAG_NAME" ]; then
	output_result "false" "No tag_name found in release response"
	exit 1
fi

# Strip prefix if configured
VERSION="$TAG_NAME"
if [ -n "$STRIP_PREFIX" ]; then
	VERSION="${TAG_NAME#"$STRIP_PREFIX"}"
fi

echo "Latest release: ${TAG_NAME} -> version ${VERSION}" >&2

# Update .version file
VERSION_FILE="${PROJECT_ROOT}/.version"
if ! echo "$VERSION" >"$VERSION_FILE"; then
	output_result "false" "Failed to write version file: ${VERSION_FILE}"
	exit 1
fi

echo "Updated ${VERSION_FILE} to ${VERSION}" >&2
output_result "true" "Synced to ${REPO} release ${TAG_NAME}" "$VERSION" "$TAG_NAME"
exit 0
