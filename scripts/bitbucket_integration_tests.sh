#!/bin/bash

set -euo pipefail

echo "Running BitBucket Integration Tests"

cp ./ghorg /usr/local/bin

BITBUCKET_WORKSPACE=ghorg

ghorg version

# clone an org with no config file
ghorg clone $BITBUCKET_WORKSPACE --token=$BITBUCKET_TOKEN --bitbucket-username=BITBUCKET_USERNAME

if [ -e $HOME/ghorg/$BITBUCKET_WORKSPACE ]
then
    echo "Pass: bitbucket org clone using no configuration file"
else
    echo "Fail: bitbucket org clone using no configuration file"
    exit 1
fi

# clone an org with no config file to a specific path
ghorg clone $BITBUCKET_WORKSPACE --token=$BITBUCKET_TOKEN --bitbucket-username=BITBUCKET_USERNAME --path=/tmp --output-dir=testing_output_dir

if [ -e /tmp/testing_output_dir ]
then
    echo "Pass: bitbucket org clone, commandline flags take overwrite conf.yaml"
else
    echo "Fail: bitbucket org clone, commandline flags take overwrite conf.yaml"
    exit 1
fi
