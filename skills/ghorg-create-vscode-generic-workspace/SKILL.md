---
name: ghorg-create-vscode-generic-workspace
description: Use when the user wants a reusable ghorg reclone "generic workspace" that clones the same set of repositories into a fresh, separately named VSCode workspace on every run, such as `ghorg reclone <prefix>-generic-workspace` producing `<prefix>-<name>` directories.
---

# Create ghorg generic workspace

A **generic workspace** is a ghorg reclone entry that acts as a template. Every run of `ghorg reclone <PREFIX>-generic-workspace` clones the same repositories into a staging directory, then a `post_exec_script` asks for a name and renames the directory to `<PREFIX>-<name>`. Each run produces a new, independent workspace (one per feature, investigation, or task).

This differs from the `ghorg-create-vscode-workspace` skill, which creates one fixed workspace that is updated in place on each reclone.

## Prerequisites

- [ghorg](https://github.com/gabrie30/ghorg) must be installed and on the system `PATH`.
- Network access to the SCM provider (GitHub, GitLab, Bitbucket, etc.).
- `bash` and `git` must be available.

## Platform support

**macOS** and **Linux**. The post-exec script is a bash script that ghorg executes directly, so on **Windows** it only works under WSL.

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| HOME_DIR | Base directory for all paths | `$HOME_DIR` if set, otherwise `$HOME` |
| SCM_ORG | SCM organization to clone from | `$SCM_ORG` if set, otherwise ask **What is the SCM organization to clone from?** (example: `my-github-org`) |

Stop with an error if HOME_DIR is empty or the user does not provide an SCM_ORG.

## Paths

| Variable | Path |
|----------|------|
| GHORG_CONFIG_DIR | HOME_DIR/.config/ghorg |
| GHORG_RECLONE_CONFIG | `$GHORG_RECLONE_PATH` if set, otherwise HOME_DIR/.config/ghorg/reclone.yaml |
| BASE_DIR | HOME_DIR/workspaces |
| RECLONE_KEY | &lt;PREFIX&gt;-generic-workspace |
| REPO_FILE | GHORG_CONFIG_DIR/&lt;PREFIX&gt;-all-repos-generic.txt |
| POST_EXEC_SCRIPT | GHORG_CONFIG_DIR/post-exec-&lt;PREFIX&gt;-generic-workspace.sh |
| TEMPLATE | `post-exec-template.sh` in this skill's directory |

## Workflow

### Step 1: Resolve environment

Resolve **HOME_DIR** and **SCM_ORG** as described above.

### Step 2: Ask for the prefix

Ask the user:

**What prefix should workspaces created from this template use?**

Example: `payments` produces workspaces like `payments-fix-refunds`.

**Validation:** Must match `^[a-z0-9-]+$`. Ask again until valid. Store as **PREFIX**.

### Step 3: Ask for repositories

Ask the user:

**List the repositories to clone into every workspace, one per line.**

Store as **REPO_LIST**. At least one repository is required.

### Step 4: Ask for files to copy (optional)

Ask the user:

**Are there local files to copy into every new workspace (for example `.env` files that are not in git)? Give one `source|destination` pair per line, or leave empty.**

- `source` is an absolute path on this machine.
- `destination` is relative to the workspace root, usually `<repo>/<path>`.

Example: `/path/to/my-app/.env|my-app/.env`

Store as **COPY_FILES**. Empty is fine.

### Step 5: Check for conflicts

Stop and report an error, without changing anything, if any of these already exist:

- A top-level key **RECLONE_KEY** in GHORG_RECLONE_CONFIG (`grep -q '^<RECLONE_KEY>:' <GHORG_RECLONE_CONFIG>`).
- The file **REPO_FILE**.
- The file **POST_EXEC_SCRIPT**.

### Step 6: Create files

1. Create **GHORG_CONFIG_DIR** and **BASE_DIR** (with parents) if missing. Create an empty GHORG_RECLONE_CONFIG if missing.
2. Write **REPO_FILE** with REPO_LIST, one repository per line.
3. Copy **TEMPLATE** to **POST_EXEC_SCRIPT** and edit the copy:
   - Replace `__PREFIX__` with PREFIX.
   - Replace `__BASE_DIR__` with the absolute BASE_DIR.
   - Add one quoted line per COPY_FILES entry inside `COPY_FILES=(` `)`, for example `  "/path/to/my-app/.env|my-app/.env"`.
4. Make it executable: `chmod +x <POST_EXEC_SCRIPT>`.
5. Check the syntax: `bash -n <POST_EXEC_SCRIPT>`.

### Step 7: Append the reclone entry

Append to **GHORG_RECLONE_CONFIG**, replacing placeholders with resolved absolute values:

```yaml
<PREFIX>-generic-workspace:
  cmd: "ghorg clone <SCM_ORG> --path=<BASE_DIR> --target-repos-path=<REPO_FILE> --output-dir=<PREFIX>-generic-workspace --stats-enabled=false"
  description: "Generic workspace for <PREFIX>. Each run creates a new <PREFIX>-<name> workspace"
  post_exec_script: "<POST_EXEC_SCRIPT>"
```

- Preserve all existing content. Add a newline first if the file does not end with one.
- Validate the file still parses: `GHORG_RECLONE_PATH="<GHORG_RECLONE_CONFIG>" ghorg reclone --list > /dev/null`. Discard the output because it can contain tokens.

### Step 8: Create the first workspace (optional)

Ask the user whether to create a workspace now. If yes, ask for its name and run:

```bash
GHORG_RECLONE_PATH="<GHORG_RECLONE_CONFIG>" GENERIC_WORKSPACE_NAME="<name>" ghorg reclone <PREFIX>-generic-workspace
```

`GENERIC_WORKSPACE_NAME` skips the interactive prompt, which needs a terminal that agent shells do not have.

ghorg still exits 0 when the post-exec script fails. Treat the run as failed if the output contains `ERROR: Running post_exec_script` or the directory from the `INFO: Workspace ready at` line does not contain a `<directory name>.code-workspace` file, and show the `post-exec logs` section to the user.

If the name was already taken, the clone stays in `BASE_DIR/<PREFIX>-generic-workspace` and the next run reuses it. Ask for a different name and run the command again.

### Step 9: Explain what was added

Tell the user the generic workspace was created, then explain what was added and how the pieces fit together, using this structure with resolved values. Include item 4 only if Step 8 created a workspace.

```markdown
**What was added**

1. **Repo list:** `<REPO_FILE>`
   - Contains: the repositories cloned into every new workspace, one per line (<N> repos: <names>).
   - Role: ghorg reads it through `--target-repos-path`. Edit it to change which repos future workspaces get. Existing workspaces are not affected.
2. **Post-exec script:** `<POST_EXEC_SCRIPT>`
   - Contains: a bash script with its settings at the top: `PREFIX`, `BASE_DIR`, `OPEN_IN_VSCODE`, and `COPY_FILES` (<N> files: <destinations>, or none).
   - Role: ghorg runs it after every clone. It asks for a name, renames the fresh clone to `<PREFIX>-<name>`, adds the marker, root git repo, and `.code-workspace` file, copies `COPY_FILES`, and opens VS Code.
3. **Reclone entry:** `<PREFIX>-generic-workspace` in `<GHORG_RECLONE_CONFIG>`
   - Contains: the `ghorg clone` command (organization, `<BASE_DIR>`, repo list, and the `<PREFIX>-generic-workspace` staging folder) plus the path to the post-exec script.
   - Role: `ghorg reclone <PREFIX>-generic-workspace` runs it. Every run creates a new workspace.
4. **Workspace folder:** `<BASE_DIR>/<PREFIX>-<slug>/`
   - Contains: the cloned repos, any copied files, `<PREFIX>-<slug>.code-workspace`, a `.git` folder at the root, and the `.ghorg-generic-workspace` marker (prefix, reclone entry, creation time).
   - Role: where you work on this task. Open the `.code-workspace` file to load every repo in one window. The marker lets the cleanup skill identify the folder.

**Workflow:** run `ghorg reclone <PREFIX>-generic-workspace` and enter a name (`fix login bug` becomes `<PREFIX>-fix-login-bug`), then work in the new folder. Every run clones into a new folder, so existing workspaces are never reset or overwritten.

**Cleaning up later:** use the `ghorg-cleanup-vscode-generic-workspace` skill to delete old workspaces or remove this template. It shows what it will remove and asks before deleting.
```

## How a run works

1. ghorg clones REPO_LIST into `BASE_DIR/<PREFIX>-generic-workspace`.
2. The post-exec script takes the name from `GENERIC_WORKSPACE_NAME` or prompts on the terminal, then slugifies it (`My Fix!` becomes `my-fix`).
3. It renames the staging directory to `BASE_DIR/<PREFIX>-<slug>` and aborts if that already exists.
4. It writes a `.ghorg-generic-workspace` marker, runs `git init` at the workspace root, creates `<PREFIX>-<slug>.code-workspace`, copies COPY_FILES, and opens the workspace in VS Code when it can find it.

To change copied files or stop VS Code from opening, edit `COPY_FILES` or `OPEN_IN_VSCODE` in POST_EXEC_SCRIPT.

### VS Code workspace file

Every workspace gets a VS Code workspace file in its root, named after the workspace folder:

**BASE_DIR/&lt;PREFIX&gt;-&lt;slug&gt;/&lt;PREFIX&gt;-&lt;slug&gt;.code-workspace**

Example: `HOME_DIR/workspaces/payments-fix-refunds/payments-fix-refunds.code-workspace`

With the following contents:

```json
{
   "folders": [
      {
         "path": "."
      }
   ],
   "settings": {}
}
```

The post-exec script writes this file on every run and overwrites any existing one. The content comes from the heredoc in `post-exec-template.sh`, so keep the two in sync when editing either.

## Failure conditions

Stop and report an error. Never silently continue when:

- HOME_DIR or SCM_ORG cannot be resolved.
- RECLONE_KEY, REPO_FILE, or POST_EXEC_SCRIPT already exists.
- A file or directory cannot be written, or `bash -n` fails.
- GHORG_RECLONE_CONFIG no longer parses after appending.
- `ghorg reclone` fails in Step 8 (non-zero exit, or a post-exec error as described there).
