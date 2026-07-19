# Don't Miss These Features

Easily overlooked flags that work across all SCM providers. Commands below omit provider specific flags for brevity; add your usual `--scm`, `--base-url`, and auth flags as needed. For provider specific setup and examples see the pages for [github](https://github.com/gabrie30/ghorg/blob/master/examples/github.md), [gitlab](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md), [bitbucket](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md), [gitea](https://github.com/gabrie30/ghorg/blob/master/examples/gitea.md), [codeberg](https://github.com/gabrie30/ghorg/blob/master/examples/codeberg.md), and [sourcehut](https://github.com/gabrie30/ghorg/blob/master/examples/sourcehut.md).

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md. You can also read this page in your terminal with `ghorg examples features`.

## Keeping Clones Safe and In Sync

1. `--dry-run` shows exactly what a clone would do without touching disk, great for testing filters and verifying your auth works

    ```
    ghorg clone <org> --match-regex=^frontend --dry-run --token=XXXXXX
    ```

1. `--protect-local` updates clean repos but skips any repo with uncommitted changes or unpushed commits, and restores the branch you had checked out

1. `--no-clean` only clones repos that are new and leaves all existing repos completely untouched. Without it, re-running a clone will `git pull` and `git clean` every repo, overwriting local changes

1. `--prune` deletes local repos that no longer exist on the remote (prompts first; add `--prune-no-confirm` to skip the prompt). `--prune-untouched` instead deletes local repos you have never modified

1. `--backup` creates bare mirror clones for backups; combine with `--clone-wiki` and `--include-submodules` for a complete archive

    ```
    ghorg clone <org> --backup --clone-wiki --include-submodules --token=XXXXXX
    ```

1. `--fetch-all` runs `git fetch --all` on every repo; add `--fetch-prune` to also remove stale remote-tracking branches, and `--fetch-git-lfs` to pull LFS content

## Filtering Which Repos Get Cloned

1. `--match-regex`/`--exclude-match-regex` and `--match-prefix`/`--exclude-match-prefix` filter repos by name

1. `--skip-archived` and `--skip-forks` drop archived repos and forks (GitHub, GitLab, Gitea, and Codeberg only)

1. `--topics` clones only repos tagged with at least one matching topic e.g. `--topics=docker,kubernetes` (GitHub, GitLab, Gitea, and Codeberg only)

1. `--target-repos-path` points at a file of exact repo names (one per line) to clone, and warns about names that don't exist, ideal for a curated list

1. A `ghorgignore` file at `$HOME/.config/ghorg/ghorgignore` is picked up automatically and skips any repo whose clone URL contains one of its lines; `ghorgonly` is the allowlist equivalent. See [selective cloning](https://github.com/gabrie30/ghorg#selective-repository-cloning)

## Speed and Large Clones

1. `--clone-depth=1` makes shallow clones, and `--git-filter=blob:none` skips binary blobs, both dramatically cut clone time and disk usage

1. `--concurrency` controls parallel clones (default 25); lower it if you hit rate limits or `too many open files`, or use `--clone-delay-seconds` to space out clones entirely

1. `--no-dir-size` skips the final directory size calculation on very large clones

## CI and Automation

1. `--quiet` reduces output to critical messages; `--exit-code-on-clone-issues` and `--exit-code-on-clone-infos` let pipelines fail on clone problems

1. `--stats-enabled` appends a row to `_ghorg_stats.csv` with per clone metrics (new commits, size, duration), useful for tracking an org over time. See [tracking clone data](https://github.com/gabrie30/ghorg#tracking-clone-data-over-time)

1. If you run the same long clone commands repeatedly, store them in a [reclone config](https://github.com/gabrie30/ghorg#reclone-command) and run them all with `ghorg reclone`

1. Use `ghorg ls` to list everything ghorg has cloned

## Tokens and Auth

1. The `--token` flag accepts a path to a file containing the token e.g. `--token=~/.config/ghorg/token.txt`, so the token itself never appears on the commandline or in shell history

1. `GHORG_TOKEN_CMD` sources your token from a command instead of storing it in cleartext, e.g. `GHORG_TOKEN_CMD: op item get gitlab --fields token` for a secrets manager

1. `--ssh-hostname` rewrites SSH clone URLs to use an alias from your `~/.ssh/config`, useful when you juggle multiple SSH identities for the same provider
