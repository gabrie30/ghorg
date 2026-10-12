---
name: ghorg-remove-workspace
description: Use when the user wants to delete or remove a ghorg reclone workspace created by the create-ghorg-workspace skill, including its reclone.yaml entry, its workspace-<name>.txt repo list, and its cloned directory.
---

# Remove ghorg reclone workspace

Removes a workspace created by the `create-ghorg-workspace` skill. That skill creates three things, and this skill removes them:

| Artifact | Location |
|----------|----------|
| Reclone entry | Top-level `<WORKSPACE_NAME>` key in reclone.yaml |
| Repo list file | HOME_DIR/.config/ghorg/workspace-&lt;WORKSPACE_NAME&gt;.txt |
| Cloned workspace | HOME_DIR/workspaces/&lt;WORKSPACE_NAME&gt; (includes the `.code-workspace` file) |

Deleting the cloned workspace destroys any uncommitted, unpushed, or stashed work in its repositories. Always run the safety check and get explicit confirmation first.

## Platform support

**macOS** and **Linux** (and WSL on Windows). Commands below are POSIX `sh` compatible, so they also run in `zsh`. On native Windows, run the equivalent `git` commands in PowerShell and use `Remove-Item -Recurse -Force` to delete.

## Environment variables and paths

| Variable | Value |
|----------|-------|
| HOME_DIR | `$HOME_DIR` if set, otherwise `$HOME` (`%USERPROFILE%` on Windows). Stop with an error if none is set. |
| GHORG_CONFIG_DIR | HOME_DIR/.config/ghorg |
| GHORG_RECLONE_CONFIG | `$GHORG_RECLONE_PATH` if set, otherwise GHORG_CONFIG_DIR/reclone.yaml |

reclone.yaml can contain tokens. Always read it through this command, which redacts `-t` and `--token` values, and never print a raw credential:

```sh
sed -E 's/(-t[= ]|--token[= ])[^ "]+/\1<redacted>/g' "<GHORG_RECLONE_CONFIG>"
```

## Workflow

### Step 1: Choose the workspace

Read **GHORG_RECLONE_CONFIG** (redacted). A top-level entry was created by `create-ghorg-workspace` when its `cmd` has a `--target-repos-path=` value ending in `workspace-<KEY>.txt` and `--output-dir=<KEY>`, where `<KEY>` is the entry's key.

If the user already named a workspace, use it. Otherwise list the matching keys with their descriptions and ask which one to remove. Store as **WORKSPACE_NAME**.

If WORKSPACE_NAME is not a top-level key, stop with an error. If the key exists but does not match the pattern above, tell the user it was not created by `create-ghorg-workspace` and ask whether to continue anyway.

From the entry's `cmd`, record:

- **REPO_FILE:** the `--target-repos-path=` value.
- **CLONE_DIR:** the `--path=` value (without a trailing slash) joined with the `--output-dir=` value. If `cmd` has no `--path=`, the clone location is unknown: tell the user and only offer to remove the config.

### Step 2: Safety check the cloned workspace

Skip this step if CLONE_DIR does not exist. Otherwise run:

```sh
WS="<CLONE_DIR>"
find "$WS" -mindepth 1 -maxdepth 1 ! -name .git ! -name .DS_Store ! -name '*.code-workspace' | sort | while IFS= read -r p; do
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

### Step 3: Explain what will be removed and confirm

Nothing has been changed yet. Before removing anything, explain to the user what each item is, what it contains, and what removing it does, using this structure with resolved values:

```markdown
**Here is what will be removed**

1. **Reclone entry:** `<WORKSPACE_NAME>` in `<GHORG_RECLONE_CONFIG>`
   - Contains: the `ghorg clone` command for this workspace (organization, clone location, repo list, and output folder name).
   - Effect: `ghorg reclone <WORKSPACE_NAME>` stops working, and a plain `ghorg reclone` no longer includes this workspace. The config is backed up to `<GHORG_RECLONE_CONFIG>.bak` first.
2. **Repo list:** `<REPO_FILE>`
   - Contains: <N> repos: <names>.
   - Effect: the record of which repos belong to this workspace is deleted.
3. **Workspace folder:** `<CLONE_DIR>/`
   - Contains: <N> cloned repos (<names>), `<WORKSPACE_NAME>.code-workspace`, and <each NOT A REPO item, or nothing else>.
   - Safety check: <each AT RISK line, or "no uncommitted, unpushed, or stashed work found">.
   - Effect: the local clones are deleted, including any work listed as at risk. Anything already pushed stays on the remote.
```

Adjust the plan to match reality:

- If another entry also uses REPO_FILE (`grep -cF '<REPO_FILE>' <GHORG_RECLONE_CONFIG>` prints more than 1), say it will be kept and name that entry.
- If CLONE_DIR does not exist or is unknown, say so in item 3 and offer only options 2 and 3 below.

Then ask the user to pick one:

1. Remove everything
2. Remove the config only (reclone entry and REPO_FILE) and keep CLONE_DIR
3. Cancel

Stop if they cancel.

### Step 4: Remove the reclone entry

Back up the config, drop the entry's block (from its top-level key up to the next top-level key), and validate. The backup keeps the original file's permissions.

```sh
CFG="<GHORG_RECLONE_CONFIG>"; KEY="<WORKSPACE_NAME>"
cp -p "$CFG" "$CFG.bak" &&
awk -v key="$KEY" '/^[^ \t#]/ { skip = ($1 == key ":") } !skip' "$CFG.bak" > "$CFG" &&
GHORG_RECLONE_PATH="$CFG" ghorg reclone --list > /dev/null &&
echo "OK: config valid"
diff "$CFG.bak" "$CFG" | sed -E 's/(-t[= ]|--token[= ])[^ "]+/\1<redacted>/g'
```

If `OK: config valid` is not printed, or the diff shows anything other than the removed entry, restore with `cp -p "$CFG.bak" "$CFG"` and stop with an error.

### Step 5: Delete files

Run these only after Step 4 succeeded.

1. Delete REPO_FILE, unless the updated config still references it (`grep -qF '<REPO_FILE>' <GHORG_RECLONE_CONFIG>`). In that case keep it and tell the user.
2. If the user chose "Remove everything" and CLONE_DIR exists, delete it with `rm -rf -- "<CLONE_DIR>"`. First verify CLONE_DIR ends with `/<WORKSPACE_NAME>` and is not HOME_DIR.

If the environment refuses a delete (for example a safety hook blocks `rm -rf`), do not work around it. Give the user the exact command to run themselves and report the item as not deleted.

### Step 6: Report

Report what actually happened, including anything kept, blocked, or failed.

```
Workspace removal summary.

Workspace name:
<WORKSPACE_NAME>

Removed from config:
<GHORG_RECLONE_CONFIG> (backup at <GHORG_RECLONE_CONFIG>.bak)

Repo list file:
<REPO_FILE> (deleted, or kept and why)

Cloned workspace:
<CLONE_DIR> (deleted, kept, not deleted with the command to run, or did not exist)
```

## Failure conditions

Stop and report an error. Never silently continue when:

- HOME_DIR or GHORG_RECLONE_CONFIG cannot be resolved or read.
- WORKSPACE_NAME is not a top-level key in GHORG_RECLONE_CONFIG.
- GHORG_RECLONE_CONFIG does not parse or the diff is wrong after editing (restore the backup first).
- CLONE_DIR fails the path check in Step 5.
- Any delete fails. Report it in Step 6 rather than claiming success.

Never delete without the user's explicit confirmation in this conversation.
