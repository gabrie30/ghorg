#!/bin/bash

set -euo pipefail

echo "Running GitHub Integration Tests"

cp ./ghorg /usr/local/bin

ghorg version

ghorg clone underdeveloped --token=$GITHUB_TOKEN --path=/tmp --output-dir=normal_clone

if [ -e /tmp/normal_clone ]
then
    echo "Pass: Normal Clone"
else
    echo "Fail: Normal Clone"
    exit 1
fi
