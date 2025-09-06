#!/bin/bash

set -euo pipefail

echo "Setting up local Bitbucket integration test environment..."

# Check if bitbucket.example.com resolves to localhost
if ! grep -q "127.0.0.1[[:space:]]*bitbucket.example.com" /etc/hosts 2>/dev/null; then
    echo "Adding bitbucket.example.com to /etc/hosts..."
    echo "This requires sudo privileges:"

    if sudo bash -c 'echo "127.0.0.1 bitbucket.example.com" >> /etc/hosts'; then
        echo "✅ Added bitbucket.example.com to /etc/hosts"
    else
        echo "❌ Failed to add entry to /etc/hosts"
        echo "Please manually add this line to /etc/hosts:"
        echo "127.0.0.1 bitbucket.example.com"
        exit 1
    fi
else
    echo "✅ bitbucket.example.com already configured in /etc/hosts"
fi

echo ""
echo "🎉 Host configuration complete!"
echo ""
echo "Next steps:"
echo "1. Start the containers: docker-compose up -d"
echo "2. Wait for Bitbucket to initialize (about 3-5 minutes)"
echo "3. Check status: docker logs bitbucket-server-local -f"
echo "4. Test access: curl http://bitbucket.example.com:7990/status"
echo "5. Run integration tests: ./integration-tests.sh"
echo ""
