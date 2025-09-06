#!/bin/bash

set -euo pipefail

# Wait for Gitea to be ready and set up admin user
# Usage: ./get_credentials.sh <GITEA_URL> <LOCAL_GITEA_GHORG_DIR>

GITEA_URL=${1:-"http://gitea.example.com:3000"}
LOCAL_GITEA_GHORG_DIR=${2:-"${HOME}/ghorg"}

echo "Waiting for Gitea to be ready at ${GITEA_URL}..."

# Wait for Gitea to be accessible
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -sf "${GITEA_URL}" > /dev/null 2>&1; then
        echo "Gitea is responding!"
        break
    fi
    echo "Attempt $((attempt + 1))/$max_attempts: Gitea not ready yet, waiting..."
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "Gitea failed to start within the expected time"
    exit 1
fi

# Wait a bit more for Gitea to fully initialize
echo "Waiting for Gitea to fully initialize..."
sleep 15

# Create the ghorg directory if it doesn't exist
mkdir -p "${LOCAL_GITEA_GHORG_DIR}"

echo "Setting up Gitea with manual database initialization..."

# Initialize the database and create admin user using the CLI
echo "Creating database and admin user via Docker exec..."
docker exec --user git gitea bash -c "
cd /data/gitea && \
/usr/local/bin/gitea migrate && \
/usr/local/bin/gitea admin user create --admin --username testuser --password testpass --email test@example.com --must-change-password=false
" || {
    echo "Admin user creation may have failed, but user might already exist"
    # Check if we can still proceed
}

# Wait a moment for everything to settle
sleep 10

# Check if the API is now available
echo "Checking API availability..."
if curl -sf "${GITEA_URL}/api/v1/version" > /dev/null 2>&1; then
    echo "API is available! Attempting to create token..."

    # Try to create an API token
    API_TOKEN_RESPONSE=$(curl -X POST "${GITEA_URL}/api/v1/users/testuser/tokens" \
      -H "Content-Type: application/json" \
      -u "testuser:testpass" \
      -d '{"name": "test-token"}' 2>/dev/null || echo '{"sha1":""}')

    API_TOKEN=$(echo "$API_TOKEN_RESPONSE" | grep -o '"sha1":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "")

    if [ -z "$API_TOKEN" ]; then
        echo "Failed to create real API token, using dummy token"
        API_TOKEN="test-token"
    else
        echo "Successfully created API token!"
    fi
else
    echo "API still not available, using basic auth approach"
    API_TOKEN="test-token"
fi

echo "API Token: ${API_TOKEN}"

# Save credentials to ghorg directory for other scripts to use
echo "testuser" > "${LOCAL_GITEA_GHORG_DIR}/gitea_username"
echo "testpass" > "${LOCAL_GITEA_GHORG_DIR}/gitea_password"
echo "${API_TOKEN}" > "${LOCAL_GITEA_GHORG_DIR}/gitea_token"

echo "Gitea setup complete!"
echo "Admin Username: testuser"
echo "Admin Password: testpass"
echo "API Token: ${API_TOKEN}"
