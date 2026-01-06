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

# Default recipe: show help
default: help

# Print this help message
help:
    @echo ""
    @echo "Usage: just [recipe]"
    @echo ""
    @echo "Available Recipes:"
    @echo ""
    @just --list
    @echo ""

# Clean and build
all: clean build

# Clean the build directory and Go cache
clean:
    @echo "* Clean the build directory and Go cache"
    rm -rf {{ build_dir }}
    {{ goclean }} -cache

# Run go-modernize with auto-fix
modernize:
    @echo "* Running go-modernize"
    modernize --fix ./...

# Run golangci-lint
lint:
    @echo "* Running golangci-lint"
    golangci-lint run ./...

# Run goreportcard-cli
reportcard:
    @echo "* Running goreportcard-cli..."
    goreportcard-cli -v

# Run all tests and print code coverage value
test:
    @echo "* Run all tests and print code coverage value."
    {{ go }} test $({{ go }} list ./... | grep -Ev 'internal/testutils') -coverprofile=coverage.txt
    @echo "* Total Coverage"
    {{ go }} tool cover -func=coverage.txt | grep total | awk '{print $3}'

# Clean go tests cache and run all tests
test-force:
    @echo "* Clean go tests cache and run all tests."
    {{ go }} clean -testcache
    just test

# Run all tests and generate coverage report.
test-coverage:
    @echo "* Run all tests and generate coverage report."
    go test -count=1 -timeout 30s $(go list ./... | grep -Ev 'internal/testutils') -covermode=atomic -coverprofile=coverage.txt

# Run all tests with race detector
test-race:
    @echo "* Running tests with race detector..."
    {{ go }} test -race $({{ go }} list ./... | grep -Ev 'internal/testutils')

# Run modernize, lint, and reportcard
check: modernize lint reportcard

# Build the binary with optimizations (reduced size)
build:
    @echo "* Building optimized binary..."
    mkdir -p {{ build_dir }}
    {{ gobuild }} {{ buildflags }} -ldflags="{{ ldflags }}" -o {{ build_dir }}/{{ app_name }} ./{{ cmd_dir }}
    @echo "* Binary size:"
    @ls -lh {{ build_dir }}/{{ app_name }} | awk '{print "  " $5}'

# Install the binary using Go install
install: check test-force
    @echo "* Install the binary using Go install"
    cd {{ cmd_dir }} && {{ go }} install {{ buildflags }} -ldflags="{{ ldflags }}" .
