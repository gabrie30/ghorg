#!/bin/bash

STOP_GITLAB_WHEN_FINISHED=${1:-'true'}
GITLAB_IMAGE_TAG=${2:-'latest'}
GITLAB_HOME=${3:-"$HOME/Desktop/ghorg/local-gitlab-ee-data-${GITLAB_IMAGE_TAG}"}
GITLAB_HOST=${4:-'gitlab.example.com'}
GITLAB_URL=${5:-'http://gitlab.example.com'}

if [ "${ENV}" == "ci" ];then
    echo "127.0.0.1 gitlab.example.com" >> /etc/hosts
fi

if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
    export STOP_GITLAB_WHEN_FINISHED=true
fi

docker stop gitlab
docker rm gitlab

echo ""
echo "To follow gitlab container logs use the following command in a new window"
echo "$ docker logs -f gitlab"
echo ""
./scripts/local-gitlab/run-ee.sh "${GITLAB_IMAGE_TAG}" "${GITLAB_HOME}" "${GITLAB_HOST}"
./scripts/local-gitlab/get_credentials.sh "${GITLAB_URL}"
