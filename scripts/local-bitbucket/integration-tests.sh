#!/bin/bash

set -euo pipefail

# Go-based integration testing script for Bitbucket Server
# Usage: ./integration-tests.sh <LOCAL_BITBUCKET_GHORG_DIR> <BITBUCKET_URL>

LOCAL_BITBUCKET_GHORG_DIR=${1:-"${HOME}/ghorg"}
BITBUCKET_URL=${2:-'http://bitbucket.example.com:7990'}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_RUNNER_DIR="${SCRIPT_DIR}/test-runner"
CONFIG_PATH="${SCRIPT_DIR}/configs/test-scenarios.json"

echo "Starting Bitbucket Server integration tests with Go-based test runner..."
echo "Bitbucket URL: ${BITBUCKET_URL}"
echo "Ghorg Dir: ${LOCAL_BITBUCKET_GHORG_DIR}"
echo "Config: ${CONFIG_PATH}"

# Build the test runner if it doesn't exist or if source files are newer
TEST_RUNNER_BINARY="${TEST_RUNNER_DIR}/bitbucket-test-runner"

# Force rebuild in CI environments or if binary doesn't exist or is newer
FORCE_BUILD=false
if [[ "${CI:-}" == "true" ]] || [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "CI environment detected - forcing clean build of test runner..."
    FORCE_BUILD=true
fi

# Rebuild main ghorg binary if not in CI (to include latest changes)
if [[ "${CI:-}" != "true" ]] && [[ "${GITHUB_ACTIONS:-}" != "true" ]]; then
    echo "Non-CI environment detected - rebuilding ghorg binary with latest changes..."

    # Find the root directory (where go.mod is)
    GHORG_ROOT="$(cd "$SCRIPT_DIR" && cd ../.. && pwd)"

    if [ -f "$GHORG_ROOT/go.mod" ]; then
        echo "Building ghorg binary from: $GHORG_ROOT"
        cd "$GHORG_ROOT"

        if go install .; then
            echo "✅ ghorg binary rebuilt successfully"
        else
            echo "❌ Failed to rebuild ghorg binary"
            exit 1
        fi

        cd "$SCRIPT_DIR"
    else
        echo "⚠️  Could not find go.mod at $GHORG_ROOT - skipping ghorg rebuild"
    fi
fi

if [ "$FORCE_BUILD" = true ] || [ ! -f "$TEST_RUNNER_BINARY" ] || [ "$TEST_RUNNER_DIR/main.go" -nt "$TEST_RUNNER_BINARY" ]; then
    echo "Building Bitbucket test runner..."
    cd "$TEST_RUNNER_DIR"

    # Clean any existing binary in CI
    if [ "$FORCE_BUILD" = true ] && [ -f "$TEST_RUNNER_BINARY" ]; then
        rm -f "$TEST_RUNNER_BINARY"
    fi

    go build -o bitbucket-test-runner main.go

    if [ $? -eq 0 ]; then
        echo "✅ Test runner built successfully"
    else
        echo "❌ Failed to build test runner"
        exit 1
    fi
    cd "$SCRIPT_DIR"
else
    echo "✅ Test runner binary is up to date"
fi

# Check if test runner binary exists and is executable
if [ ! -f "$TEST_RUNNER_BINARY" ]; then
    echo "❌ Test runner binary not found at: $TEST_RUNNER_BINARY"
    exit 1
fi

if [ ! -x "$TEST_RUNNER_BINARY" ]; then
    echo "Making test runner executable..."
    chmod +x "$TEST_RUNNER_BINARY"
fi

# Verify environment variables are set
if [ -z "${GHORG_SCM_BASE_URL:-}" ]; then
    echo "Setting GHORG_SCM_BASE_URL to ${BITBUCKET_URL}"
    export GHORG_SCM_BASE_URL="$BITBUCKET_URL"
fi

if [ -z "${GHORG_SCM_TYPE:-}" ]; then
    echo "Setting GHORG_SCM_TYPE to bitbucket"
    export GHORG_SCM_TYPE="bitbucket"
fi

if [ -z "${GHORG_BITBUCKET_USERNAME:-}" ]; then
    echo "Setting default GHORG_BITBUCKET_USERNAME to admin"
    export GHORG_BITBUCKET_USERNAME="admin"
fi

if [ -z "${GHORG_BITBUCKET_APP_PASSWORD:-}" ]; then
    echo "Setting default GHORG_BITBUCKET_APP_PASSWORD to admin"
    export GHORG_BITBUCKET_APP_PASSWORD="admin"
fi

if [ -z "${GHORG_INSECURE_BITBUCKET_CLIENT:-}" ]; then
    echo "Setting GHORG_INSECURE_BITBUCKET_CLIENT to true"
    export GHORG_INSECURE_BITBUCKET_CLIENT="true"
fi

if [ -z "${GHORG_CLONE_PROTOCOL:-}" ]; then
    echo "Setting GHORG_CLONE_PROTOCOL to https"
    export GHORG_CLONE_PROTOCOL="https"
fi

# Set concurrency to 1 for Bitbucket Server stability
if [ -z "${GHORG_CONCURRENCY:-}" ]; then
    echo "Setting GHORG_CONCURRENCY to 1 for Bitbucket Server stability"
    export GHORG_CONCURRENCY="1"
fi

echo ""
echo "Environment variables:"
echo "  GHORG_SCM_TYPE: ${GHORG_SCM_TYPE}"
echo "  GHORG_SCM_BASE_URL: ${GHORG_SCM_BASE_URL}"
echo "  GHORG_BITBUCKET_USERNAME: ${GHORG_BITBUCKET_USERNAME}"
echo "  GHORG_BITBUCKET_APP_PASSWORD: [REDACTED]"
echo "  GHORG_INSECURE_BITBUCKET_CLIENT: ${GHORG_INSECURE_BITBUCKET_CLIENT}"
echo "  GHORG_CLONE_PROTOCOL: ${GHORG_CLONE_PROTOCOL}"
echo ""

# Run the test runner
echo "Running integration tests..."
if "$TEST_RUNNER_BINARY" -config "$CONFIG_PATH" -ghorg-dir "$LOCAL_BITBUCKET_GHORG_DIR" -base-url "$BITBUCKET_URL" -token "${GHORG_BITBUCKET_APP_PASSWORD}"; then
    echo ""
    echo "🎉 All Bitbucket Server integration tests passed!"
    exit 0
else
    echo ""
    echo "❌ Bitbucket Server integration tests failed"
    exit 1
fi
