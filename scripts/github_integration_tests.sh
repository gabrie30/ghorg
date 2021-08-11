#!/bin/bash

set -euo pipefail

echo "Running GitHub Integration Tests"

cp ./ghorg /usr/local/bin

GITHUB_ORG=underdeveloped

ghorg version

sed -i 's/GHORG_OUTPUT_DIR:/GHORG_OUTPUT_DIR: testing_conf_is_set/g' $HOME/.config/ghorg/conf.yaml

ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN

if [ -e $HOME/ghorg/testing_conf_is_set ]
then
    echo "Pass: github org clone using conf.yaml"
else
    echo "Fail: github org clone using conf.yaml"
    exit 1
fi

ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --path=/tmp --output-dir=testing_output_dir

if [ -e /tmp/testing_output_dir ]
then
    echo "Pass: github org clone, commandline flags take overwrite conf.yaml"
else
    echo "Fail: github org clone, commandline flags take overwrite conf.yaml"
    exit 1
fi

sed -i 's/GHORG_OUTPUT_DIR: testing_conf_is_set/GHORG_OUTPUT_DIR:/g' $HOME/.config/ghorg/conf.yaml
