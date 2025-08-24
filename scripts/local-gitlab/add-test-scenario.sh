#!/bin/bash

set -euo pipefail

# Utility script to add a new test scenario
# Usage: ./add-test-scenario.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_PATH="${SCRIPT_DIR}/configs/test-scenarios.json"
TEMP_CONFIG="${CONFIG_PATH}.tmp"

echo "=== Add New Test Scenario ==="
echo ""

# Prompt for test details
read -p "Test name (kebab-case, e.g., 'my-new-test'): " test_name
read -p "Test description: " test_description
read -p "Ghorg command (use {{.BaseURL}}, {{.Token}}, {{.GhorgDir}} for templating): " test_command

echo ""
echo "Run command twice? (for testing clone then pull)"
read -p "Run twice (y/n): " run_twice_input
run_twice=$(if [[ "$run_twice_input" =~ ^[Yy]$ ]]; then echo "true"; else echo "false"; fi)

echo ""
echo "Expected structure (relative paths from ghorg directory):"
echo "Enter paths one by one, empty line to finish:"
expected_structure=()
while true; do
    read -p "Path (or empty to finish): " path
    if [[ -z "$path" ]]; then
        break
    fi
    expected_structure+=("$path")
done

echo ""
echo "Setup commands (optional, executed before main command):"
echo "Enter commands one by one, empty line to finish:"
setup_commands=()
while true; do
    read -p "Setup command (or empty to finish): " cmd
    if [[ -z "$cmd" ]]; then
        break
    fi
    setup_commands+=("$cmd")
done

echo ""
echo "Verification commands (optional, executed after main command):"
echo "Enter commands one by one, empty line to finish:"
verify_commands=()
while true; do
    read -p "Verify command (or empty to finish): " cmd
    if [[ -z "$cmd" ]]; then
        break
    fi
    verify_commands+=("$cmd")
done

# Create the new test scenario JSON
cat > /tmp/new_scenario.json << EOF
{
  "name": "$test_name",
  "description": "$test_description",
  "command": "$test_command",
  "run_twice": $run_twice,
EOF

# Add setup commands if any
if [[ ${#setup_commands[@]} -gt 0 ]]; then
    echo '  "setup_commands": [' >> /tmp/new_scenario.json
    for i in "${!setup_commands[@]}"; do
        if [[ $i -eq $((${#setup_commands[@]} - 1)) ]]; then
            echo "    \"${setup_commands[$i]}\"" >> /tmp/new_scenario.json
        else
            echo "    \"${setup_commands[$i]}\"," >> /tmp/new_scenario.json
        fi
    done
    echo '  ],' >> /tmp/new_scenario.json
fi

# Add verify commands if any
if [[ ${#verify_commands[@]} -gt 0 ]]; then
    echo '  "verify_commands": [' >> /tmp/new_scenario.json
    for i in "${!verify_commands[@]}"; do
        if [[ $i -eq $((${#verify_commands[@]} - 1)) ]]; then
            echo "    \"${verify_commands[$i]}\"" >> /tmp/new_scenario.json
        else
            echo "    \"${verify_commands[$i]}\"," >> /tmp/new_scenario.json
        fi
    done
    echo '  ],' >> /tmp/new_scenario.json
fi

# Add expected structure
echo '  "expected_structure": [' >> /tmp/new_scenario.json
for i in "${!expected_structure[@]}"; do
    if [[ $i -eq $((${#expected_structure[@]} - 1)) ]]; then
        echo "    \"${expected_structure[$i]}\"" >> /tmp/new_scenario.json
    else
        echo "    \"${expected_structure[$i]}\"," >> /tmp/new_scenario.json
    fi
done
echo '  ]' >> /tmp/new_scenario.json
echo '}' >> /tmp/new_scenario.json

echo ""
echo "=== Preview of New Test Scenario ==="
cat /tmp/new_scenario.json
echo ""

read -p "Add this test scenario to the configuration? (y/n): " confirm

if [[ "$confirm" =~ ^[Yy]$ ]]; then
    # Parse the current config and add the new scenario
    python3 << EOF
import json

# Read current config
with open('$CONFIG_PATH', 'r') as f:
    config = json.load(f)

# Read new scenario
with open('/tmp/new_scenario.json', 'r') as f:
    new_scenario = json.load(f)

# Add to scenarios
config['test_scenarios'].append(new_scenario)

# Write back
with open('$CONFIG_PATH', 'w') as f:
    json.dump(config, f, indent=2)

print(f"Added test scenario '{new_scenario['name']}' to configuration")
EOF

    echo "Test scenario added successfully!"
    echo "You can now run it with:"
    echo "  ./integration-tests.sh # (runs all tests)"
    echo "  or"
    echo "  ./test-runner/gitlab-test-runner -test=\"$test_name\" # (runs specific test)"
else
    echo "Test scenario was not added."
fi

# Clean up
rm -f /tmp/new_scenario.json
