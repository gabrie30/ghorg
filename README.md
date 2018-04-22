# ghorg

## Purpose

Github search is terrible. The idea here is quickly clone all org repos into a single directory and use something like ack for searching the org. Can also be used for setting up kuve <https://github.com/wrsinc/kuve>

## Use

- $ ghorg nameoforg

## Setup

- $ git clone
- $ cd ghorg
- $ cp .env-sample .env
- update your .env
- $ make install
- $ go install

## Auth
- If org is behind SSO add it to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
