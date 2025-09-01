#!/bin/bash

set -e

BITBUCKET_URL="${1:-http://bitbucket.example.com:7990}"
OUTPUT_DIR="${2:-/tmp/test-ghorg-output}"

# Default admin credentials (set in bitbucket.properties)
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin"
ADMIN_EMAIL="admin@bitbucket.local"

echo "Waiting for Bitbucket Server to be ready at ${BITBUCKET_URL}..."

# Function to check if Bitbucket is responding
check_bitbucket_ready() {
    curl -s --max-time 10 "${BITBUCKET_URL}/status" > /dev/null 2>&1
}

# Wait for Bitbucket to be ready (up to 5 minutes)
timeout=300
elapsed=0
while ! check_bitbucket_ready && [ $elapsed -lt $timeout ]; do
    echo "Checking if Bitbucket is ready... (${elapsed}s/${timeout}s)"
    sleep 10
    elapsed=$((elapsed + 10))
done

if ! check_bitbucket_ready; then
    echo "❌ Bitbucket Server did not become ready within ${timeout} seconds"
    echo "Check container logs: docker logs bitbucket"
    exit 1
fi

echo "Bitbucket Server is responding!"

# Give additional time for complete initialization with pre-configured database and admin user
echo "Waiting for Bitbucket Server to fully initialize with H2 database..."
echo "Background setup automation script is also running..."
echo "This may take up to 3 minutes for complete automated setup..."
# Reduced initialization wait
sleep 30

# Check if setup is complete by testing the API
echo "Checking Bitbucket Server setup status..."

# Test if we can access the API without authentication (should fail with 401)
api_response=$(curl -s -w "%{http_code}" "${BITBUCKET_URL}/rest/api/1.0/repos" -o /dev/null)

if [ "$api_response" = "401" ]; then
    echo "✅ Bitbucket Server API is responding correctly (401 Unauthorized as expected)"
    setup_complete=true
elif [ "$api_response" = "200" ]; then
    echo "⚠️  Bitbucket Server API is accessible without authentication - this may indicate setup issues"
    setup_complete=true
else
    echo "⚠️  Bitbucket Server API returned status: $api_response"

    # Check if we're still seeing setup pages
    if curl -s "${BITBUCKET_URL}" | grep -qi "setup\|database\|welcome.*bitbucket"; then
        echo "🔧 Bitbucket Server is still showing setup pages"
        setup_complete=false
    else
        echo "✅ Bitbucket Server appears to be ready"
        setup_complete=true
    fi
fi

if [ "$setup_complete" = false ]; then
    if [ "${ENV:-}" = "ci" ]; then
        echo "🤖 CI Mode: Setup not complete but proceeding with authentication test"
        echo "This validates that ghorg's Bitbucket Server client implementation is working"
    else
        echo ""
        echo "💡 MANUAL SETUP REQUIRED:"
        echo "1. Open ${BITBUCKET_URL} in your browser"
        echo "2. Complete the setup wizard:"
        echo "   - Choose 'Internal' for database (H2)"
        echo "   - Create admin user with credentials:"
        echo "     Username: admin"
        echo "     Password: admin"
        echo "     Email: admin@bitbucket.local"
        echo "3. Skip license (evaluation mode)"
        echo "4. Complete setup and re-run tests"
        echo ""
        exit 1
    fi
fi

# Test authentication with admin credentials
echo "Testing authentication with admin credentials..."

auth_test=$(curl -s -w "%{http_code}" -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" \
    "${BITBUCKET_URL}/rest/api/1.0/users/${ADMIN_USERNAME}" -o /tmp/auth_test.json)

if [ "$auth_test" = "200" ]; then
    echo "✅ Authentication successful with admin:admin"

    # Validate the response contains user information
    if grep -q "\"name\".*\"${ADMIN_USERNAME}\"" /tmp/auth_test.json; then
        echo "✅ User information retrieved successfully"

        # Output credentials for other scripts
        mkdir -p "${OUTPUT_DIR}"
        cat > "${OUTPUT_DIR}/bitbucket_credentials.env" << EOF
export GHORG_SCM_TYPE=bitbucket
export GHORG_SCM_BASE_URL=${BITBUCKET_URL}
export GHORG_BITBUCKET_USERNAME=${ADMIN_USERNAME}
export GHORG_BITBUCKET_APP_PASSWORD=${ADMIN_PASSWORD}
export GHORG_INSECURE_BITBUCKET_CLIENT=true
EOF

        echo "✅ Credentials saved to ${OUTPUT_DIR}/bitbucket_credentials.env"

        # Test a few more API endpoints to ensure full functionality
        echo "Testing additional API endpoints..."

        # Test projects endpoint
        projects_test=$(curl -s -w "%{http_code}" -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" \
            "${BITBUCKET_URL}/rest/api/1.0/projects" -o /dev/null)

        if [ "$projects_test" = "200" ]; then
            echo "✅ Projects API endpoint working"
        else
            echo "⚠️  Projects API returned status: $projects_test"
        fi

        # Test repositories endpoint
        repos_test=$(curl -s -w "%{http_code}" -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" \
            "${BITBUCKET_URL}/rest/api/1.0/repos" -o /dev/null)

        if [ "$repos_test" = "200" ]; then
            echo "✅ Repositories API endpoint working"
        else
            echo "⚠️  Repositories API returned status: $repos_test"
        fi

        echo ""
        echo "🎉 Bitbucket Server is ready for testing!"
        echo "   Base URL: ${BITBUCKET_URL}"
        echo "   Username: ${ADMIN_USERNAME}"
        echo "   Password: ${ADMIN_PASSWORD}"
        echo ""

    else
        echo "⚠️  Authentication succeeded but user data appears invalid"
        cat /tmp/auth_test.json
    fi

elif [ "$auth_test" = "401" ]; then
    echo "❌ Authentication failed with admin:admin"
    echo "API request failed with status 401: Authentication failed"

    # Check if setup is still in progress or if we can try API-based admin creation
    setup_page_content=$(curl -s "${BITBUCKET_URL}" 2>/dev/null)

    if echo "$setup_page_content" | grep -qi "setup\|database\|admin.*user"; then
        echo "⚠️  Bitbucket Server setup may still be in progress"
        echo "Attempting API-based admin user creation as fallback..."

        # Try to create admin user via setup API if available
        admin_creation_response=$(curl -s -w "%{http_code}" \
            -H "Content-Type: application/json" \
            -d '{
                "username": "'${ADMIN_USERNAME}'",
                "password": "'${ADMIN_PASSWORD}'",
                "displayName": "Administrator",
                "emailAddress": "'${ADMIN_EMAIL}'"
            }' \
            "${BITBUCKET_URL}/rest/api/1.0/admin/users" 2>/dev/null)

        if [[ "$admin_creation_response" == *"201" ]] || [[ "$admin_creation_response" == *"409" ]]; then
            echo "✅ Admin user creation via API successful (or user already exists)"
            sleep 30  # Give time for user to be fully set up

            # Retry authentication
            retry_auth_test=$(curl -s -w "%{http_code}" -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" \
                "${BITBUCKET_URL}/rest/api/1.0/users/${ADMIN_USERNAME}" -o /tmp/retry_auth_test.json)

            if [ "$retry_auth_test" = "200" ]; then
                echo "✅ Authentication successful after API user creation"
                if grep -q "\"name\".*\"${ADMIN_USERNAME}\"" /tmp/retry_auth_test.json; then
                    echo "✅ User information retrieved successfully"
                    mkdir -p "${OUTPUT_DIR}"
                    cat > "${OUTPUT_DIR}/bitbucket_credentials.env" << EOF
export GHORG_SCM_TYPE=bitbucket
export GHORG_SCM_BASE_URL=${BITBUCKET_URL}
export GHORG_BITBUCKET_USERNAME=${ADMIN_USERNAME}
export GHORG_BITBUCKET_APP_PASSWORD=${ADMIN_PASSWORD}
export GHORG_INSECURE_BITBUCKET_CLIENT=true
EOF
                    echo "✅ Credentials saved to ${OUTPUT_DIR}/bitbucket_credentials.env"
                    echo "🎉 Fully automated Bitbucket Server setup complete!"
                    exit 0
                fi
            fi
        else
            echo "⚠️  API-based admin user creation failed: $admin_creation_response"
        fi

        echo "Waiting additional 60 seconds for properties-based setup to complete..."
        sleep 60

        # Final retry with original credentials
        final_retry_auth_test=$(curl -s -w "%{http_code}" -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}" \
            "${BITBUCKET_URL}/rest/api/1.0/users/${ADMIN_USERNAME}" -o /tmp/final_retry_auth_test.json)

        if [ "$final_retry_auth_test" = "200" ]; then
            echo "✅ Authentication successful after extended wait"
            if grep -q "\"name\".*\"${ADMIN_USERNAME}\"" /tmp/final_retry_auth_test.json; then
                echo "✅ User information retrieved successfully"
                mkdir -p "${OUTPUT_DIR}"
                cat > "${OUTPUT_DIR}/bitbucket_credentials.env" << EOF
export GHORG_SCM_TYPE=bitbucket
export GHORG_SCM_BASE_URL=${BITBUCKET_URL}
export GHORG_BITBUCKET_USERNAME=${ADMIN_USERNAME}
export GHORG_BITBUCKET_APP_PASSWORD=${ADMIN_PASSWORD}
export GHORG_INSECURE_BITBUCKET_CLIENT=true
EOF
                echo "✅ Credentials saved to ${OUTPUT_DIR}/bitbucket_credentials.env"
                echo "🎉 Fully automated Bitbucket Server setup complete!"
                exit 0
            fi
        fi
    fi

    echo ""
    echo "❌ Unattended setup failed. Possible issues:"
    echo "1. Check Docker logs: docker logs bitbucket"
    echo "2. Verify bitbucket.properties configuration"
    echo "3. Ensure sufficient startup time for initialization"
    echo ""
    exit 1

else
    echo "❌ Authentication test failed with status: $auth_test"
    exit 1
fi

# Clean up
rm -f /tmp/auth_test.json /tmp/bb_cookies

echo "✅ Bitbucket Server setup and authentication verification complete!"
