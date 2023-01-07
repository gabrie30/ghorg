#!/bin/bash

set -e

STOP_GITLAB_WHEN_FINISHED=${1:-'true'}
PERSIST_GITLAB_LOCALLY=${2:-'false'}
GITLAB_IMAGE_TAG=${3:-'latest'}
GITLAB_HOME=${4:-"$HOME/Desktop/ghorg/local-gitlab-ee-data-${GITLAB_IMAGE_TAG}"}
GITLAB_HOST=${5:-'gitlab.example.com'}
GITLAB_URL=${6:-'http://gitlab.example.com'}

if [ "${ENV}" == "ci" ];then
    echo "127.0.0.1 gitlab.example.com" >> /etc/hosts
fi

docker rm gitlab --force --volumes

rm -rf $HOME/Desktop/ghorg/local-gitlab-v15-*

echo ""
echo "To follow gitlab container logs use the following command in a new window"
echo "$ docker logs -f gitlab"
echo ""

./scripts/local-gitlab/run-ee.sh "${GITLAB_IMAGE_TAG}" "${GITLAB_HOME}" "${GITLAB_HOST}" "${PERSIST_GITLAB_LOCALLY}"
./scripts/local-gitlab/get_credentials.sh "${GITLAB_URL}"

if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
    docker rm gitlab --force --volumes
fi
