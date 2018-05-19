# ghorg

Github search is terrible. The idea with ghorg is to quickly clone all org repos into a single directory and use something like ack to search.

## Use

> NOTE: When ran ghorg will overwrite any local changes. If you are using ghorg to create a directory to work out of, make sure you rename the directory before running a second time.

```
$ ghorg org
```



## Setup

1.  $ go get -u github.com/gabrie30/ghorg
1.  $ cd $HOME/go/src/github.com/gabrie30/ghorg
1. $ cp .env-sample .env
1. update your .env
1. $ make install
1. $ go install

## Get Existing GitHub Token

- $ security find-internet-password -s github.com  | grep "acct" | awk -F\" '{ print $4 }'


## Auth through SSO

- If org is behind SSO a normal token will not work. You will need to add SSO to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
