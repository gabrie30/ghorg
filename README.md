# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a> [![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

ghorg allows you to quickly clone all of an orgs, or users repos into a single directory. This can be useful in many situations including

1. Searching an orgs/users codebase with ack, silver searcher, grep etc..
2. Bash scripting
3. Creating backups
4. Onboarding
5. Performing Audits

> When running ghorg a second time, all local changes in your *_ghorg directory will be overwritten by whats on GitHub. If you are working out of this directory, make sure you rename it before running a second time otherwise all of you changes will be lost.

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
$ mkdir -p $HOME/ghorg
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > $HOME/ghorg/conf.yaml
```

### Golang

```bash
$ go get -u github.com/gabrie30/ghorg
$ cd $HOME/go/src/github.com/gabrie30/ghorg
$ make install
# Update conf if neeed
$ vi ~/ghorg/conf.yaml
$ go install
```

## Use

```bash
$ ghorg clone org-you-want-to-clone
$ ghorg clone user-you-want-to-clone
$ ghorg clone --help
```

> ghorg defaults to master branch however, for gitflows you can run on develop by setting GHORG_BRANCH=develop or similar

## Configuration

Configuration can be set in two ways. The first is in `$HOME/ghorg/conf.yaml`. This file will be created from the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) and copied into `$HOME/ghorg/conf.yaml`. The second is setting flags via the cli, run `$ ghorg clone --help` for a list of flags

## Default GitHub Token Used

```bash
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
```

> If running this does not return the correct key you will need to generate a token via GithHub and add it to your $HOME/ghorg/conf.yaml, or see Troubleshooting section below.

> To view all other default environment variables see sample-conf.yaml

## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)

## Troubleshooting

- Make sure your `$ git --version` is >= 2.19.0
- You may need to increase your ulimits if cloning a large org
- If cloning via HTTPS make sure the osxkeychain has your github access token. This can be determined by running the `security` command above.
    - If this command does not return anything either switch to cloning via ssh (update your `$HOME/ghorg/conf.yaml`) or set it up by following this [GitHub Documentation](https://help.github.com/en/articles/caching-your-github-password-in-git)
    - If your GitHub account is behind 2fa follow this [StackOverflow Post](https://stackoverflow.com/questions/31305945/git-clone-from-github-over-https-with-two-factor-authentication) or this [Github Documentation](https://github.blog/2013-09-03-two-factor-authentication/#how-does-it-work-for-command-line-git) as noted in comments be sure to use your token as your username and give a blank password.

### Updating brew tap
- [See Readme](https://github.com/gabrie30/homebrew-utils/blob/master/README.md)
