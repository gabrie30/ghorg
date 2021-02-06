# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a> [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

ghorg allows you to quickly clone all of an orgs, or users repos into a single directory. This can be useful in many situations including

1. Searching an orgs/users codebase with ack, silver searcher, grep etc..
2. Bash scripting
3. Creating backups
4. Onboarding
5. Performing Audits

> When running ghorg a second time, all local changes in your *_ghorg directory will be overwritten by whats on GitHub. If you are working out of this directory, make sure you rename it before running a second time otherwise all of your changes will be lost.

<p align="center">
  <img width="648" alt="ghorg cli example" src="https://user-images.githubusercontent.com/1512282/63229247-5459f880-c1b3-11e9-9e5d-d20723046946.png">
</p>

## Supported Providers
- GitHub
- GitLab
- Bitbucket
- Gitea

> The terminology used in ghorg is that of GitHub, mainly orgs/repos. GitLab and BitBucket use different terminology. There is a handy chart thanks to GitLab that translates terminology [here](https://about.gitlab.com/images/blogimages/gitlab-terminology.png)

## Install

### Homebrew

> optional

```bash
$ brew update
$ brew upgrade git
```
> required

```bash
$ brew install gabrie30/utils/ghorg
$ mkdir -p $HOME/.config/ghorg # (optional but recommended)
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml # (optional but recommended)
$ vi $HOME/.config/ghorg/conf.yaml # (optional but recommended)
```

### Golang

```bash
# ensure $HOME/go/bin is in your path ($ echo $PATH | grep $HOME/go/bin)
$ go get github.com/gabrie30/ghorg
$ mkdir -p $HOME/.config/ghorg # (optional but recommended)
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml # (optional but recommended)
$ vi $HOME/.config/ghorg/conf.yaml # (optional but recommended)
```

## Use

```bash
# note: to view/set all available flags/features see sample-conf.yaml and for more examples see ./examples
$ ghorg clone someorg
$ ghorg clone someorg --concurrency=50 --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
$ ghorg clone someuser --clone-type=user --protocol=ssh --branch=develop --color=off
$ ghorg clone gitlab-group --scm=gitlab --base-url=https://gitlab.internal.yourcompany.com --preserve-dir
$ ghorg clone gitlab-group/gitlab-subgroup --scm=gitlab --base-url=https://gitlab.internal.yourcompany.com
$ ghorg clone --help
# view cloned resources
$ ghorg ls
$ ghorg ls someorg
```

## Setup and Configuration

Configuration for each clone can be set in two ways. The first is in `$HOME/.config/ghorg/conf.yaml`. This file should be created from the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) by copying into `$HOME/.config/ghorg/conf.yaml`. The second method of configuration is setting flags via the cli, run `$ ghorg clone --help` for a list of flags. A flag set on the command line will overwrite any setting in the conf.yaml

### github setup
1. Create [Personal Access Token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) with all `repo` scopes. Update `GHORG_TOKEN` in your `ghorg/conf.yaml`, as a cli flag, or add to your [osx keychain](https://help.github.com/en/github/using-git/caching-your-github-password-in-git).

### gitlab setup

1. Create [Personal Access Token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) with the `read_api` scope (or `api` for self-managed GitLab older than 12.10). This token can be added to your `ghorg/conf.yaml`, as a cli flag, or your [osx keychain](https://help.github.com/en/github/using-git/caching-your-github-password-in-git).
1. Update the `GitLab Specific` config in your `ghorg/conf.yaml` or via cli flags
1. Update `GHORG_SCM` to `gitlab` in your `ghorg/conf.yaml` or via cli flags
1. See [examples/gitlab.md](https://github.com/gabrie30/ghorg/blob/master/examples/gitlab.md) on how to run

### gitea setup

1. Create [Access Token](https://docs.gitea.io/en-us/api-usage/) (Settings -> Applications -> Generate Token)
1. Update `GHORG_TOKEN` in your `ghorg/conf.yaml` or use the (--token, -t) flag.
1. Update `GHORG_SCM` to `gitea` in your `ghorg/conf.yaml` or via cli flags

### bitbucket setup

1. To configure with bitbucket you will need to create a new [app password](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) and update your `$HOME/.config/ghorg/conf.yaml` [here](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml#L37-L47) or use the (--token, -t) and (--bitbucket-username) flags.
1. Update [SCM type](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml#L54-L57) to `bitbucket` in your `ghorg/conf.yaml` or via cli flags

### osx default github/gitlab token used

> NOTE: cloning via https rather than ssh is the ghorg default, this is because a token must be present to retreive the list of repos. However, if you run into trouble cloning via https and genearlly clone via ssh, try switching `--protocol ssh`

```bash
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
$ security find-internet-password -s gitlab.com  | grep "acct" | awk -F\" '{ print $4 }'
```

> It's recommended to store github/gitlab tokens in the osxkeychain, if this command returns anything other than your token see Troubleshooting section below. However, you can always add your token to the $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flags.

## Ignoring Repos
- To ignore any archived repos while cloning use the `--skip-archived` flag (not bitbucket)
- To ignore specific repos create a `ghorgignore` file inside `$HOME/.config/ghorg`. Each line in this file is considered a substring and will be compared against each repos clone url. If the clone url contains a substring in the `ghorgignore` it will be excluded from cloning. To prevent accidentally excluding a repo, you should make each line as specific as possible, eg. `https://github.com/gabrie30/ghorg.git` or `git@github.com:gabrie30/ghorg.git` depending on how you clone.

  ```bash
  # Create ghorgignore
  touch $HOME/.config/ghorg/ghorgignore

  # update file
  vi $HOME/.config/ghorg/ghorgignore
  ```

## Known issues

- When cloning if you see something like `Username for 'https://gitlab.com': ` the command won't finish. I haven't been able to identify the reason for this occuring. The fix for this is to make sure your token is in the osxkeychain. See the troubleshooting section for how to set this up, or try cloning via ssh (--protocol=ssh).
- If you are cloning a large org you may see `Error: open /dev/null: too many open files` which means you need to increase your ulimits, there are lots of docs online for this. For mac the quick and dirty is below

  ```
  # reset the soft and hard file limit boundaries
  $ sudo launchctl limit maxfiles 65536 200000

  # actually now set the ulimit boundary
  $ ulimit -n 20000
  ```

  Another solution is to decrease the number of concurrent clones. Use the `--concurrency` flag to set to lower than 25 (the default)

## Troubleshooting

- If the `security` command does not return your token, follow this [GitHub Documentation](https://help.github.com/en/articles/caching-your-github-password-in-git). For GitHub tokens you will need to set your token as your username and set nothing as the password when prompted. For GitLab you will need to set your token for both the username and password when prompted. This will correctly store your credentials in the keychain. If you are still having problems see this [StackOverflow Post](https://stackoverflow.com/questions/31305945/git-clone-from-github-over-https-with-two-factor-authentication)
- If your GitHub account is behind 2fa follow this [Github Documentation](https://github.blog/2013-09-03-two-factor-authentication/#how-does-it-work-for-command-line-git)
- GitHub Personal Access Token only finding public repos - Give your token all the repo permissions
- Make sure your `$ git --version` is >= 2.19.0

### Updating brew tap
- [See Readme](https://github.com/gabrie30/homebrew-utils/blob/master/README.md)
