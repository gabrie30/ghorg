---
name: create-ghorg-workspace
description: Creates a new ghorg reclone command for a VSCode workspace. Works on macOS, Linux, and Windows. Appends to the ghorg reclone.yaml config and creates a workspace repo list file.
---

# Create ghorg reclone workspace

Creates a **ghorg reclone workspace configuration** and its **repository list file**, then clones the repositories. After completion the workspace is ready to use.

## Prerequisites

- [ghorg](https://github.com/gabrie30/ghorg) must be installed and available on the system `PATH`.
- The user must have network access to the SCM provider (e.g. GitHub, GitLab, Bitbucket).

## Platform support

This skill works on **macOS**, **Linux**, and **Windows**. All file paths are constructed using the resolved environment variables below. Path separators are determined by the host operating system — use `/` on macOS and Linux, and `\` on Windows. When running shell commands, use the appropriate shell for the platform (`bash`/`zsh` on macOS/Linux, `cmd`/`powershell` on Windows).

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| HOME_DIR | Home directory used as the base for all paths | Detected from the operating system (see below) |
| SCM_ORG | SCM organization to clone from | **Required** — must be provided by the user |

### Resolving HOME_DIR

Resolve **HOME_DIR** using the following strategy:

1. If the environment variable `$HOME_DIR` (or `%HOME_DIR%` on Windows) is set, use that value.
2. Otherwise, detect the home directory from the operating system:
   - **macOS / Linux:** Use the `$HOME` environment variable (typically `/Users/<username>` on macOS, `/home/<username>` on Linux).
   - **Windows:** Use the `%USERPROFILE%` environment variable (typically `C:\Users\<username>`).

### Resolving SCM_ORG

Resolve **SCM_ORG** using the following strategy:

1. If the environment variable `$SCM_ORG` (or `%SCM_ORG%` on Windows) is set, use that value.
2. Otherwise, prompt the user:

**What is the SCM organization to clone from?**

Example: `my-github-org`

Store as **SCM_ORG**. This value is required and cannot be empty.

## Paths

All paths are relative to **HOME_DIR**. Use the platform-appropriate path separator.

| Variable | Path |
|----------|------|
| GHORG_CONFIG_DIR | HOME_DIR/.config/ghorg |
| GHORG_RECLONE_CONFIG | HOME_DIR/.config/ghorg/reclone.yaml |
| WORKSPACE_DIR | HOME_DIR/workspaces |
| REPO_FILE_TEMPLATE | HOME_DIR/.config/ghorg/workspace-&lt;WORKSPACE_NAME&gt;.txt |

> **Windows note:** On Windows the ghorg config directory may be at `%USERPROFILE%\.config\ghorg`. Verify by checking if the directory exists or by running `ghorg ls`. Adjust paths accordingly.

## Workflow

### Step 1 — Resolve environment

Resolve **HOME_DIR** and **SCM_ORG** as described in the [Environment variables](#environment-variables) section above. Both values must be resolved before continuing.

### Step 2 — Ask for workspace name

Prompt the user:

**What should the new reclone workspace be called?**

Example: `my-reclone`

**Validation:** Must match regex `^[a-z0-9-]+$` (lowercase letters, numbers, hyphens only).

Store as **WORKSPACE_NAME**. If invalid, ask again until valid.

### Step 3 — Ask for repositories

Prompt the user:

**List any repositories to include in this workspace.**

Enter one repository per line. Press enter on an empty line when finished.

Example input:
```
repo-one
repo-two
repo-three
```

Store as **REPO_LIST**. If none provided, continue with an empty list (empty file).

### Step 4 — Ensure ghorg config exists

- Ensure **HOME_DIR/.config/ghorg** exists; create the directory (including parents) if it does not.
- Ensure **HOME_DIR/.config/ghorg/reclone.yaml** exists; create an empty file if it does not.

### Step 5 — Create workspace repo list file

Create the file:

**HOME_DIR/.config/ghorg/workspace-&lt;WORKSPACE_NAME&gt;.txt**

Example: `HOME_DIR/.config/ghorg/workspace-my-reclone.txt`

- Write the repository list exactly as provided (one repo per line).
- **If the file already exists, do not overwrite it** — stop and report an error.

### Step 6 — Append workspace configuration

Append the following YAML block to **HOME_DIR/.config/ghorg/reclone.yaml**:

```yaml
<WORKSPACE_NAME>:
  cmd: "ghorg clone <SCM_ORG> --path=<HOME_DIR>/workspaces --target-repos-path=<HOME_DIR>/.config/ghorg/workspace-<WORKSPACE_NAME>.txt --output-dir=<WORKSPACE_NAME>"
  description: "Workspace for <WORKSPACE_NAME>"
```

Replace `<WORKSPACE_NAME>`, `<SCM_ORG>`, and `<HOME_DIR>` with the resolved values.

**Rules:**
- **Check first:** If a top-level key matching the workspace name already exists in reclone.yaml, stop and report an error. Do not overwrite.
- Preserve all existing YAML content.
- Append the new workspace to the end of the file.
- Maintain valid YAML formatting (e.g. ensure a newline before the new top-level key if the file was non-empty).

### Step 7 — Run ghorg reclone

Run the following command:

```
ghorg reclone <WORKSPACE_NAME>
```

This clones the repositories into `HOME_DIR/workspaces/<WORKSPACE_NAME>`.

### Step 8 — Create VS Code / Cursor workspace file

Create the file:

**HOME_DIR/workspaces/&lt;WORKSPACE_NAME&gt;/&lt;WORKSPACE_NAME&gt;.code-workspace**

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

- **If the file already exists, overwrite it** (this is safe to regenerate).

### Step 9 — Confirmation

After all steps complete, output:

```
Workspace created and cloned successfully.

Workspace name:
<WORKSPACE_NAME>

SCM organization:
<SCM_ORG>

Repo list file:
<HOME_DIR>/.config/ghorg/workspace-<WORKSPACE_NAME>.txt

Config updated:
<HOME_DIR>/.config/ghorg/reclone.yaml

Cloned to:
<HOME_DIR>/workspaces/<WORKSPACE_NAME>

VS Code workspace file:
<HOME_DIR>/workspaces/<WORKSPACE_NAME>/<WORKSPACE_NAME>.code-workspace
```

Replace `<WORKSPACE_NAME>`, `<SCM_ORG>`, and `<HOME_DIR>` with the resolved values.

## Failure conditions

Stop and report an error (do not silently continue) if:

- **HOME_DIR** cannot be resolved from the environment.
- **SCM_ORG** is not set and the user does not provide a value when prompted.
- The workspace name already exists as a top-level key in `reclone.yaml`.
- The repo list file **HOME_DIR/.config/ghorg/workspace-&lt;WORKSPACE_NAME&gt;.txt** already exists.
- The `ghorg reclone` command fails (non-zero exit code).
- Filesystem write fails (directory or file creation).
- YAML formatting cannot be preserved when appending.

Never silently continue after an error.
