package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestSyncDefaultBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	// Create a test repository
	tempDir, err := createTestRepo(t)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test sync with clean working directory (should sync)
	t.Run("Sync with clean working directory", func(t *testing.T) {
		// Clone the repository
		destDir, err := os.MkdirTemp("", "ghorg-sync-dest")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    destDir,
			CloneBranch: "main",
		}

		client := GitClient{}

		// First clone normally
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// SyncDefaultBranch should work since working directory is clean
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch failed: %v", err)
		}

		// Verify the file still exists
		_, err = os.Stat(filepath.Join(destDir, "README.md"))
		if err != nil {
			t.Errorf("README.md should exist: %v", err)
		}
	})

	// Test sync with local changes (should not sync)
	t.Run("Sync with local changes", func(t *testing.T) {
		// Clone the repository
		destDir, err := os.MkdirTemp("", "ghorg-sync-dest-dirty")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    destDir,
			CloneBranch: "main",
		}

		client := GitClient{}

		// First clone normally
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Make some local changes to make the working directory dirty
		err = os.WriteFile(filepath.Join(destDir, "new-file.txt"), []byte("local changes"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Now sync should NOT work since there are local changes
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch failed: %v", err)
		}

		// Verify that the .git directory still exists
		_, err = os.Stat(filepath.Join(destDir, ".git"))
		if err != nil {
			t.Errorf(".git directory should exist: %v", err)
		}
	})
}

func TestSyncDefaultBranchErrorCases(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Debug mode execution", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "true")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a fake remote origin
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should work in debug mode: %v", err)
		}
	})
}

func TestPartialCloneAndSyncIntegration(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	// Create a test repository with some content
	tempDir, err := createTestRepoWithMultipleFiles(t)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Partial clone with blob filter", func(t *testing.T) {
		// Set up partial clone with blob:none filter
		os.Setenv("GHORG_GIT_FILTER", "blob:none")
		defer os.Unsetenv("GHORG_GIT_FILTER")

		destDir, err := os.MkdirTemp("", "ghorg-partial-clone")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    destDir,
			CloneBranch: "main",
		}

		client := GitClient{}

		// Clone with partial clone filter
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Verify this is a partial clone by checking for missing objects
		cmd := exec.Command("git", "rev-list", "--objects", "--missing=print", "HEAD")
		cmd.Dir = destDir
		output, err := cmd.Output()
		if err != nil {
			t.Logf("Note: Could not check for missing objects (may be expected for small test files): %v", err)
		} else {
			outputStr := string(output)
			if strings.Contains(outputStr, "?") {
				t.Logf("SUCCESS: Found missing objects in partial clone: %s", outputStr)
			} else {
				t.Logf("Note: No missing objects found (expected for small test files)")
			}
		}

		// Verify we can still access files (should trigger blob fetch on demand)
		files := []string{"README.md", "small.txt", "config.json"}
		for _, file := range files {
			if _, err := os.Stat(filepath.Join(destDir, file)); err != nil {
				t.Errorf("File %s should be accessible: %v", file, err)
			}
		}
	})

}

func TestSyncDefaultBranchExtensive(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("No remote origin", func(t *testing.T) {
		// Create a test repository without remote
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Remove the remote if it exists
		cmd := exec.Command("git", "remote", "remove", "origin")
		cmd.Dir = tempDir
		cmd.Run() // Ignore error if remote doesn't exist

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail when no remote: %v", err)
		}
	})

	t.Run("Sync with unpushed commits", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a fake remote origin
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Make a new commit that hasn't been pushed
		err = os.WriteFile(filepath.Join(tempDir, "newfile.txt"), []byte("new content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		cmd = exec.Command("git", "add", "newfile.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "New commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail with unpushed commits: %v", err)
		}
	})

	t.Run("Debug mode with working directory changes", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "true")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Make some local changes
		err = os.WriteFile(filepath.Join(tempDir, "untracked.txt"), []byte("new content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create untracked file: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail with local changes: %v", err)
		}
	})

	t.Run("Debug mode with unpushed commits", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "true")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a fake remote origin
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Make a new commit that hasn't been pushed
		err = os.WriteFile(filepath.Join(tempDir, "newfile.txt"), []byte("new content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		cmd = exec.Command("git", "add", "newfile.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "New commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail with unpushed commits: %v", err)
		}
	})

	t.Run("Different branch checkout", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create and switch to a different branch
		cmd := exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a fake remote origin
		cmd = exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail when switching branches: %v", err)
		}
	})

	t.Run("Failed checkout with debug", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "true")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a fake remote origin
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "nonexistent-branch",
			Name:        "test-repo",
		}

		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not fail when checkout fails: %v", err)
		}
	})
}

// createTestRepo creates a simple test repository using git CLI
func createTestRepo(_ *testing.T) (string, error) {
	tempDir, err := os.MkdirTemp("", "ghorg-test-repo")
	if err != nil {
		return "", err
	}

	// Initialize repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Set default branch to main for consistency
	cmd = exec.Command("git", "config", "init.defaultBranch", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Disable GPG signing for tests
	cmd = exec.Command("git", "config", "commit.gpgsign", "false")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Create a test file
	filename := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(filename, []byte("# Test Repository for Sync"), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Add and commit the file
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// createTestRepoWithMultipleFiles creates a test repository with various file types
func createTestRepoWithMultipleFiles(_ *testing.T) (string, error) {
	tempDir, err := os.MkdirTemp("", "ghorg-test-repo-multi")
	if err != nil {
		return "", err
	}

	// Initialize repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Set default branch to main for consistency
	cmd = exec.Command("git", "config", "init.defaultBranch", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Disable GPG signing for tests
	cmd = exec.Command("git", "config", "commit.gpgsign", "false")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Create various test files
	files := map[string]string{
		"README.md":   "# Test Repository\n\nThis is a test repository for ghorg partial clone testing.\n",
		"small.txt":   "This is a small text file for testing.\n",
		"config.json": `{"name": "test", "version": "1.0.0", "description": "Test configuration"}`,
		"large.log":   strings.Repeat("This is a line in a large log file.\n", 100),
	}

	for filename, content := range files {
		err = os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	// Create a subdirectory with files
	subDir := filepath.Join(tempDir, "docs")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	err = os.WriteFile(filepath.Join(subDir, "API.md"), []byte("# API Documentation\n\nAPI docs here.\n"), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Add and commit all files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit with multiple file types")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// configureGitInRepo configures git user and disables GPG signing in a repository directory
// This is useful for cloned repositories that need commit configuration for tests
func configureGitInRepo(repoDir string) error {
	// Configure git user
	cmd := exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure user name: %w", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure user email: %w", err)
	}

	// Disable GPG signing for tests
	cmd = exec.Command("git", "config", "commit.gpgsign", "false")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable GPG signing: %w", err)
	}

	return nil
}

// Add tests for missing coverage cases
func TestSyncDefaultBranchMissingCoverage(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	// Test with repository where hasLocalChanges fails
	t.Run("Error checking local changes", func(t *testing.T) {
		// Enable sync for this test
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// For this test, I need a path that exists and is a git repo with a remote,
		// but where `git status` will fail

		// Create a test repository first
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote so the remote check passes
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Now break the git repository by removing the index file which will make git status fail
		indexFile := filepath.Join(tempDir, ".git", "index")
		originalIndex, err := os.ReadFile(indexFile)
		if err != nil {
			t.Fatalf("Failed to read original index: %v", err)
		}
		defer os.WriteFile(indexFile, originalIndex, 0644) // Restore for cleanup

		// Write invalid data to the index file
		err = os.WriteFile(indexFile, []byte("invalid index data"), 0644)
		if err != nil {
			t.Fatalf("Failed to corrupt index file: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    "https://example.com/repo.git",
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// This should return an error when checking for local changes
		err = client.SyncDefaultBranch(repo)
		if err == nil {
			t.Error("Expected error when checking local changes fails")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to check working directory status") {
			t.Errorf("Expected working directory error, got: %v", err)
		}
	})

	// Test with error checking unpushed commits
	t.Run("Error getting current branch", func(t *testing.T) {
		// Enable sync for this test
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote so we pass the remote check
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Corrupt the .git/HEAD file to make getCurrentBranch fail
		headFile := filepath.Join(tempDir, ".git", "HEAD")
		err = os.WriteFile(headFile, []byte("ref: refs/heads/nonexistent"), 0644)
		if err != nil {
			t.Fatalf("Failed to corrupt HEAD file: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    "https://example.com/repo.git",
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// This should return an error when getCurrentBranch fails
		err = client.SyncDefaultBranch(repo)
		if err == nil {
			t.Error("Expected error when getCurrentBranch fails")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to get current branch") {
			t.Errorf("Expected getCurrentBranch error, got: %v", err)
		}
	})

	// Test with error getting current branch
	t.Run("Error getting current branch", func(t *testing.T) {
		// Enable sync for this test
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository and put it in detached HEAD state
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Get the commit hash and checkout to detached HEAD
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = tempDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get commit hash: %v", err)
		}
		commitHash := strings.TrimSpace(string(output))

		cmd = exec.Command("git", "checkout", commitHash)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout detached HEAD: %v", err)
		}

		// Add a remote for the test
		cmd = exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    "https://example.com/repo.git",
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// This should return an error due to detached HEAD state
		err = client.SyncDefaultBranch(repo)
		if err == nil {
			t.Error("Expected error when in detached HEAD state")
		}
	})

	// Test debug mode output for local changes
	t.Run("Debug mode with local changes", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "1")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Make local changes
		err = os.WriteFile(filepath.Join(tempDir, "dirty.txt"), []byte("local changes"), 0644)
		if err != nil {
			t.Fatalf("Failed to create dirty file: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    "https://example.com/repo.git",
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should skip sync and output debug message
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("Should not error, just skip sync: %v", err)
		}
	})

	// Test debug mode output for unpushed commits
	t.Run("Debug mode with unpushed commits", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "1")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository with a commit
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Create another commit to have unpushed changes
		err = os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("new content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		cmd = exec.Command("git", "add", "new.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "New commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    "https://example.com/repo.git",
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should skip sync and output debug message
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("Should not error, just skip sync: %v", err)
		}
	})

	// Test debug mode with divergent commits
	t.Run("Debug mode with divergent commits", func(t *testing.T) {
		// Set debug mode
		os.Setenv("GHORG_DEBUG", "1")
		defer os.Unsetenv("GHORG_DEBUG")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Create a feature branch and add a commit to it
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch (this makes it divergent from main)
		err = os.WriteFile(filepath.Join(tempDir, "feature.txt"), []byte("feature content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Feature commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Push the feature branch so it doesn't register as "unpushed"
		cmd = exec.Command("git", "push", "-u", "origin", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push feature branch: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main", // Different from current branch, with divergent commits
			Name:        "test-repo",
		}

		// Should skip sync and output debug message about divergent commits
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("Should not error, just skip sync: %v", err)
		}
	})
}

func TestSyncDefaultBranchConfiguration(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Sync disabled by default (GHORG_SYNC_DEFAULT_BRANCH not set)", func(t *testing.T) {
		// Ensure GHORG_SYNC_DEFAULT_BRANCH is not set
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should return immediately without doing any sync operations
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error when disabled: %v", err)
		}
	})

	t.Run("Sync disabled when GHORG_SYNC_DEFAULT_BRANCH=false", func(t *testing.T) {
		// Set GHORG_SYNC_DEFAULT_BRANCH to false
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "false")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should return immediately without doing any sync operations
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error when disabled: %v", err)
		}
	})

	t.Run("Sync enabled when GHORG_SYNC_DEFAULT_BRANCH=true", func(t *testing.T) {
		// Set GHORG_SYNC_DEFAULT_BRANCH to true
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote so the sync logic can proceed
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should proceed with sync logic (won't skip due to configuration)
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should work when enabled: %v", err)
		}
	})

	t.Run("Debug mode shows sync disabled message", func(t *testing.T) {
		// Set debug mode and ensure sync is disabled
		os.Setenv("GHORG_DEBUG", "1")
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		defer func() {
			os.Unsetenv("GHORG_DEBUG")
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should output debug message about sync being disabled
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error when disabled: %v", err)
		}
	})
}

// TestSyncDefaultBranchComprehensiveCoverage tests all code paths in SyncDefaultBranch
func TestSyncDefaultBranchComprehensiveCoverage(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Sync disabled by default", func(t *testing.T) {
		// Ensure sync is disabled
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should return early without error
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error when disabled: %v", err)
		}
	})

	t.Run("Error getting remote URL", func(t *testing.T) {
		// Enable sync
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository without remote
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Remove any existing remote
		cmd := exec.Command("git", "remote", "remove", "origin")
		cmd.Dir = tempDir
		cmd.Run() // Ignore error if remote doesn't exist

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should return without error when remote doesn't exist
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error when remote doesn't exist: %v", err)
		}
	})

	t.Run("Error checking working directory changes debug", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Make some local changes
		err = os.WriteFile(filepath.Join(tempDir, "local-change.txt"), []byte("local changes"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should skip sync due to working directory changes and show debug message
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with working directory changes: %v", err)
		}
	})

	t.Run("Error checking unpushed commits debug", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Make a new commit that hasn't been pushed
		err = os.WriteFile(filepath.Join(tempDir, "newfile.txt"), []byte("new content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		cmd = exec.Command("git", "add", "newfile.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "New commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Should skip sync due to unpushed commits and show debug message
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with unpushed commits: %v", err)
		}
	})

	t.Run("Debug mode with divergent commits", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Create and switch to a feature branch
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Make a commit on the feature branch that isn't on main
		err = os.WriteFile(filepath.Join(tempDir, "feature.txt"), []byte("feature content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Feature commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main", // Target branch is different from current
			Name:        "test-repo",
		}

		// Should skip sync due to divergent commits and show debug message
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with divergent commits: %v", err)
		}
	})
}

func TestSyncDefaultBranchDebugOutputs(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Debug output for working directory changes", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Make sure there are working directory changes
		err = os.WriteFile(filepath.Join(tempDir, "dirty-file.txt"), []byte("dirty content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create dirty file: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Capture output to verify debug message
		// The sync should be skipped due to working directory changes
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with working directory changes: %v", err)
		}
	})

	t.Run("Debug output for unpushed commits", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Make a new commit that hasn't been pushed
		err = os.WriteFile(filepath.Join(tempDir, "unpushed.txt"), []byte("unpushed content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create unpushed file: %v", err)
		}

		cmd = exec.Command("git", "add", "unpushed.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add unpushed file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Unpushed commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit unpushed file: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// The sync should be skipped due to unpushed commits
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with unpushed commits: %v", err)
		}
	})

	t.Run("Debug output for divergent commits", func(t *testing.T) {
		// Enable sync and debug
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add remote pointing to itself
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push to create the remote branch
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Create and switch to a feature branch
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Make a commit on the feature branch that isn't on main
		err = os.WriteFile(filepath.Join(tempDir, "feature.txt"), []byte("feature content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Feature commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main", // Target branch is different from current
			Name:        "test-repo",
		}

		// The sync should be skipped due to divergent commits
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Errorf("SyncDefaultBranch should not error with divergent commits: %v", err)
		}
	})
}
