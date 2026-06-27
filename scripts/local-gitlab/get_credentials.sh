#!/bin/bash

set -xv

# poll until gitlab has started

GITLAB_URL=$1
LOCAL_GITLAB_GHORG_DIR=$2
started="0"
counter=0
# ~30 minutes at 10s per attempt; override with GITLAB_READY_MAX_ATTEMPTS if needed
max_attempts=${GITLAB_READY_MAX_ATTEMPTS:-180}

dump_diagnostics() {
  echo "=== Diagnostics: GitLab did not become ready ==="
  echo "--- docker ps -a ---"
  docker ps -a || true
  echo "--- last response headers from ${GITLAB_URL}/user/sign_in ---"
  curl -I -sS -L "${GITLAB_URL}/user/sign_in" || true
  echo "--- docker logs gitlab (last 200 lines) ---"
  docker logs --tail 200 gitlab || true
}

until [ "$started" -eq "1" ]
do
  if [ "$counter" -gt "$max_attempts" ]; then
    echo "GitLab isn't starting properly, exiting"
    dump_diagnostics
    exit 1
  fi

  # Use the returned HTTP status code directly so the readiness check is agnostic
  # to the HTTP version in the status line (HTTP/1.1 vs HTTP/2) and to header
  # formatting differences across curl versions.
  resp=$(curl -s -o /dev/null -L -w "%{http_code}" "${GITLAB_URL}/user/sign_in" || true)

  if [ "${resp}" = "200" ]; then
    started="1"
    echo "GitLab is fully up and running..."
    break
  fi

  # Periodically surface the current status code and recent container logs so a
  # hung or crashing boot is visible in CI output instead of a silent wait.
  if [ $((counter % 12)) -eq 0 ]; then
    echo "Latest HTTP status from ${GITLAB_URL}/user/sign_in: '${resp:-no response}'"
    docker logs --tail 5 gitlab 2>&1 || true
  fi

  sleep 10
  ((counter++)) || true
  echo "GitLab has not started, sleeping...count: ${counter} (status: ${resp:-none})"
done

set -x

docker exec gitlab gitlab-rails runner "token = User.find_by_username('root').personal_access_tokens.create(scopes: [:api, :read_api, :sudo], name: 'CI Test Token', expires_at: 365.days.from_now); token.set_token('password'); token.save!"
