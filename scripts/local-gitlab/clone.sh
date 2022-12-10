#!/bin/bash

set -ex

TOKEN=$1
GITLAB_URL=$2

export GHORG_INSECURE_GITLAB_CLIENT=true

# run twice, once for clone then pull
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat

ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm
ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1

ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-all-users
