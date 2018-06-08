# ghorg

[![Go Report Card](https://goreportcard.com/badge/github.com/gabrie30/ghorg)](https://goreportcard.com/report/github.com/gabrie30/ghorg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) <a href="https://godoc.org/github.com/gabrie30/ghorg"><img src="https://godoc.org/github.com/gabrie30/ghorg?status.svg" alt="GoDoc"></a>

GitHub search is terrible. The idea with ghorg is to quickly clone all org repos into a single directory and use something like ack to search.

> When running ghorg a second time, all local changes in your *_ghorg directory will be overwritten by whats on GitHub. If you are working out of this directory, make sure you rename it before running a second time.

## Setup

### Homebrew

```bash
$ brew tap gabrie30/utils
$ brew install ghorg
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
$ ghorg org
```

> ghorg defaults to master however, for gitflows you can run on develop by setting GHORG_BRANCH=develop or similar

## Default GitHub Token Used

```bash
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
```

> If running this does not return the correct key you will need to generate a token via GithHub and add it to your $HOME/.ghorg

> To view all other default environment variables see .env-sample

## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)

## Troubleshooting
- You may need to increase your ulimits if cloning a large org
- Other issues can most likely be resolved by adding a `.ghorg` to your users home directory and setting the necessary values defined in the `.env-sample`
