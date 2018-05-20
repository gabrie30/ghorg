# ghorg

GitHub search is terrible. The idea with ghorg is to quickly clone all org repos into a single directory and use something like ack to search.

> NOTE: When running ghorg a second time, all local changes in your ghorg directory will be overwritten by whats on GitHub. If you are working out of this directory, make sure you rename it before running a second time.

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

```
$ ghorg org
```

## Default GitHub Token Used

- $ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'


## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
