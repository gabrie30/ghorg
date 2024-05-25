#! /bin/bash

set -x

# Note: you will need to stop manually
# docker stop gitlab
# docker rm gitlab
# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

# Tags at https://hub.docker.com/r/gitlab/gitlab-ce/tags for CE, EE is only latest

# make sure 127.0.0.1 gitlab.example.com is added to your /etc/hosts

GITLAB_IMAGE_TAG=$1
GITLAB_HOME=$2
GITLAB_HOST=$3
PERSIST_GITLAB_LOCALLY=$4

echo ""
echo "Starting fresh install of GitLab Enterprise Edition, using tag: ${GITLAB_IMAGE_TAG}"


if [ "${PERSIST_GITLAB_LOCALLY}" == "true" ];then
  echo "Removing any previous install at path: ${GITLAB_HOME}"
  echo ""

  rm -rf "${GITLAB_HOME}"

  docker run \
    -d=true \
    --hostname "${GITLAB_HOST}" \
    --publish 443:443 --publish 80:80 --publish 2222:22 \
    --name gitlab \
    gitlab/gitlab-ee:"${GITLAB_IMAGE_TAG}"
else

  docker run \
    -d=true \
    --hostname "${GITLAB_HOST}" \
    --publish 443:443 --publish 80:80 --publish 2222:22 \
    --name gitlab \
    gitlab/gitlab-ee:"${GITLAB_IMAGE_TAG}"
fi

echo ""
