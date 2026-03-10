# Variables

export LOG_STYLE := "emoji"
logger := "scripts/lib/logger.sh"

# Go commands

go := "go"
gobuild := go + " build"
goclean := go + " clean"

# Binary name

app_name := "sley"

# Directories

build_dir := "build"
cmd_dir := "cmd/" + app_name

# Build optimization flags
# -s: Omit the symbol table and debug information
# -w: Omit the DWARF symbol table

ldflags := "-s -w"

# -trimpath: Remove file system paths from binary

buildflags := "-trimpath"

# Default - show help
default:
    @just --list

# === Build Recipes ===

# Build the binary with optimizations (reduced size)
build:
    @. {{ logger }} && log_info "Building optimized binary..."
    mkdir -p {{ build_dir }}
    {{ gobuild }} {{ buildflags }} -ldflags="{{ ldflags }}" -o {{ build_dir }}/{{ app_name }} ./{{ cmd_dir }}
    @. {{ logger }} && log_info "Binary size:"
    @ls -lh {{ build_dir }}/{{ app_name }} | awk '{print "  " $5}'

# Install the binary using Go install
install: check test-force
    @. {{ logger }} && log_info "Install the binary using Go install"
    cd {{ cmd_dir }} && {{ go }} install {{ buildflags }} -ldflags="{{ ldflags }}" .

# Clean and build
all: clean build

# Clean the build directory and Go cache
clean:
    @. {{ logger }} && log_info "Clean the build directory and Go cache"
    rm -rf {{ build_dir }}
    rm -f coverage.out coverage.html
    {{ goclean }} -cache

# === Code Quality ===

# Format code
fmt:
    @. {{ logger }} && log_info "Running fmt and gofumpt"
    {{ go }} fmt ./...

# Run go-modernize with auto-fix
modernize:
    @. {{ logger }} && log_info "Running go-modernize"
    modernize --fix ./...

# Run golangci-lint
lint:
    @. {{ logger }} && log_info "Running golangci-lint"
    golangci-lint run ./...

# Run goreportcard-cli
reportcard:
    @. {{ logger }} && log_info "Running goreportcard-cli"
    goreportcard-cli -v

# Run govulncheck
security-scan:
    @. {{ logger }} && log_info "Running govulncheck"
    govulncheck ./...

# Run modernize, lint, and reportcard
check: fmt modernize lint reportcard

# Run go mod tidy
tidy:
    @. {{ logger }} && log_info "Running go mod tidy"
    {{ go }} mod tidy

# Run go mod download
deps:
    @. {{ logger }} && log_info "Running go mod download"
    {{ go }} mod download

# === Test Recipes ===

# Run all tests and print code coverage value
test:
    @. {{ logger }} && log_info "Run all tests"
    {{ go }} test $({{ go }} list ./... | grep -Ev 'internal/testutils') -coverprofile=coverage.txt
    @. {{ logger }} && log_info "Total Coverage"
    {{ go }} tool cover -func=coverage.txt | grep total | awk '{print $3}'

# Clean go tests cache and run all tests
test-force:
    @. {{ logger }} && log_info "Clean go tests cache and run all tests"
    {{ go }} clean -testcache
    just test

# Run all tests and generate coverage report.
test-coverage:
    @. {{ logger }} && log_info "Run all tests and generate coverage report"
    {{ go }} test -count=1 -timeout 30s $({{ go }} list ./... | grep -Ev 'internal/testutils') -covermode=atomic -coverprofile=coverage.txt

# Run all tests with race detector
test-race:
    @. {{ logger }} && log_info "Running tests with race detector"
    {{ go }} test -race $({{ go }} list ./... | grep -Ev 'internal/testutils')

# === Utilities ===

# Update dependencies
deps-update:
    @. {{ logger }} && log_info "Running go update deps"
    {{ go }} get -u ./...
    {{ go }} mod tidy
