#!/bin/bash

set -euo pipefail

echo "Running GitLab Cloud Integration Tests"

cp ./ghorg /usr/local/bin

# https://gitlab.com/gitlab-examples
GITLAB_ORG=gitlab-examples
GITLAB_SUB_GROUP=wayne-enterprises

ghorg clone $GITLAB_ORG --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=examples-flat

if [ -e "${HOME}"/ghorg/examples-flat/microservice ]
then
    echo "Pass: gitlab org clone flat file"
else
    echo "Fail: gitlab org clone flat file"
    exit 1
fi

ghorg clone $GITLAB_ORG --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=examples --preserve-dir

if [ -e "${HOME}"/ghorg/examples/"${GITLAB_SUB_GROUP}"/wayne-industries/microservice ]
then
    echo "Pass: gitlab org clone preserve directories"
else
    echo "Fail: gitlab org clone preserve directories"
    exit 1
fi

ghorg clone $GITLAB_ORG/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab

if [ -e "${HOME}"/ghorg/"${GITLAB_SUB_GROUP}"/microservice ]
then
    echo "Pass: gitlab subgroup clone flat file"
else
    echo "Fail: gitlab subgroup clone flat file"
    exit 1
fi

ghorg clone $GITLAB_ORG --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir

if [ -e "${HOME}"/ghorg/"${GITLAB_SUB_GROUP}"/wayne-industries/microservice ]
then
    echo "Pass: gitlab subgroup clone preserve directories"
else
    echo "Fail: gitlab subgroup clone preserve directories"
    exit 1
fi
