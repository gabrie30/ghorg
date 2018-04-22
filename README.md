# ghorg

### Purpose
Github search is terrible. The idea here is quickly clone all org repos into a single directory and use something like ack for searching the org. Can also be used for setting up kuve <https://github.com/wrsinc/kuve>

### Use
`$ ghorg nameoforg`

### Setup
1. $ git clone
1. $ cd ghorg
1. $cp .env-sample .env
2. update your .env
2. $ make install
3. $ go install

### Auth
- If org is behind SSO add it to the [Github token](https://help.github.com/articles/authorizing-a-personal-access-token-for-use-with-a-saml-single-sign-on-organization/)
