# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a> [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![WakeMeOps](https://docs.wakemeops.com/badges/ghorg.svg)](https://docs.wakemeops.com//packages/ghorg)

Pronounced [gore-guh]; similar to [gorge](https://www.dictionary.com/browse/gorge). You can use ghorg to gorge on orgs.

Use ghorg to quickly clone all of an orgs, or users repos into a single directory. This can be useful in many situations including

1. Searching an orgs/users codebase with ack, silver searcher, grep etc..
1. Bash scripting
1. Creating backups
1. Onboarding new team members (cloning all team repos)
1. Performing Audits

> With default configuration ghorg performs two actions.
> 1. Will clone a repo if its not inside the clone directory.
> 2. If repo does exists locally in the clone directory it will perform a git pull and git clean on the repo.

> So when running ghorg a second time on the same org/user, all local changes in the cloned directory by default will be overwritten by what's on GitHub. If you want to work out of this directory, make sure you either rename the directory or set the `--no-clean` flag on all future clones to prevent losing your changes locally.

<p align="center">
  <img width="648" alt="ghorg cli example" src="https://user-images.githubusercontent.com/1512282/63229247-5459f880-c1b3-11e9-9e5d-d20723046946.png">
</p>

## Supported Providers
- GitHub (Self Hosted & Cloud)
  - [Install](https://github.com/gabrie30/ghorg#installation) | [Setup](https://github.com/gabrie30/ghorg#github-setup) | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/github.md)
- GitLab (Self Hosted & Cloud)
  - [Install](https://github.com/gabrie30/ghorg#installation) | [Setup](https://github.com/gabrie30/ghorg#gitlab-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md)
- Bitbucket (Cloud & Self-hosted Server)
  - [Install](https://github.com/gabrie30/ghorg#installation) | [Setup](https://github.com/gabrie30/ghorg#bitbucket-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md)
- Gitea (Self Hosted Only)
  - [Install](https://github.com/gabrie30/ghorg#installation) | [Setup](https://github.com/gabrie30/ghorg#gitea-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/gitea.md)

> The terminology used in ghorg is that of GitHub, mainly orgs/repos. GitLab and BitBucket use different terminology. There is a handy chart thanks to GitLab that translates terminology [here](https://about.gitlab.com/images/blogimages/gitlab-terminology.png). Note, some features may be different for certain providers.

## High Level Features

- [Filter](#selective-repository-cloning) or select specific repositories for cloning
- Create [backups](#creating-backups) of repositories
- Simplify complex clone commands using [reclone](#reclone-command) shortcuts
- Initiate clone operations via [HTTP server](#reclone-server-command)
- Schedule cloning tasks using [cron](#reclone-cron-command)
- Monitor and track clone [metrics](#tracking-clone-data-over-time) over time

## Installation

There are a installation methods available, please choose the one that suits your fancy:
- [Prebuilt Binaries](#prebuilt-binaries)
- [Homebrew](#homebrew)
- [Golang](#golang)
- [Docker](#docker)
- [Windows Support](#windows-support)

For each installation method, optionally create a ghorg configuration file. See the [configuration](#configuration) section for more details.

```bash
mkdir -p $HOME/.config/ghorg
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
vi $HOME/.config/ghorg/conf.yaml # To update your configuration
```

### Prebuilt Binaries

See [latest release](https://github.com/gabrie30/ghorg/releases/latest) to download directly for

- Mac (Darwin)
- Windows
- Linux

If you don't know which to choose its likely going to be the x86_64 version for your operating system.

### Homebrew

```bash
brew install gabrie30/utils/ghorg
```

### Golang

```bash
# ensure $HOME/go/bin is in your path ($ echo $PATH | grep $HOME/go/bin)

# if using go 1.16+ locally
go install github.com/gabrie30/ghorg@latest

# older go versions can run
go get github.com/gabrie30/ghorg
```

## Configuration

Precedence for configuration is first given to the flags set on the command-line, then to what's set in your `$HOME/.config/ghorg/conf.yaml`. This file comes from the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) and can be installed by performing the following.

```bash
mkdir -p $HOME/.config/ghorg
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
vi $HOME/.config/ghorg/conf.yaml # To update your configuration
```

If no configuration file is found ghorg will use its defaults and try to clone a GitHub Org, however an api token is always required.

You can have multiple configuration files which is useful if you clone from multiple SCM providers with different tokens and settings. Alternative configuration files can only be referenced as a command-line flag `--config`.

If you have multiple different orgs/users/configurations to clone see the `ghorg reclone` command as a way to manage them.

Note: ghorg will respect the `XDG_CONFIG_HOME` [environment variable](https://wiki.archlinux.org/title/XDG_Base_Directory) if set.

## SCM Provider Setup

> Note: if you are running into issues, read the troubleshooting and known issues section below

### GitHub Setup
1. Create [Personal Access Token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) with all `repo` scopes. Update `GHORG_GITHUB_TOKEN` in your `ghorg/conf.yaml` or as a cli flag or place it in a file and add the path to `GHORG_GITHUB_TOKEN`. If your org has Saml SSO in front you will need to give your token those permissions as well, see [this doc](https://docs.github.com/en/github/authenticating-to-github/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).
1. For cloning GitHub Enterprise (self hosted github instances) repos you must set `--base-url` e.g. `ghorg clone <github_org> --base-url=https://internal.github.com`
1. See [examples/github.md](https://github.com/gabrie30/ghorg/blob/master/examples/github.md) on how to run

#### GitHub App Authentication (Advanced)

1. [Create a GitHub App](https://docs.github.com/en/apps/creating-github-apps/setting-up-a-github-app/creating-a-github-app) in your Organization. You only need to fill out the required fields. Make sure to give Repository Permissions ->  contents -> read only permissions
1. Install the GitHub App into your Organization
1. Generate a a private key from the GitHub App, set the location of the key to `GHORG_GITHUB_APP_PEM_PATH`
1. Locate the GitHub App ID from the GitHub App, set the value to `GHORG_GITHUB_APP_ID`
1. Locate the GitHub Installation ID from the URL of the GitHub app, set the value to `GHORG_GITHUB_APP_INSTALLATION_ID`. NOTE: you will need to use the actual GitHub url to get this ID, go to your GitHub Organization Settings Page -> Third Party Access -> GitHub Apps -> Configure -> Get ID from URL

### GitLab Setup

1. Create [Personal Access Token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) with the `read_api` scope (or `api` for self-managed GitLab older than 12.10). This token can be added to your `ghorg/conf.yaml` or as a cli flag.
1. Update the `GitLab Specific` config in your `ghorg/conf.yaml` or via cli flags or place it in a file and add the path to `GHORG_GITLAB_TOKEN`
1. Update `GHORG_SCM_TYPE` to `gitlab` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/gitlab.md](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md) on how to run

### Gitea Setup

1. Create [Access Token](https://docs.gitea.io/en-us/api-usage/) (Settings -> Applications -> Generate Token)
1. Update `GHORG_GITEA_TOKEN` in your `ghorg/conf.yaml` or use the (--token, -t) flag or place it in a file and add the path to `GHORG_GITEA_TOKEN`.
1. Update `GHORG_SCM_TYPE` to `gitea` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/gitea.md](https://github.com/gabrie30/ghorg/blob/master/examples/gitea.md) on how to run

### Bitbucket Setup

> Note: ghorg supports both Bitbucket Cloud and Bitbucket Server (self-hosted instances)

#### App Passwords

1. To configure with bitbucket you will need to create a new [app password](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) and update your `$HOME/.config/ghorg/conf.yaml` or use the (--token, -t) and (--bitbucket-username) flags.
1. Update [SCM type](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml#L54-L57) to `bitbucket` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/bitbucket.md](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md) on how to run

#### PAT/OAuth token

1. Create a [PAT](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html)
1. Set the token with `GHORG_BITBUCKET_OAUTH_TOKEN` in your `$HOME/.config/ghorg/conf.yaml` or using the `--token` flag. Make sure you do not have `--bitbucket-username` set.
1. Update SCM TYPE to `bitbucket` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/bitbucket.md](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md) on how to run

#### Bitbucket Server (Self-hosted)

1. To configure with Bitbucket Server you will need to provide your instance URL via `GHORG_SCM_BASE_URL` in your `$HOME/.config/ghorg/conf.yaml` or use the `--base-url` flag.
1. Create credentials (username/password or app password) and update your configuration or use the `--bitbucket-username` and `--token` flags.
1. For insecure connections (HTTP), set `GHORG_INSECURE_BITBUCKET_CLIENT=true`
1. Update [SCM type](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml#L54-L57) to `bitbucket` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/bitbucket.md](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md) on how to run

## How to Use

See [examples](https://github.com/gabrie30/ghorg/tree/master/examples) directory for more SCM specific docs or use the examples command e.g. `ghorg examples gitlab`

```bash
$ ghorg clone kubernetes --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
# Example how to use --token with a file path
$ ghorg clone kubernetes --token=~/.config/ghorg/gitlab-token.txt
$ ghorg clone davecheney --clone-type=user --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone gitlab-examples --scm=gitlab --preserve-dir --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone all-groups --scm=gitlab --base-url=https://gitlab.internal.yourcompany.com --preserve-dir
$ ghorg clone --help
# view cloned resources
$ ghorg ls
$ ghorg ls someorg
$ ghorg ls someorg | xargs -I %s mv %s bar/
```

## Changing Clone Directories

1. By default ghorg will clone the org or user repos into a directory like `$HOME/ghorg/org`. If you want to clone the org to a different directory use the `--path` flag or set `GHORG_ABSOLUTE_PATH_TO_CLONE_TO` in your ghorg conf. **This value must be an absolute path**. For example if you wanted to clone the kubernetes org to `/tmp/ghorg` you would run the following command.

    ```
    $ ghorg clone kubernetes --path=/tmp/ghorg
    ```
    which would create...

    ```
    /tmp/ghorg
    └── kubernetes
        ├── apimachinery
        ├── gengo
        ├── git-sync
        ├── kubeadm
        ├── kubernetes-template-project
        ├── ...
    ```

1. If you want to change the name of the directory the repos get cloned into, set the `GHORG_OUTPUT_DIR` in your ghorg conf or set the `--output-dir` flag. For example to clone only the repos starting with `sig-` from the kubernetes org into a direcotry called `kubernetes-sig-only`. You would run the following command.

    ```
    $ ghorg clone kubernetes --match-regex=^sig- --output-dir=kubernetes-sig-only
    ```
    which would create...

    ```
    $HOME/ghorg
    └── kubernetes-sig-only
        ├── sig-release
        ├── sig-security
        └── sig-testing
    ```
## Selective Repository Cloning
- To only clone repos that match regex use `--match-regex` flag or exclude cloning repos that match regex with `--exclude-match-regex`
- To only clone repos that match prefix(s) use `--match-prefix` flag or exclude cloning repos that match prefix(s) with `--exclude-match-prefix`
- To filter out any archived repos while cloning use the `--skip-archived` flag (not bitbucket)
- To filter out any forked repos while cloning use the `--skip-forks` flag
- Filter by specific repo [topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics) `GHORG_TOPICS` or `--topics` will clone only repos with a matching topic. GitHub/GitLab/Gitea only
- To clone a specific set of repositories, create a file listing the names of these repositories, one per line. Then, use the `GHORG_TARGET_REPOS` or `--target-repos-path` flag to specify the path to this file.
- To exclude specific repositories from being cloned, you can create a `ghorgignore` file in the `$HOME/.config/ghorg` directory. Each line in this file should contain a unique identifier of the repository, which is considered a substring. This substring is then compared against each repository's clone URL during the cloning process. If the clone URL contains the substring listed in the `ghorgignore` file, that repository will be skipped and not cloned. To avoid unintentionally excluding a repository, ensure that each line in the `ghorgignore` file is as specific as possible. For instance, you could use `https://github.com/gabrie30/ghorg.git` or `git@github.com:gabrie30/ghorg.git`, depending on your cloning method. This feature is particularly useful for permanently excluding certain repositories from the cloning process. If you wish to use multiple `ghorgignore` files or store them in a different location, you can use the `--ghorgignore-path` flag to specify an alternative path.
  ```bash
  # Create ghorgignore
  touch $HOME/.config/ghorg/ghorgignore

  # Update file
  vi $HOME/.config/ghorg/ghorgignore
  ```

- To clone only specific repositories matching certain patterns, you can create a `ghorgonly` file in the `$HOME/.config/ghorg` directory. Each line in this file should contain a substring pattern. Only repositories whose clone URLs contain these patterns will be cloned. This is useful when you want to clone a specific subset of repositories from a large organization. The `ghorgonly` filter is applied first, followed by `ghorgignore`, allowing you to combine both for fine-grained control (e.g., clone only `api-*` repos but exclude `api-legacy`). Like `ghorgignore`, you can specify an alternative path using the `--ghorgonly-path` flag.
  ```bash
  # Create ghorgonly
  touch $HOME/.config/ghorg/ghorgonly

  # Update file with patterns (one per line)
  vi $HOME/.config/ghorg/ghorgonly
  ```

## Creating Backups

When taking backups the notable flags are `--backup`, `--clone-wiki`, and `--include-submodules`. The `--backup` flag will clone the repo with [git clone --mirror](https://www.git-scm.com/docs/git-clone#Documentation/git-clone.txt---mirror). The `--clone-wiki` flag will include any wiki pages the repo has. If you want to include any submodules you will need `--include-submodules`. Lastly, if you want to exclude any binary files use the the flag `--git-filter=blob:none` to prevent them from being cloned.

```
ghorg clone kubernetes --backup --clone-wiki --include-submodules
```

This will create a kubernetes_backup directory for the org. Each folder inside will contain the .git contents for the source repo. To restore the code from the .git contents you would move all contents into a .git dir, then run `git init` inside the dir, then checkout branch e.g.

```sh
# inside kubernetes_backup dir, to restore kubelet source code
cd kubelet
mkdir .git
mv -f * .git # moves all contents into .git directory
git init
git checkout master
```

## Reclone Command

The `ghorg reclone` command is a way to store all your `ghorg clone` commands in one configuration file and makes calling long or multiple `ghorg clone` commands easier.

Once your [reclone.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-reclone.yaml) configuration is set you can call `ghorg reclone` to clone each entry individually or clone all at once, see examples below.

Each reclone entry can have:
- `cmd`: The ghorg clone command to execute (required)
- `description`: A description of what the command does (optional)
- `post_exec_script`: Path to a script that will be called after the clone command finishes (optional). The script will always be called, regardless of success or failure, and receives two arguments: the status (`success` or `fail`) and the name of the reclone entry. This allows you to implement custom notifications, monitoring, or other automation (optional)

Example `reclone.yaml` entry:

```yaml
gitlab-examples:
  cmd: "ghorg clone gitlab-examples --scm=gitlab --token=XXXXXXX"
  post_exec_script: "/path/to/notify.sh"
```

Example script for `post_exec_script` (e.g. `/path/to/notify.sh`):

```sh
#!/bin/sh
STATUS="$1"
NAME="$2"

if [ "$STATUS" = "success" ]; then
  # Success webhook
  curl -fsS https://hc-ping.com/your-uuid-here
else
  # Failure webhook
  curl -fsS https://hc-ping.com/your-uuid-here/fail
fi
```

```
# To clone all the entries in your reclone.yaml omit any arguments
ghorg reclone
```

```
# To run one or more entries you can pass arguments
ghorg reclone kubernetes-sig-staging kubernetes-sig
```

```
# To view all your reclone commands
# NOTE: This command prints tokens to stdout
ghorg reclone --list
```

<p align="center">
  <img width="648" alt="ghorg reclone example" src="https://user-images.githubusercontent.com/1512282/183263986-50e56b86-12b9-479b-9c52-b1c74129228c.png">
</p>

#### Setup

Add a [reclone.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-reclone.yaml) to your `$HOME/.config/ghorg` directory. You can use the following command to set it for you with examples to use as a template

```
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-reclone.yaml > $HOME/.config/ghorg/reclone.yaml
```

Update file with the commands you wish to run.

## Reclone Server Command

The `reclone-server` command starts a server that allows you to trigger ad hoc reclone commands via HTTP requests.

### Usage

```sh
ghorg reclone-server [flags]
```

### Flags

- `--port`: Specify the port on which the server will run. If not specified, the server will use the default port.

### Endpoints

- **`/trigger/reclone`**: Triggers the reclone command. To prevent resource exhaustion, only one request can processed at a time.
  - **Query Parameters**:
    - `cmd`: Optional. Allows you to call a specific reclone, otherwise all reclones are ran.
  - **Responses**:
    - `200 OK`: Command started successfully.
    - `429 Too Many Requests`: Server is currently running a reclone command, you will need to wait until its completed before starting another one.

- **`/stats`**: Returns the statistics of the reclone operations in JSON format. `GHORG_STATS_ENABLED=true` or `--stats-enabled` must be set to work.
  - **Responses**:
    - `200 OK`: Statistics returned successfully.
    - `428 Precondition required`: Ghorg stats is not enabled.
    - `500 Internal Server Error`: Unable to read the statistics file.

- **`/health`**: Health check endpoint.
  - **Responses**:
    - `200 OK`: Server is healthy.

### Examples

Starting the server. The default port is `8080` but you can optionally start the server on different port using the `--port` flag:

```sh
ghorg reclone-server
```

Trigger reclone command, this will run all cmds defined in your `reclone.yaml`:

```sh
curl "http://localhost:8080/trigger/reclone"
```

Trigger a specific reclone command:

```sh
curl "http://localhost:8080/trigger/reclone?cmd=your-reclone-command"
```

Get the statistics:

```sh
curl "http://localhost:8080/stats"
```

Check the server health:

```sh
curl "http://localhost:8080/health"
```

## Reclone Cron Command

The `reclone-cron` command sets up a simple cron job that triggers the reclone command at specified minute intervals indefinitely.

### Usage

```sh
ghorg reclone-cron [flags]
```

### Flags

- `--minutes`: Specify the interval in minutes at which the reclone command will be triggered. Default is every 60 minutes.

### Example

Set up a cron job to trigger the reclone command every day:

```sh
ghorg reclone-cron --minutes 1440
```

### Environment Variables

- `GHORG_CRON_TIMER_MINUTES`: The interval in minutes for the cron job. This can be set via the `--minutes` flag. Default is 60 minutes.

## Using Docker

The provided images are built for both `amd64` and `arm64` architectures and are available solely on Github Container Registry [ghcr.io](https://github.com/gabrie30/ghorg/pkgs/container/ghorg).

```shell
# Should print help message
# You can also specify a version as the tag, such as ghcr.io/gabrie30/ghorg:v1.9.9
docker run --rm ghcr.io/gabrie30/ghorg:latest
```

> Note: There are also tags available for the latest on trunk, such as `master` or `master-<commit SHA 7 chars>`, but these **are not recommended**.

The commands for ghorg are parsed as docker commands. The entrypoint is the `ghorg` binary, hence you only need to enter remaining arguments as follows:

```shell
docker run --rm ghcr.io/gabrie30/ghorg \
    clone kubernetes --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

The image ships with the following environment variables set:

```shell
GHORG_CONFIG=/config/conf.yaml
GHORG_RECLONE_PATH=/config/reclone.yaml
GHORG_ABSOLUTE_PATH_TO_CLONE_TO=/data
```

These can be overriden, if necessary, by including the `-e` flag to the docker run comand, e.g. `-e GHORG_GITHUB_TOKEN=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2`.

### Persisting Data on the Host

In order to store data on the host, it is required to bind mount a volume:
- `$HOME/.config/ghorg:/config`: Mounts your config directory inside the container, to access `config.yaml` and `reclone.yaml`.
- `$HOME/repositories:/data`: Mounts your local data directory inside the container, where repos will be downloaded by default.

```shell
docker run --rm \
        -e GHORG_GITHUB_TOKEN=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2 \
        -v $HOME/.config/ghorg:/config `# optional` \
        -v $HOME/repositories:/data \
        ghcr.io/gabrie30/ghorg:latest \
        clone kubernetes --match-regex=^sig
```

> Note: Altering `GHORG_ABSOLUTE_PATH_TO_CLONE_TO` will require changing the mount location from `/data` to the new location inside the container.

A shell alias might make this more practical:

```shell
alias ghorg="docker run --rm -v $HOME/.config/ghorg:/config -v $HOME/repositories:/data ghcr.io/gabrie30/ghorg:latest"

# Using the alias: creates and cleans up the container
ghorg clone kubernetes --match-regex=^sig
```

## Tracking Clone Data Over Time

To track data on your clones over time, you can use the ghorg stats feature. It is recommended to enable ghorg stats in your configuration file by setting `GHORG_STATS_ENABLED=true`. This ensures that each clone operation is logged automatically without needing to set the command line flag `--stats-enabled` every time. **The ghorg stats feature is disabled by default and needs to be enabled.**

When ghorg stats is enabled, the CSV file `_ghorg_stats.csv` is created in the directory specified by `GHORG_ABSOLUTE_PATH_TO_CLONE_TO`. This file contains detailed information about each clone operation, which is useful for auditing and tracking purposes such as the size of the clone and the number of new commits over time.

Below are the headers and their descriptions. Note that these headers may change over time. If there are any changes in the headers, a new file named `_ghorg_stats_new_header_${sha256HashOfHeader}.csv` will be created to prevent incorrect data from being added to your CSV.

- **datetime**: Date and time of the clone in YYYY-MM-DD hh:mm:ss format
- **clonePath**: Location of the clone directory
- **scm**: Name of the source control used
- **cloneType**: Either user or org clone
- **cloneTarget**: What is specified after the clone command `ghorg clone <target>`
- **totalCount**: Total number of resources expected to be cloned or pulled
- **newClonesCount**: Sum of all new repos cloned
- **existingResourcesPulledCount**: Sum of all repos that were pulled
- **dirSizeInMB**: The size in megabytes of the output dir
- **newCommits**: Sum of all new commits in all repos pulled
- **cloneInfosCount**: Number of clone Info messages
- **cloneErrorsCount**: Number of clone Issues/Errors
- **updateRemoteCount**: Number of remotes updated
- **pruneCount**: Number of repos pruned
- **hasCollisions**: If there were any name collisions, only can happen with gitlab clones
- **ghorgignore**: If a ghorgignore was used in the clone
- **ghorgonly**: If a ghorgonly was used in the clone
- **totalDurationSeconds**: Total time in seconds for the entire clone operation
- **ghorgVersion**: Version of ghorg used in the clone

#### Converting CSV to JSON

```bash
go install github.com/gabrie30/csvToJson@latest && \
csvToJson _ghorg_stats.csv
```

## Windows support

Windows is supported when built with golang or as a [prebuilt binary](https://github.com/gabrie30/ghorg/releases/latest) however, the readme and other documentation is not geared towards Windows users.

Alternatively, Windows users can also install ghorg using [scoop](https://scoop.sh/#/)

  ```
  scoop bucket add main
  scoop install ghorg
  ```

## Troubleshooting

- If you are having trouble cloning repos. Try to clone one of the repos locally e.g. manually running `git clone https://github.com/your_private_org/your_private_repo.git` if this does not work, ghorg will also not work. Your git client must first be setup to clone the target repos. If you normally clone using an ssh key use the `--protocol=ssh` flag with ghorg. This will fetch the ssh clone urls instead of the https clone urls.
- If you are cloning a large org you may see `Error: open /dev/null: too many open files` which means you need to increase your ulimits, there are lots of docs online for this. Another solution is to decrease the number of concurrent clones. Use the `--concurrency` flag to set to lower than 25 (the default)
- If your GitHub org is behind SSO, you will need to authorize your token, see [here](https://docs.github.com/en/github/authenticating-to-github/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on)
- If your GitHub Personal Access Token is only finding public repos, give your token all the repos permissions
- Make sure your `$ git --version` is >= 2.19.0
- Check for other software, such as anti-malware, that could interfere with ghorgs ability to create large number of connections, see [issue 132](https://github.com/gabrie30/ghorg/issues/132#issuecomment-889357960). You can also lower the concurrency with `--concurrency=n` default is 25.
- To debug yourself you can call ghorg with the GHORG_DEBUG=true env e.g `GHORG_DEBUG=true ghorg clone kubernetes`. Note, when this env is set concurrency is set to a value of 1 and will expose the api key used to stdout.
- If you've gotten this far and still have an issue feel free to raise an issue
- If you’re cloning using https, but you have submodules which are configured to use ssh, you can force git to pull these submodules as well via https by running these commands before running ghorg:
  ```
  git config --global url."https://github.com/".insteadOf git@github.com:
  git config --global credential.https://github.com/.helper '! f() { echo username=x-access-token; echo password=$GHORG_GITHUB_TOKEN; };f'
  ```
