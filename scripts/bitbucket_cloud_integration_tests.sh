#!/bin/bash

set -euo pipefail

echo "Running BitBucket Integration Tests"

cp ./ghorg /usr/local/bin

BITBUCKET_WORKSPACE=ghorg
export GHORG_EXIT_CODE_ON_CLONE_ISSUES=0

# ==========================================
# API Token Authentication Tests (Recommended)
# Note: API token must have all read scopes (Account, Workspace, Projects, Repositories)
# ==========================================
echo ""
echo "=== Testing API Token Authentication (New Method) ==="

# Test 1: Clone using API token with email
if [ -n "${BITBUCKET_API_TOKEN:-}" ] && [ -n "${BITBUCKET_API_EMAIL:-}" ]; then
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_API_TOKEN}" --bitbucket-api-email="${BITBUCKET_API_EMAIL}" --scm=bitbucket --output-dir=bb-api-token-test-1

    if [ -e "${HOME}"/ghorg/bb-api-token-test-1 ]
    then
        echo "Pass: bitbucket org clone using API token authentication"
    else
        echo "Fail: bitbucket org clone using API token authentication"
        exit 1
    fi

    # Test 2: Clone using API token with environment variables
    export GHORG_BITBUCKET_API_TOKEN="${BITBUCKET_API_TOKEN}"
    export GHORG_BITBUCKET_API_EMAIL="${BITBUCKET_API_EMAIL}"
    ghorg clone $BITBUCKET_WORKSPACE --scm=bitbucket --output-dir=bb-api-token-test-2

    if [ -e "${HOME}"/ghorg/bb-api-token-test-2 ]
    then
        echo "Pass: bitbucket org clone using API token via environment variables"
    else
        echo "Fail: bitbucket org clone using API token via environment variables"
        exit 1
    fi

    # Clean up env vars for subsequent tests
    unset GHORG_BITBUCKET_API_TOKEN
    unset GHORG_BITBUCKET_API_EMAIL
else
    echo "Skipping API token tests: BITBUCKET_API_TOKEN or BITBUCKET_API_EMAIL not set"
fi

# ==========================================
# Legacy App Password Authentication Tests
# ==========================================
echo ""
echo "=== Testing App Password Authentication (Legacy Method) ==="

if [ -n "${BITBUCKET_TOKEN:-}" ] && [ -n "${BITBUCKET_USERNAME:-}" ]; then
    # Test 3: Clone using app password (legacy method)
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --scm=bitbucket --output-dir=bb-app-password-test-1

    if [ -e "${HOME}"/ghorg/bb-app-password-test-1 ]
    then
        echo "Pass: bitbucket org clone using app password authentication"
    else
        echo "Fail: bitbucket org clone using app password authentication"
        exit 1
    fi

    # Test 4: Clone to a specific path with app password
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --path=/tmp --output-dir=bb-app-password-test-2 --scm=bitbucket

    if [ -e /tmp/bb-app-password-test-2 ]
    then
        echo "Pass: bitbucket org clone with custom path using app password"
    else
        echo "Fail: bitbucket org clone with custom path using app password"
        exit 1
    fi

    # Test 5: Preserve SCM hostname with app password
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --path=/tmp --output-dir=bb-app-password-test-3 --scm=bitbucket --preserve-scm-hostname

    if [ -e /tmp/bitbucket.com/bb-app-password-test-3 ]
    then
        echo "Pass: bitbucket org clone with preserve scm hostname"
    else
        echo "Fail: bitbucket org clone with preserve scm hostname"
        exit 1
    fi
else
    echo "Skipping app password tests: BITBUCKET_TOKEN or BITBUCKET_USERNAME not set"
fi

echo ""
echo "All BitBucket integration tests completed successfully!"
