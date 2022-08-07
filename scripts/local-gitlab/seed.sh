#! /bin/bash

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=${1:-'yYPQd9zVy3hvMqsuK13-'}
BASE_URL="http://gitlab.example.com"

# Create 3 groups, namespace_id will start at 4
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group1", "name": "group1" }' \
    "${BASE_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group2", "name": "group2" }' \
    "${BASE_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group3", "name": "group3" }' \
    "${BASE_URL}/api/v4/groups"

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "group4", "name": "group4" }' \
    "${BASE_URL}/api/v4/groups"

# create repos for user
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${BASE_URL}/api/v4/projects?name=baz${a}&initialize_with_readme=true"
done

# create repos in group1
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${BASE_URL}/api/v4/projects?name=baz${a}&namespace_id=4&initialize_with_readme=true"
done

# create repos in group2
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${BASE_URL}/api/v4/projects?name=baz${a}&namespace_id=5&initialize_with_readme=true"
done

# create repos in group3
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${BASE_URL}/api/v4/projects?name=baz${a}&namespace_id=6&initialize_with_readme=true"
done

# create repos in group3
for ((a=0; a <= 10 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${BASE_URL}/api/v4/projects?name=baz${a}&namespace_id=7&initialize_with_readme=true&wiki_enabled=true"
done
