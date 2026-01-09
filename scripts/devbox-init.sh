#!/usr/bin/env bash
# devbox-init.sh  -  one-time setup for local dev environments
set -eu

# -------- Config --------
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"/lib/common.sh

# -------- Start --------
h1 'Welcome to the sley devbox!'

log_default ""

# Go dependencies and tools
log_info 'Setting up Go dependencies and tools...'
if [ -f "go.mod" ]; then
  log_info 'Downloading Go modules...'
  (go mod download)
  log_success 'Go modules downloaded'
else
  log_warning 'go.mod not found - skipping Go module download'
fi

if command -v go >/dev/null 2>&1; then
  log_info 'Installing Go tools...'

  if go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest; then
    log_success 'go-modernize installed'
  else
    log_warning 'Failed to install go-modernize'
  fi

  if go install golang.org/x/vuln/cmd/govulncheck@latest; then
    log_success 'govulncheck installed'
  else
    log_warning 'Failed to install govulncheck'
  fi

  # goreportcard-cli requires manual installation:
  # git clone https://github.com/gojp/goreportcard.git && cd goreportcard && make install && go install ./cmd/goreportcard-cli
  if command -v goreportcard-cli >/dev/null 2>&1; then
    log_success 'goreportcard-cli already installed'
  else
    log_faint 'goreportcard-cli not installed (optional) - see: https://github.com/gojp/goreportcard'
  fi
else
  log_warning 'Go not available - skipping Go tools installation'
fi

log_default ""

# Git hooks (prek)
log_info 'Setting up Git hooks with prek...'
# Ensure custom hooks are executable
if [ -f scripts/githooks/commit-msg ]; then
  chmod +x scripts/githooks/commit-msg
  log_success 'Custom hooks made executable'
else
  log_warning 'scripts/githooks/commit-msg not found'
fi

if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  if command -v prek >/dev/null 2>&1; then
    # Install prek hooks (workspace mode for monorepo)
    if prek install -- --workspace; then
      log_success 'prek hooks installed in workspace mode'
    else
      log_warning 'Failed to install prek hooks'
    fi
  else
    log_warning 'prek not installed  -  run: pip install prek'
  fi
else
  log_warning 'not a git repository  -  skipping hooks installation'
fi

log_default ""
# Helpful commands
h3 'Available just commands:'
cat <<'TXT'
  just help        - Show help message
  just all         - Clean and build
  just clean       - Clean the build directory and Go cache
  just test        - Run all tests and generate coverage report
  just test-force  - Clean go tests cache and run all tests
  just modernize   - Run go-modernize with auto-fix
  just check       - Run modernize, lint, and test
  just lint        - Run golangci-lint
  just build       - Build the binary to build/sley
  just install     - Install the binary using Go install

Quick start: `just check` to run all quality checks!
TXT

# End
printf '\n%s\n\n' '===================================='
