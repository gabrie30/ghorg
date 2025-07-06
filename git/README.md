# Git Sync Functionality Documentation

This document provides comprehensive documentation for the Git synchronization functionality in ghorg, which enables automatic synchronization of local repositories with their remote default branches.

## Overview

The sync functionality provides automatic synchronization of local default branches with their remote counterparts. This is particularly useful when recloning repositories that may have received updates since the last clone operation.

## Configuration

### Environment Variable

- **Name**: `GHORG_SYNC_DEFAULT_BRANCH`
- **Type**: Boolean (string representation)
- **Default**: `false` (sync disabled by default)
- **Valid Values**: `"true"`, `"false"`, or unset

### Command Line Flag

- **Flag**: `--sync-default-branch`
- **Type**: Boolean flag
- **Default**: `false` (sync disabled by default)
- **Effect**: When used, sets `GHORG_SYNC_DEFAULT_BRANCH=true`

### Configuration File

In `sample-conf.yaml`:
```yaml
GHORG_SYNC_DEFAULT_BRANCH: false
```

## Usage Examples

### Enable sync via command line flag
```bash
ghorg clone organization --sync-default-branch
```

### Enable sync via environment variable
```bash
export GHORG_SYNC_DEFAULT_BRANCH=true
ghorg clone organization
```

### Disable sync (default behavior)
```bash
ghorg clone organization
# or explicitly:
export GHORG_SYNC_DEFAULT_BRANCH=false
ghorg clone organization
```

## Safety Philosophy

The sync functionality follows a **safety-first approach** to prevent data loss and maintain repository integrity:

### Safety Checks

1. **Working Directory Changes**: Sync is skipped if the working directory has uncommitted changes
2. **Unpushed Commits**: Sync is skipped if the current branch has commits not present on the remote
3. **Divergent Commits**: Sync is skipped if the current branch has commits not on the default branch
4. **Remote Accessibility**: Sync is skipped if the remote repository is not accessible

### Default Behavior

- Sync is **disabled by default** to prevent unexpected changes
- Must be explicitly enabled via flag or configuration
- When disabled, repositories are cloned but not synchronized

## Technical Implementation

### Core Functions

- `SyncDefaultBranch(repo scm.Repo) error`: Main sync function
- `hasLocalChanges(repo scm.Repo) (bool, error)`: Checks for uncommitted changes
- `hasUnpushedCommits(repo scm.Repo) (bool, error)`: Checks for unpushed commits
- `hasCommitsNotOnDefaultBranch(repo scm.Repo) (bool, error)`: Checks for divergent commits
- `getCurrentBranch(repo scm.Repo) (string, error)`: Gets current branch name

### Sync Process

1. Check if sync is enabled via `GHORG_SYNC_DEFAULT_BRANCH`
2. Verify remote origin exists and is accessible
3. Check for working directory changes
4. Check for unpushed commits
5. Check for commits not on default branch
6. If all safety checks pass, perform git operations:
   - Fetch latest changes from remote
   - Reset local default branch to match remote

### Error Handling

- All safety checks are performed before any destructive operations
- Detailed error messages are provided for troubleshooting
- Debug mode provides additional logging information

## Debug Mode

Enable debug output to see detailed sync information:

```bash
export GHORG_DEBUG=true
export GHORG_SYNC_DEFAULT_BRANCH=true
ghorg clone organization
```

Debug output includes:
- Sync enable/disable status
- Safety check results
- Skip reasons when sync is not performed
- Git command execution details

## Integration with Other Features

### Partial Clone Support

Sync functionality works with:
- Blob filters (`--filter=blob:none`)
- Sparse checkout patterns
- Shallow clones with depth limitations

### Compatibility

- Works with all supported Git providers (GitHub, GitLab, Bitbucket, Gitea)
- Compatible with all clone types (org, user, repo)
- Supports both SSH and HTTPS protocols

## Testing

Comprehensive test coverage includes:

### Unit Tests
- Configuration handling
- Safety check functions
- Error conditions
- Edge cases

### Integration Tests
- Flag processing
- Environment variable handling
- Complete sync workflows
- Multi-repository scenarios

### Test Helpers
- `setupTestRepo()`: Creates temporary test repositories
- `createCommitInRepo()`: Adds commits for testing
- `configureGitIdentity()`: Sets up git user/email for tests
- `disableGPGSigning()`: Disables GPG signing in test repos

## Troubleshooting

### Common Issues

1. **Sync not happening**
   - Verify `GHORG_SYNC_DEFAULT_BRANCH=true` is set
   - Check for working directory changes
   - Verify remote accessibility

2. **Sync skipped due to local changes**
   - Commit or stash local changes
   - Or disable sync for this operation

3. **Sync skipped due to unpushed commits**
   - Push commits to remote
   - Or disable sync for this operation

4. **Permission errors**
   - Verify git credentials are correctly configured
   - Check repository access permissions

### Debug Information

Enable debug mode to get detailed information:
```bash
export GHORG_DEBUG=true
```

This will show:
- Whether sync is enabled/disabled
- Which safety checks are being performed
- Reasons for skipping sync operations
- Git command outputs

## Best Practices

1. **Start with sync disabled** (default) when first using ghorg
2. **Enable sync only when needed** for automated environments
3. **Use debug mode** when troubleshooting sync issues
4. **Commit local changes** before enabling sync
5. **Push commits** before enabling sync to avoid conflicts
6. **Test sync behavior** in non-production environments first

## Migration Notes

### From Previous Versions

If upgrading from a version without sync functionality:
- No action required - sync is disabled by default
- Existing clone operations continue to work unchanged
- Enable sync explicitly when ready to use the feature

### Configuration Changes

- `GHORG_SYNC_DEFAULT` was renamed to `GHORG_SYNC_DEFAULT_BRANCH`
- All references have been updated in code, tests, and documentation
- Old environment variable name is no longer supported

## See Also

- Main ghorg README for general usage
- `sample-conf.yaml` for configuration examples
- Test files (`sync_test.go`, `clone_test.go`) for usage examples
- Source code in `git/sync.go` for implementation details
