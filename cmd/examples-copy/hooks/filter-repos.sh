#!/bin/sh
# Example ghorg repo filter hook using jq.
#
# ghorg writes the repo list as a JSON array to this script's stdin and uses
# whatever JSON array the script writes to stdout as the final clone list.
# Anything printed to stderr is shown in ghorg's output while the hook runs.
#
# Usage:
#   ghorg clone my-org --repo-filter-hook=/path/to/filter-repos.sh
#
# This example keeps only repos whose name does not end in -archive.

jq '[.[] | select(.name | endswith("-archive") | not)]'
