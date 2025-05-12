#!/bin/bash

set -euo pipefail

echo "Running BitBucket Integration Tests"

cp ./ghorg /usr/local/bin

BITBUCKET_WORKSPACE=ghorg

# clone an org with no config file
ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --scm=bitbucket --output-dir=bb-test-1

if [ -e "${HOME}"/ghorg/bb-test-1 ]
then
    echo "Pass: bitbucket org clone using no configuration file"
else
    echo "Fail: bitbucket org clone using no configuration file"
    exit 1
fi

# clone an org with no config file to a specific path
ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --path=/tmp --output-dir=testing_output_dir --scm=bitbucket

if [ -e /tmp/testing_output_dir ]
then
    echo "Pass: bitbucket org clone, commandline flags take overwrite conf.yaml"
else
    echo "Fail: bitbucket org clone, commandline flags take overwrite conf.yaml"
    exit 1
fi

# preserve scm hostname
ghorg clone $BITBUCKET_WORKSPACE --token="${BITBUCKET_TOKEN}" --bitbucket-username="${BITBUCKET_USERNAME}" --path=/tmp --output-dir=testing_output_dir --scm=bitbucket --preserve-scm-hostname

if [ -e /tmp/testing_output_dir ]
then
    echo "Pass: bitbucket org clone, preserve scm hostname"
else
    echo "Fail: bitbucket org clone, preserve scm hostname"
    exit 1
fi
