#!/bin/bash

set -x

# poll until gitlab has started

GITLAB_URL=$1
LOCAL_GITLAB_GHORG_DIR=$2
started="0"
counter=0

until [ $started -eq "1" ]
do
  resp=$(curl -I -s -L "${GITLAB_URL}"/user/sign_in | grep "HTTP/1.1 200 OK" | cut -d$' ' -f2)

  if [ $counter -gt 100 ]; then
    echo "GitLab isn't starting properly, exiting"
    exit 1
  fi

  if [ "${resp}" = "200" ]; then
    started="1"
    echo "GitLab is fully up and running..."
  fi
  sleep 10
  ((counter++))
  echo "GitLab has not started, sleeping...count: ${counter}"
done

set -x

docker exec -it gitlab gitlab-rails runner "token = User.find_by_username('root').personal_access_tokens.create(scopes: [:api, :read_api, :sudo], name: 'CI Test Token'); token.set_token('password'); token.save!"

API_TOKEN="password"

# seed new instance using
./scripts/local-gitlab/seed.sh "${API_TOKEN}" "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
