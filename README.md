# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a> [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![WakeMeOps](https://docs.wakemeops.com/badges/ghorg.svg)](https://docs.wakemeops.com//packages/ghorg)

Pronounced [gore-guh]; similar to [gorge](https://www.dictionary.com/browse/gorge). You can use ghorg to gorge on orgs. Time to ghorg!

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
  - [Install](https://github.com/gabrie30/ghorg#install) | [Setup](https://github.com/gabrie30/ghorg#github-setup) | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/github.md)
- GitLab (Self Hosted & Cloud)
  - [Install](https://github.com/gabrie30/ghorg#install) | [Setup](https://github.com/gabrie30/ghorg#gitlab-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md)
- Bitbucket (Cloud Only)
  - [Install](https://github.com/gabrie30/ghorg#install) | [Setup](https://github.com/gabrie30/ghorg#bitbucket-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md)
- Gitea (Self Hosted Only)
  - [Install](https://github.com/gabrie30/ghorg#install) | [Setup](https://github.com/gabrie30/ghorg#gitea-setup)  | [Examples](https://github.com/gabrie30/ghorg/blob/master/examples/gitea.md)

> The terminology used in ghorg is that of GitHub, mainly orgs/repos. GitLab and BitBucket use different terminology. There is a handy chart thanks to GitLab that translates terminology [here](https://about.gitlab.com/images/blogimages/gitlab-terminology.png). Note, some features may be different for certain providers.

## Windows support

Windows is supported when built with golang or as a [prebuilt binary](https://github.com/gabrie30/ghorg/releases/latest) however, the readme and other documentation is not geared towards windows users.

## Configuration

Precedence for configuration is first given to the flags set on the command-line, then to what's set in your `$HOME/.config/ghorg/conf.yaml`. This file comes from the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml).

Although it's optional, it is recommended to add a `$HOME/.config/ghorg/conf.yaml` following the instructions in the install section.

You can have multiple configuration files which is useful if you clone from multiple SCM providers. Alternative configuration files can only be referenced as a command-line flag `--config`.

If you have multiple different orgs/users/configurations to clone see the `ghorg reclone` command as a way to manage them.

Note: ghorg will respect the `XDG_CONFIG_HOME` [environment variable](https://wiki.archlinux.org/title/XDG_Base_Directory) if set.

## Install

### Prebuilt Binaries

See [latest release](https://github.com/gabrie30/ghorg/releases/latest) to download directly for

- Mac (Darwin)
- Windows
- Linux

If you don't know which to choose its likely going to be the x86_64 version for your operating system

### Homebrew

> optional but recommended

```bash
mkdir -p $HOME/.config/ghorg
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
vi $HOME/.config/ghorg/conf.yaml # To update your configuration
```
> required

```bash
brew install gabrie30/utils/ghorg
```

### Golang

> optional but recommended

```bash
mkdir -p $HOME/.config/ghorg
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
vi $HOME/.config/ghorg/conf.yaml # To update your configuration
```

> required

```bash
# ensure $HOME/go/bin is in your path ($ echo $PATH | grep $HOME/go/bin)

# if using go 1.16+ locally
go install github.com/gabrie30/ghorg@latest

# older go versions can run
go get github.com/gabrie30/ghorg
```

## SCM Provider Setup

> Note: if you are running into issues, read the troubleshooting and known issues section below

### GitHub Setup
1. Create [Personal Access Token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) with all `repo` scopes. Update `GHORG_GITHUB_TOKEN` in your `ghorg/conf.yaml` or as a cli flag. If your org has Saml SSO in front you will need to give your token those permissions as well, see [this doc](https://docs.github.com/en/github/authenticating-to-github/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).
1. For cloning GitHub Enterprise (self hosted github instances) repos you must set `--base-url` e.g. `ghorg clone <github_org> --base-url=https://internal.github.com`
1. See [examples/github.md](https://github.com/gabrie30/ghorg/blob/master/examples/github.md) on how to run

### GitLab Setup

1. Create [Personal Access Token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) with the `read_api` scope (or `api` for self-managed GitLab older than 12.10). This token can be added to your `ghorg/conf.yaml` or as a cli flag.
1. Update the `GitLab Specific` config in your `ghorg/conf.yaml` or via cli flags
1. Update `GHORG_SCM_TYPE` to `gitlab` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/gitlab.md](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md) on how to run

#### GitLab Specific Notes

1. With GitLab, if ghorg detects repo naming collisions with repos being cloned from different groups/subgroups, ghorg will automatically append the group/subgroup path to the repo name. You will be notified in the output if this occurs.

1. There are different commands for hosted gitlab instances vs gitlab cloud read below for the differences.

##### Hosted GitLab Instances

1. To clone all the groups at once use the keyword "all-groups". Note, all-groups requires a GitLab 13.0.1 or greater and will only clone from groups/repos your user has permissions to.

    ```sh
    $ ghorg clone all-groups --base-url=https://${your.hosted.gitlab.com} --scm=gitlab --token=XXXX --preserve-dir
    ```

1. For all versions of GitLab you can clone groups or sub groups individually

    ```sh
    # cloning a top level group
    $ ghorg clone mygroup --base-url=https://${your.hosted.gitlab.com} --scm=gitlab --token=XXXX --preserve-dir

    # cloning a subgroup
    $ ghorg clone mygroup/mysubgroup --base-url=https://${your.hosted.gitlab.com} --scm=gitlab --token=XXXX --preserve-dir
    ```

1. You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

##### GitLab Cloud

To clone all repos you can use the top level group name e.g. to clone `gitlab-examples` on GitLab cloud https://gitlab.com/gitlab-examples

```sh
$ ghorg clone gitlab-examples --scm=gitlab --token=XXXX --preserve-dir
```

### Gitea Setup

1. Create [Access Token](https://docs.gitea.io/en-us/api-usage/) (Settings -> Applications -> Generate Token)
1. Update `GHORG_GITEA_TOKEN` in your `ghorg/conf.yaml` or use the (--token, -t) flag.
1. Update `GHORG_SCM_TYPE` to `gitea` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/gitea.md](https://github.com/gabrie30/ghorg/blob/master/examples/gitea.md) on how to run

### Bitbucket Setup

#### App Passwords

1. To configure with bitbucket you will need to create a new [app password](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) and update your `$HOME/.config/ghorg/conf.yaml` or use the (--token, -t) and (--bitbucket-username) flags.
1. Update [SCM type](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml#L54-L57) to `bitbucket` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/bitbucket.md](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md) on how to run

#### PAT/OAuth token

1. Create a [PAT](https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html)
1. Set the token with `GHORG_BITBUCKET_OAUTH_TOKEN` in your `$HOME/.config/ghorg/conf.yaml` or using the `--token` flag. Make sure you do not have `--bitbucket-username` set.
1. Update SCM TYPE to `bitbucket` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/bitbucket.md](https://github.com/gabrie30/ghorg/blob/master/examples/bitbucket.md) on how to run


## How to Use

See [examples](https://github.com/gabrie30/ghorg/tree/master/examples) dir for more SCM specific docs

```bash
# note: to view/set all available flags/features see sample-conf.yaml
$ ghorg clone kubernetes --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone davecheney --clone-type=user --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone gitlab-examples --scm=gitlab --preserve-dir --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone all-groups --scm=gitlab --base-url=https://gitlab.internal.yourcompany.com --preserve-dir
$ ghorg clone --help
# view cloned resources
$ ghorg ls
$ ghorg ls someorg
```

### With Docker

> This is only recommended for testing due to resource constraints

1. Clone repo then `cd ghorg`
1. Build the image `docker build . -t ghorg-docker`
1. Run in docker

```bash
# using your local ghorg configuration file, cloning in container
docker run -v $HOME/.config/ghorg/conf.yaml:/root/.config/ghorg/conf.yaml ghorg-docker ./ghorg clone kubernetes

# using flags, cloning in container
docker run ghorg-docker ./ghorg clone kubernetes --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2

# using flags, cloning to your machine
docker run -v $HOME/ghorg/:/root/ghorg/ ghorg-docker ./ghorg clone kubernetes --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2 --output-dir=cloned-from-docker
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
## Filtering Repos
- To only clone repos that match regex use `--match-regex` flag or exclude cloning repos that match regex with `--exclude-match-regex`
- To only clone repos that match prefix(s) use `--match-prefix` flag or exclude cloning repos that match prefix(s) with `--exclude-match-prefix`
- To filter out any archived repos while cloning use the `--skip-archived` flag (not bitbucket)
- To filter out any forked repos while cloning use the `--skip-forks` flag
- Filter by specific repo [topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics) `GHORG_TOPICS` or `--topics` will clone only repos with a matching topic. GitHub/GitLab/Gitea only
- To ignore specific repos create a `ghorgignore` file inside `$HOME/.config/ghorg`. Each line in this file is considered a substring and will be compared against each repos clone url. If the clone url contains a substring in the `ghorgignore` it will be excluded from cloning. To prevent accidentally excluding a repo, you should make each line as specific as possible, eg. `https://github.com/gabrie30/ghorg.git` or `git@github.com:gabrie30/ghorg.git` depending on how you clone. This is useful for permanently ignoring certain repos.

  ```bash
  # Create ghorgignore
  touch $HOME/.config/ghorg/ghorgignore

  # Update file
  vi $HOME/.config/ghorg/ghorgignore
  ```

## Creating Backups

When taking backups the two notable flags are `--backup`, `--clone-wiki`, and `--include-submodules`. The `--backup` flag will clone the repo with [git clone --mirror](https://www.git-scm.com/docs/git-clone#Documentation/git-clone.txt---mirror). The `--clone-wiki` flag will include any wiki pages the repo has. If you want to include any submodules you will need `--include-submodules`. Lastly, if you want to exclude any binary files use the the flag `--git-filter=blob:none` to prevent them from being cloned.

```
ghorg clone kubernetes --backup --clone-wiki --include-submodules --git-filter=blob:none
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

Once your [reclone.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-reclone.yaml) configuration is set you can call `ghorg reclone` to clone each entry individually or clone all at once.

To use, add a [reclone.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-reclone.yaml) to your `$HOME/.config/ghorg` directory. You can use the following command to set it for you with examples to use as a template

```
curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-reclone.yaml > $HOME/.config/ghorg/reclone.yaml
```

After updating your `reclone.yaml` you can run

```
# To clone all the entries in your reclone.yaml omit any arguments
ghorg reclone
```

```
# To run one or more entries you can pass arguments
ghorg reclone kubernetes-sig-staging kubernetes-sig
```

<p align="center">
  <img width="648" alt="ghorg reclone example" src="https://user-images.githubusercontent.com/1512282/183263986-50e56b86-12b9-479b-9c52-b1c74129228c.png">
</p>

## Troubleshooting

- If you are having trouble cloning repos. Try to clone one of the repos locally e.g. manually running `git clone https://github.com/your_private_org/your_private_repo.git` if this does not work, ghorg will also not work. Your git client must first be setup to clone the target repos. If you normally clone using an ssh key use the `--protocol=ssh` flag with ghorg. This will fetch the ssh clone urls instead of the https clone urls.
- If you are cloning a large org you may see `Error: open /dev/null: too many open files` which means you need to increase your ulimits, there are lots of docs online for this. Another solution is to decrease the number of concurrent clones. Use the `--concurrency` flag to set to lower than 25 (the default)
- If your GitHub org is behind SSO, you will need to authorize your token, see [here](https://docs.github.com/en/github/authenticating-to-github/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on)
- If your GitHub Personal Access Token is only finding public repos, give your token all the repos permissions
- Make sure your `$ git --version` is >= 2.19.0
- Check for other software, such as anti-malware, that could interfere with ghorgs ability to create large number of connections, see [issue 132](https://github.com/gabrie30/ghorg/issues/132#issuecomment-889357960). You can also lower the concurrency with `--concurrency=n` default is 25.
- To debug yourself you can call ghorg with the GHORG_DEBUG=true env e.g `GHORG_DEBUG=true ghorg clone kubernetes --concurrency=1`
- If you've gotten this far and still have an issue feel free to raise an issue
