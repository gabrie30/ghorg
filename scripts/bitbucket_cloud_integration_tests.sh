#!/bin/bash

set -euo pipefail

echo "Running BitBucket Integration Tests"

cp ./ghorg /usr/local/bin

BITBUCKET_WORKSPACE=ghorg
export GHORG_EXIT_CODE_ON_CLONE_ISSUES=0

# ==========================================
# API Token Authentication Tests (Recommended)
#
# API tokens are the supported replacement for App Passwords, which Bitbucket
# Cloud permanently disables on July 28, 2026. All functional coverage lives
# here so it keeps running after App Passwords are removed.
#
# Note: API token must have all read scopes (Account, Workspace, Projects, Repositories)
# ==========================================
echo ""
echo "=== Testing API Token Authentication (Recommended Method) ==="

if [ -n "${BITBUCKET_API_TOKEN:-}" ] && [ -n "${BITBUCKET_API_EMAIL:-}" ]; then
    # Test 1: Clone using API token with email
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

    # Test 3: Clone to a specific path with API token
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_API_TOKEN}" --bitbucket-api-email="${BITBUCKET_API_EMAIL}" --path=/tmp --output-dir=bb-api-token-test-3 --scm=bitbucket

    if [ -e /tmp/bb-api-token-test-3 ]
    then
        echo "Pass: bitbucket org clone with custom path using API token"
    else
        echo "Fail: bitbucket org clone with custom path using API token"
        exit 1
    fi

    # Test 4: Preserve SCM hostname with API token
    ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_API_TOKEN}" --bitbucket-api-email="${BITBUCKET_API_EMAIL}" --path=/tmp --output-dir=bb-api-token-test-4 --scm=bitbucket --preserve-scm-hostname

    if [ -e /tmp/bitbucket.com/bb-api-token-test-4 ]
    then
        echo "Pass: bitbucket org clone with preserve scm hostname using API token"
    else
        echo "Fail: bitbucket org clone with preserve scm hostname using API token"
        exit 1
    fi
else
    echo "Skipping API token tests: BITBUCKET_API_TOKEN or BITBUCKET_API_EMAIL not set"
fi

# ==========================================
# Legacy App Password Authentication Tests
#
# DEPRECATED: Bitbucket Cloud App Passwords are being phased out via brownouts
# (June 9 - July 27, 2026) and are permanently disabled on July 28, 2026.
# These tests are best-effort only: failures are reported as warnings and do
# NOT fail the suite, so CI keeps passing once App Passwords stop working.
# Functional coverage is exercised by the API token tests above.
# ==========================================
echo ""
echo "=== Testing App Password Authentication (Legacy / Deprecated Method) ==="
echo "WARNING: Bitbucket Cloud App Passwords are deprecated and permanently disabled on July 28, 2026. Migrate to API Tokens."

if [ -n "${BITBUCKET_TOKEN:-}" ] && [ -n "${BITBUCKET_USERNAME:-}" ]; then
    # Test 5: Clone using app password (legacy method)
    if ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --scm=bitbucket --output-dir=bb-app-password-test-1 \
        && [ -e "${HOME}"/ghorg/bb-app-password-test-1 ]
    then
        echo "Pass: bitbucket org clone using app password authentication"
    else
        echo "Warn: bitbucket org clone using app password authentication failed (App Passwords are deprecated, non-fatal)"
    fi
else
    echo "Skipping app password tests: BITBUCKET_TOKEN or BITBUCKET_USERNAME not set"
fi

echo ""
echo "All BitBucket integration tests completed successfully!"
