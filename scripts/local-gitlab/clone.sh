#!/bin/bash

set -ex

TOKEN=${1:-'password'}
GITLAB_URL=${2:-'http://gitlab.example.com'}

export GHORG_INSECURE_GITLAB_CLIENT=true

############                                                          ############
############ CLONE AND TEST ALL-GROUPS PRESERVING DIRECTORY STRUCTURE ############
############                                                          ############

# run twice, once for clone then pull
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos

GOT=$( ghorg ls local-gitlab-v15-repos/group1 | grep -o 'local-gitlab-v15-repos/group1.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/group1/baz0
local-gitlab-v15-repos/group1/baz1
local-gitlab-v15-repos/group1/baz2
local-gitlab-v15-repos/group1/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "ALL_GROUPS GROUP1 PRESERVE DIR TEST FAILED"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos/group2 | grep -o 'local-gitlab-v15-repos/group2.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/group2/baz0
local-gitlab-v15-repos/group2/baz1
local-gitlab-v15-repos/group2/baz2
local-gitlab-v15-repos/group2/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "ALL_GROUPS GROUP2 PRESERVE DIR TEST FAILED"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos/group3/subgroup-a | grep -o 'local-gitlab-v15-repos/group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/group3/subgroup-a/subgroup_repo_0
local-gitlab-v15-repos/group3/subgroup-a/subgroup_repo_1
local-gitlab-v15-repos/group3/subgroup-a/subgroup_repo_2
local-gitlab-v15-repos/group3/subgroup-a/subgroup_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "ALL_GROUPS GROUP3/SUBGROUP-A PRESERVE DIR TEST FAILED"
exit 1
fi

############                            ############
############ CLONE AND TEST ALL-GROUPS  ############
############                            ############

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat

GOT=$( ghorg ls local-gitlab-v15-repos-flat | grep -o 'local-gitlab-v15-repos-flat.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-flat/Monitoring
local-gitlab-v15-repos-flat/group1_baz0
local-gitlab-v15-repos-flat/group1_baz1
local-gitlab-v15-repos-flat/group1_baz2
local-gitlab-v15-repos-flat/group1_baz3
local-gitlab-v15-repos-flat/group2_baz0
local-gitlab-v15-repos-flat/group2_baz1
local-gitlab-v15-repos-flat/group2_baz2
local-gitlab-v15-repos-flat/group2_baz3
local-gitlab-v15-repos-flat/subgroup_repo_0
local-gitlab-v15-repos-flat/subgroup_repo_1
local-gitlab-v15-repos-flat/subgroup_repo_2
local-gitlab-v15-repos-flat/subgroup_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "ALL_GROUPS FLAT TEST FAILED"
exit 1
fi

ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm
ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm

ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup

ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1
ghorg clone group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1

############                                                         ############
############ CLONE AND TEST ALL-USERS PRESERVING DIRECTORY STRUCTURE ############
############                                                         ############


ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-all-users-preserve --preserve-dir

GOT=$(ghorg ls local-gitlab-v15-all-users-preserve/root | grep -o 'local-gitlab-v15-all-users-preserve/root.*')
WANT=$(cat <<EOF
local-gitlab-v15-all-users-preserve/root/rootrepos0
local-gitlab-v15-all-users-preserve/root/rootrepos1
local-gitlab-v15-all-users-preserve/root/rootrepos2
local-gitlab-v15-all-users-preserve/root/rootrepos3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "ALL_USERS PRESERVE DIR TEST FAILED"
exit 1
fi

############                          ############
############ CLONE AND TEST ALL-USERS ############
############                          ############

ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-all-users

TEST_ALL_USERS_GOT=$(ghorg ls local-gitlab-v15-all-users | grep -o 'local-gitlab-v15-all-users.*')
TEST_ALL_USERS_WANT=$(cat <<EOF
local-gitlab-v15-all-users/rootrepos0
local-gitlab-v15-all-users/rootrepos1
local-gitlab-v15-all-users/rootrepos2
local-gitlab-v15-all-users/rootrepos3
EOF
)

if [ "${TEST_ALL_USERS_WANT}" != "${TEST_ALL_USERS_GOT}" ]
then
echo "ALL_USERS TEST FAILED"
exit 1
fi
