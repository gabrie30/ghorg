#!/bin/bash

set -ex

TOKEN=${1:-'password'}
GITLAB_URL=${2:-'http://gitlab.example.com'}
LOCAL_GITLAB_GHORG_DIR=${3:-"${HOME}/Desktop/ghorg"}

export GHORG_INSECURE_GITLAB_CLIENT=true

# NOTE run all clones twice to test once for clone then pull



   ##   #      #             ####  #####   ####  #    # #####   ####
  #  #  #      #            #    # #    # #    # #    # #    # #
 #    # #      #      ##### #      #    # #    # #    # #    #  ####
 ###### #      #            #  ### #####  #    # #    # #####       #
 #    # #      #            #    # #   #  #    # #    # #      #    #
 #    # ###### ######        ####  #    #  ####   ####  #       ####



############ CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos

GOT=$( ghorg ls local-gitlab-v15-repos/local-gitlab-group1 | grep -o 'local-gitlab-v15-repos/local-gitlab-group1.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/local-gitlab-group1/baz0
local-gitlab-v15-repos/local-gitlab-group1/baz1
local-gitlab-v15-repos/local-gitlab-group1/baz2
local-gitlab-v15-repos/local-gitlab-group1/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos/local-gitlab-group2 | grep -o 'local-gitlab-v15-repos/local-gitlab-group2.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/local-gitlab-group2/baz0
local-gitlab-v15-repos/local-gitlab-group2/baz1
local-gitlab-v15-repos/local-gitlab-group2/baz2
local-gitlab-v15-repos/local-gitlab-group2/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos/local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-v15-repos/local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos/local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-v15-repos/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-repos/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-repos/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-repos/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED"
exit 1
fi

############ CLONE AND TEST ALL-GROUPS, OUTPUT DIR  ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat

GOT=$( ghorg ls local-gitlab-v15-repos-flat | grep -o 'local-gitlab-v15-repos-flat.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-flat/Monitoring
local-gitlab-v15-repos-flat/local-gitlab-group1_baz0
local-gitlab-v15-repos-flat/local-gitlab-group1_baz1
local-gitlab-v15-repos-flat/local-gitlab-group1_baz2
local-gitlab-v15-repos-flat/local-gitlab-group1_baz3
local-gitlab-v15-repos-flat/local-gitlab-group2_baz0
local-gitlab-v15-repos-flat/local-gitlab-group2_baz1
local-gitlab-v15-repos-flat/local-gitlab-group2_baz2
local-gitlab-v15-repos-flat/local-gitlab-group2_baz3
local-gitlab-v15-repos-flat/subgroup_a_repo_0
local-gitlab-v15-repos-flat/subgroup_a_repo_1
local-gitlab-v15-repos-flat/subgroup_a_repo_2
local-gitlab-v15-repos-flat/subgroup_a_repo_3
local-gitlab-v15-repos-flat/subgroup_b_repo_0
local-gitlab-v15-repos-flat/subgroup_b_repo_1
local-gitlab-v15-repos-flat/subgroup_b_repo_2
local-gitlab-v15-repos-flat/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR"
exit 1
fi

############ CLONE ALL-GROUPS, BACKUP, CLONE WIKI, OUTPUT DIR  ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --clone-wiki --output-dir=local-gitlab-v15-backup

  #####  ### #     #  #####  #       #######    #     #  #####  ####### ######
 #     #  #  ##    # #     # #       #          #     # #     # #       #     #
 #        #  # #   # #       #       #          #     # #       #       #     #
  #####   #  #  #  # #  #### #       #####      #     #  #####  #####   ######
       #  #  #   # # #     # #       #          #     #       # #       #   #
 #     #  #  #    ## #     # #       #          #     # #     # #       #    #
  #####  ### #     #  #####  ####### #######     #####   #####  ####### #     #


############ CLONE SINGLE USER, OUTPUT DIR, PRUNE, PRUNE NO CONFIRM ############
ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm
ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm

 ####### ####### ######     #       ####### #     # ####### #           #####  ######  ####### #     # ######
    #    #     # #     #    #       #       #     # #       #          #     # #     # #     # #     # #     #
    #    #     # #     #    #       #       #     # #       #          #       #     # #     # #     # #     #
    #    #     # ######     #       #####   #     # #####   #          #  #### ######  #     # #     # ######
    #    #     # #          #       #        #   #  #       #          #     # #   #   #     # #     # #
    #    #     # #          #       #         # #   #       #          #     # #    #  #     # #     # #
    #    ####### #          ####### #######    #    ####### #######     #####  #     # #######  #####  #


############ CLONE TOP LEVEL GROUP, BACKUP, OUTPUT DIR ############
ghorg clone local-gitlab-group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup
ghorg clone local-gitlab-group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --backup --output-dir=local-gitlab-v15-group1-backup

############ CLONE TOP LEVEL GROUP, OUTPUT DIR ############
ghorg clone local-gitlab-group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1
ghorg clone local-gitlab-group1 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group1

############ CLONE AND TEST TOP LEVEL GROUP  ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}"
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}"

GOT=$(ghorg ls local-gitlab-group3 | grep -o 'local-gitlab-group3.*')
WANT=$(cat <<EOF
local-gitlab-group3/subgroup_a_repo_0
local-gitlab-group3/subgroup_a_repo_1
local-gitlab-group3/subgroup_a_repo_2
local-gitlab-group3/subgroup_a_repo_3
local-gitlab-group3/subgroup_b_repo_0
local-gitlab-group3/subgroup_b_repo_1
local-gitlab-group3/subgroup_b_repo_2
local-gitlab-group3/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST TOP LEVEL GROUP FAILED"
exit 1
fi

# FAILING
############ CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, PRESERVE DIR, OUTPUT DIR ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-preserve
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-preserve

GOT=$(ghorg ls local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-group3-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, PRESERVE DIR, OUTPUT DIR TEST FAILED"
exit 1
fi

############ CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, OUTPUT DIR ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3

GOT=$(ghorg ls local-gitlab-v15-group3 | grep -o 'local-gitlab-v15-group3.*')
WANT=$(cat <<EOF
local-gitlab-v15-group3/subgroup_a_repo_0
local-gitlab-v15-group3/subgroup_a_repo_1
local-gitlab-v15-group3/subgroup_a_repo_2
local-gitlab-v15-group3/subgroup_a_repo_3
local-gitlab-v15-group3/subgroup_b_repo_0
local-gitlab-v15-group3/subgroup_b_repo_1
local-gitlab-v15-group3/subgroup_b_repo_2
local-gitlab-v15-group3/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, OUTPUT DIR FAILED"
exit 1
fi

############ CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, PRESERVE DIR ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir

GOT=$(ghorg ls local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "TEST GROUP WITH SUBGROUP WITH PRESERVE DIR NO OUTPUT DIR FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3


  #####  #     # ######      #####  ######  ####### #     # ######
 #     # #     # #     #    #     # #     # #     # #     # #     #
 #       #     # #     #    #       #     # #     # #     # #     #
  #####  #     # ######     #  #### ######  #     # #     # ######
       # #     # #     #    #     # #   #   #     # #     # #
 #     # #     # #     #    #     # #    #  #     # #     # #
  #####   #####  ######      #####  #     # #######  #####  #


############ CLONE AND TEST SUBGROUP WITH NESTED SUBGROUP  ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}"
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}"

GOT=$(ghorg ls local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-group3/subgroup-a/subgroup_a_repo_3
local-gitlab-group3/subgroup-a/subgroup_b_repo_0
local-gitlab-group3/subgroup-a/subgroup_b_repo_1
local-gitlab-group3/subgroup-a/subgroup_b_repo_2
local-gitlab-group3/subgroup-a/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP WITH NESTED SUBGROUP FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3

############ CLONE AND TEST SUBGROUP WITH NESTED SUBGROUB, PRESERVE DIR ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir

GOT=$(ghorg ls local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP WITH NESTED SUBGROUB, PRESERVE DIR FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3

############ CLONE AND TEST SUBGROUP, PRESERVE DIR, OUTPUT DIR ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-subgroup-a-preserve
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-subgroup-a-preserve

GOT=$(ghorg ls local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3.*')
WANT=$(cat <<EOF
local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-group3-subgroup-a-preserve/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP, PRESERVE DIR, OUTPUT DIR FAILED"
exit 1
fi

############ CLONE AND TEST SUBGROUP, OUTPUT DIR ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3-subgroup-a
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3-subgroup-a

GOT=$(ghorg ls local-gitlab-v15-group3-subgroup-a | grep -o 'local-gitlab-v15-group3-subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_0
local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_1
local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_2
local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_3
local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_0
local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_1
local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_2
local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP, OUTPUT DIR FAILED"
exit 1
fi

    #    #       #             #     #  #####  ####### ######   #####
   # #   #       #             #     # #     # #       #     # #     #
  #   #  #       #             #     # #       #       #     # #
 #     # #       #       ##### #     #  #####  #####   ######   #####
 ####### #       #             #     #       # #       #   #         #
 #     # #       #             #     # #     # #       #    #  #     #
 #     # ####### #######        #####   #####  ####### #     #  #####


############ CLONE AND TEST ALL-USERS, PRESERVE DIR, OUTPUT DIR ############
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
echo "CLONE AND TEST ALL-USERS, PRESERVE DIR, OUTPUT DIR"
exit 1
fi

############ CLONE AND TEST ALL-USERS, OUTPUT DIR ############
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
echo "CLONE AND TEST ALL-USERS, OUTPUT DIR FAILED"
exit 1
fi

echo "INTEGRATOIN TESTS FINISHED..."
