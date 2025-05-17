#!/bin/bash

set -xv

STOP_GITLAB_WHEN_FINISHED=${1:-'true'}
PERSIST_GITLAB_LOCALLY=${2:-'false'}
GITLAB_IMAGE_TAG=${3:-'latest'}
GITLAB_HOME=${4:-"$HOME/ghorg/local-gitlab-ee-data-${GITLAB_IMAGE_TAG}"}
GITLAB_HOST=${5:-'gitlab.example.com'}
GITLAB_URL=${6:-'http://gitlab.example.com'}
LOCAL_GITLAB_GHORG_DIR=${7:-"${HOME}/ghorg"}
API_TOKEN="password"

if [ "${ENV}" == "ci" ];then
    echo "127.0.0.1 gitlab.example.com" >> /etc/hosts
fi

docker rm gitlab --force --volumes

rm -rf $HOME/ghorg/local-gitlab-*

echo ""
echo "To follow gitlab container logs use the following command in a new window"
echo "$ docker logs -f gitlab"
echo ""

./scripts/local-gitlab/run-ee.sh "${GITLAB_IMAGE_TAG}" "${GITLAB_HOME}" "${GITLAB_HOST}" "${PERSIST_GITLAB_LOCALLY}"
if [ $? -ne 0 ]; then
    exit 1
fi

./scripts/local-gitlab/get_credentials.sh "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
if [ $? -ne 0 ]; then
    exit 1
fi

# seed new instance using
./scripts/local-gitlab/seed.sh "${API_TOKEN}" "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
if [ $? -ne 0 ]; then
    exit 1
fi

./scripts/local-gitlab/integration-tests.sh "${LOCAL_GITLAB_GHORG_DIR}" "${TOKEN}" "${GITLAB_URL}"
if [ $? -ne 0 ]; then
    exit 1
fi

if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
    docker rm gitlab --force --volumes
fi
