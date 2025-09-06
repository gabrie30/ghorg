#!/bin/bash

set -euo pipefail

# Bitbucket Server integration test script - AUTOMATED VERSION
# Usage: ./start-auto.sh [STOP_BITBUCKET_WHEN_FINISHED] [PERSIST_BITBUCKET_LOCALLY]

STOP_BITBUCKET_WHEN_FINISHED=${1:-'true'}
PERSIST_BITBUCKET_LOCALLY=${2:-'false'}
BITBUCKET_IMAGE_TAG=${3:-'latest'}
BITBUCKET_HOME=${4:-"${HOME}/ghorg/bitbucket-server-data"}
BITBUCKET_HOST=${5:-'bitbucket.example.com'}
BITBUCKET_URL=${6:-'http://bitbucket.example.com:7990'}
LOCAL_BITBUCKET_GHORG_DIR=${7:-"${HOME}/ghorg"}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Bitbucket Server Integration Test - AUTOMATED ==="
echo "🗑️  Database will be wiped clean on startup"
echo "Stop when finished: ${STOP_BITBUCKET_WHEN_FINISHED}"
echo "Persist locally: ${PERSIST_BITBUCKET_LOCALLY}"
echo "Image tag: ${BITBUCKET_IMAGE_TAG}"
echo "Bitbucket home: ${BITBUCKET_HOME}"
echo "Bitbucket host: ${BITBUCKET_HOST}"
echo "Bitbucket URL: ${BITBUCKET_URL}"
echo "Local ghorg directory: ${LOCAL_BITBUCKET_GHORG_DIR}"
echo ""

# Complete cleanup - remove all data for fresh start
echo "🧹 Performing complete cleanup for fresh test run..."

# Stop and remove existing containers and volumes
if docker ps -aq -f name=bitbucket > /dev/null || docker ps -aq -f name=bitbucket-postgres > /dev/null; then
    echo "Stopping and removing existing Bitbucket and PostgreSQL containers..."
    cd "${SCRIPT_DIR}"
    docker-compose down -v || true
    sleep 2
fi

# Remove persistent data directories
echo "Cleaning up persistent data directories..."
if [ -d "${BITBUCKET_HOME}" ]; then
    echo "Removing Bitbucket home directory: ${BITBUCKET_HOME}"
    rm -rf "${BITBUCKET_HOME}" || true
fi

# Remove any Docker volumes related to Bitbucket
echo "Removing Docker volumes..."
docker volume ls -q | grep -i bitbucket | xargs -r docker volume rm || true
docker volume ls -q | grep -i postgres | xargs -r docker volume rm || true

# Clean up any leftover test directories
echo "Cleaning up test output directories..."
rm -rf "${HOME}/ghorg/local-bitbucket-"* || true
rm -rf "${HOME}/ghorg/bitbucket.example.com" || true
rm -rf "/tmp/bitbucket-custom-path" || true

# Prune unused Docker resources
echo "Pruning unused Docker resources..."
docker system prune -f || true

echo "✅ Cleanup completed - ready for fresh installation"

# Ensure fresh Bitbucket home directory exists
echo "Creating fresh Bitbucket home directory: ${BITBUCKET_HOME}"
mkdir -p "${BITBUCKET_HOME}"

echo "Starting Bitbucket Server..."
cd "${SCRIPT_DIR}"
./run.sh "${BITBUCKET_IMAGE_TAG}" "${BITBUCKET_HOME}" "${BITBUCKET_HOST}" "${PERSIST_BITBUCKET_LOCALLY}"

echo ""
echo "Waiting for Bitbucket Server to be ready..."
OUTPUT_DIR="${HOME}/ghorg-test-output"
mkdir -p "${OUTPUT_DIR}"

# Wait for server to be ready
echo "Waiting for Bitbucket Server to start..."
sleep 10

# Check if server is running
echo "Checking server status..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -s "${BITBUCKET_URL}/status" | grep -q "RUNNING"; then
        echo "✅ Bitbucket Server is running!"
        break
    fi
    echo "⏳ Waiting for server... (attempt $((attempt + 1))/$max_attempts)"
    sleep 5
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Server failed to start within expected time"
    exit 1
fi

# Set up environment variables
export GHORG_SCM_TYPE=bitbucket
export GHORG_SCM_BASE_URL="${BITBUCKET_URL}"
export GHORG_BITBUCKET_USERNAME=admin
export GHORG_BITBUCKET_APP_PASSWORD=admin
export GHORG_INSECURE_BITBUCKET_CLIENT=true

# Save credentials for other scripts
mkdir -p "${OUTPUT_DIR}"
cat > "${OUTPUT_DIR}/bitbucket_credentials.env" << EOF
export GHORG_SCM_TYPE=bitbucket
export GHORG_SCM_BASE_URL=${BITBUCKET_URL}
export GHORG_BITBUCKET_USERNAME=admin
export GHORG_BITBUCKET_APP_PASSWORD=admin
export GHORG_INSECURE_BITBUCKET_CLIENT=true
EOF
echo "✅ Credentials saved"

echo ""
echo "Seeding Bitbucket Server with test data..."
if ./seed.sh; then
    echo "✅ Seeding completed"

    echo ""
    echo "Running integration tests..."
    if ./integration-tests.sh; then
        echo "🎉 All integration tests passed!"
        test_result=0
    else
        echo "❌ Some integration tests failed"
        test_result=1
    fi
else
    echo "⚠️  Seeding failed, but running basic tests..."
    echo ""
    echo "Running basic integration tests..."
    if ./integration-tests.sh; then
        echo "🎉 Basic integration tests passed!"
        test_result=0
    else
        echo "❌ Integration tests failed"
        test_result=1
    fi
fi

# Cleanup
if [ "${STOP_BITBUCKET_WHEN_FINISHED}" = "true" ]; then
    echo ""
    echo "Stopping Bitbucket Server and PostgreSQL (keeping volumes for next run)..."
    cd "${SCRIPT_DIR}"
    docker-compose down || true
fi

echo ""
if [ $test_result -eq 0 ]; then
    echo "🎉 Bitbucket Server integration tests completed successfully!"
else
    echo "❌ Bitbucket Server integration tests failed"
fi

exit $test_result
