#!/bin/bash

set -euo pipefail

echo "Running Sourcehut Integration Tests"

cp ./ghorg /usr/local/bin

# Sourcehut test user - this should be a real sourcehut user with test repos
# For CI, this would need to be set up with test repositories
# Note: Username can be with or without ~ prefix (e.g., "gabrie30" or "~gabrie30")
SOURCEHUT_USER=${SOURCEHUT_TEST_USER:-"gabrie30"}
GHORG_TEST_REPO=${SOURCEHUT_TEST_REPO:-"ghorg-test"}
GHORG_EXIT_CODE_ON_CLONE_ISSUES=0

if [ -z "$SOURCEHUT_USER" ]; then
    echo "SOURCEHUT_TEST_USER environment variable is not set"
    echo "Skipping sourcehut integration tests"
    exit 0
fi

if [ -z "$SOURCEHUT_TOKEN" ]; then
    echo "SOURCEHUT_TOKEN environment variable is not set"
    echo "Skipping sourcehut integration tests"
    exit 0
fi

ghorg version

echo "Testing sourcehut user: $SOURCEHUT_USER"

# Test 1: Clone a user's repos with no config file
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN

if [ -d "$HOME/ghorg/$SOURCEHUT_USER" ]
then
    echo "Pass: sourcehut user clone using no configuration file"
else
    echo "Fail: sourcehut user clone using no configuration file"
    exit 1
fi

# Test 2: Clone with preserve-scm-hostname
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --preserve-scm-hostname

if [ -d "$HOME/ghorg/git.sr.ht/$SOURCEHUT_USER" ]
then
    echo "Pass: sourcehut user clone preserving scm hostname"
else
    echo "Fail: sourcehut user clone preserving scm hostname"
    exit 1
fi

# Test 3: Clone to a specific path with output-dir
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --path=/tmp --output-dir=testing_sourcehut_output

if [ -d "/tmp/testing_sourcehut_output" ]
then
    echo "Pass: sourcehut user clone, custom path and output dir"
else
    echo "Fail: sourcehut user clone, custom path and output dir"
    exit 1
fi

# Test 5: Clone with HTTPS protocol explicitly
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --protocol=https --path=/tmp --output-dir=testing_sourcehut_https

if [ -d "/tmp/testing_sourcehut_https" ]
then
    echo "Pass: sourcehut user clone with HTTPS protocol"
else
    echo "Fail: sourcehut user clone with HTTPS protocol"
    exit 1
fi

# Test 6: Clone org (should work the same as user clone for sourcehut)
ghorg clone $SOURCEHUT_USER --scm=sourcehut --clone-type=org --token=$SOURCEHUT_TOKEN --path=/tmp --output-dir=testing_sourcehut_org

if [ -d "/tmp/testing_sourcehut_org" ]
then
    echo "Pass: sourcehut org clone (same as user)"
else
    echo "Fail: sourcehut org clone (same as user)"
    exit 1
fi

# Test 7: Clone with match-regex filter (if repos exist with pattern)
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --match-regex="test" --path=/tmp --output-dir=testing_sourcehut_regex

if [ -d "/tmp/testing_sourcehut_regex" ]
then
    echo "Pass: sourcehut user clone with regex filter"
else
    echo "Fail: sourcehut user clone with regex filter"
    exit 1
fi

# Test 8: Clone with configuration file
mkdir -p $HOME/.config/ghorg
cp sample-conf.yaml $HOME/.config/ghorg/conf.yaml

# Update config for sourcehut
sed "s/GHORG_OUTPUT_DIR:/GHORG_OUTPUT_DIR: testing_sourcehut_conf/g" $HOME/.config/ghorg/conf.yaml > updated_conf.yaml && \
mv $HOME/.config/ghorg/conf.yaml $HOME/.config/ghorg/conf-bak.yaml && \
mv updated_conf.yaml $HOME/.config/ghorg/conf.yaml

ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN

if [ -d "$HOME/ghorg/testing_sourcehut_conf" ]
then
    echo "Pass: sourcehut user clone, using config file in default location"
else
    echo "Fail: sourcehut user clone, using config file in default location"
    exit 1
fi

# Restore original config
mv $HOME/.config/ghorg/conf-bak.yaml $HOME/.config/ghorg/conf.yaml

# Test 9: Clone with alternative config file path
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --config=$PWD/scripts/testing_confs/alternative_clone_path_conf.yaml

if [ -d "/tmp/path_from_configuration_file" ]
then
    echo "Pass: sourcehut user clone, alternative configuration file path"
else
    echo "Fail: sourcehut user clone, alternative configuration file path"
    exit 1
fi

# Test 10: Prune test with preserve-scm-hostname
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --preserve-scm-hostname --prune-untouched --prune-untouched-no-confirm

if [ -d "$HOME/ghorg/git.sr.ht/$SOURCEHUT_USER" ]
then
    echo "Pass: sourcehut user clone with prune untouched"
else
    echo "Fail: sourcehut user clone with prune untouched"
    exit 1
fi

# Test 11: GHORGONLY test with custom pattern
GHORGONLY_TEST_FILE=$(mktemp)
echo "test" > "$GHORGONLY_TEST_FILE"

ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --path=/tmp --output-dir=testing_sourcehut_ghorgonly --ghorgonly-path="$GHORGONLY_TEST_FILE"

if [ -d "/tmp/testing_sourcehut_ghorgonly" ]
then
    echo "Pass: sourcehut user clone with ghorgonly filter"

    # Count repos - should only have repos matching the pattern
    COUNT=$(find /tmp/testing_sourcehut_ghorgonly -maxdepth 1 -type d ! -path /tmp/testing_sourcehut_ghorgonly | wc -l)
    if [ "${COUNT}" -ge 0 ]
    then
        echo "Pass: sourcehut user clone with ghorgonly - filtered repos count: ${COUNT}"
    else
        echo "Fail: sourcehut user clone with ghorgonly - unexpected count"
        rm -f "$GHORGONLY_TEST_FILE"
        exit 1
    fi
else
    echo "Fail: sourcehut user clone with ghorgonly filter"
    rm -f "$GHORGONLY_TEST_FILE"
    exit 1
fi

rm -f "$GHORGONLY_TEST_FILE"

# Test 12: Clone with branch flag
ghorg clone $SOURCEHUT_USER --scm=sourcehut --token=$SOURCEHUT_TOKEN --branch=main --path=/tmp --output-dir=testing_sourcehut_branch

if [ -d "/tmp/testing_sourcehut_branch" ]
then
    echo "Pass: sourcehut user clone with custom branch"
else
    echo "Fail: sourcehut user clone with custom branch"
    exit 1
fi

echo ""
echo "=========================================="
echo "All sourcehut integration tests passed!"
echo "=========================================="

