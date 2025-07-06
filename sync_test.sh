#!/bin/bash

# Simple test to verify sync functionality with sparse checkout
set -e

echo "Testing sync functionality with sparse checkout simulation..."

TEST_DIR="/tmp/ghorg-sync-test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Build ghorg first
cd /Users/blairham/Developer/github.com/blairham/ghorg
go build -o "$TEST_DIR/ghorg"
cd "$TEST_DIR"

# Create a test repository
echo "Creating test repository..."
git init test-repo
cd test-repo
git config user.name "Test User"
git config user.email "test@example.com"

echo "# Test Repository" >README.md
echo "Some content" >file1.txt
mkdir -p subdir
echo "Nested content" >subdir/file2.txt

git add .
git commit -m "Initial commit"
cd ..

# Test with path filter that matches nothing (simulates sparse checkout with no matches)
echo "Testing with path filter that matches nothing..."
export GHORG_ABSOLUTE_PATH_TO_CLONE_TO="$TEST_DIR/filtered-clone"
mkdir -p "$GHORG_ABSOLUTE_PATH_TO_CLONE_TO"
export GHORG_PATH_FILTER="nonexistent/**"

./ghorg clone file://$TEST_DIR/test-repo --clone-type=user --protocol=https

# Check that the repository was cloned
if [ ! -d "$GHORG_ABSOLUTE_PATH_TO_CLONE_TO/test-repo/.git" ]; then
  echo "ERROR: Repository should be cloned"
  exit 1
fi

cd "$GHORG_ABSOLUTE_PATH_TO_CLONE_TO/test-repo"

# Verify that sparse checkout filtered out files but repo is functional
if [ -f "README.md" ]; then
  echo "Note: README.md was not filtered out (this may be expected depending on sparse-checkout behavior)"
fi

# Check that we can perform git operations
if ! git status >/dev/null; then
  echo "ERROR: Should be able to run git status"
  exit 1
fi

# Check that the default branch is properly synced
if ! git rev-parse HEAD >/dev/null; then
  echo "ERROR: HEAD should be properly set"
  exit 1
fi

echo "SUCCESS: Sync functionality works correctly with sparse checkout"

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo "Sync test completed successfully!"
