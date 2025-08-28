#!/bin/bash

set -euo pipefail

# Go-based integration testing script for Gitea
# Usage: ./integration-tests.sh <LOCAL_GITEA_GHORG_DIR> <TOKEN> <GITEA_URL>

LOCAL_GITEA_GHORG_DIR=${1:-"${HOME}/ghorg"}
TOKEN=${2:-$(cat "${LOCAL_GITEA_GHORG_DIR}/gitea_token" 2>/dev/null || echo "test-token")}
GITEA_URL=${3:-'http://gitea.example.com:3000'}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_RUNNER_DIR="${SCRIPT_DIR}/test-runner"
CONFIG_PATH="${SCRIPT_DIR}/configs/test-scenarios.json"

echo "Starting Gitea integration tests with Go-based test runner..."
echo "Gitea URL: ${GITEA_URL}"
echo "Ghorg Dir: ${LOCAL_GITEA_GHORG_DIR}"
echo "Config: ${CONFIG_PATH}"

# Build the test runner if it doesn't exist or if source files are newer
TEST_RUNNER_BINARY="${TEST_RUNNER_DIR}/gitea-test-runner"

# Force rebuild in CI environments or if binary doesn't exist or is newer
FORCE_BUILD=false
if [[ "${CI:-}" == "true" ]] || [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "CI environment detected - forcing clean build of test runner..."
    FORCE_BUILD=true
fi

if [[ ! -f "${TEST_RUNNER_BINARY}" ]] || [[ "${TEST_RUNNER_DIR}/main.go" -nt "${TEST_RUNNER_BINARY}" ]] || [[ "${FORCE_BUILD}" == "true" ]]; then
    echo "Building Gitea test runner..."
    cd "${TEST_RUNNER_DIR}"

    # Remove existing binary to ensure clean build
    rm -f gitea-test-runner

    go mod download
    go build -o gitea-test-runner main.go

    # Verify binary was created and is executable
    if [[ ! -f "gitea-test-runner" ]]; then
        echo "Error: Failed to build gitea-test-runner binary"
        exit 1
    fi

    chmod +x gitea-test-runner
    cd -
fi

# Install ghorg binary for testing if not in CI
if [[ "${CI:-}" != "true" ]] && [[ "${GITHUB_ACTIONS:-}" != "true" ]]; then
    echo "Installing ghorg binary for testing..."
    GHORG_PROJECT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
    cd "${GHORG_PROJECT_DIR}"
    go install .
    cd -
    echo "Using ghorg binary: $(which ghorg)"
    echo "Ghorg version: $(ghorg version)"
fi

# Run the integration tests
echo "Running Gitea integration tests..."
"${TEST_RUNNER_BINARY}" \
    -token="${TOKEN}" \
    -base-url="${GITEA_URL}" \
    -ghorg-dir="${LOCAL_GITEA_GHORG_DIR}" \
    -config="${CONFIG_PATH}"

if [[ $? -eq 0 ]]; then
    echo "Gitea integration tests completed successfully!"
else
    echo "Gitea integration tests failed!"
    exit 1
fi
