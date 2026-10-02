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
# To see every field available to this script, run ghorg with GHORG_DEBUG=true.
# ghorg prints the JSON it passes to the script just before running it. Add
# --include-scm-data to also get each repo's SCM provider API data as scm_data.
# Warning: GHORG_DEBUG also prints your API token to stdout. It can appear
# inside each clone_url when cloning over HTTPS, so keep the output private.
#   GHORG_DEBUG=true ghorg clone my-org --repo-filter-hook=/path/to/filter-repos.sh --dry-run
#
# This example keeps only repos whose name does not end in -archive.

jq '[.[] | select(.name | endswith("-archive") | not)]'
