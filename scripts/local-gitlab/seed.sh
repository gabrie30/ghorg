#! /bin/bash

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=$1
GITLAB_URL=$2

# Create 3 groups, namespace_id will start at 4
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group1", "name": "group1" }' \
    "${GITLAB_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group2", "name": "group2" }' \
    "${GITLAB_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group3", "name": "group3" }' \
    "${GITLAB_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group4", "name": "group4" }' \
    "${GITLAB_URL}/api/v4/groups"

# create repos for user
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&initialize_with_readme=true"
done

# create repos in group1
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=4&initialize_with_readme=true"
done

# create repos in group2
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=5&initialize_with_readme=true"
done

# create repos in group3
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=6&initialize_with_readme=true"
done

# create repos in group3
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=7&initialize_with_readme=true&wiki_enabled=true"
done

./scripts/local-gitlab/clone.sh "${TOKEN}" "${GITLAB_URL}"
