# ghorg

GitHub search is terrible. The idea with ghorg is to quickly clone all org repos into a single directory and use something like ack to search.

> When running ghorg a second time, all local changes in your *_ghorg directory will be overwritten by whats on GitHub. If you are working out of this directory, make sure you rename it before running a second time.

## Setup

```bash
$ go get -u github.com/gabrie30/ghorg
$ cd $HOME/go/src/github.com/gabrie30/ghorg
$ cp .env-sample .env
# update your .env
# If GITHUB_TOKEN is not set in .ghorg, defaults to keychain, see below
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

> If running this does not return the correct key you will need to generate a token via GithHub and add it to your .env

## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
