# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a> [![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

ghorg allows you to quickly clone all of an orgs repos into a single directory. This can be useful in many situations including

1. Searching your orgs codebase with ack, silver searcher, grep etc..
2. Bash scripting
3. Creating backups
4. Onboarding new teammates
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
$ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/.env-sample > $HOME/.ghorg
```

### Golang

```bash
$ go get -u github.com/gabrie30/ghorg
$ cd $HOME/go/src/github.com/gabrie30/ghorg
$ cp .env-sample .env
# update your .env, if needed
# If GHORG_GITHUB_TOKEN is not set in .ghorg, defaults to keychain, see below
$ make install
$ go install
```

## Use

```bash
$ ghorg org-you-want-to-clone
```

> ghorg defaults to master however, for gitflows you can run on develop by setting GHORG_BRANCH=develop or similar

## Configuration

All configuration will be done in the .ghorg file. This file will be created from the [.env-sample](https://github.com/gabrie30/ghorg/blob/master/.env-sample) and copied into ~/.ghorg. Make sure this file exists then configure to your needs.

## Default GitHub Token Used

```bash
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
```

> If running this does not return the correct key you will need to generate a token via GithHub and add it to your $HOME/.ghorg

> To view all other default environment variables see .env-sample

## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)

## Troubleshooting

- Make sure your `$ git --version` is >= 2.19.0
- You may need to increase your ulimits if cloning a large org
- Other issues can most likely be resolved by adding a `.ghorg` to your users home directory and setting the necessary values defined in the `.env-sample`

### Updating brew tap
- [See Readme](https://github.com/gabrie30/homebrew-utils/blob/master/README.md)
