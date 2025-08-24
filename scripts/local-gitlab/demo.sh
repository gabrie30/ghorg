#!/bin/bash

set -euo pipefail

# Demo script showing the refactored GitLab integration system
# This doesn't start GitLab, just demonstrates the components

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== GitLab Integration Test System Demo ==="
echo ""
echo "This demo shows the refactored components without starting GitLab."
echo ""

echo "📁 Directory Structure:"
echo "configs/"
echo "├── seed-data.json          # GitLab resources to create"
echo "└── test-scenarios.json     # Integration test scenarios"
echo ""
echo "seeder/"
echo "├── main.go                 # Go-based seeder"
echo "└── go.mod"
echo ""
echo "test-runner/"
echo "├── main.go                 # Go-based test runner"
echo "└── go.mod"
echo ""

echo "📋 Available seed data configuration:"
if [[ -f "${SCRIPT_DIR}/configs/seed-data.json" ]]; then
    echo "Groups:"
    python3 -c "
import json
with open('${SCRIPT_DIR}/configs/seed-data.json', 'r') as f:
    data = json.load(f)
for group in data['groups']:
    print(f'  - {group[\"name\"]}: {len(group.get(\"repositories\", []))} repositories')
    if 'subgroups' in group:
        for subgroup in group['subgroups']:
            print(f'    └── {subgroup[\"name\"]}: {len(subgroup.get(\"repositories\", []))} repositories')
            if 'subgroups' in subgroup:
                for nested in subgroup['subgroups']:
                    print(f'        └── {nested[\"name\"]}: {len(nested.get(\"repositories\", []))} repositories')
print(f'Users: {len(data[\"users\"])}')
print(f'Root user repositories: {len(data[\"root_user\"][\"repositories\"])}')
print(f'Root snippets: {len(data[\"root_snippets\"])}')
"
else
    echo "❌ No seed data configuration found"
fi

echo ""
echo "🧪 Available test scenarios:"
if [[ -f "${SCRIPT_DIR}/configs/test-scenarios.json" ]]; then
    python3 -c "
import json
with open('${SCRIPT_DIR}/configs/test-scenarios.json', 'r') as f:
    data = json.load(f)
for i, scenario in enumerate(data['test_scenarios'], 1):
    print(f'  {i}. {scenario[\"name\"]}')
    print(f'     {scenario[\"description\"]}')
    print(f'     Expected results: {len(scenario[\"expected_structure\"])} paths')
    if scenario.get('run_twice'):
        print('     🔄 Runs twice (clone + pull)')
    print()
"
else
    echo "❌ No test scenarios configuration found"
fi

echo "🔧 Building components..."

# Build seeder
if [[ -f "${SCRIPT_DIR}/seeder/main.go" ]]; then
    echo "Building seeder..."
    cd "${SCRIPT_DIR}/seeder"
    if go build -o gitlab-seeder main.go 2>/dev/null; then
        echo "✅ Seeder built successfully"
        echo "   Usage: ./gitlab-seeder -token=TOKEN -base-url=URL -config=CONFIG"
    else
        echo "❌ Failed to build seeder (Go dependencies may need to be installed)"
    fi
    cd - > /dev/null
else
    echo "❌ Seeder source not found"
fi

# Build test runner
if [[ -f "${SCRIPT_DIR}/test-runner/main.go" ]]; then
    echo "Building test runner..."
    cd "${SCRIPT_DIR}/test-runner"
    if go build -o gitlab-test-runner main.go 2>/dev/null; then
        echo "✅ Test runner built successfully"
        echo "   Usage: ./gitlab-test-runner -token=TOKEN -base-url=URL -ghorg-dir=DIR"
        echo "   List tests: ./gitlab-test-runner -list -token=TOKEN"
        echo "   Run specific test: ./gitlab-test-runner -test=NAME -token=TOKEN"
    else
        echo "❌ Failed to build test runner"
    fi
    cd - > /dev/null
else
    echo "❌ Test runner source not found"
fi

echo ""
echo "🚀 Ready to use!"
echo ""
echo "To run the full integration test:"
echo "  ./start-ee.sh"
echo ""
echo "To add a new test scenario:"
echo "  ./add-test-scenario.sh"
echo ""
echo "To see detailed documentation:"
echo "  cat README.md"

echo ""
echo "💡 Key Benefits of Refactored System:"
echo "  ✅ Configuration-driven (easy to modify)"
echo "  ✅ Modular components (seeder + test runner)"
echo "  ✅ Better error handling and logging"
echo "  ✅ Reusable test scenarios"
echo "  ✅ Easy to extend with new tests"
echo "  ✅ Clear separation of concerns"
echo ""
echo "Demo completed! 🎉"
