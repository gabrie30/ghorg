# Local GitLab

Allows you to spin up GitLab locally with Docker to test cloning. Would eventually like to turn these into integration tests.

For enterprise GitLab, start docker then run `./scripts/local-gitlab/start-ee.sh false` from the root of the repo

TODO: Do the same for the community edition of GitLab

If running locally you'll also need to update your /etc/hosts

`echo "127.0.0.1 gitlab.example.com" >> /etc/hosts`

Once github is running you can vist

http://gitlab.example.com in your browser

You can get the root token by running

```
docker exec -it gitlab grep 'Password:' /etc/gitlab/initial_root_password | awk '{print $2}'
```
