#! /bin/bash

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=$1
GITLAB_URL=$2

# Create 3 groups, namespace_id will start at 4 (same thing as Group ID you can find in the UI)
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
    --data '{"path": "subgroup-a", "name": "subgroup-a" }' \
    "${GITLAB_URL}/api/v4/groups?parent_id=18"

# Create 2 users
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser1@example.com", "password": "adminadmin1","name": "testuser1","reset_password": "true" }'

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser2@example.com", "password": "adminadmin1","name": "testuser2","reset_password": "true" }'

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

# create repos in group3/subgroup-a
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=subgroup_repo_${a}&namespace_id=20&initialize_with_readme=true"
done

./scripts/local-gitlab/clone.sh "${TOKEN}" "${GITLAB_URL}"
