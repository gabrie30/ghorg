#! /bin/bash

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=$1
GITLAB_URL=$2

# Create 2 groups, namespace_id will start at 4
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group1", "name": "group1" }' \
    "${GITLAB_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group2", "name": "group2" }' \
    "${GITLAB_URL}/api/v4/groups"

sleep 1

# create repos for user
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&initialize_with_readme=true"
done

sleep 1

# create repos in group1
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=4&initialize_with_readme=true"
done

sleep 1

# create repos in group2
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=5&initialize_with_readme=true"
done

./scripts/local-gitlab/clone.sh "${TOKEN}" "${GITLAB_URL}"
