#!/bin/bash

set -euo pipefail

# Quick test script to verify Bitbucket Server setup
# Usage: ./test-setup.sh [BITBUCKET_URL]

BITBUCKET_URL=${1:-"http://bitbucket.example.com:7990"}
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin"

echo "=== Testing Bitbucket Server Setup ==="
echo "URL: $BITBUCKET_URL"
echo "Testing with credentials: $ADMIN_USERNAME:$ADMIN_PASSWORD"
echo ""

# Test 1: Check if Bitbucket is responding
echo "1. Testing server response..."
if curl -f -s "$BITBUCKET_URL/status" > /dev/null 2>&1; then
    echo "✅ Bitbucket Server is responding"
else
    echo "❌ Bitbucket Server is not responding"
    echo "   Make sure Docker container is running: docker ps | grep bitbucket"
    exit 1
fi

# Test 2: Check if setup is needed
echo ""
echo "2. Checking setup status..."
if curl -f -s "$BITBUCKET_URL" | grep -q "setup"; then
    echo "⚠️  Setup is required!"
    echo ""
    echo "MANUAL SETUP STEPS:"
    echo "1. Visit: $BITBUCKET_URL"
    echo "2. Choose 'Internal' database (H2)"
    echo "3. Create admin user:"
    echo "   - Username: $ADMIN_USERNAME"
    echo "   - Password: $ADMIN_PASSWORD"
    echo "   - Email: admin@bitbucket.local"
    echo "4. Skip license (evaluation mode)"
    echo "5. Complete setup and re-run this test"
    echo ""
    echo "After setup, run: ./test-setup.sh"
else
    echo "✅ Setup appears complete"
fi

# Test 3: Check authentication
echo ""
echo "3. Testing API authentication..."

auth_success=false
endpoints=("/rest/api/1.0/projects" "/rest/api/1.0/users" "/status")

for endpoint in "${endpoints[@]}"; do
    echo "   Testing: $endpoint"
    response=$(curl -s -w "%{http_code}" -u "$ADMIN_USERNAME:$ADMIN_PASSWORD" "$BITBUCKET_URL$endpoint" 2>/dev/null || echo "000")
    http_code="${response: -3}"

    case "$http_code" in
        200|403)
            echo "   ✅ Auth working ($http_code)"
            auth_success=true
            break
            ;;
        401)
            echo "   ❌ Auth failed (401) - setup incomplete"
            ;;
        *)
            echo "   ⚠️  Unexpected response: $http_code"
            ;;
    esac
done

# Final result
echo ""
if [ "$auth_success" = true ]; then
    echo "🎉 SUCCESS: Bitbucket Server is properly set up!"
    echo ""
    echo "You can now run the integration tests:"
    echo "  ./start.sh"
    echo ""
    echo "Or test ghorg manually:"
    echo "  GHORG_INSECURE_BITBUCKET_CLIENT=true ghorg clone PROJECT_KEY --scm=bitbucket --base-url=$BITBUCKET_URL --bitbucket-username=$ADMIN_USERNAME --token=$ADMIN_PASSWORD --dry-run"
else
    echo "❌ FAILED: Authentication not working"
    echo ""
    echo "Complete the manual setup steps above and re-run this test."
fi
