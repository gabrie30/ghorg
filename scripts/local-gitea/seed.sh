#!/bin/bash

set -euo pipefail

# Go-based seeding script for Gitea
# Usage: ./seed.sh [API_TOKEN] [GITEA_URL] [LOCAL_GITEA_GHORG_DIR]

LOCAL_GITEA_GHORG_DIR=${3:-"${HOME}/ghorg"}
API_TOKEN=${1:-$(cat "${LOCAL_GITEA_GHORG_DIR}/gitea_token" 2>/dev/null || echo "test-token")}
GITEA_URL=${2:-"http://gitea.example.com:3000"}

# Also read username and password for fallback authentication
ADMIN_USERNAME=$(cat "${LOCAL_GITEA_GHORG_DIR}/gitea_username" 2>/dev/null || echo "testuser")
ADMIN_PASSWORD=$(cat "${LOCAL_GITEA_GHORG_DIR}/gitea_password" 2>/dev/null || echo "testpass")

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDER_DIR="${SCRIPT_DIR}/seeder"
CONFIG_PATH="${SCRIPT_DIR}/configs/seed-data.json"

echo "Starting Gitea seeding with Go-based seeder..."
echo "Gitea URL: ${GITEA_URL}"
echo "Ghorg Dir: ${LOCAL_GITEA_GHORG_DIR}"
echo "Config: ${CONFIG_PATH}"

# Build the seeder if it doesn't exist or if source files are newer
SEEDER_BINARY="${SEEDER_DIR}/gitea-seeder"

# Force rebuild in CI environments or if binary doesn't exist or is newer
FORCE_BUILD=false
if [[ "${CI:-}" == "true" ]] || [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "CI environment detected - forcing clean build of seeder..."
    FORCE_BUILD=true
fi

if [[ ! -f "${SEEDER_BINARY}" ]] || [[ "${SEEDER_DIR}/main.go" -nt "${SEEDER_BINARY}" ]] || [[ "${FORCE_BUILD}" == "true" ]]; then
    echo "Building Gitea seeder..."
    cd "${SEEDER_DIR}"

    # Remove existing binary to ensure clean build
    rm -f gitea-seeder

    go mod download
    go build -o gitea-seeder main.go

    # Verify binary was created and is executable
    if [[ ! -f "gitea-seeder" ]]; then
        echo "Error: Failed to build gitea-seeder binary"
        exit 1
    fi

    chmod +x gitea-seeder
    cd -
fi

# Run the seeder
echo "Seeding Gitea instance..."
echo "Using admin credentials: ${ADMIN_USERNAME}:${ADMIN_PASSWORD}"
echo "Using API token: ${API_TOKEN}"

"${SEEDER_BINARY}" \
    -token="${API_TOKEN}" \
    -username="${ADMIN_USERNAME}" \
    -password="${ADMIN_PASSWORD}" \
    -base-url="${GITEA_URL}" \
    -config="${CONFIG_PATH}"

if [[ $? -eq 0 ]]; then
    echo "Gitea seeding completed successfully!"
else
    echo "Gitea seeding encountered some issues, but may have partially succeeded"
    echo "Check the logs above for details"
    # Don't exit with error since partial success is acceptable for testing
fi
