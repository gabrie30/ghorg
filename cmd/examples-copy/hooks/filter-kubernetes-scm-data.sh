#!/bin/sh
# Example ghorg repo filter hook using jq and scm_data, written for the
# kubernetes GitHub org.
#
# With --include-scm-data, ghorg adds the object GitHub's API returned for each
# repo to the JSON as scm_data, so this script can filter on fields ghorg has no
# flags for. Field names and units match GitHub's API.
#
# Usage:
#   ghorg clone kubernetes --include-scm-data --repo-filter-hook=/path/to/filter-kubernetes-scm-data.sh
#
# To see every field available to this script, run ghorg with GHORG_DEBUG=true.
# ghorg prints the JSON it passes to the script just before running it.
# Warning: GHORG_DEBUG also prints your API token to stdout. It can appear
# inside each clone_url when cloning over HTTPS, so keep the output private.
#   GHORG_DEBUG=true ghorg clone kubernetes --include-scm-data --repo-filter-hook=/path/to/filter-kubernetes-scm-data.sh --dry-run
#
# This example keeps the actively developed repos a contributor can open pull
# requests against. It skips:
#   - archived repos
#   - read-only staging mirrors such as client-go and apimachinery. They are
#     published from kubernetes/kubernetes/staging, and pull requests opened
#     against them are ignored. Kubernetes does not mark them consistently, so a
#     repo counts as a mirror if any of these is true:
#       - issues are turned off (has_issues is false)
#       - it has the k8s-staging topic
#       - its description mentions kubernetes/kubernetes/staging
#   - repos with no pushes in the last year
#
# Wikis carry their parent repo's scm_data, so the wiki of a skipped repo is
# skipped too.

input=$(cat)

# scm_data is only present when ghorg runs with --include-scm-data. Exiting
# non-zero makes ghorg abort instead of cloning a list that was never filtered.
if ! printf '%s' "$input" | jq -e 'all(.[]; .scm_data != null)' > /dev/null; then
  echo "filter-kubernetes-scm-data.sh: scm_data is missing, run ghorg with --include-scm-data" >&2
  exit 1
fi

# GitHub timestamps are UTC strings like 2026-10-02T15:04:05Z, so comparing
# them as strings orders them by date.
printf '%s' "$input" | jq '
  def staging_mirror:
    .scm_data.has_issues == false
    or ((.scm_data.topics // []) | any(. == "k8s-staging"))
    or ((.scm_data.description // "") | contains("kubernetes/kubernetes/staging"));

  (now - 365 * 24 * 60 * 60 | todate) as $one_year_ago
  | [.[] | select(
      .scm_data.archived != true
      and (staging_mirror | not)
      and .scm_data.pushed_at >= $one_year_ago
    )]
'
