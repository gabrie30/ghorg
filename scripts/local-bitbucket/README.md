# Bitbucket Server Integration Tests

This directory contains integration tests for ghorg's Bitbucket Server support.

These tests are not in CI due to the image bitbucket provides requires some manual steps unfortunately.

## Overview

These tests verify that ghorg can correctly:
- Connect to Bitbucket Server instances (self-hosted)
- Clone projects and user repositories
- Handle various command-line options and configurations
- Work with both HTTP and HTTPS connections

## Prerequisites

- Docker and Docker Compose installed and running
- Go 1.24.6 or later
- ghorg binary built and in PATH (will be rebuilt automatically)

## Running Tests

### Quick Start

```bash
./start.sh
```

This will:
1. Start a local Bitbucket Server Docker container
2. **Pause for manual setup** (due to licensing requirements)
3. Wait for you to complete the setup in your browser
4. Continue with automated seeding and testing
5. Clean up resources

### Manual Setup Process

**Important**: Bitbucket Server requires manual setup due to licensing requirements that cannot be automated.

When you run `./start.sh`, it will:
1. Start Bitbucket Server container
2. Wait 60 seconds for startup
3. Display instructions and pause for manual setup:

```
🔧 MANUAL SETUP REQUIRED
========================================
Bitbucket Server is now running at: http://bitbucket.example.com:7990

Please complete the setup manually:
1. Open http://bitbucket.example.com:7990 in your browser
   - Create admin user with credentials:
     Username: admin
     Password: admin
     Email: admin@bitbucket.local
1. Enter your Bitbucket Data Center license key
1. Complete setup

Once setup is complete, press ENTER to continue with seeding and testing...
```

4. After you complete setup and press ENTER, it will automatically:
   - Verify the setup worked
   - Seed the instance with test data
   - Run comprehensive integration tests

### Individual Components

You can run individual components separately:

```bash
# Start Bitbucket Server only
./run.sh latest /tmp/bitbucket-data bitbucket.example.com false

# Seed with test data (after manual setup)
./seed.sh admin http://bitbucket.example.com:7990

# Run integration tests only
./integration-tests.sh
```

## Test Structure

### Test Scenarios

The integration tests cover 19 different scenarios including:
- Single workspace cloning
- User repository cloning
- Regex filtering
- Skip archived/forks flags
- Concurrent cloning
- Backup functionality
- Large workspace pagination
- Various ghorg command-line options

### Test Data

Test data is defined in `configs/seed-data.json` and includes:
- Multiple workspaces/projects with repositories
- Different repository types and configurations
- User accounts and permissions
- Large datasets for pagination testing (105+ repositories)

## Configuration

- **Bitbucket Host**: bitbucket.example.com (added to /etc/hosts)
- **Port**: 7990 (HTTP), 7999 (SSH), 22 (SSH)
- **Admin Credentials**: admin/admin
- **Database**: Internal H2 (for testing only)

## Troubleshooting

### Common Issues

1. **Manual Setup Required**:
   - This is expected! Bitbucket Server licensing prevents full automation
   - Complete the setup wizard in your browser as instructed
   - Use admin/admin credentials and Internal (H2) database

2. **Setup Verification Failed**:
   - Ensure you completed all setup steps
   - Verify admin user can log in at the web interface
   - Check that setup wizard is fully complete (no more setup pages)

3. **Connection Refused**:
   - Ensure Docker is running
   - Check container status: `docker ps`
   - Verify /etc/hosts entry: `127.0.0.1 bitbucket.example.com`

4. **Tests Failing**:
   - Ensure manual setup was completed successfully
   - Verify admin/admin credentials work
   - Check Docker logs: `docker logs bitbucket`

### Environment Variables

After manual setup, these are automatically configured:
- `GHORG_SCM_TYPE=bitbucket`
- `GHORG_SCM_BASE_URL=http://bitbucket.example.com:7990`
- `GHORG_BITBUCKET_USERNAME=admin`
- `GHORG_BITBUCKET_APP_PASSWORD=admin`
- `GHORG_INSECURE_BITBUCKET_CLIENT=true`

## CI/GitHub Actions

Due to the manual setup requirement, Bitbucket Server integration tests are **not included** in CI pipelines. These tests are designed for local development and testing only.

The GitHub Actions workflow for Bitbucket integration tests has been removed since automation is not possible.

## Why Manual Setup?

Bitbucket Server requires:
1. **License acceptance**: Cannot be automated due to legal/licensing constraints
2. **Admin user creation**: Tied to license acceptance process
3. **Database initialization**: Part of the setup wizard flow

This is different from GitLab/Gitea which allow API-based or configuration file automation.

## Directory Structure

```
local-bitbucket/
├── README.md                 # This file
├── start.sh                 # Main orchestration script (with manual setup pause)
├── run.sh                   # Docker container startup (simplified)
├── get_credentials.sh       # Credential verification (simplified)
├── seed.sh                  # Run the Go seeder
├── integration-tests.sh     # Run the Go test runner
├── configs/
│   ├── seed-data.json      # Test data configuration
│   └── test-scenarios.json # Test scenarios configuration
├── seeder/
│   ├── go.mod              # Go module for seeder
│   └── main.go             # Bitbucket seeder implementation
└── test-runner/
    ├── go.mod              # Go module for test runner
    └── main.go             # Test runner implementation
```

## Cleanup

```bash
# Stop and remove containers (from the local-bitbucket directory)
cd scripts/local-bitbucket
docker-compose down

# Remove volumes (optional - removes database data)
docker-compose down -v

# Remove test data
rm -rf /tmp/ghorg-*
```

## Troubleshooting

### Server Crashes During Integration Tests

**Problem**: Bitbucket Server crashes when running integration tests.

**Root Cause**: Previously caused by Bitbucket Data Center license incompatibility with H2 database. Now resolved by using PostgreSQL.

**Solution**: The integration tests now use PostgreSQL database which is fully compatible with Bitbucket Data Center licenses, eliminating the previous H2 compatibility issues.

**Database Configuration**: PostgreSQL runs in a separate Docker container and is automatically configured for integration testing.

### Common Issues

- **HTTP 500 errors**: Usually indicates server overload or database issues
- **Server unresponsive**: Check Docker container status with `docker ps`
- **API authentication failures**: Verify admin:admin credentials are set correctly
