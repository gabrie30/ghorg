#!/bin/bash

set -euo pipefail

# Refactored GitLab EE integration test script
# Usage: ./start-ee.sh [STOP_GITLAB_WHEN_FINISHED] [PERSIST_GITLAB_LOCALLY] [GITLAB_IMAGE_TAG] [GITLAB_HOME] [GITLAB_HOST] [GITLAB_URL] [LOCAL_GITLAB_GHORG_DIR]

STOP_GITLAB_WHEN_FINISHED=${1:-'true'}
PERSIST_GITLAB_LOCALLY=${2:-'false'}
GITLAB_IMAGE_TAG=${3:-'latest'}
GITLAB_HOME=${4:-"$HOME/ghorg/local-gitlab-ee-data-${GITLAB_IMAGE_TAG}"}
GITLAB_HOST=${5:-'gitlab.example.com'}
GITLAB_URL=${6:-'http://gitlab.example.com'}
LOCAL_GITLAB_GHORG_DIR=${7:-"${HOME}/ghorg"}
API_TOKEN="password"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== GitLab EE Integration Test (Refactored) ==="
echo "Stop when finished: ${STOP_GITLAB_WHEN_FINISHED}"
echo "Persist locally: ${PERSIST_GITLAB_LOCALLY}"
echo "GitLab tag: ${GITLAB_IMAGE_TAG}"
echo "GitLab home: ${GITLAB_HOME}"
echo "GitLab host: ${GITLAB_HOST}"
echo "GitLab URL: ${GITLAB_URL}"
echo "Ghorg dir: ${LOCAL_GITLAB_GHORG_DIR}"

if [ "${ENV:-}" == "ci" ];then
    echo "127.0.0.1 gitlab.example.com" >> /etc/hosts
fi

echo "Stopping and removing any existing GitLab containers..."
docker rm gitlab --force --volumes || true

echo "Cleaning up old data..."
rm -rf "$HOME/ghorg/local-gitlab-*" || true

echo ""
echo "To follow gitlab container logs use the following command in a new window:"
echo "$ docker logs -f gitlab"
echo ""

echo "=== Starting GitLab Container ==="
"${SCRIPT_DIR}/run-ee.sh" "${GITLAB_IMAGE_TAG}" "${GITLAB_HOME}" "${GITLAB_HOST}" "${PERSIST_GITLAB_LOCALLY}"
if [ $? -ne 0 ]; then
    echo "Failed to start GitLab container"
    exit 1
fi

echo "=== Waiting for GitLab to be Ready and Getting Credentials ==="
"${SCRIPT_DIR}/get_credentials.sh" "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to get GitLab credentials"
    exit 1
fi

echo "=== Seeding GitLab Instance (Using Go Seeder) ==="
"${SCRIPT_DIR}/seed.sh" "${API_TOKEN}" "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to seed GitLab instance"
    exit 1
fi

echo "=== Running Integration Tests (Using Go Test Runner) ==="
"${SCRIPT_DIR}/integration-tests.sh" "${LOCAL_GITLAB_GHORG_DIR}" "${API_TOKEN}" "${GITLAB_URL}"
if [ $? -ne 0 ]; then
    echo "Integration tests failed"
    if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
        docker rm gitlab --force --volumes
    fi
    exit 1
fi

echo "=== Integration Tests Completed Successfully ==="

if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
    echo "Stopping and removing GitLab container..."
    docker rm gitlab --force --volumes
    echo "GitLab container stopped and removed"
else
    echo "GitLab container is still running. You can access it at: ${GITLAB_URL}"
    echo "To stop it manually, run: docker stop gitlab && docker rm gitlab"
fi

echo ""
echo "🎉 GitLab EE integration tests completed successfully with refactored framework!"
