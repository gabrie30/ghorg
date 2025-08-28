#!/bin/bash

set -euo pipefail

# Gitea integration test script
# Usage: ./start.sh [STOP_GITEA_WHEN_FINISHED] [PERSIST_GITEA_LOCALLY] [GITEA_IMAGE_TAG] [GITEA_HOME] [GITEA_HOST] [GITEA_URL] [LOCAL_GITEA_GHORG_DIR]

STOP_GITEA_WHEN_FINISHED=${1:-'true'}
PERSIST_GITEA_LOCALLY=${2:-'false'}
GITEA_IMAGE_TAG=${3:-'latest'}
GITEA_HOME=${4:-"$HOME/ghorg/local-gitea-data-${GITEA_IMAGE_TAG}"}
GITEA_HOST=${5:-'gitea.example.com'}
GITEA_URL=${6:-'http://gitea.example.com:3000'}
LOCAL_GITEA_GHORG_DIR=${7:-"${HOME}/ghorg"}
API_TOKEN="test-token"  # Default token - will be set during setup

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Gitea Integration Test ==="
echo "Stop when finished: ${STOP_GITEA_WHEN_FINISHED}"
echo "Persist locally: ${PERSIST_GITEA_LOCALLY}"
echo "Gitea tag: ${GITEA_IMAGE_TAG}"
echo "Gitea home: ${GITEA_HOME}"
echo "Gitea host: ${GITEA_HOST}"
echo "Gitea URL: ${GITEA_URL}"
echo "Ghorg dir: ${LOCAL_GITEA_GHORG_DIR}"

if [ "${ENV:-}" == "ci" ];then
    echo "127.0.0.1 gitea.example.com" >> /etc/hosts
fi

echo "Stopping and removing any existing Gitea containers..."
docker rm gitea --force --volumes || true

echo "Cleaning up old data..."
rm -rf "$HOME/ghorg/local-gitea-*" || true

echo ""
echo "To follow gitea container logs use the following command in a new window:"
echo "$ docker logs -f gitea"
echo ""

echo "=== Starting Gitea Container ==="
"${SCRIPT_DIR}/run.sh" "${GITEA_IMAGE_TAG}" "${GITEA_HOME}" "${GITEA_HOST}" "${PERSIST_GITEA_LOCALLY}"
if [ $? -ne 0 ]; then
    echo "Failed to start Gitea container"
    exit 1
fi

echo "=== Waiting for Gitea to be Ready and Setting Up Credentials ==="
"${SCRIPT_DIR}/get_credentials.sh" "${GITEA_URL}" "${LOCAL_GITEA_GHORG_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to set up Gitea credentials"
    exit 1
fi

echo "=== Seeding Gitea Instance (Using Go Seeder) ==="
"${SCRIPT_DIR}/seed.sh" "${API_TOKEN}" "${GITEA_URL}" "${LOCAL_GITEA_GHORG_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to seed Gitea instance"
    exit 1
fi

echo "=== Running Integration Tests (Using Go Test Runner) ==="
"${SCRIPT_DIR}/integration-tests.sh" "${LOCAL_GITEA_GHORG_DIR}" "${API_TOKEN}" "${GITEA_URL}"
if [ $? -ne 0 ]; then
    echo "Integration tests failed"
    if [ "${STOP_GITEA_WHEN_FINISHED}" == "true" ];then
        docker rm gitea --force --volumes
    fi
    exit 1
fi

echo "=== Integration Tests Completed Successfully ==="

if [ "${STOP_GITEA_WHEN_FINISHED}" == "true" ];then
    echo "Stopping and removing Gitea container..."
    docker rm gitea --force --volumes
    echo "Gitea container stopped and removed"
else
    echo "Gitea container is still running. You can access it at: ${GITEA_URL}"
    echo "To stop it manually, run: docker stop gitea && docker rm gitea"
fi

echo ""
echo "🎉 Gitea integration tests completed successfully!"
