#!/bin/bash

set -xv

LOCAL_GITLAB_GHORG_DIR=${1:-"${HOME}/ghorg"}
TOKEN=${2:-'password'}
GITLAB_URL=${3:-'http://gitlab.example.com'}


# Delete all folders that start with local-gitlab-v15- in the LOCAL_GITLAB_GHORG_DIR
for dir in "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-*; do
    if [ -d "$dir" ]; then
        rm -rf "$dir"
    fi
done


export GHORG_INSECURE_GITLAB_CLIENT=true
# export GHORG_DEBUG=true

# NOTE run all clones twice to test once for clone then pull



   ##   #      #             ####  #####   ####  #    # #####   ####
  #  #  #      #            #    # #    # #    # #    # #    # #
 #    # #      #      ##### #      #    # #    # #    # #    #  ####
 ###### #      #            #  ### #####  #    # #    # #####       #
 #    # #      #            #    # #   #  #    # #    # #      #    #
 #    # ###### ######        ####  #    #  ####   ####  #       ####



############ CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS ############
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
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED local-gitlab-group1"
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
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED local-gitlab-group2"
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
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR TEST FAILED local-gitlab-group3/subgroup-a"
exit 1
fi

############ CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos-snippets --clone-snippets
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos-snippets --clone-snippets

GOT=$( ghorg ls local-gitlab-v15-repos-snippets | grep -o 'local-gitlab-v15-repos-snippets.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-snippets/_ghorg_root_level_snippets
local-gitlab-v15-repos-snippets/local-gitlab-group1
local-gitlab-v15-repos-snippets/local-gitlab-group2
local-gitlab-v15-repos-snippets/local-gitlab-group3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS TEST FAILED local-gitlab-group1"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos-snippets/local-gitlab-group1 | grep -o 'local-gitlab-v15-repos-snippets/local-gitlab-group1.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-snippets/local-gitlab-group1/baz0
local-gitlab-v15-repos-snippets/local-gitlab-group1/baz1
local-gitlab-v15-repos-snippets/local-gitlab-group1/baz2
local-gitlab-v15-repos-snippets/local-gitlab-group1/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS TEST FAILED local-gitlab-group1"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos-snippets/local-gitlab-group2 | grep -o 'local-gitlab-v15-repos-snippets/local-gitlab-group2.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz0
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz0.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz1
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz1.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz2
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz2.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz3
local-gitlab-v15-repos-snippets/local-gitlab-group2/baz3.snippets
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS TEST FAILED local-gitlab-group2"
exit 1
fi

GOT=$( ghorg ls local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_0.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_1.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_2.snippets
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_3.snippets
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS TEST FAILED local-gitlab-group3/subgroup-a"
exit 1
fi

############ CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS, PERSERVE SCM HOSTNAME ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos-snippets --clone-snippets --preserve-scm-hostname
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --output-dir=local-gitlab-v15-repos-snippets --clone-snippets --preserve-scm-hostname

GOT=$( ghorg ls gitlab.example.com/local-gitlab-v15-repos-snippets | grep -o 'gitlab.example.com/local-gitlab-v15-repos-snippets.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-repos-snippets/_ghorg_root_level_snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS, PRESERVE SCM HOSTNAME TEST FAILED local-gitlab-group1"
exit 1
fi

GOT=$( ghorg ls gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1 | grep -o 'gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1/baz0
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1/baz1
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1/baz2
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group1/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS, PRESERVE SCM HOSTNAME TEST FAILED local-gitlab-group1"
exit 1
fi

GOT=$( ghorg ls gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2 | grep -o 'gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz0
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz0.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz1
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz1.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz2
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz2.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz3
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group2/baz3.snippets
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS, PRESERVE SCM HOSTNAME TEST FAILED local-gitlab-group2"
exit 1
fi

GOT=$( ghorg ls gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a | grep -o 'gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup-b
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_0.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_1.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_2.snippets
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
gitlab.example.com/local-gitlab-v15-repos-snippets/local-gitlab-group3/subgroup-a/subgroup_a_repo_3.snippets
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, OUTPUT DIR, SNIPPETS, PRESERVE SCM HOSTNAME, TEST FAILED local-gitlab-group3/subgroup-a"
exit 1
fi

############ CLONE AND TEST ALL-GROUPS, PRESERVE DIR, NO OUTPUT DIR ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir

GOT=$( ghorg ls gitlab.example.com/local-gitlab-group1 | grep -o 'gitlab.example.com/local-gitlab-group1.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-group1/baz0
gitlab.example.com/local-gitlab-group1/baz1
gitlab.example.com/local-gitlab-group1/baz2
gitlab.example.com/local-gitlab-group1/baz3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, PRESERVE DIR, NO OUTPUT DIR TEST FAILED local-gitlab-group1"
exit 1
fi

############ CLONE AND TEST ALL-GROUPS, OUTPUT DIR  ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-flat

GOT=$( ghorg ls local-gitlab-v15-repos-flat | grep -o 'local-gitlab-v15-repos-flat.*')
WANT=$(cat <<EOF
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

############ CLONE AND TEST ALL-GROUPS, OUTPUT DIR, SNIPPETS ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-all-groups-snippets --clone-snippets
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-repos-all-groups-snippets --clone-snippets

GOT=$( ghorg ls local-gitlab-v15-repos-all-groups-snippets | grep -o 'local-gitlab-v15-repos-all-groups-snippets.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-all-groups-snippets/_ghorg_root_level_snippets
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group1_baz0
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group1_baz1
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group1_baz2
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group1_baz3
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz0
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz0.snippets
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz1
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz1.snippets
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz2
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz2.snippets
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz3
local-gitlab-v15-repos-all-groups-snippets/local-gitlab-group2_baz3.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_0
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_0.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_1
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_1.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_2
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_2.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_3
local-gitlab-v15-repos-all-groups-snippets/subgroup_a_repo_3.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_0
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_0.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_1
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_1.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_2
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_2.snippets
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_3
local-gitlab-v15-repos-all-groups-snippets/subgroup_b_repo_3.snippets
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR"
exit 1
fi

########### CLONE AND TEST ALL-GROUPS, OUTPUT DIR, WIKI  ############
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-wiki --output-dir=local-gitlab-v15-repos-flat-wiki
ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-wiki --output-dir=local-gitlab-v15-repos-flat-wiki

GOT=$( ghorg ls local-gitlab-v15-repos-flat-wiki | grep -o 'local-gitlab-v15-repos-flat-wiki.*')
WANT=$(cat <<EOF
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz0
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz0.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz1
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz1.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz2
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz2.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz3
local-gitlab-v15-repos-flat-wiki/local-gitlab-group1_baz3.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz0
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz0.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz1
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz1.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz2
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz2.wiki
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz3
local-gitlab-v15-repos-flat-wiki/local-gitlab-group2_baz3.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_0
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_0.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_1
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_1.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_2
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_2.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_3
local-gitlab-v15-repos-flat-wiki/subgroup_a_repo_3.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_0
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_0.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_1
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_1.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_2
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_2.wiki
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_3
local-gitlab-v15-repos-flat-wiki/subgroup_b_repo_3.wiki
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR, WIKI"
exit 1
fi

# ########### CLONE AND TEST ALL-GROUPS, OUTPUT DIR, WIKI, SNIPPETS  ############
# ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-wiki --clone-snippets --output-dir=local-gitlab-v15-repos-flat-wiki-snippets
# ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-wiki --clone-snippets --output-dir=local-gitlab-v15-repos-flat-wiki-snippets

# GOT=$( ghorg ls local-gitlab-v15-repos-flat-wiki-snippets | grep -o 'local-gitlab-v15-repos-flat-wiki-snippets.*')
# WANT=$(cat <<EOF
# local-gitlab-v15-repos-flat-wiki-snippets/_ghorg_root_level_snippets
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz0
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz0.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz1
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz1.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz2
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz2.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz3
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group1_baz3.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz0
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz0.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz0.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz1
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz1.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz1.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz2
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz2.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz2.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz3
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz3.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/local-gitlab-group2_baz3.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_0
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_0.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_0.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_1
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_1.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_1.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_2
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_2.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_2.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_3
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_3.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_a_repo_3.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_0
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_0.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_0.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_1
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_1.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_1.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_2
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_2.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_2.wiki
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_3
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_3.snippets
# local-gitlab-v15-repos-flat-wiki-snippets/subgroup_b_repo_3.wiki
# EOF
# )

# if [ "${WANT}" != "${GOT}" ]
# then
# echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR, WIKI, AND SNIPPETS"
# exit 1
# fi

# ############ CLONE AND TEST ALL-GROUPS, OUTPUT DIR, SNIPPETS, ROOT LEVEL  ############
# ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --clone-snippets --output-dir=local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups
# ghorg clone all-groups --scm=gitlab --base-url="${GITLAB_URL}" --token="$TOKEN" --preserve-dir --clone-snippets --output-dir=local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups

# # Test root level snippets
# GOT=$( ghorg ls local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups/_ghorg_root_level_snippets | grep -o 'local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups.*')
# WANT=$(cat <<EOF
# local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups/_ghorg_root_level_snippets/snippet1-1
# local-gitlab-v15-snippets-preserve-dir-output-dir-all-groups/_ghorg_root_level_snippets/snippet2-2
# EOF
# )

# if [ "${WANT}" != "${GOT}" ]
# then
# echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR, SNIPPETS, ROOT LEVEL FAILED"
# exit 1
# fi

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

# ############ CLONE SINGLE USER, OUTPUT DIR, SNIPPETS ############
# ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-snippets --output-dir=local-gitlab-v15-root-user-repos-snippets
# ghorg clone root --clone-type=user --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --clone-snippets --output-dir=local-gitlab-v15-root-user-repos-snippets

# # Test root level snippets
# GOT=$( ghorg ls local-gitlab-v15-root-user-repos-snippets | grep -o 'local-gitlab-v15-root-user-repos-snippets.*')
# WANT=$(cat <<EOF
# local-gitlab-v15-root-user-repos-snippets/rootrepos0
# local-gitlab-v15-root-user-repos-snippets/rootrepos1
# local-gitlab-v15-root-user-repos-snippets/rootrepos1.snippets
# local-gitlab-v15-root-user-repos-snippets/rootrepos2
# local-gitlab-v15-root-user-repos-snippets/rootrepos3
# EOF
# )

# if [ "${WANT}" != "${GOT}" ]
# then
# echo "CLONE AND TEST ALL-GROUPS, OUTPUT DIR, SNIPPETS, ROOT LEVEL FAILED"
# exit 1
# fi

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
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-top-level-group
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-top-level-group

GOT=$(ghorg ls local-gitlab-v15-top-level-group | grep -o 'local-gitlab-v15-top-level-group.*')
WANT=$(cat <<EOF
local-gitlab-v15-top-level-group/subgroup_a_repo_0
local-gitlab-v15-top-level-group/subgroup_a_repo_1
local-gitlab-v15-top-level-group/subgroup_a_repo_2
local-gitlab-v15-top-level-group/subgroup_a_repo_3
local-gitlab-v15-top-level-group/subgroup_b_repo_0
local-gitlab-v15-top-level-group/subgroup_b_repo_1
local-gitlab-v15-top-level-group/subgroup_b_repo_2
local-gitlab-v15-top-level-group/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST TOP LEVEL GROUP FAILED"
exit 1
fi

############ CLONE AND TEST TOP LEVEL GROUP, PRESERVE SCM HOSTNAME ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-top-level-group --preserve-scm-hostname
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-top-level-group --preserve-scm-hostname

GOT=$(ghorg ls gitlab.example.com/local-gitlab-v15-top-level-group | grep -o 'gitlab.example.com/local-gitlab-v15-top-level-group.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_a_repo_0
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_a_repo_1
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_a_repo_2
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_a_repo_3
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_b_repo_0
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_b_repo_1
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_b_repo_2
gitlab.example.com/local-gitlab-v15-top-level-group/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST TOP LEVEL GROUP PRESERVE SCM HOSTNAME FAILED"
exit 1
fi

############ CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, PRESERVE DIR, OUTPUT DIR ############
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-preserve
ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-group3-preserve

GOT=$(ghorg ls local-gitlab-v15-group3-preserve/subgroup-a | grep -o 'local-gitlab-v15-group3-preserve/subgroup-a.*')
WANT=$(cat <<EOF
local-gitlab-v15-group3-preserve/subgroup-a/subgroup-b
local-gitlab-v15-group3-preserve/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-group3-preserve/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-group3-preserve/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-group3-preserve/subgroup-a/subgroup_a_repo_3
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

# ############ CLONE AND TEST TOP LEVEL GROUP WITH NESTED SUBGROUP, PRESERVE DIR, SNIPPETS ############
# ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --clone-snippets --output-dir=local-gitlab-v15-group-3-perserve-snippets
# ghorg clone local-gitlab-group3 --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --clone-snippets --output-dir=local-gitlab-v15-group-3-perserve-snippets

# GOT=$(ghorg ls local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-group3/subgroup-a.*')
# WANT=$(cat <<EOF
# local-gitlab-group3/subgroup-a/subgroup-b
# local-gitlab-group3/subgroup-a/subgroup_a_repo_0
# local-gitlab-group3/subgroup-a/subgroup_a_repo_1
# local-gitlab-group3/subgroup-a/subgroup_a_repo_2
# local-gitlab-group3/subgroup-a/subgroup_a_repo_3
# EOF
# )

# if [ "${WANT}" != "${GOT}" ]
# then
# echo "TEST GROUP WITH SUBGROUP WITH PRESERVE DIR OUTPUT DIR SNIPPETS FAILED"
# exit 1
# fi


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

############ CLONE AND TEST SUBGROUP, NESTED SUBGROUB, OUTPUT DIR ############
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
echo "CLONE AND TEST SUBGROUP, NESTED SUBGROUB, OUTPUT DIR FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3

############ CLONE AND TEST SUBGROUP, NESTED SUBGROUB, OUTPUT DIR, PRESERVE SCM HOSTNAME ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3-subgroup-a --preserve-scm-hostname
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-group3-subgroup-a --preserve-scm-hostname

GOT=$(ghorg ls gitlab.example.com/local-gitlab-v15-group3-subgroup-a | grep -o 'gitlab.example.com/local-gitlab-v15-group3-subgroup-a.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_0
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_1
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_2
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_a_repo_3
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_0
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_1
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_2
gitlab.example.com/local-gitlab-v15-group3-subgroup-a/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP, NESTED SUBGROUB, OUTPUT DIR, PRESERVE SCM HOSTNAME FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3

############ CLONE AND TEST SUBGROUP, NESTED SUBGROUB, NO OUTPUT DIR, PRESERVE SCM HOSTNAME ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-scm-hostname
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-scm-hostname

GOT=$(ghorg ls gitlab.example.com/local-gitlab-group3/subgroup-a | grep -o 'gitlab.example.com/local-gitlab-group3/subgroup-a.*')
WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_b_repo_0
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_b_repo_1
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_b_repo_2
gitlab.example.com/local-gitlab-group3/subgroup-a/subgroup_b_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP, NESTED SUBGROUB, NO OUTPUT DIR, PRESERVE SCM HOSTNAME FAILED"
exit 1
fi

rm -rf "${LOCAL_GITLAB_GHORG_DIR}"/local-gitlab-group3

############ CLONE AND TEST SUBGROUP, NESTED SUBGROUPS, PRESERVE DIR, OUTPUT DIR ############
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-subgroups-preserve-output
ghorg clone local-gitlab-group3/subgroup-a --scm=gitlab --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --output-dir=local-gitlab-v15-subgroups-preserve-output

GOT=$(ghorg ls local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a | grep -o 'local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3.*')
WANT=$(cat <<EOF
local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a/subgroup-b
local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a/subgroup_a_repo_0
local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a/subgroup_a_repo_1
local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a/subgroup_a_repo_2
local-gitlab-v15-subgroups-preserve-output/local-gitlab-group3/subgroup-a/subgroup_a_repo_3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST SUBGROUP, NESTED SUBGROUPS, PRESERVE DIR, OUTPUT DIR  FAILED"
exit 1
fi

    #    #       #             #     #  #####  ####### ######   #####
   # #   #       #             #     # #     # #       #     # #     #
  #   #  #       #             #     # #       #       #     # #
 #     # #       #       ##### #     #  #####  #####   ######   #####
 ####### #       #             #     #       # #       #   #         #
 #     # #       #             #     # #     # #       #    #  #     #
 #     # ####### #######        #####   #####  ####### #     #  #####


############ CLONE AND TEST ALL-USERS, PRESERVE DIR ############
ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir

GOT=$(ghorg ls gitlab.example.com/root | grep -o 'gitlab.example.com/root.*')
WANT=$(cat <<EOF
gitlab.example.com/root/rootrepos0
gitlab.example.com/root/rootrepos1
gitlab.example.com/root/rootrepos2
gitlab.example.com/root/rootrepos3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-USERS, PRESERVE DIR, OUTPUT DIR"
exit 1
fi

############ CLONE AND TEST ALL-USERS, PRESERVE DIR ############
ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --preserve-dir --preserve-scm-hostname

GOT=$(ghorg ls gitlab.example.com/root | grep -o 'gitlab.example.com/root.*')
WANT=$(cat <<EOF
gitlab.example.com/root/rootrepos0
gitlab.example.com/root/rootrepos1
gitlab.example.com/root/rootrepos2
gitlab.example.com/root/rootrepos3
EOF
)

if [ "${WANT}" != "${GOT}" ]
then
echo "CLONE AND TEST ALL-USERS, PRESERVE DIR, OUTPUT DIR"
exit 1
fi

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
local-gitlab-v15-all-users/testuser1-repo
EOF
)

if [ "${TEST_ALL_USERS_WANT}" != "${TEST_ALL_USERS_GOT}" ]
then
echo "CLONE AND TEST ALL-USERS, OUTPUT DIR FAILED"
exit 1
fi

############ CLONE AND TEST ALL-USERS, OUTPUT DIR, SNIPPETS ############
ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-all-users-snippets --clone-snippets

TEST_ALL_USERS_SNIPPETS_GOT=$(ghorg ls local-gitlab-v15-all-users-snippets | grep -o 'local-gitlab-v15-all-users-snippets.*')
TEST_ALL_USERS_SNIPPETS_WANT=$(cat <<EOF
local-gitlab-v15-all-users-snippets/_ghorg_root_level_snippets
local-gitlab-v15-all-users-snippets/rootrepos0
local-gitlab-v15-all-users-snippets/rootrepos1
local-gitlab-v15-all-users-snippets/rootrepos2
local-gitlab-v15-all-users-snippets/rootrepos3
local-gitlab-v15-all-users-snippets/testuser1-repo
local-gitlab-v15-all-users-snippets/testuser1-repo.snippets
EOF
)

if [ "${TEST_ALL_USERS_SNIPPETS_WANT}" != "${TEST_ALL_USERS_SNIPPETS_GOT}" ]
then
echo "CLONE AND TEST ALL-USERS, OUTPUT DIR SNIPPETS FAILED"
exit 1
fi

############ CLONE AND TEST ALL-USERS, OUTPUT DIR, SNIPPETS, PRESERVE SCM HOSTNAME ############
ghorg clone all-users --scm=gitlab --clone-type=user --base-url="${GITLAB_URL}" --token="${TOKEN}" --output-dir=local-gitlab-v15-all-users-snippets --clone-snippets --preserve-scm-hostname

TEST_ALL_USERS_SNIPPETS_GOT=$(ghorg ls gitlab.example.com/local-gitlab-v15-all-users-snippets | grep -o 'gitlab.example.com/local-gitlab-v15-all-users-snippets.*')
TEST_ALL_USERS_SNIPPETS_WANT=$(cat <<EOF
gitlab.example.com/local-gitlab-v15-all-users-snippets/_ghorg_root_level_snippets
gitlab.example.com/local-gitlab-v15-all-users-snippets/rootrepos0
gitlab.example.com/local-gitlab-v15-all-users-snippets/rootrepos1
gitlab.example.com/local-gitlab-v15-all-users-snippets/rootrepos2
gitlab.example.com/local-gitlab-v15-all-users-snippets/rootrepos3
gitlab.example.com/local-gitlab-v15-all-users-snippets/testuser1-repo
gitlab.example.com/local-gitlab-v15-all-users-snippets/testuser1-repo.snippets
EOF
)

if [ "${TEST_ALL_USERS_SNIPPETS_WANT}" != "${TEST_ALL_USERS_SNIPPETS_GOT}" ]
then
echo "CLONE AND TEST ALL-USERS, OUTPUT DIR SNIPPETS, PRESERVE SCM HOSTNAME FAILED"
exit 1
fi

echo "INTEGRATOIN TESTS FINISHED..."
