#!/bin/bash

set -ex

TOKEN=$1
GITLAB_URL=$2

# run twice, once for clone then pull

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos --insecure-gitlab-client
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos --insecure-gitlab-client

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat --insecure-gitlab-client
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat --insecure-gitlab-client

ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm --insecure-gitlab-client
ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm --insecure-gitlab-client

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup --insecure-gitlab-client
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup --insecure-gitlab-client

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup --insecure-gitlab-client
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup --insecure-gitlab-client

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1 --insecure-gitlab-client
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1 --insecure-gitlab-client

if [ "${STOP_GITLAB_WHEN_FINISHED}" == "true" ];then
    docker rm gitlab --force --volumes
fi
