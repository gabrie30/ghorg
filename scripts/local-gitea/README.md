# Gitea Integration Tests

This directory contains the Gitea integration test system, modeled after the GitLab integration test system.

## Directory Structure

```
scripts/local-gitea/
├── configs/
│   ├── seed-data.json          # Defines Gitea resources to create
│   └── test-scenarios.json     # Defines integration test scenarios
├── seeder/
│   ├── main.go                 # Go-based seeder implementation
│   └── go.mod                  # Seeder dependencies
├── test-runner/
│   ├── main.go                 # Go-based test runner implementation
│   └── go.mod                  # Test runner dependencies
├── start.sh                   # Main entry point
├── run.sh                     # Gitea container startup script
├── get_credentials.sh         # Setup admin user and credentials
├── seed.sh                    # Seeding script using Go seeder
├── integration-tests.sh       # Test script using Go test runner
└── README.md                  # This file
```

## Quick Start

### Running All Tests

```bash
# Run the Gitea integration tests
./start.sh

# Or with custom parameters
./start.sh true false latest
```

## Script Arguments

### Quick Reference

| **Script** | **Arguments** | **Purpose** |
|---|---|---|
| `start.sh` | 7 optional args | Main entry point - runs entire test suite |
| `seed.sh` | 3 optional args | Seeds Gitea with test data |
| `integration-tests.sh` | 3 optional args | Runs integration tests only |
| `run.sh` | 4 optional args | Starts Gitea container (internal) |

### `start.sh` Arguments

The main entry point script accepts up to 7 optional arguments. All arguments have sensible defaults if not provided.

**Usage:**
```bash
./start.sh [STOP_GITEA_WHEN_FINISHED] [PERSIST_GITEA_LOCALLY] [GITEA_IMAGE_TAG] [GITEA_HOME] [GITEA_HOST] [GITEA_URL] [LOCAL_GITEA_GHORG_DIR]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `STOP_GITEA_WHEN_FINISHED` | `'true'` | Whether to stop and remove the Gitea container after tests complete. Set to `'false'` to keep Gitea running for debugging. |
| `PERSIST_GITEA_LOCALLY` | `'false'` | Whether to persist Gitea data locally across container restarts. Set to `'true'` to keep data between runs. |
| `GITEA_IMAGE_TAG` | `'latest'` | Gitea Docker image tag to use. Can be specific version like `'1.20.0'` or `'latest'`. |
| `GITEA_HOME` | `"$HOME/ghorg/local-gitea-data-${GITEA_IMAGE_TAG}"` | Directory where Gitea stores persistent data on the host machine. |
| `GITEA_HOST` | `'gitea.example.com'` | Hostname for the Gitea instance. Used for container networking and /etc/hosts entries. |
| `GITEA_URL` | `'http://gitea.example.com:3000'` | Full URL to access the Gitea instance. Used by ghorg and the test tools. |
| `LOCAL_GITEA_GHORG_DIR` | `"${HOME}/ghorg"` | Local directory where ghorg will clone repositories and store its working files. |

**Examples:**

```bash
# Default behavior - run tests and clean up
./start.sh

# Keep Gitea running after tests for debugging
./start.sh false

# Use specific Gitea version and keep it running
./start.sh false false 1.20.0

# Full custom configuration
./start.sh true true latest /tmp/gitea-data gitea.local http://gitea.local:3000 /tmp/ghorg
```

**Common Scenarios:**

```bash
# Development - keep Gitea running for multiple test iterations
./start.sh false false latest

# CI/CD - use clean environment and cleanup afterwards (default)
./start.sh true false latest

# Testing specific Gitea version
./start.sh true false 1.19.0

# Custom data persistence for repeated testing
./start.sh false true latest /data/gitea-persistent
```

### Individual Component Arguments

#### `seed.sh` Arguments

Seeds the Gitea instance with test data using the Go-based seeder.

**Usage:**
```bash
./seed.sh [API_TOKEN] [GITEA_URL] [LOCAL_GITEA_GHORG_DIR]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `API_TOKEN` | From `${LOCAL_GITEA_GHORG_DIR}/gitea_token` | Gitea API token for authentication |
| `GITEA_URL` | `"http://gitea.example.com:3000"` | Full URL to the Gitea instance |
| `LOCAL_GITEA_GHORG_DIR` | `"${HOME}/ghorg"` | Directory where ghorg stores its configuration and temp files |

**Example:**
```bash
# Use defaults
./seed.sh

# Custom parameters
./seed.sh "my-token" "http://gitea.local:3000" "/tmp/ghorg"
```

#### `integration-tests.sh` Arguments

Runs the integration tests using the Go-based test runner.

**Usage:**
```bash
./integration-tests.sh [LOCAL_GITEA_GHORG_DIR] [API_TOKEN] [GITEA_URL]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `LOCAL_GITEA_GHORG_DIR` | `"${HOME}/ghorg"` | Directory where ghorg will clone repositories for testing |
| `API_TOKEN` | From `${LOCAL_GITEA_GHORG_DIR}/gitea_token` | Gitea API token for authentication |
| `GITEA_URL` | `"http://gitea.example.com:3000"` | Full URL to the Gitea instance |

**Example:**
```bash
# Use defaults
./integration-tests.sh

# Custom parameters
./integration-tests.sh "/tmp/ghorg" "my-token" "http://gitea.local:3000"
```

#### `run.sh` Arguments (Internal)

Starts the Gitea Docker container. Called internally by `start.sh`.

**Usage:**
```bash
./run.sh [GITEA_IMAGE_TAG] [GITEA_HOME] [GITEA_HOST] [PERSIST_GITEA_LOCALLY]
```

| **Argument** | **Default** | **Description** |
|---|---|---|
| `GITEA_IMAGE_TAG` | `"latest"` | Gitea Docker image tag |
| `GITEA_HOME` | Dynamic | Host directory for Gitea data persistence |
| `GITEA_HOST` | `"gitea.example.com"` | Container hostname |
| `PERSIST_GITEA_LOCALLY` | `"false"` | Whether to persist data between container restarts |

#### Go Tool Arguments (Direct Usage)

For advanced usage, you can run the Go tools directly:

**Seeder (`seeder/gitea-seeder`)**:
```bash
./gitea-seeder [flags]
  -config string
        Path to seed data configuration file (default "configs/seed-data.json")
  -token string
        Gitea API token (required)
  -base-url string
        Gitea base URL (required)
```

**Test Runner (`test-runner/gitea-test-runner`)**:
```bash
./gitea-test-runner [flags]
  -config string
        Path to test scenarios configuration file (default "configs/test-scenarios.json")
  -token string
        Gitea API token (required)
  -base-url string
        Gitea base URL (required)
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
./test-runner/gitea-test-runner -list -token="your-token"

# Run specific test
./test-runner/gitea-test-runner -test="all-orgs-basic" -token="your-token" -base-url="http://gitea.example.com:3000"

# Seed with custom config
./seeder/gitea-seeder -config="my-seed-data.json" -token="your-token" -base-url="http://gitea.example.com:3000"
```

### Running Individual Components

```bash
# Seed Gitea instance only
./seed.sh "your-token" "http://gitea.example.com:3000" "${HOME}/ghorg"

# Run integration tests only (assumes seeded instance)
./integration-tests.sh "${HOME}/ghorg" "your-token" "http://gitea.example.com:3000"
```

## Configuration

### Seed Data Configuration (`configs/seed-data.json`)

Defines the Gitea resources to create during seeding:

```json
{
  "organizations": [
    {
      "name": "My Organization",
      "username": "my-org",
      "description": "My test organization",
      "repositories": [
        {
          "name": "my-repo",
          "initialize_with_readme": true,
          "description": "My test repository"
        }
      ]
    }
  ],
  "users": [
    {
      "username": "testuser",
      "email": "test@example.com",
      "password": "password123",
      "full_name": "Test User",
      "repositories": [...]
    }
  ],
  "root_user": {
    "repositories": [...]
  }
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
      "command": "ghorg clone all-orgs --scm=gitea --base-url={{.BaseURL}} --token={{.Token}} --output-dir=test-output",
      "run_twice": true,
      "setup_commands": ["git init {{.GhorgDir}}/test-setup"],
      "verify_commands": ["test -d '{{.GhorgDir}}/test-output'"],
      "expected_structure": [
        "test-output/org1/repo1",
        "test-output/org2/repo2"
      ]
    }
  ]
}
```

## Adding New Seed Data

1. **Edit the configuration**: Modify `configs/seed-data.json` to add new organizations, repositories, or users
2. **Test the changes**: Run `./seed.sh` to verify the new seed data is created correctly

### Example: Adding a New Organization

```json
{
  "name": "New Organization",
  "username": "new-org",
  "description": "Description of the new organization",
  "repositories": [
    {
      "name": "new-repo",
      "initialize_with_readme": true,
      "description": "New repository description"
    }
  ]
}
```

## Adding New Test Scenarios

### Manual Configuration

1. Edit `configs/test-scenarios.json`
2. Add a new test scenario object to the `test_scenarios` array
3. Test with: `./test-runner/gitea-test-runner -test="your-test-name"`

### Programmatically

```bash
# Build the test runner
cd test-runner && go build -o gitea-test-runner main.go

# List available tests
./gitea-test-runner -list -token="your-token"

# Run a specific test
./gitea-test-runner -test="specific-test-name" -token="your-token" -base-url="http://gitea.example.com:3000"
```

## Template Variables

Both seeder and test runner support template variables:

- `{{.BaseURL}}` - Gitea base URL
- `{{.Token}}` - Gitea API token
- `{{.GhorgDir}}` - Ghorg directory path

## Development

### Building the Components

```bash
# Build seeder
cd seeder && go build -o gitea-seeder main.go

# Build test runner
cd test-runner && go build -o gitea-test-runner main.go
```

### Running Tests in Development

```bash
# Run specific test scenario
cd test-runner
go run main.go -test="all-orgs-basic" -token="your-token" -base-url="http://gitea.example.com:3000"

# List all available test scenarios
go run main.go -list -token="your-token"
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
# Check Gitea is accessible
curl -I http://gitea.example.com:3000

# Verify seeding completed
./seeder/gitea-seeder -token="your-token" -base-url="http://gitea.example.com:3000"

# Run specific failing test
./test-runner/gitea-test-runner -test="failing-test-name" -token="your-token"
```

### Configuration Issues
```bash
# Validate JSON configuration
python3 -m json.tool configs/seed-data.json
python3 -m json.tool configs/test-scenarios.json
```

## Differences from GitLab Integration Tests

- Uses **organizations** instead of **groups** (Gitea terminology)
- Gitea runs on port **3000** by default (vs GitLab's 80/443)
- Different API endpoints and authentication mechanisms
- Simplified user management (no complex namespace handling)
- Uses the `code.gitea.io/sdk/gitea` Go SDK instead of GitLab's SDK

## GitHub Actions Integration

The Gitea integration tests can be run in GitHub Actions via the workflow file at `.github/workflows/gitea-integration-tests.yml`. This workflow:

1. Sets up Go and Docker
2. Builds ghorg
3. Adds necessary host entries
4. Runs the full Gitea integration test suite

The tests run automatically on pull requests to ensure ghorg's Gitea functionality remains working.
