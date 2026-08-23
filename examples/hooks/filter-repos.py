#!/usr/bin/env python3
"""Example ghorg repo filter hook in python.

ghorg writes the repo list as a JSON array to stdin and uses whatever JSON
array this script writes to stdout as the final clone list. The hook can
remove repos and can also modify repo fields such as clone_branch.

Usage:
    ghorg clone my-org --repo-filter-hook=/path/to/filter-repos.py
"""

import json
import sys

repos = json.load(sys.stdin)

kept = []
for repo in repos:
    # Drop repos with a legacy- prefix.
    if repo["name"].startswith("legacy-"):
        continue

    # Mutation example: clone the gh-pages branch for the docs repo.
    if repo["name"] == "docs":
        repo["clone_branch"] = "gh-pages"

    kept.append(repo)

print(f"filter hook kept {len(kept)} of {len(repos)} repos", file=sys.stderr)
json.dump(kept, sys.stdout)
