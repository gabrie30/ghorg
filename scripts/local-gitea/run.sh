#!/bin/bash

set -xv

# Start Gitea Docker container
# https://docs.gitea.io/en-us/install-with-docker/

# make sure 127.0.0.1 gitea.example.com is added to your /etc/hosts

GITEA_IMAGE_TAG=$1
GITEA_HOME=$2
GITEA_HOST=$3
PERSIST_GITEA_LOCALLY=$4

echo ""
echo "Starting fresh install of Gitea, using tag: ${GITEA_IMAGE_TAG}"

if [ "${GHORG_GHA_CI:-}" == "true" ]; then
  GHORG_SSH_PORT=2223
else
  GHORG_SSH_PORT=22
fi

if [ "${PERSIST_GITEA_LOCALLY}" == "true" ];then
  echo "Removing any previous install at path: ${GITEA_HOME}"
  echo ""

  rm -rf "${GITEA_HOME}"
  mkdir -p "${GITEA_HOME}"

  docker run \
    -d=true \
    --hostname "${GITEA_HOST}" \
    --publish 3000:3000 --publish "${GHORG_SSH_PORT}":22 \
    --name gitea \
    -v "${GITEA_HOME}:/data" \
    -e GITEA__database__DB_TYPE=sqlite3 \
    -e GITEA__database__PATH=/data/gitea/gitea.db \
    -e GITEA__repository__ROOT=/data/git/repositories \
    -e GITEA__server__DOMAIN="${GITEA_HOST}" \
    -e GITEA__server__SSH_DOMAIN="${GITEA_HOST}" \
    -e GITEA__server__ROOT_URL="http://${GITEA_HOST}:3000/" \
    -e GITEA__server__HTTP_PORT=3000 \
    -e GITEA__server__SSH_PORT=22 \
    -e GITEA__server__LFS_START_SERVER=true \
    -e GITEA__lfs__PATH=/data/git/lfs \
    -e GITEA__log__ROOT_PATH=/data/gitea/log \
    -e GITEA__log__MODE=console \
    -e GITEA__log__LEVEL=info \
    -e GITEA__service__DISABLE_REGISTRATION=false \
    -e GITEA__service__REQUIRE_SIGNIN_VIEW=false \
    -e GITEA__service__DEFAULT_ALLOW_CREATE_ORGANIZATION=true \
    -e GITEA__service__DEFAULT_ENABLE_TIMETRACKING=true \
    -e GITEA__security__INSTALL_LOCK=true \
    -e GITEA__security__SECRET_KEY=abcd1234567890abcd1234567890abcd1234567890abcd \
    -e GITEA__security__PASSWORD_COMPLEXITY=off \
    -e GITEA__mailer__ENABLED=false \
    -e GITEA__session__PROVIDER=file \
    -e GITEA__picture__DISABLE_GRAVATAR=false \
    -e GITEA__picture__ENABLE_FEDERATED_AVATAR=true \
    -e GITEA__openid__ENABLE_OPENID_SIGNIN=true \
    -e GITEA__openid__ENABLE_OPENID_SIGNUP=true \
    gitea/gitea:"${GITEA_IMAGE_TAG}"
else
  docker run \
    -d=true \
    --hostname "${GITEA_HOST}" \
    --publish 3000:3000 --publish "${GHORG_SSH_PORT}":22 \
    --name gitea \
    -e GITEA__database__DB_TYPE=sqlite3 \
    -e GITEA__database__PATH=/data/gitea/gitea.db \
    -e GITEA__repository__ROOT=/data/git/repositories \
    -e GITEA__server__DOMAIN="${GITEA_HOST}" \
    -e GITEA__server__SSH_DOMAIN="${GITEA_HOST}" \
    -e GITEA__server__ROOT_URL="http://${GITEA_HOST}:3000/" \
    -e GITEA__server__HTTP_PORT=3000 \
    -e GITEA__server__SSH_PORT=22 \
    -e GITEA__server__LFS_START_SERVER=true \
    -e GITEA__lfs__PATH=/data/git/lfs \
    -e GITEA__log__ROOT_PATH=/data/gitea/log \
    -e GITEA__log__MODE=console \
    -e GITEA__log__LEVEL=info \
    -e GITEA__service__DISABLE_REGISTRATION=false \
    -e GITEA__service__REQUIRE_SIGNIN_VIEW=false \
    -e GITEA__service__DEFAULT_ALLOW_CREATE_ORGANIZATION=true \
    -e GITEA__service__DEFAULT_ENABLE_TIMETRACKING=true \
    -e GITEA__security__INSTALL_LOCK=true \
    -e GITEA__security__SECRET_KEY=abcd1234567890abcd1234567890abcd1234567890abcd \
    -e GITEA__security__PASSWORD_COMPLEXITY=off \
    -e GITEA__mailer__ENABLED=false \
    -e GITEA__session__PROVIDER=file \
    -e GITEA__picture__DISABLE_GRAVATAR=false \
    -e GITEA__picture__ENABLE_FEDERATED_AVATAR=true \
    -e GITEA__openid__ENABLE_OPENID_SIGNIN=true \
    -e GITEA__openid__ENABLE_OPENID_SIGNUP=true \
    gitea/gitea:"${GITEA_IMAGE_TAG}"
fi

echo ""
