# Git Package Documentation

This document provides comprehensive documentation for the Git package in ghorg, which handles all Git repository operations including cloning, synchronization, and repository management through a clean interface-based design.

## Overview

The Git package provides a unified interface for Git operations in ghorg through the `Gitter` interface. This includes:

- Repository cloning with various options (depth, filters, submodules)
- Automatic synchronization of local default branches with remote counterparts
- Working directory and commit status checking
- Branch management and checkout operations
- Remote repository URL handling

## Architecture

### Gitter Interface

The package is built around the `Gitter` interface, which defines all Git operations:

```go
type Gitter interface {
    Clone(repo scm.Repo) error
    SyncDefaultBranch(repo scm.Repo) error
    HasLocalChanges(repo scm.Repo) (bool, error)
    HasUnpushedCommits(repo scm.Repo) (bool, error)
    HasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error)
    GetCurrentBranch(repo scm.Repo) (string, error)
    GetRemoteURL(repo scm.Repo, remoteName string) (string, error)
    Checkout(repo scm.Repo) error
    FetchCloneBranch(repo scm.Repo) error
    UpdateRef(repo scm.Repo, refName, commitRef string) error
}
```

### GitClient Implementation

The `GitClient` struct implements the `Gitter` interface and provides the concrete implementation for all Git operations using the git CLI.

```go
type GitClient struct{}

func NewGit() Gitter {
    return GitClient{}
}
```

### Interface Benefits

The interface-based design provides several advantages:

- **Testability**: Easy mocking for unit tests
- **Flexibility**: Multiple implementations possible (e.g., different Git backends)
- **Consistency**: All Git operations go through the same interface
- **Maintainability**: Clear separation of concerns between Git operations and business logic

### Usage Pattern

Throughout ghorg, Git operations are accessed via the interface:

```go
gitClient := git.NewGit()
err := gitClient.Clone(repo)
if err != nil {
    return fmt.Errorf("clone failed: %w", err)
}

// Sync is automatically called if enabled
```

## Clone Functionality

### Supported Features

- **Standard Cloning**: Full repository clones with complete history
- **Shallow Clones**: Limited history depth via `GHORG_CLONE_DEPTH`
- **Partial Clones**: Blob filtering via `GHORG_GIT_FILTER` (e.g., `blob:none`)
- **Submodules**: Optional submodule inclusion via `GHORG_INCLUDE_SUBMODULES`
- **Branch Selection**: Clone specific branches via `GHORG_BRANCH`
- **Protocol Support**: Both HTTPS and SSH protocols

### Integration with Sync

The `Clone` method automatically calls `SyncDefaultBranch` when `GHORG_SYNC_DEFAULT_BRANCH=true`, providing seamless integration between cloning and synchronization.

## Sync Configuration

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

### Interface Methods

#### Core Git Operations

- `Clone(repo scm.Repo) error`: Complete repository cloning with all configured options
- `SyncDefaultBranch(repo scm.Repo) error`: Safe synchronization with remote default branch

#### Repository State Checking

- `HasLocalChanges(repo scm.Repo) (bool, error)`: Detects uncommitted changes in working directory
- `HasUnpushedCommits(repo scm.Repo) (bool, error)`: Checks for local commits not pushed to remote
- `HasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error)`: Detects divergent commits
- `GetCurrentBranch(repo scm.Repo) (string, error)`: Returns current branch name

#### Repository Management

- `GetRemoteURL(repo scm.Repo, remoteName string) (string, error)`: Retrieves remote repository URLs
- `Checkout(repo scm.Repo) error`: Switches to target branch
- `FetchCloneBranch(repo scm.Repo) error`: Fetches latest changes from remote
- `UpdateRef(repo scm.Repo, refName, commitRef string) error`: Updates branch references

### Sync Process

1. **Configuration Check**: Verify sync is enabled via `GHORG_SYNC_DEFAULT_BRANCH`
2. **Remote Validation**: Confirm remote origin exists and is accessible
3. **Safety Checks**: Perform all safety validations:
   - Check for working directory changes
   - Check for unpushed commits
   - Check for commits not on default branch
4. **Branch Operations**: If all checks pass:
   - Switch to target branch if necessary
   - Fetch latest changes from remote
   - Update local branch reference to match remote

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

### Integration with Clone Process

The sync functionality is seamlessly integrated into the clone workflow:

```go
func (g GitClient) Clone(repo scm.Repo) error {
    // ... clone operations ...
    
    // Automatically sync if enabled
    if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") == "true" {
        if err := g.SyncDefaultBranch(repo); err != nil {
            // Log error but don't fail the clone
        }
    }
    
    return nil
}
```

## Integration with Other Features

### Partial Clone Support

Sync functionality works seamlessly with:
- **Blob filters**: `GHORG_GIT_FILTER=blob:none` for excluding file contents
- **Tree filters**: `GHORG_GIT_FILTER=tree:0` for commit and tree objects only
- **Shallow clones**: `GHORG_CLONE_DEPTH` for limited history depth
- **Submodules**: `GHORG_INCLUDE_SUBMODULES` for recursive submodule handling

### Compatibility

- Works with all supported Git providers (GitHub, GitLab, Bitbucket, Gitea)
- Compatible with all clone types (org, user, repo)
- Supports both SSH and HTTPS protocols

## Testing

The package maintains high test coverage with well-organized test suites:

### Test Organization

- **`git_test.go`**: Unit tests for individual Gitter interface methods
- **`sync_test.go`**: Integration tests and comprehensive sync functionality testing
- **Coverage**: 70.2% code coverage for `sync.go` with targeted testing of all major paths

### Test Categories

#### Unit Tests (`git_test.go`)
- Individual method testing for each interface method
- Error condition handling
- Edge case validation
- Mock-friendly interface testing

#### Integration Tests (`sync_test.go`)
- End-to-end sync workflows
- Multi-repository scenarios
- Configuration testing
- Debug output validation
- Safety check verification

#### Test Infrastructure
- `createTestRepo()`: Creates temporary Git repositories for testing
- `createTestRepoWithMultipleFiles()`: Creates repositories with various file types
- `configureGitInRepo()`: Sets up Git configuration in test repositories
- Comprehensive cleanup and isolation between tests

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

## Go-Git Implementation & CLI Fallback

### Current Architecture

ghorg includes a secondary Git implementation using the go-git library in addition to the primary CLI-based implementation. This provides flexibility and demonstrates different approaches to Git operations:

- **Primary**: `GitClient` (CLI-based) - Production implementation in `git.go`
- **Secondary**: `GoGitClient` (go-git library) - Alternative implementation in `gogit.go`

### Go-Git Filter Fallback

The go-git implementation automatically falls back to the git CLI when filtering is requested, since go-git v5 does not support native filtering:

- **Without filtering**: Uses go-git for faster cloning without shell dependencies
- **With filtering**: Automatically falls back to git CLI for proper filter support

When `GHORG_GIT_FILTER` is set, the `GoGitClient` will:
1. Detect the filter environment variable
2. Log the fallback (if `GHORG_DEBUG` is enabled)
3. Use the `GitClient` implementation for proper filter support

### Migration to go-git v6

When go-git v6 becomes stable with native filter support in `CloneOptions`, the migration path is:

1. **Update** `gogit.go`:
   - Add back `"github.com/go-git/go-git/v6/plumbing/protocol/packp"` import
   - Replace the filter fallback section in the `Clone` method with:
     ```go
     // Use native go-git v6 filter support
     if gitFilter != "" {
         filter, err := parseGitFilterToPackpFilter(gitFilter)
         if err == nil && filter != "" {
             cloneOptions.Filter = filter // v6 native support
             if os.Getenv("GHORG_DEBUG") != "" {
                 fmt.Printf("Using native go-git v6 filter: %s\n", filter)
             }
         }
     }
     ```
2. **Add** native `parseGitFilterToPackpFilter` function back to `gogit.go`
3. **Remove** CLI fallback logic
4. **Update** dependencies to go-git v6

This architecture ensures a clean migration path while maintaining full filter functionality with the current go-git v5 limitations.

### Filter Support Matrix

| Filter Type | CLI Implementation | go-git v5 | go-git v6 (Future) |
|-------------|-------------------|-----------|-------------------|
| `blob:none` | ✅ Native | ⚠️ CLI Fallback | ✅ Native |
| `blob:limit=<size>` | ✅ Native | ⚠️ CLI Fallback | ✅ Native |
| `tree:<depth>` | ✅ Native | ⚠️ CLI Fallback | ✅ Native |
| Combined filters | ✅ Native | ⚠️ CLI Fallback | ✅ Native |

## Package Structure

```
git/
├── git.go           # Gitter interface and GitClient implementation (CLI-based)
├── gogit.go         # GoGitClient implementation (go-git library-based)
├── sync.go          # Sync functionality implementation
├── git_test.go      # Unit tests for git.go methods
├── gogit_test.go    # Unit tests for gogit.go methods
├── sync_test.go     # Integration tests for sync functionality
└── README.md        # This documentation
```

### Key Files

- **`git.go`**: Defines the `Gitter` interface and provides the `GitClient` implementation with all core Git operations (CLI-based)
- **`gogit.go`**: Provides the `GoGitClient` implementation using the go-git library (alternative implementation with CLI fallback for filtering)
- **`sync.go`**: Contains the `SyncDefaultBranch` method implementation with safety checks and sync logic
- **Test files**: Comprehensive test coverage ensuring reliability and maintainability


## See Also

- **Main ghorg README**: General usage and configuration
- **`sample-conf.yaml`**: Complete configuration examples
- **`cmd/root.go`**: Command-line flag definitions and defaults
- **Interface Usage**: All ghorg components use `git.NewGit()` for Git operations
- **Source Code**: 
  - `git/git.go`: Interface definition and implementation
  - `git/sync.go`: Sync functionality details
  - `git/*_test.go`: Comprehensive test examples
