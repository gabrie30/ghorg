# GitLab Integration Tests

This directory contains the refactored GitLab integration test system.

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

### Running All Tests

```bash
# Run the refactored integration tests
./start-ee.sh

# Or with custom parameters
./start-ee.sh true false latest
```

## Script Arguments

### Quick Reference

| **Script** | **Arguments** | **Purpose** |
|---|---|---|
| `start-ee.sh` | 7 optional args | Main entry point - runs entire test suite |
| `seed.sh` | 3 optional args | Seeds GitLab with test data |
| `integration-tests.sh` | 3 optional args | Runs integration tests only |
| `run-ee.sh` | 4 optional args | Starts GitLab container (internal) |

### `start-ee.sh` Arguments

The main entry point script accepts up to 7 optional arguments. All arguments have sensible defaults if not provided.

**Usage:**
```bash
./start-ee.sh [STOP_GITLAB_WHEN_FINISHED] [PERSIST_GITLAB_LOCALLY] [GITLAB_IMAGE_TAG] [GITLAB_HOME] [GITLAB_HOST] [GITLAB_URL] [LOCAL_GITLAB_GHORG_DIR]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `STOP_GITLAB_WHEN_FINISHED` | `'true'` | Whether to stop and remove the GitLab container after tests complete. Set to `'false'` to keep GitLab running for debugging. |
| `PERSIST_GITLAB_LOCALLY` | `'false'` | Whether to persist GitLab data locally across container restarts. Set to `'true'` to keep data between runs. |
| `GITLAB_IMAGE_TAG` | `'latest'` | GitLab Docker image tag to use. Can be specific version like `'16.4.0-ce.0'` or `'latest'`. |
| `GITLAB_HOME` | `"$HOME/ghorg/local-gitlab-ee-data-${GITLAB_IMAGE_TAG}"` | Directory where GitLab stores persistent data on the host machine. |
| `GITLAB_HOST` | `'gitlab.example.com'` | Hostname for the GitLab instance. Used for container networking and /etc/hosts entries. |
| `GITLAB_URL` | `'http://gitlab.example.com'` | Full URL to access the GitLab instance. Used by ghorg and the test tools. |
| `LOCAL_GITLAB_GHORG_DIR` | `"${HOME}/ghorg"` | Local directory where ghorg will clone repositories and store its working files. |

**Examples:**

```bash
# Default behavior - run tests and clean up
./start-ee.sh

# Keep GitLab running after tests for debugging
./start-ee.sh false

# Use specific GitLab version and keep it running
./start-ee.sh false false 16.4.0-ce.0

# Full custom configuration
./start-ee.sh true true latest /tmp/gitlab-data gitlab.local http://gitlab.local /tmp/ghorg
```

**Common Scenarios:**

```bash
# Development - keep GitLab running for multiple test iterations
./start-ee.sh false false latest

# CI/CD - use clean environment and cleanup afterwards (default)
./start-ee.sh true false latest

# Testing specific GitLab version
./start-ee.sh true false 16.3.0-ce.0

# Custom data persistence for repeated testing
./start-ee.sh false true latest /data/gitlab-persistent
```

### Individual Component Arguments

#### `seed.sh` Arguments

Seeds the GitLab instance with test data using the Go-based seeder.

**Usage:**
```bash
./seed.sh [API_TOKEN] [GITLAB_URL] [LOCAL_GITLAB_GHORG_DIR]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `API_TOKEN` | `"password"` | GitLab API token for authentication (default root password) |
| `GITLAB_URL` | `"http://gitlab.example.com"` | Full URL to the GitLab instance |
| `LOCAL_GITLAB_GHORG_DIR` | `"${HOME}/ghorg"` | Directory where ghorg stores its configuration and temp files |

**Example:**
```bash
# Use defaults
./seed.sh

# Custom parameters
./seed.sh "my-token" "http://gitlab.local:8080" "/tmp/ghorg"
```

#### `integration-tests.sh` Arguments

Runs the integration tests using the Go-based test runner.

**Usage:**
```bash
./integration-tests.sh [LOCAL_GITLAB_GHORG_DIR] [API_TOKEN] [GITLAB_URL]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `LOCAL_GITLAB_GHORG_DIR` | `"${HOME}/ghorg"` | Directory where ghorg will clone repositories for testing |
| `API_TOKEN` | `"password"` | GitLab API token for authentication |
| `GITLAB_URL` | `"http://gitlab.example.com"` | Full URL to the GitLab instance |

**Example:**
```bash
# Use defaults
./integration-tests.sh

# Custom parameters
./integration-tests.sh "/tmp/ghorg" "my-token" "http://gitlab.local:8080"
```

#### `run-ee.sh` Arguments (Internal)

Starts the GitLab Docker container. Called internally by `start-ee.sh`.

**Usage:**
```bash
./run-ee.sh [GITLAB_IMAGE_TAG] [GITLAB_HOME] [GITLAB_HOST] [PERSIST_GITLAB_LOCALLY]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `GITLAB_IMAGE_TAG` | `"latest"` | GitLab Docker image tag |
| `GITLAB_HOME` | Dynamic | Host directory for GitLab data persistence |
| `GITLAB_HOST` | `"gitlab.example.com"` | Container hostname |
| `PERSIST_GITLAB_LOCALLY` | `"false"` | Whether to persist data between container restarts |

#### Go Tool Arguments (Direct Usage)

For advanced usage, you can run the Go tools directly:

**Seeder (`seeder/gitlab-seeder`)**:
```bash
./gitlab-seeder [flags]
  -config string
        Path to seed data configuration file (default "configs/seed-data.json")
  -token string
        GitLab API token (required)
  -base-url string
        GitLab base URL (required)
```

**Test Runner (`test-runner/gitlab-test-runner`)**:
```bash
./gitlab-test-runner [flags]
  -config string
        Path to test scenarios configuration file (default "configs/test-scenarios.json")
  -token string
        GitLab API token (required)
  -base-url string
        GitLab base URL (required)
  -ghorg-dir string
        Ghorg directory path (default "${HOME}/ghorg")
  -test string
        Run specific test by name (optional)
  -list
        List all available tests and exit
```

**Examples:**
```bash
# List all available test scenarios
./test-runner/gitlab-test-runner -list -token="password"

# Run specific test
./test-runner/gitlab-test-runner -test="all-groups-preserve-dir-output-dir" -token="password" -base-url="http://gitlab.example.com"

# Seed with custom config
./seeder/gitlab-seeder -config="my-seed-data.json" -token="password" -base-url="http://gitlab.example.com"
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
