#!/bin/bash

set -e

# poll until gitlab has started

GITLAB_URL=$1
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

# Once running get pw with
pw=$(docker exec -it gitlab grep 'Password:' /etc/gitlab/initial_root_password | awk '{print $2}')

echo "grant_type=password&username=root&password=${pw}" > auth.txt

BEARER_TOKEN_JSON=$(curl -s --data "@auth.txt" --request POST "${GITLAB_URL}/oauth/token")

echo "${BEARER_TOKEN_JSON}"

BEARER_TOKEN=$(echo "${BEARER_TOKEN_JSON}" | jq -r '.access_token' | tr -d '\n')

rm auth.txt

TOKEN_NUMS=$(echo "${RANDOM}")

API_TOKEN=$(curl -s --request POST --header "Authorization: Bearer ${BEARER_TOKEN}" --data "name=admintoken-${TOKEN_NUMS}" --data "expires_at=2050-04-04" --data "scopes[]=api" "${GITLAB_URL}/api/v4/users/1/personal_access_tokens" | jq -r '.token' | tr -d '\n')

# seed new instance using
./scripts/local-gitlab/seed.sh "${API_TOKEN}" "${GITLAB_URL}"
