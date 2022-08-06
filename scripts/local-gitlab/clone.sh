#!/bin/bash

BASE_URL="http://gitlab.example.com"
TOKEN=${1:-'EVvbmz5qb28ok-rU-zo5'}

ghorg clone all-groups --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --preserve-dir --concurrency=25 --output-dir=local-gitlab-v15-repos

ghorg clone all-groups --scm=gitlab --base-url=${BASE_URL} --token=$TOKEN --concurrency=25 --output-dir=local-gitlab-v15-repos-flat
