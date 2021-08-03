#!/bin/bash

set -euo pipefail

echo "Running GitHub Integration Tests"

cp ./ghorg /usr/local/bin

GITHUB_ORG=underdeveloped

ghorg version

ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN

if [ -e $HOME/ghorg/$GITHUB_ORG ]
then
    echo "Pass: github org clone"
else
    echo "Fail: github org clone"
    exit 1
fi

ghorg clone $GITHUB_ORG --token=$GITHUB_TOKEN --path=/tmp --output-dir=testing_output_dir

if [ -e /tmp/testing_output_dir ]
then
    echo "Pass: github org clone custom output directories"
else
    echo "Fail: github org clone custom output directories"
    exit 1
fi
