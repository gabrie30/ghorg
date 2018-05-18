# ghorg

Github search is terrible. The idea here is quickly clone all org repos into a single directory and use something like ack for searching the org. Can also be used for setting up kuve <https://github.com/wrsinc/kuve>

## Use

- $ ghorg nameoforg

## Setup

- $ go get -u github.com/gabrie30/ghorg
- $ cd $HOME/go/src/github.com/gabrie30/ghorg
- $ cp .env-sample .env
- update your .env
- $ make install
- $ go install

## Get Existing GitHub Token

```
$ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'
```

## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
