#!/bin/bash

set -euo pipefail

echo "Running Codeberg Integration Tests"

cp ./ghorg /usr/local/bin

# Codeberg test org - a real Codeberg org containing test repos.
# For CI this is configured via repository secrets.
#
# Expected layout of the test org:
#   test         - private repo (validates token auth clone)
#   frontend-app - public repo, topic: ghorgtest
#   backend-app  - public repo, topic: ghorgtest
#   archived-app - public repo, archived (validates --skip-archived)
#   wiki-app     - public repo with wiki enabled and at least one wiki page
#   forked-app   - fork of another repo (validates --skip-forks)
CODEBERG_ORG=${CODEBERG_TEST_ORG:-"ghorg"}
TOTAL_REPOS=6

if [ -z "${CODEBERG_TOKEN:-}" ]; then
    echo "CODEBERG_TOKEN environment variable is not set"
    echo "Skipping codeberg integration tests"
    exit 0
fi

ghorg version

echo "Testing codeberg org: $CODEBERG_ORG"

# count_repos <dir> - count top level directories (cloned repos) in a clone target
count_repos() {
    find "$1" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' '
}

# Test 1: Clone an org with no config file (base URL defaults to codeberg.org)
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN

if [ -d "$HOME/ghorg/$CODEBERG_ORG" ] && [ "$(count_repos $HOME/ghorg/$CODEBERG_ORG)" -eq "$TOTAL_REPOS" ]
then
    echo "Pass: codeberg org clone using no configuration file ($TOTAL_REPOS repos)"
else
    echo "Fail: codeberg org clone using no configuration file, expected $TOTAL_REPOS repos got $(count_repos $HOME/ghorg/$CODEBERG_ORG 2>/dev/null || echo 0)"
    exit 1
fi

# Test 2: Private repo clone worked (token auth)
if [ -d "$HOME/ghorg/$CODEBERG_ORG/test" ]
then
    echo "Pass: codeberg private repo cloned with token auth"
else
    echo "Fail: codeberg private repo cloned with token auth"
    exit 1
fi

# Test 3: Clone with preserve-scm-hostname
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --preserve-scm-hostname

if [ -d "$HOME/ghorg/codeberg.org/$CODEBERG_ORG" ]
then
    echo "Pass: codeberg org clone preserving scm hostname"
else
    echo "Fail: codeberg org clone preserving scm hostname"
    exit 1
fi

# Test 4: Clone to a specific path with output-dir
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --path=/tmp --output-dir=testing_codeberg_output

if [ -d "/tmp/testing_codeberg_output" ]
then
    echo "Pass: codeberg org clone, custom path and output dir"
else
    echo "Fail: codeberg org clone, custom path and output dir"
    exit 1
fi

# Test 5: Clone with an explicit base-url pointing at codeberg.org
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --base-url=https://codeberg.org --path=/tmp --output-dir=testing_codeberg_baseurl

if [ -d "/tmp/testing_codeberg_baseurl" ]
then
    echo "Pass: codeberg org clone with explicit base-url"
else
    echo "Fail: codeberg org clone with explicit base-url"
    exit 1
fi

# Test 6: Clone with match-regex filter, only frontend-app should match
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --match-regex="^frontend" --path=/tmp --output-dir=testing_codeberg_regex

if [ -d "/tmp/testing_codeberg_regex/frontend-app" ] && [ "$(count_repos /tmp/testing_codeberg_regex)" -eq 1 ]
then
    echo "Pass: codeberg org clone with regex filter"
else
    echo "Fail: codeberg org clone with regex filter, expected only frontend-app"
    exit 1
fi

# Test 7: Skip archived repos, archived-app should be excluded
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --skip-archived --path=/tmp --output-dir=testing_codeberg_skip_archived

if [ ! -d "/tmp/testing_codeberg_skip_archived/archived-app" ] && [ "$(count_repos /tmp/testing_codeberg_skip_archived)" -eq "$((TOTAL_REPOS - 1))" ]
then
    echo "Pass: codeberg org clone skipping archived repos"
else
    echo "Fail: codeberg org clone skipping archived repos, archived-app should be excluded"
    exit 1
fi

# Test 8: Skip forks, forked-app should be excluded
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --skip-forks --path=/tmp --output-dir=testing_codeberg_skip_forks

if [ ! -d "/tmp/testing_codeberg_skip_forks/forked-app" ] && [ "$(count_repos /tmp/testing_codeberg_skip_forks)" -eq "$((TOTAL_REPOS - 1))" ]
then
    echo "Pass: codeberg org clone skipping forks"
else
    echo "Fail: codeberg org clone skipping forks, forked-app should be excluded"
    exit 1
fi

# Test 9: Filter by topic, only frontend-app and backend-app have topic ghorgtest
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --topics=ghorgtest --path=/tmp --output-dir=testing_codeberg_topics

if [ -d "/tmp/testing_codeberg_topics/frontend-app" ] && [ -d "/tmp/testing_codeberg_topics/backend-app" ] && [ "$(count_repos /tmp/testing_codeberg_topics)" -eq 2 ]
then
    echo "Pass: codeberg org clone filtering by topic"
else
    echo "Fail: codeberg org clone filtering by topic, expected only frontend-app and backend-app"
    exit 1
fi

# Test 10: Clone wikis, wiki-app has a wiki enabled
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --clone-wiki --path=/tmp --output-dir=testing_codeberg_wiki

if [ -d "/tmp/testing_codeberg_wiki/wiki-app.wiki" ]
then
    echo "Pass: codeberg org clone with wikis"
else
    echo "Fail: codeberg org clone with wikis, expected wiki-app.wiki"
    exit 1
fi

# Test 10b: Re-clone with wikis, exercises the pull path of an existing wiki
# clone. Regression test for wikis on a main (not master) default branch.
if GHORG_EXIT_CODE_ON_CLONE_INFOS=1 GHORG_EXIT_CODE_ON_CLONE_ISSUES=1 ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --clone-wiki --path=/tmp --output-dir=testing_codeberg_wiki
then
    echo "Pass: codeberg org re-clone with wikis (pull path)"
else
    echo "Fail: codeberg org re-clone with wikis (pull path)"
    exit 1
fi

# Test 11: Clone with configuration file in default location
mkdir -p $HOME/.config/ghorg
cp sample-conf.yaml $HOME/.config/ghorg/conf.yaml

sed "s/GHORG_OUTPUT_DIR:/GHORG_OUTPUT_DIR: testing_codeberg_conf/g" $HOME/.config/ghorg/conf.yaml > updated_conf.yaml && \
mv $HOME/.config/ghorg/conf.yaml $HOME/.config/ghorg/conf-bak.yaml && \
mv updated_conf.yaml $HOME/.config/ghorg/conf.yaml

ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN

if [ -d "$HOME/ghorg/testing_codeberg_conf" ]
then
    echo "Pass: codeberg org clone, using config file in default location"
else
    echo "Fail: codeberg org clone, using config file in default location"
    exit 1
fi

# Restore original config
mv $HOME/.config/ghorg/conf-bak.yaml $HOME/.config/ghorg/conf.yaml

# Test 12: Clone with alternative config file path
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --config=$PWD/scripts/testing_confs/alternative_clone_path_conf.yaml

if [ -d "/tmp/path_from_configuration_file" ]
then
    echo "Pass: codeberg org clone, alternative configuration file path"
else
    echo "Fail: codeberg org clone, alternative configuration file path"
    exit 1
fi

# Test 13: Prune test with preserve-scm-hostname
ghorg clone $CODEBERG_ORG --scm=codeberg --token=$CODEBERG_TOKEN --preserve-scm-hostname --prune-untouched --prune-untouched-no-confirm

if [ -d "$HOME/ghorg/codeberg.org/$CODEBERG_ORG" ]
then
    echo "Pass: codeberg org clone with prune untouched"
else
    echo "Fail: codeberg org clone with prune untouched"
    exit 1
fi

echo ""
echo "=========================================="
echo "All codeberg integration tests passed!"
echo "=========================================="
