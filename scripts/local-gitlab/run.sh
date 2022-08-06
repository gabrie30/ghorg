#! /bin/bash

# Note: you will need to stop manually
# docker stop gitlab
# docker rm gitlab
# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

# Tags at https://hub.docker.com/r/gitlab/gitlab-ce/tags for CE, EE is only latest

# make sure 127.0.0.1 gitlab.example.com is added to your /etc/hosts

export GITLAB_IMAGE_TAG=latest
export GITLAB_HOME=$HOME/Desktop/ghorg/local-gitlab-ee-data-$GITLAB_IMAGE_TAG

docker run \
  --hostname gitlab.example.com \
  --publish 443:443 --publish 80:80 --publish 22:22 \
  --name gitlab \
  --restart always \
  --volume $GITLAB_HOME/config:/etc/gitlab \
  --volume $GITLAB_HOME/logs:/var/log/gitlab \
  --volume $GITLAB_HOME/data:/var/opt/gitlab \
  gitlab/gitlab-ee:$GITLAB_IMAGE_TAG

# Once running get pw with
docker exec -it gitlab grep 'Password:' /etc/gitlab/initial_root_password

# Login at

# http://gitlab.example.com/users/sign_in
# username: root

# Create an api token
# http://gitlab.example.com/-/profile/personal_access_tokens

# seed new instance using
# ./scripts/local-gitlab/seed.sh $TOKEN
