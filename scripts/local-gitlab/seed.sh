#!/bin/bash

set -euo pipefail

# New Go-based seeding script
# Usage: ./seed.sh <TOKEN> <GITLAB_URL> <LOCAL_GITLAB_GHORG_DIR>

TOKEN=$1
GITLAB_URL=$2
LOCAL_GITLAB_GHORG_DIR=${3:-"${HOME}/ghorg"}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEEDER_DIR="${SCRIPT_DIR}/seeder"
CONFIG_PATH="${SCRIPT_DIR}/configs/seed-data.json"

echo "Starting GitLab seeding with Go-based seeder..."
echo "GitLab URL: ${GITLAB_URL}"
echo "Config: ${CONFIG_PATH}"

# Build the seeder if it doesn't exist or if source files are newer
SEEDER_BINARY="${SEEDER_DIR}/gitlab-seeder"
if [[ ! -f "${SEEDER_BINARY}" ]] || [[ "${SEEDER_DIR}/main.go" -nt "${SEEDER_BINARY}" ]]; then
    echo "Building GitLab seeder..."
    cd "${SEEDER_DIR}"
    go mod download
    go build -o gitlab-seeder main.go
    cd -
fi

# Run the seeder
echo "Running GitLab seeder..."
"${SEEDER_BINARY}" \
    -token="${TOKEN}" \
    -base-url="${GITLAB_URL}" \
    -config="${CONFIG_PATH}"

if [[ $? -eq 0 ]]; then
    echo "GitLab seeding completed successfully!"
    echo "Sleeping 5 seconds to ensure all resources are ready..."
    sleep 5
else
    echo "GitLab seeding failed!"
    exit 1
fi
