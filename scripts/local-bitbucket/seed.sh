#!/bin/bash

set -euo pipefail

# Seed Bitbucket Server with test data
# Usage: ./seed.sh <TOKEN> <BITBUCKET_URL> <LOCAL_BITBUCKET_GHORG_DIR>

TOKEN=${1:-'admin'}
BITBUCKET_URL=${2:-'http://bitbucket.example.com:7990'}
LOCAL_BITBUCKET_GHORG_DIR=${3:-"${HOME}/ghorg"}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDER_DIR="${SCRIPT_DIR}/seeder"
CONFIG_PATH="${SCRIPT_DIR}/configs/seed-data.json"

echo "Starting Bitbucket Server seeding process..."
echo "Bitbucket URL: ${BITBUCKET_URL}"
echo "Config: ${CONFIG_PATH}"

# Build the seeder if it doesn't exist or if source files are newer
SEEDER_BINARY="${SEEDER_DIR}/bitbucket-seeder"

# Force rebuild in CI environments or if binary doesn't exist or is newer
FORCE_BUILD=false
if [[ "${CI:-}" == "true" ]] || [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "CI environment detected - forcing clean build of seeder..."
    FORCE_BUILD=true
fi

if [[ ! -f "${SEEDER_BINARY}" ]] || [[ "${SEEDER_DIR}/main.go" -nt "${SEEDER_BINARY}" ]] || [[ "${FORCE_BUILD}" == "true" ]]; then
    echo "Building Bitbucket seeder..."
    cd "${SEEDER_DIR}"

    # Remove existing binary to ensure clean build
    rm -f bitbucket-seeder

    go mod download
    go build -o bitbucket-seeder main.go

    # Verify binary was created and is executable
    if [[ ! -f "bitbucket-seeder" ]]; then
        echo "Error: Failed to build bitbucket-seeder binary"
        exit 1
    fi

    chmod +x bitbucket-seeder
    cd -
fi

# Get credentials from stored files
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin"

if [[ -f "${LOCAL_BITBUCKET_GHORG_DIR}/bitbucket_username" ]]; then
    ADMIN_USERNAME=$(cat "${LOCAL_BITBUCKET_GHORG_DIR}/bitbucket_username")
fi

if [[ -f "${LOCAL_BITBUCKET_GHORG_DIR}/bitbucket_password" ]]; then
    ADMIN_PASSWORD=$(cat "${LOCAL_BITBUCKET_GHORG_DIR}/bitbucket_password")
fi

echo "Using admin credentials: ${ADMIN_USERNAME}:${ADMIN_PASSWORD}"

# Run the seeder
echo "Running Bitbucket seeder..."
"${SEEDER_BINARY}" \
    -username="${ADMIN_USERNAME}" \
    -password="${ADMIN_PASSWORD}" \
    -base-url="${BITBUCKET_URL}" \
    -config="${CONFIG_PATH}"

if [[ $? -eq 0 ]]; then
    echo "Bitbucket seeding completed successfully!"
else
    echo "Bitbucket seeding failed!"
    exit 1
fi

echo "Verifying seeded data..."

# Test that we can access the created workspaces/projects via API
echo "Testing API access to seeded data..."

# Check if we can list projects
if curl -f -s -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" "${BITBUCKET_URL}/rest/api/1.0/projects" > /dev/null; then
    echo "✅ API access to projects successful"
else
    echo "⚠️  Could not access projects API - seeding may have been incomplete"
fi

# Check if we can list repositories for a test workspace
if curl -f -s -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" "${BITBUCKET_URL}/rest/api/1.0/projects/LBP1/repos" > /dev/null; then
    echo "✅ API access to project repositories successful"
else
    echo "⚠️  Could not access project repositories - seeding may have been incomplete"
fi

echo "Bitbucket seeding verification completed!"
echo "You can access the Bitbucket Server at: ${BITBUCKET_URL}"
echo "Admin credentials: ${ADMIN_USERNAME}:${ADMIN_PASSWORD}"
