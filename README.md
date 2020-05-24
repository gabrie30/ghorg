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
- Bitbucket (see bitbucket setup)

> The terminology used in ghorg is that of GitHub, mainly orgs/repos. GitLab and BitBucket use different terminology. There is a handy chart thanks to GitLab that translates terminology [here](https://about.gitlab.com/images/blogimages/gitlab-terminology.png)

## Setup

### Homebrew

> optional

```bash
$ brew update
$ brew upgrade git
```
> required

```bash
$ brew install gabrie30/utils/ghorg
$ mkdir -p $HOME/.config/ghorg
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
$ vi $HOME/.config/ghorg/conf.yaml # (optional but recommended)
```

### Golang

```bash
# ensure $HOME/go/bin is in your path ($ echo $PATH | grep $HOME/go/bin)
$ go get -u github.com/gabrie30/ghorg
$ mkdir -p $HOME/.config/ghorg
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/.config/ghorg/conf.yaml
$ vi $HOME/.config/ghorg/conf.yaml # (optional but recommended)
```

## Use

```bash
# note: to view all available flags/features see sample-conf.yaml
$ ghorg clone someorg
$ ghorg clone someorg --concurrency=50
$ ghorg clone someuser --clone-type=user --protocol=ssh --branch=develop --color=off
$ ghorg clone gitlab-org --scm=gitlab --namespace=gitlab-org/security-products
$ ghorg clone gitlab-org --base-url=https://gitlab.internal.yourcompany.com --preserve-dir
$ ghorg clone --help
```

## Configuration

Configuration can be set in two ways. The first is in `$HOME/.config/ghorg/conf.yaml`. This file will be created from the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) and copied into `$HOME/.config/ghorg/conf.yaml`. The second is setting flags via the cli, run `$ ghorg clone --help` for a list of flags. Any flag set on the command line will overwrite anything in the conf.yaml

## Default GitHub/GitLab Token Used

> NOTE: cloning via https rather than ssh is the ghorg default, this is because a token must be present to retreive the list of repos. However, if you run into trouble cloning via https and genearlly clone via ssh, try switching `--protocol ssh`

```bash
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
$ security find-internet-password -s gitlab.com  | grep "acct" | awk -F\" '{ print $4 }'
```

> It's recommended to store github/gitlab tokens in the osxkeychain, if this command returns anything other than your token see Troubleshooting section below. However, you can always add your token to the $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flags.


## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)

## Bitbucket Setup

To configure with bitbucket you will need to create a new [app password](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) and update your `$HOME/ghorg/conf.yaml` or use the (--token, -t) and (--bitbucket-username) flags.

## Ignoring Repos
- To ignore any archived repos while cloning use the `--skip-archived` flag (github/gitlab only)
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
