# Refactored GitLab Integration Tests

This directory contains the refactored GitLab integration test system that replaces the monolithic bash scripts with modular, maintainable Go-based tools.

## Overview

The refactored system consists of:

1. **Configuration-based Seeding**: JSON configuration files define the seed data
2. **Go-based Seeder**: A Go tool that reads configuration and creates GitLab resources
3. **Test Framework**: A Go-based test runner that executes configurable test scenarios
4. **Modular Scripts**: Clean shell scripts that orchestrate the components

## Directory Structure

```
scripts/local-gitlab/
├── configs/
│   ├── seed-data.json          # Defines GitLab resources to create
│   └── test-scenarios.json     # Defines integration test scenarios
├── seeder/
│   ├── main.go                 # Go-based seeder implementation
│   └── go.mod                  # Seeder dependencies
├── test-runner/
│   ├── main.go                 # Go-based test runner implementation
│   └── go.mod                  # Test runner dependencies
├── start-ee.sh                # Refactored main entry point
├── seed.sh                    # New seeding script using Go seeder
├── integration-tests.sh       # New test script using Go test runner
├── add-test-scenario.sh       # Utility to add new test scenarios
└── README-refactored.md       # This file
```

## Quick Start

### Running All Tests (Refactored Version)

```bash
# Run the refactored integration tests
./start-ee.sh

# Or with custom parameters
./start-ee.sh true false latest
```

### Running Individual Components

```bash
# Seed GitLab instance only
./seed.sh "password" "http://gitlab.example.com" "${HOME}/ghorg"

# Run integration tests only (assumes seeded instance)
./integration-tests.sh "${HOME}/ghorg" "password" "http://gitlab.example.com"
```

## Configuration

### Seed Data Configuration (`configs/seed-data.json`)

Defines the GitLab resources to create during seeding:

```json
{
  "groups": [
    {
      "name": "my-group",
      "path": "my-group",
      "description": "My test group",
      "repositories": [
        {
          "name": "my-repo",
          "initialize_with_readme": true,
          "snippets": [
            {
              "title": "My Snippet",
              "file_name": "test.txt",
              "content": "Test content",
              "visibility": "public"
            }
          ]
        }
      ],
      "subgroups": [...]
    }
  ],
  "users": [...],
  "root_user": {...},
  "root_snippets": [...]
}
```

### Test Scenarios Configuration (`configs/test-scenarios.json`)

Defines the integration test scenarios:

```json
{
  "test_scenarios": [
    {
      "name": "my-test-scenario",
      "description": "Test description",
      "command": "ghorg clone all-groups --scm=gitlab --base-url={{.BaseURL}} --token={{.Token}} --output-dir=test-output",
      "run_twice": true,
      "setup_commands": ["git init {{.GhorgDir}}/test-setup"],
      "verify_commands": ["test -d '{{.GhorgDir}}/test-output'"],
      "expected_structure": [
        "test-output/group1/repo1",
        "test-output/group2/repo2"
      ]
    }
  ]
}
```

## Adding New Seed Data

1. **Edit the configuration**: Modify `configs/seed-data.json` to add new groups, repositories, users, or snippets
2. **Test the changes**: Run `./seed.sh` to verify the new seed data is created correctly

### Example: Adding a New Group

```json
{
  "name": "new-group",
  "path": "new-group",
  "description": "Description of the new group",
  "repositories": [
    {
      "name": "new-repo",
      "initialize_with_readme": true
    }
  ]
}
```

## Adding New Test Scenarios

### Method 1: Use the Helper Script

```bash
./add-test-scenario.sh
```

This interactive script will guide you through creating a new test scenario.

### Method 2: Manual Configuration

1. Edit `configs/test-scenarios.json`
2. Add a new test scenario object to the `test_scenarios` array
3. Test with: `./test-runner/gitlab-test-runner -test="your-test-name"`

### Method 3: Programmatically

```bash
# Build the test runner
cd test-runner && go build -o gitlab-test-runner main.go

# List available tests
./gitlab-test-runner -list

# Run a specific test
./gitlab-test-runner -test="specific-test-name" -token="password" -base-url="http://gitlab.example.com"
```

## Template Variables

Both seeder and test runner support template variables:

- `{{.BaseURL}}` - GitLab base URL
- `{{.Token}}` - GitLab API token
- `{{.GhorgDir}}` - Ghorg directory path

## Development

### Building the Components

```bash
# Build seeder
cd seeder && go build -o gitlab-seeder main.go

# Build test runner
cd test-runner && go build -o gitlab-test-runner main.go
```

### Running Tests in Development

```bash
# Run specific test scenario
cd test-runner
go run main.go -test="all-groups-preserve-dir-output-dir" -token="password" -base-url="http://gitlab.example.com"

# List all available test scenarios
go run main.go -list -token="password"
```

## Advantages of Refactored System

1. **Maintainability**: Configuration-driven approach makes it easy to modify tests and seed data
2. **Modularity**: Separate components for seeding and testing
3. **Reusability**: Test scenarios can be easily copied and modified
4. **Better Error Handling**: Go-based tools provide clearer error messages
5. **Extensibility**: Easy to add new test scenarios or seed data configurations
6. **Documentation**: Clear separation of concerns and self-documenting configuration

## Migration from Old System

The refactored system is designed to be fully backward-compatible. The original scripts (`seed.sh`, `integration-tests.sh`, `start-ee.sh`) remain unchanged and continue to work.

To migrate to the refactored system:

1. Use `start-ee.sh` for the refactored system
2. All existing test scenarios have been converted to the new configuration format
3. The test results should be identical between old and new systems

## Troubleshooting

### Build Errors
```bash
# Ensure Go modules are downloaded
cd seeder && go mod download
cd test-runner && go mod download
```

### Test Failures
```bash
# Check GitLab is accessible
curl -I http://gitlab.example.com

# Verify seeding completed
./seeder/gitlab-seeder -token="password" -base-url="http://gitlab.example.com"

# Run specific failing test
./test-runner/gitlab-test-runner -test="failing-test-name" -token="password"
```

### Configuration Issues
```bash
# Validate JSON configuration
python3 -m json.tool configs/seed-data.json
python3 -m json.tool configs/test-scenarios.json
```
