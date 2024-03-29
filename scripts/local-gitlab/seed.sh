#! /bin/bash

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=$1
GITLAB_URL=$2
LOCAL_GITLAB_GHORG_DIR=$3

# Create 3 groups, namespace_id will start at 4 (same thing as Group ID you can find in the UI)
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group1", "name": "local-gitlab-group1" }' \
    "${GITLAB_URL}/api/v4/groups"

sleep 5

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group2", "name": "local-gitlab-group2" }' \
    "${GITLAB_URL}/api/v4/groups"

sleep 5

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group3", "name": "local-gitlab-group3" }' \
    "${GITLAB_URL}/api/v4/groups"

sleep 5

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "subgroup-a", "name": "subgroup-a" }' \
    "${GITLAB_URL}/api/v4/groups?parent_id=6"

sleep 5

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "subgroup-b", "name": "subgroup-b" }' \
    "${GITLAB_URL}/api/v4/groups?parent_id=7"

sleep 5

# Create 2 users
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser1@example.com", "password": "adminadmin1","name": "testuser1","username": "testuser1",reset_password": "true" }' \
    "${GITLAB_URL}/api/v4/users"

sleep 5

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser2@example.com", "password": "adminadmin1","name": "testuser2","username": "testuser2","reset_password": "true" }' \
    "${GITLAB_URL}/api/v4/users"

sleep 5

# create repos for root user
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=rootrepos${a}&initialize_with_readme=true"
done

sleep 5

# create repos in group1
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=4&initialize_with_readme=true"
done

sleep 5

# create repos in group2
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=5&initialize_with_readme=true"
done

sleep 5

# create repos in group3/subgroup-a
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=subgroup_a_repo_${a}&namespace_id=7&initialize_with_readme=true"
done

sleep 5

# create repos in group3/subgroup-a/subgroup-b
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=subgroup_b_repo_${a}&namespace_id=8&initialize_with_readme=true"
done

sleep 5

./scripts/local-gitlab/integration-tests.sh "${TOKEN}" "${GITLAB_URL}" "${LOCAL_GITLAB_GHORG_DIR}"
