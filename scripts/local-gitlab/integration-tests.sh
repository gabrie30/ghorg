#!/bin/bash

set -euo pipefail

# New Go-based integration testing script
# Usage: ./integration-tests.sh <LOCAL_GITLAB_GHORG_DIR> <TOKEN> <GITLAB_URL>

LOCAL_GITLAB_GHORG_DIR=${1:-"${HOME}/ghorg"}
TOKEN=${2:-'password'}
GITLAB_URL=${3:-'http://gitlab.example.com'}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_RUNNER_DIR="${SCRIPT_DIR}/test-runner"
CONFIG_PATH="${SCRIPT_DIR}/configs/test-scenarios.json"

echo "Starting GitLab integration tests with Go-based test runner..."
echo "GitLab URL: ${GITLAB_URL}"
echo "Ghorg Dir: ${LOCAL_GITLAB_GHORG_DIR}"
echo "Config: ${CONFIG_PATH}"

# Build the test runner if it doesn't exist or if source files are newer
TEST_RUNNER_BINARY="${TEST_RUNNER_DIR}/gitlab-test-runner"

# Force rebuild in CI environments or if binary doesn't exist or is newer
FORCE_BUILD=false
if [[ "${CI:-}" == "true" ]] || [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "CI environment detected - forcing clean build of test runner..."
    FORCE_BUILD=true
fi

if [[ ! -f "${TEST_RUNNER_BINARY}" ]] || [[ "${TEST_RUNNER_DIR}/main.go" -nt "${TEST_RUNNER_BINARY}" ]] || [[ "${FORCE_BUILD}" == "true" ]]; then
    echo "Building GitLab test runner..."
    cd "${TEST_RUNNER_DIR}"
    
    # Remove existing binary to ensure clean build
    rm -f gitlab-test-runner
    
    go mod download
    go build -o gitlab-test-runner main.go
    
    # Verify binary was created and is executable
    if [[ ! -f "gitlab-test-runner" ]]; then
        echo "Error: Failed to build gitlab-test-runner binary"
        exit 1
    fi
    
    chmod +x gitlab-test-runner
    cd -
fi

# Run the integration tests
echo "Running GitLab integration tests..."
"${TEST_RUNNER_BINARY}" \
    -token="${TOKEN}" \
    -base-url="${GITLAB_URL}" \
    -ghorg-dir="${LOCAL_GITLAB_GHORG_DIR}" \
    -config="${CONFIG_PATH}"

if [[ $? -eq 0 ]]; then
    echo "GitLab integration tests completed successfully!"
else
    echo "GitLab integration tests failed!"
    exit 1
fi
