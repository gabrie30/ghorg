#!/bin/bash

set -euo pipefail

echo "Running GitHub Integration Tests"

cp ./ghorg /usr/local/bin

GITHUB_ORG=forcePushToProduction
GHORG_TEST_REPO=ghorg-ci-test

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

# clone an org with no config file to a specific path
ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --path=/tmp --output-dir=testing_output_dir

if [ -e /tmp/testing_output_dir/$GHORG_TEST_REPO ]
then
    echo "Pass: github org clone, commandline flags take overwrite conf.yaml"
else
    echo "Fail: github org clone, commandline flags take overwrite conf.yaml"
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
