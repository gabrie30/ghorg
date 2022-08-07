#!/bin/bash

BASE_URL="http://gitlab.example.com"
TOKEN=${1:-'yYPQd9zVy3hvMqsuK13-'}

ghorg clone all-groups --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --preserve-dir --output-dir=local-gitlab-v15-repos

ghorg clone all-groups --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --output-dir=local-gitlab-v15-repos-flat

ghorg clone root --clone-type=user --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --output-dir=local-gitlab-v15-root-user-repos --prune --prune-no-confirm

ghorg clone all-groups --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --backup --clone-wiki --output-dir=local-gitlab-v15-backup

ghorg clone group1 --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --backup --output-dir=local-gitlab-v15-group1-backup

ghorg clone group1 --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --output-dir=local-gitlab-v15-group1
