#!/usr/bin/env bash
set -euo pipefail

# Read JSON input from stdin
INPUT=$(cat)

# Extract values using parameter expansion (no jq dependency)
VERSION=$(echo "$INPUT" | grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*:.*"\([^"]*\)".*/\1/')
PROJECT_ROOT=$(echo "$INPUT" | grep -o '"project_root"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*:.*"\([^"]*\)".*/\1/')

if [[ -z "$VERSION" || -z "$PROJECT_ROOT" ]]; then
	echo '{"success": false, "message": "Missing version or project_root in input"}'
	exit 1
fi

# Files to update
TEST_FILE="${PROJECT_ROOT}/internal/version/version_test.go"

ERRORS=""

# Update version_test.go
if [[ -f "$TEST_FILE" ]]; then
	if ! sed -i '' "s/expectedVersion := \"[^\"]*\"/expectedVersion := \"${VERSION}\"/" "$TEST_FILE"; then
		ERRORS="${ERRORS}Failed to update ${TEST_FILE}. "
	fi
else
	ERRORS="${ERRORS}File not found: ${TEST_FILE}. "
fi

if [[ -n "$ERRORS" ]]; then
	echo "{\"success\": false, \"message\": \"${ERRORS}\"}"
	exit 1
fi

echo "{\"success\": true, \"message\": \"Updated version to ${VERSION} in version_test.go\"}"
