---
name: ghorg-cleanup-generic-workspaces
description: Use when the user wants to clean up ghorg generic workspaces, such as deleting old `<prefix>-<name>` workspace directories created by `ghorg reclone <prefix>-generic-workspace`, or removing a generic workspace template (its reclone.yaml entry, post-exec script, and repo list).
---

# Clean up ghorg generic workspaces

Cleans up after the `ghorg-create-generic-workspace` skill. It supports two kinds of cleanup, and the user picks one or both:

- **Delete workspaces:** remove `<PREFIX>-<name>` directories created by runs of the generic workspace.
- **Remove the template:** remove the `<PREFIX>-generic-workspace` reclone entry, its post-exec script, and its repo list file. Existing workspaces are left alone unless the user also picks "Delete workspaces".

Deleting a workspace destroys any uncommitted, unpushed, or stashed work in its repositories. Always run the safety check and get explicit confirmation before deleting anything.

## Platform support

**macOS** and **Linux** (and WSL on Windows). Commands below are POSIX `sh` compatible, so they also run in `zsh`.

## Environment variables and paths

| Variable | Value |
|----------|-------|
| HOME_DIR | `$HOME_DIR` if set, otherwise `$HOME`. Stop with an error if neither is set. |
| GHORG_CONFIG_DIR | HOME_DIR/.config/ghorg |
| GHORG_RECLONE_CONFIG | `$GHORG_RECLONE_PATH` if set, otherwise GHORG_CONFIG_DIR/reclone.yaml |
| MARKER_FILE | `.ghorg-generic-workspace` at the root of each workspace |

reclone.yaml can contain tokens. Always read it through this command, which redacts `-t` and `--token` values, and never print a raw credential:

```sh
sed -E 's/(-t[= ]|--token[= ])[^ "]+/\1<redacted>/g' "<GHORG_RECLONE_CONFIG>"
```

## Workflow

### Step 1: Find generic workspace templates

Read **GHORG_RECLONE_CONFIG** (redacted). A top-level entry is a generic workspace template when its `cmd` contains `--output-dir=<PREFIX>-generic-workspace` and it has a `post_exec_script`. For each template record:

- **RECLONE_KEY:** the top-level key.
- **PREFIX:** from `--output-dir=<PREFIX>-generic-workspace`.
- **BASE_DIR:** the `--path=` value in `cmd`, without a trailing slash.
- **REPO_FILE:** the `--target-repos-path=` value in `cmd`.
- **POST_EXEC_SCRIPT:** the `post_exec_script` value.

If there is more than one, ask which template to clean up.

If none are found, or the user wants to clean up workspaces whose template was already removed, ask for the **PREFIX** and **BASE_DIR** (default `HOME_DIR/workspaces`) and only offer option 1 in Step 2.

### Step 2: Ask what to clean up

Ask the user to pick one:

1. Delete workspaces created from this template
2. Remove the template itself
3. Both

### Step 3: Find workspaces (options 1 and 3)

Run, substituting BASE_DIR and PREFIX:

```sh
BASE_DIR="<BASE_DIR>"; PREFIX="<PREFIX>"
find "$BASE_DIR" -mindepth 1 -maxdepth 1 -type d | sort | while IFS= read -r d; do
  name=$(basename "$d")
  if [ -f "$d/.ghorg-generic-workspace" ]; then
    grep -qx "prefix=$PREFIX" "$d/.ghorg-generic-workspace" &&
      echo "VERIFIED $d (created $(sed -n 's/^created=//p' "$d/.ghorg-generic-workspace"))"
  elif [ "$name" = "$PREFIX-generic-workspace" ]; then
    echo "STAGING $d"
  else
    case "$name" in
      "$PREFIX"-*) [ -f "$d/$name.code-workspace" ] && echo "UNVERIFIED $d" ;;
    esac
  fi
done
```

- **VERIFIED:** has a marker for this PREFIX. Safe to offer.
- **STAGING:** left behind by a run that failed or was interrupted before renaming.
- **UNVERIFIED:** matches the naming pattern but has no marker. It may predate the marker or belong to a template with a longer prefix (prefix `app` also matches `app-api-fix`). Only delete one if the user names it explicitly.

Show the list and ask which workspaces to delete. "All verified" is a valid answer. If none are found, say so. For option 1, stop. For option 3, skip Step 4 and show only the template in Step 5.

### Step 4: Safety check each selected workspace

For each selected workspace, run:

```sh
WS="<WORKSPACE_DIR>"
find "$WS" -mindepth 1 -maxdepth 1 ! -name .git ! -name .DS_Store ! -name .ghorg-generic-workspace ! -name '*.code-workspace' | sort | while IFS= read -r p; do
  if [ ! -d "$p" ] || [ ! -e "$p/.git" ]; then echo "NOT A REPO: $p"; continue; fi
  dirty=$(git -C "$p" status --porcelain | wc -l | tr -d ' ')
  unpushed=$(git -C "$p" log --branches --not --remotes --oneline | wc -l | tr -d ' ')
  stashed=$(git -C "$p" stash list | wc -l | tr -d ' ')
  if [ "$dirty" != 0 ] || [ "$unpushed" != 0 ] || [ "$stashed" != 0 ]; then
    echo "AT RISK: $p ($dirty uncommitted, $unpushed unpushed commits, $stashed stashes)"
  fi
done
```

`NOT A REPO` lines are files or folders the user added at the workspace root. `AT RISK` lines are repositories with work that only exists locally. No output means nothing would be lost.

### Step 5: Explain what will be removed and confirm

Nothing has been changed yet. Before removing anything, explain to the user what each item is, what it contains, and what removing it does. Use this structure with resolved values, including only the sections for the option the user picked:

```markdown
**Here is what will be removed**

Workspaces (one item per selected workspace):

1. **Workspace:** `<WORKSPACE_DIR>/` (created <date from the marker, or "unknown">)
   - Contains: <N> cloned repos (<names>), `<name>.code-workspace`, the `.ghorg-generic-workspace` marker, a root `.git` folder, and <each NOT A REPO item, or nothing else>.
   - Safety check: <each AT RISK line, or "no uncommitted, unpushed, or stashed work found">.
   - Effect: the folder and everything in it are deleted, including any work listed as at risk and files copied in through `COPY_FILES`. Anything already pushed stays on the remote.

Template:

- **Reclone entry:** `<RECLONE_KEY>` in `<GHORG_RECLONE_CONFIG>`
  - Contains: the `ghorg clone` command for this template and the path to its post-exec script.
  - Effect: `ghorg reclone <RECLONE_KEY>` stops working, so no new `<PREFIX>-<name>` workspaces can be created. The config is backed up to `<GHORG_RECLONE_CONFIG>.bak` first.
- **Post-exec script:** `<POST_EXEC_SCRIPT>`
  - Contains: the setup script ghorg runs after each clone, with `COPY_FILES` set to <N files: sources, or none>.
  - Effect: deleted. The source files listed in `COPY_FILES` are not touched.
- **Repo list:** `<REPO_FILE>`
  - Contains: <N> repos: <names>.
  - Effect: deleted.

**Not affected:** <each workspace that will be kept, such as unselected or unverified ones, or "nothing else">.
```

Adjust the plan to match reality:

- If the updated config would still reference POST_EXEC_SCRIPT or REPO_FILE through another entry, say that file will be kept and name the entry.
- For option 2, add: existing workspaces are not deleted. To delete them later, run this skill again with prefix `<PREFIX>`.

Then ask the user to pick one:

1. Remove everything listed
2. Remove everything except the at-risk workspaces (offer only if any are at risk; the template is still removed if listed)
3. Cancel

Stop if they cancel.

### Step 6: Delete workspaces (options 1 and 3)

Delete each confirmed directory with `rm -rf -- "<WORKSPACE_DIR>"`. Before each delete, verify `[ "$(dirname "<WORKSPACE_DIR>")" = "<BASE_DIR>" ]` and that the path is not HOME_DIR.

If the environment refuses a delete (for example a safety hook blocks `rm -rf`), do not work around it. Give the user the exact command to run themselves and report the directory as not deleted.

### Step 7: Remove the template (options 2 and 3)

Back up the config, drop the entry's block (from its top-level key up to the next top-level key), and validate. The backup keeps the original file's permissions.

```sh
CFG="<GHORG_RECLONE_CONFIG>"; KEY="<RECLONE_KEY>"
cp -p "$CFG" "$CFG.bak" &&
awk -v key="$KEY" '/^[^ \t#]/ { skip = ($1 == key ":") } !skip' "$CFG.bak" > "$CFG" &&
GHORG_RECLONE_PATH="$CFG" ghorg reclone --list > /dev/null &&
echo "OK: config valid"
diff "$CFG.bak" "$CFG" | sed -E 's/(-t[= ]|--token[= ])[^ "]+/\1<redacted>/g'
```

If `OK: config valid` is not printed, or the diff shows anything other than the removed entry, restore with `cp -p "$CFG.bak" "$CFG"` and stop with an error.

Only after that succeeds, delete POST_EXEC_SCRIPT and REPO_FILE. Keep any file the updated config still references (`grep -qF '<path>' <GHORG_RECLONE_CONFIG>`) and tell the user it was kept.

### Step 8: Report

Report what actually happened, including anything kept, blocked, or failed.

```
Generic workspace cleanup summary.

Deleted workspaces:
<one path per line, or "none">

Not deleted (blocked or failed, with the command to run):
<one path per line, or "none">

Kept (not selected or at risk):
<one path per line, or "none">

Template removed:
<RECLONE_KEY, POST_EXEC_SCRIPT, REPO_FILE, or "no">

reclone.yaml backup:
<GHORG_RECLONE_CONFIG>.bak (only if Step 7 ran)
```

## Failure conditions

Stop and report an error. Never silently continue when:

- HOME_DIR or GHORG_RECLONE_CONFIG cannot be resolved or read.
- A delete target is not a direct child of BASE_DIR.
- GHORG_RECLONE_CONFIG does not parse or the diff is wrong after editing (restore the backup first).
- Any delete fails. Report it in Step 8 rather than claiming success.

Never delete without the user's explicit confirmation in this conversation.
