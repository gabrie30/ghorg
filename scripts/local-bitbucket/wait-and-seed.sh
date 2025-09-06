#!/bin/bash

set -euo pipefail

BITBUCKET_URL="http://bitbucket.example.com:7990"
MAX_WAIT=600  # 10 minutes max wait
ELAPSED=0

echo "⏳ Waiting for Bitbucket Server to be fully ready..."
echo "This may take several minutes with PostgreSQL setup..."

while [ $ELAPSED -lt $MAX_WAIT ]; do
    STATUS=$(curl -s "$BITBUCKET_URL/status" | grep -o '"state":"[^"]*"' | cut -d'"' -f4 || echo "UNKNOWN")

    case $STATUS in
        "RUNNING")
            echo "✅ Bitbucket Server is RUNNING!"
            break
            ;;
        "STARTING")
            echo "⏳ Server starting... ($ELAPSED/${MAX_WAIT}s)"
            ;;
        *)
            echo "⚠️  Server status: $STATUS ($ELAPSED/${MAX_WAIT}s)"
            ;;
    esac

    sleep 10
    ELAPSED=$((ELAPSED + 10))
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo "❌ Server failed to start within $MAX_WAIT seconds"
    exit 1
fi

# Test API access
echo "🔍 Testing API access..."
if curl -s -u admin:admin "$BITBUCKET_URL/rest/api/1.0/projects" > /dev/null; then
    echo "✅ API access confirmed"
else
    echo "❌ API access failed"
    exit 1
fi

# Run the improved seeder
echo "🌱 Running improved seeder with git initialization..."
./seed.sh

echo "🎉 Seeder completed! Testing a clone operation..."

# Test if repositories are now cloneable
if git clone http://admin:admin@bitbucket.example.com:7990/scm/lbp1/baz0.git /tmp/test-fixed-repo; then
    echo "🎉 SUCCESS! Repository cloning now works!"
    rm -rf /tmp/test-fixed-repo
else
    echo "❌ Repository cloning still has issues"
    exit 1
fi
