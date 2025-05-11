#!/bin/bash

set -euo pipefail

echo "Running GitLab Cloud Integration Tests"

cp ./ghorg /usr/local/bin

# https://gitlab.com/gitlab-examples
GITLAB_GROUP=gitlab-examples
GITLAB_SUB_GROUP=wayne-enterprises

GITLAB_GROUP_2=ghorg-test-group

ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=examples-flat

if [ -e "${HOME}"/ghorg/examples-flat/microservice ]
then
    echo "Pass: gitlab org clone flat file"
else
    echo "Fail: gitlab org clone flat file"
    exit 1
fi

# preserve scm hostname
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=examples-flat --preserve-scm-hostname

if [ -e "${HOME}"/ghorg/gitlab.com/examples-flat/microservice ]
then
    echo "Pass: gitlab org clone flat file, preserve scm hostname"
else
    echo "Fail: gitlab org clone flat file, preserve scm hostname"
    exit 1
fi

#
# TOP LEVEL GROUP TESTS
#

# NO FLAGS
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/microservice ]
then
    echo "Pass: gitlab org clone"
    rm -rf "${HOME}/ghorg/gitlab-examples"
else
    echo "Fail: gitlab org clone"
    exit 1
fi

# OUTPUT DIR
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=examples

if [ -e "${HOME}"/ghorg/examples/microservice ]
then
    echo "Pass: gitlab org clone output dir"
    rm -rf "${HOME}/ghorg/examples"
else
    echo "Fail: gitlab org clone output dir"
    exit 1
fi


# PRESERVE DIR
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/wayne-enterprises/wayne-industries/microservice ]
then
    echo "Pass: gitlab org clone preserve dir"
    rm -rf "${HOME}/ghorg/${GITLAB_GROUP}"
else
    echo "Fail: gitlab org clone preserve dir"
    exit 1
fi


# OUTPUT DIR AND PRESERVE DIR
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir --output-dir=group-output-perserve

if [ -e "${HOME}"/ghorg/group-output-perserve/wayne-enterprises/wayne-industries/microservice ]
then
    echo "Pass: gitlab org clone preserve dir, output dir"
    rm -rf "${HOME}/ghorg/${GITLAB_GROUP}"
else
    echo "Fail: gitlab org clone preserve dir, output dir"
    exit 1
fi

# PRUNE AND PRESERVE DIR
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --prune --preserve-dir --prune-no-confirm
git init "${HOME}"/ghorg/"${GITLAB_GROUP}"/wayne-enterprises/wayne-industries/prunable
ghorg clone $GITLAB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --prune --preserve-dir --prune-no-confirm

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/wayne-enterprises/wayne-industries/microservice ] && \
    [ ! -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/wayne-enterprises/wayne-industries/prunable ]
then
    echo "Pass: gitlab org clone preserve dir, prune"
    rm -rf "${HOME}/ghorg/${GITLAB_GROUP}"
else
    echo "Fail: gitlab org clone preserve dir, prune"
    exit 1
fi

# REPO NAME COLLISION
ghorg clone $GITLAB_GROUP_2 --token="${GITLAB_TOKEN}" --scm=gitlab

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP_2}"/_subgroup-1_foobar ]
then
    echo "Pass: gitlab group clone with colliding repo names"
else
    echo "Fail: gitlab group clone with colliding repo names"
    exit 1
fi

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP_2}"/_subgroup-2_foobar ]
then
    echo "Pass: gitlab group clone with colliding repo names"
else
    echo "Fail: gitlab group clone with colliding repo names"
    exit 1
fi

# SNIPPETS
ghorg clone $GITLAB_GROUP_2 --token="${GITLAB_TOKEN}" --scm=gitlab --clone-snippets --preserve-dir

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP_2}"/subgroup-2/foobar.snippets/test-snippet-2-3711655 ]
then
    echo "Pass: gitlab group clone snippet 2 with preserve dir"
else
    echo "Fail: gitlab group clone snippet 2 with preserve dir"
    exit 1
fi

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP_2}"/subgroup-2/foobar.snippets/test-snippet-1-3711654 ]
then
    echo "Pass: gitlab group clone snippet 1 with preserve dir"
else
    echo "Fail: gitlab group clone snippet 1 with preserve dir"
    exit 1
fi

ghorg clone $GITLAB_GROUP_2 --token="${GITLAB_TOKEN}" --scm=gitlab --clone-snippets --preserve-dir --preserve-scm-hostname

if [ -e "${HOME}"/ghorg/gitlab.com/"${GITLAB_GROUP_2}"/subgroup-2/foobar.snippets/test-snippet-2-3711655 ]
then
    echo "Pass: gitlab group clone snippet 2 with preserve dir, preserve scm hostname"
else
    echo "Fail: gitlab group clone snippet 2 with preserve dir, preserve scm hostname"
    exit 1
fi

if [ -e "${HOME}"/ghorg/gitlab.com/"${GITLAB_GROUP_2}"/subgroup-2/foobar.snippets/test-snippet-1-3711654 ]
then
    echo "Pass: gitlab group clone snippet 1 with preserve dir, preserve scm hostname"
else
    echo "Fail: gitlab group clone snippet 1 with preserve dir, preserve scm hostname"
    exit 1
fi

#
# SUBGROUP TESTS
#

# NO FLAGS
ghorg clone $GITLAB_GROUP/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/"${GITLAB_SUB_GROUP}"/microservice ]
then
    echo "Pass: gitlab subgroup clone flat file"
    rm -rf "${HOME}/ghorg/${GITLAB_GROUP}"
else
    echo "Fail: gitlab subgroup clone flat file"
    exit 1
fi

# OUTPUT DIR
ghorg clone $GITLAB_GROUP/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --output-dir=example-output

if [ -e "${HOME}"/ghorg/example-output/microservice ]
then
    echo "Pass: gitlab subgroup output dir"
else
    echo "Fail: gitlab subgroup output dir"
    exit 1
fi

# PRESERVE DIR
ghorg clone $GITLAB_GROUP/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir

if [ -e "${HOME}"/ghorg/"${GITLAB_GROUP}"/"${GITLAB_SUB_GROUP}"/wayne-industries/microservice ]
then
    echo "Pass: gitlab subgroup clone preserve directories"
    rm -rf "${HOME}/ghorg/${GITLAB_GROUP}"
else
    echo "Fail: gitlab subgroup clone preserve directories"
    exit 1
fi

# PRESERVE DIR, PRESERVE SCM HOSTNAME
ghorg clone $GITLAB_GROUP/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir --preserve-scm-hostname

if [ -e "${HOME}"/ghorg/gitlab.com/"${GITLAB_GROUP}"/"${GITLAB_SUB_GROUP}"/wayne-industries/microservice ]
then
    echo "Pass: gitlab subgroup clone preserve directories, preserve scm hostname"
    rm -rf "${HOME}/ghorg/gitlab.com/${GITLAB_GROUP}"
else
    echo "Fail: gitlab subgroup clone preserve directories, preserve scm hostname"
    exit 1
fi

# OUTPUT DIR AND PRESERVE DIR
ghorg clone $GITLAB_GROUP/$GITLAB_SUB_GROUP --token="${GITLAB_TOKEN}" --scm=gitlab --preserve-dir --output-dir=examples-subgroup-preserve-output

if [ -e "${HOME}"/ghorg/examples-subgroup-preserve-output/"${GITLAB_GROUP}"/"${GITLAB_SUB_GROUP}"/wayne-industries/microservice ]
then
    echo "Pass: gitlab subgroup clone preserve directories and output dir"
    rm -rf "${HOME}/ghorg/examples-subgroup-preserve-output"
else
    echo "Fail: gitlab subgroup clone preserve directories and output dir"
    exit 1
fi
