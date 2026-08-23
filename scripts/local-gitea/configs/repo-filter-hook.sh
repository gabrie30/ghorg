#!/bin/sh
# Integration test fixture for GHORG_REPO_FILTER_HOOK.
# Keeps only the repos named baz0 and baz1. Requires jq, which is
# preinstalled on the GitHub Actions ubuntu runners these tests use.
jq '[.[] | select(.name == "baz0" or .name == "baz1")]'
