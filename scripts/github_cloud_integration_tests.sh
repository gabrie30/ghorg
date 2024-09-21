#!/bin/bash

set -euo pipefail

echo "Running GitHub Integration Tests"

cp ./ghorg /usr/local/bin

GITHUB_ORG=forcepushtoproduction
GHORG_TEST_REPO=ghorg-ci-test
GHORG_TEST_SELF_PRIVATE_REPO=ghorg_testing_private
REPO_WITH_TESTING_TOPIC=ghorg-repo-with-topic-of-testing
GITHUB_SELF=gabrie30

ghorg version

# clone an org with no config file
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN

if [ -e $HOME/ghorg/$GITHUB_ORG/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone using no configuration file"
else
    echo "Fail: github org clone using no configuration file"
    exit 1
fi

# clone an org preserving scm hostname
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --preserve-scm-hostname

if [ -e $HOME/ghorg/github.com/$GITHUB_ORG/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone preserving scm hostname"
else
    echo "Fail: github org clone preserving scm hostname"
    exit 1
fi

# clone an org with no config file to a specific path
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --path=/tmp --output-dir=testing_output_dir

if [ -e /tmp/testing_output_dir/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone, commandline flags take overwrite conf.yaml"
else
    echo "Fail: github org clone, commandline flags take overwrite conf.yaml"
    exit 1
fi

# user cloning selfs private repo
ghorg clone $GITHUB_SELF --clone-type=user --topics=ghogtestprivate --token=$GITHUB_TOKEN --path=/tmp --output-dir=testing_self_private_repo

if [ -e /tmp/testing_self_private_repo/$GHORG_TEST_SELF_PRIVATE_REPO ]
then
    echo "Pass: github self private repos clone"
else
    echo "Fail: github self private repos clone"
    exit 1
fi

# clone an org with configuration file set by config flag
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --config=$PWD/scripts/testing_confs/alternative_clone_path_conf.yaml

if [ -e /tmp/path_from_configuration_file/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone, alternative configuration file path"
else
    echo "Fail: github org clone, alternative configuration file path"
    exit 1
fi

mkdir -p $HOME/.config/ghorg
cp sample-conf.yaml $HOME/.config/ghorg/conf.yaml

# hack to allow sed to be ran on both mac and ubuntu
sed "s/GHORG_OUTPUT_DIR:/GHORG_OUTPUT_DIR: testing_conf_is_set/g" $HOME/.config/ghorg/conf.yaml >updated_conf.yaml && \
mv $HOME/.config/ghorg/conf.yaml $HOME/.config/ghorg/conf-bak.yaml && \
mv updated_conf.yaml $HOME/.config/ghorg/conf.yaml

# clone an org with configuration set at the default location using latest sample-config.yaml
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN

if [ -e $HOME/ghorg/testing_conf_is_set/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone, using config file in default location"
else
    echo "Fail: github org clone, using config file in default location"
    exit 1
fi

# Move back to original conf but keep updated_conf if we want to use it again
mv $HOME/.config/ghorg/conf.yaml $HOME/.config/ghorg/updated_conf.yaml
mv $HOME/.config/ghorg/conf-bak.yaml $HOME/.config/ghorg/conf.yaml

# RECLONE BASIC

# hack to allow sed to be ran on both mac and ubuntu
sed "s/XTOKEN/${GITHUB_TOKEN}/g" $PWD/scripts/testing_confs/reclone-basic.yaml > $PWD/scripts/testing_confs/updated_reclone.yaml

ghorg reclone --reclone-path=$PWD/scripts/testing_confs/updated_reclone.yaml

if [ -e /tmp/testing_reclone_with_tag/$REPO_WITH_TESTING_TOPIC ]
then
    echo "Pass: github reclone testing-topic-in-tmp-dir file exists"
else
    echo "Fail: github reclone testing-topic-in-tmp-dir file does not exist"
    exit 1
fi

COUNT=$(ls /tmp/testing_reclone_with_tag | wc -l)

if [ "${COUNT}" -eq 1 ]
then
    echo "Pass: github reclone testing_reclone_with_tag"
else
    echo "Fail: github reclone testing_reclone_with_tag too many files found"
    exit 1
fi

if [ -e /tmp/all-repos/$REPO_WITH_TESTING_TOPIC ]
then
    echo "Pass: github reclone all-repos"
else
    echo "Fail: github reclone all-repos"
    exit 1
fi

COUNT=$(ls /tmp/all-repos | wc -l)

if [ "${COUNT}" -ge 3 ]
then
    echo "Pass: github reclone all-repos count"
else
    echo "Fail: github reclone all-repos count too low"
    exit 1
fi

