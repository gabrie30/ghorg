#! /bin/bash

set -xv

# https://docs.gitlab.com/ee/install/docker.html#install-gitlab-using-docker-engine

TOKEN=$1
GITLAB_URL=$2
LOCAL_GITLAB_GHORG_DIR=$3

# Create 3 groups, namespace_id will start at 2 (same thing as Group ID you can find in the UI)
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group1", "name": "local-gitlab-group1" }' \
    "${GITLAB_URL}/api/v4/groups"

echo ""
echo ""
echo ""
sleep 1

GROUP1_NAMESPACE_ID=$(curl --request GET --header "PRIVATE-TOKEN: $TOKEN" \
    "${GITLAB_URL}/api/v4/namespaces/local-gitlab-group1" | jq '.id')


echo ""
echo ""
echo ""
sleep 1

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group2", "name": "local-gitlab-group2" }' \
    "${GITLAB_URL}/api/v4/groups"

echo ""
echo ""
echo ""
sleep 1

GROUP2_NAMESPACE_ID=$(curl --request GET --header "PRIVATE-TOKEN: $TOKEN" \
    "${GITLAB_URL}/api/v4/namespaces/local-gitlab-group2" | jq '.id')


echo ""
echo ""
echo ""
sleep 1

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "local-gitlab-group3", "name": "local-gitlab-group3" }' \
    "${GITLAB_URL}/api/v4/groups"

echo ""
echo ""
echo ""
sleep 1

GROUP3_NAMESPACE_ID=$(curl --request GET --header "PRIVATE-TOKEN: $TOKEN" \
    "${GITLAB_URL}/api/v4/namespaces/local-gitlab-group3" | jq '.id')

echo ""
echo ""
echo ""
sleep 1

# group3/subgroup-a
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "subgroup-a", "name": "subgroup-a" }' \
    "${GITLAB_URL}/api/v4/groups?parent_id=${GROUP3_NAMESPACE_ID}"

echo ""
echo ""
echo ""
sleep 1

GROUP3_SUBGROUPA_NAMESPACE_ID=$(curl --request GET --header "PRIVATE-TOKEN: $TOKEN" \
    "${GITLAB_URL}/api/v4/namespaces/local-gitlab-group3%2Fsubgroup-a" | jq '.id')

echo ""
echo ""
echo ""
sleep 1

# group3/subgroup-a/subgroup-b
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"path": "subgroup-b", "name": "subgroup-b" }' \
    "${GITLAB_URL}/api/v4/groups?parent_id=${GROUP3_SUBGROUPA_NAMESPACE_ID}"

echo ""
echo ""
echo ""
sleep 2

GROUP3_SUBGROUPA_SUBGROUPB_NAMESPACE_ID=$(curl --request GET --header "PRIVATE-TOKEN: $TOKEN" \
    "${GITLAB_URL}/api/v4/namespaces/local-gitlab-group3%2Fsubgroup-a%2Fsubgroup-b" | jq '.id')

echo ""
echo ""
echo ""
sleep 1

# Create 2 users
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser1@example.com", "password": "adminadmin1","name": "testuser1","username": "testuser1"}' \
    "${GITLAB_URL}/api/v4/users"

echo ""
echo ""
echo ""
sleep 1

curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data '{"email": "testuser2@example.com", "password": "adminadmin1","name": "testuser2","username": "testuser2"}' \
    "${GITLAB_URL}/api/v4/users"

echo ""
echo ""
echo ""
sleep 1

# create repos for root user
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=rootrepos${a}&initialize_with_readme=true"
done

echo ""
echo ""
echo ""
sleep 1

# create a repo for testuser1, this user has an id of 2
curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects/user/2?name=testuser1-repo&initialize_with_readme=true"

echo -e "\n\n\n"
sleep 1

# create a snippet for testuser1's repo
SNIPPET_DATA='{"title": "my-first-snippet", "file_name": "snippet.txt", "content": "This is my first snippet", "visibility": "public"}'
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
    --header "Content-Type: application/json" \
    --data "${SNIPPET_DATA}" \
    "${GITLAB_URL}/api/v4/projects/testuser1%2Ftestuser1-repo/snippets"

echo -e "\n\n\n"
sleep 1

# create repos in group1
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=${GROUP1_NAMESPACE_ID}&initialize_with_readme=true"
done

echo ""
echo ""
echo ""
sleep 1

# create snippets at the root level
for ((a=1; a <= 2 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/snippets?title=snippet${a}&file_name=file${a}&content=content${a}&description=description${a}&visibility=public"
done

echo ""
echo ""
echo ""
sleep 1

# create repos and snippets in group2
for ((a=0; a <= 3 ; a++))
do
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=baz${a}&namespace_id=${GROUP2_NAMESPACE_ID}&initialize_with_readme=true"
    sleep 1
    # Create non-empty snippet for the repo
    curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
        --header "Content-Type: application/json" \
        --data '{"title": "Snippet for subgroup_a_repo_'${a}'", "file_name": "snippet.txt", "content": "This is a snippet for subgroup_a_repo_'${a}'", "visibility": "public"}' \
        "${GITLAB_URL}/api/v4/projects/local-gitlab-group2%2Fbaz${a}/snippets"
done

echo ""
echo ""
echo ""
sleep 1

# create repos and snippets in group3/subgroup-a
for ((a=0; a <= 3 ; a++))
do
    # Create repo
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=subgroup_a_repo_${a}&namespace_id=${GROUP3_SUBGROUPA_NAMESPACE_ID}&initialize_with_readme=true"
    echo ""
    sleep 1
    # Create non-empty snippet for the repo
    curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
        --header "Content-Type: application/json" \
        --data '{"title": "Snippet for subgroup_a_repo_'${a}'", "file_name": "snippet.txt", "content": "This is a snippet for subgroup_a_repo_'${a}'", "visibility": "public"}' \
        "${GITLAB_URL}/api/v4/projects/local-gitlab-group3%2Fsubgroup-a%2Fsubgroup_a_repo_${a}/snippets"
done

echo ""
echo ""
echo ""
sleep 1

# create repos and snippets in group3/subgroup-a/subgroup-b
for ((a=0; a <= 3 ; a++))
do
    # Create repo
    curl --header "PRIVATE-TOKEN: $TOKEN" -X POST "${GITLAB_URL}/api/v4/projects?name=subgroup_b_repo_${a}&namespace_id=${GROUP3_SUBGROUPA_SUBGROUPB_NAMESPACE_ID}&initialize_with_readme=true"
    echo ""
    sleep 1
    # Create non-empty snippet for the repo
    curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
        --header "Content-Type: application/json" \
        --data '{"title": "Snippet for subgroup_b_repo_'${a}'", "file_name": "snippet.txt", "content": "This is a snippet for subgroup_b_repo_'${a}'", "visibility": "public"}' \
        "${GITLAB_URL}/api/v4/projects/local-gitlab-group3%2Fsubgroup-a%2Fsubgroup-b%2Fsubgroup_b_repo_${a}/snippets"
done

echo ""
echo ""
echo ""
sleep 1

echo "sleeping before running integration tests, to ensure all resources are created"
sleep 5

./scripts/local-gitlab/integration-tests.sh "${LOCAL_GITLAB_GHORG_DIR}" "${TOKEN}" "${GITLAB_URL}"
