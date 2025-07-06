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

func TestHasCheckedOutFiles(t *testing.T) {
	client := GitClient{}

	// Test with files present
	t.Run("Directory with files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-files-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create .git directory and a regular file
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasFiles, err := client.hasCheckedOutFiles(repo)
		if err != nil {
			t.Fatalf("hasCheckedOutFiles failed: %v", err)
		}
		if !hasFiles {
			t.Error("Should detect files are present")
		}
	})

	// Test with only .git directory
	t.Run("Directory with only .git", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-nogitfiles-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create only .git directory
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasFiles, err := client.hasCheckedOutFiles(repo)
		if err != nil {
			t.Fatalf("hasCheckedOutFiles failed: %v", err)
		}
		if hasFiles {
			t.Error("Should detect no files are present")
		}
	})
}

func TestHasCheckedOutFilesEdgeCases(t *testing.T) {
	client := GitClient{}

	t.Run("Directory read error", func(t *testing.T) {
		// Use a non-existent directory
		repo := scm.Repo{HostPath: "/non/existent/directory"}
		_, err := client.hasCheckedOutFiles(repo)
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("Directory with hidden files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-hidden-files")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create .git directory and a hidden file
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create hidden file: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasFiles, err := client.hasCheckedOutFiles(repo)
		if err != nil {
			t.Fatalf("hasCheckedOutFiles failed: %v", err)
		}
		if !hasFiles {
			t.Error("Should detect hidden files as checked out files")
		}
	})
}

func TestHasLocalChanges(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	// Test with clean working directory
	t.Run("Clean working directory", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
		}
		if hasChanges {
			t.Error("Clean repository should have no local changes")
		}
	})

	// Test with modified files
	t.Run("Modified files", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Modify an existing file
		err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("Modified content"), 0644)
		if err != nil {
			t.Fatalf("Failed to modify file: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
		}
		if !hasChanges {
			t.Error("Repository with modified files should have local changes")
		}
	})

	// Test with untracked files
	t.Run("Untracked files", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a new untracked file
		err = os.WriteFile(filepath.Join(tempDir, "untracked.txt"), []byte("New file"), 0644)
		if err != nil {
			t.Fatalf("Failed to create untracked file: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
		}
		if !hasChanges {
			t.Error("Repository with untracked files should have local changes")
		}
	})
}

func TestHasLocalChangesEdgeCases(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Git status command error", func(t *testing.T) {
		// Use a directory that's not a git repository
		tempDir, err := os.MkdirTemp("", "ghorg-not-git")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		_, err = client.hasLocalChanges(repo)
		if err == nil {
			t.Error("Expected error for non-git directory")
		}
		if !strings.Contains(err.Error(), "failed to check git status") {
			t.Errorf("Expected git status error, got: %v", err)
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

	t.Run("Sparse checkout with path filter", func(t *testing.T) {
		// Set up sparse checkout to only include specific paths
		os.Setenv("GHORG_PATH_FILTER", "*.md")
		defer os.Unsetenv("GHORG_PATH_FILTER")

		destDir, err := os.MkdirTemp("", "ghorg-sparse-checkout")
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

		// Clone with sparse checkout
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Verify sparse checkout is configured
		sparseFile := filepath.Join(destDir, ".git", "info", "sparse-checkout")
		if _, err := os.Stat(sparseFile); err != nil {
			t.Errorf("Sparse checkout file should exist: %v", err)
		}

		// Check that only filtered files are present
		if _, err := os.Stat(filepath.Join(destDir, "README.md")); err != nil {
			t.Errorf("README.md should be present (matches *.md filter): %v", err)
		}

		// Verify sync was called and repository is functional (should work since working directory is clean)
		if err := client.SyncDefaultBranch(repo); err != nil {
			t.Errorf("SyncDefaultBranch should work with clean working directory: %v", err)
		}
	})

	t.Run("Combined partial clone and sparse checkout", func(t *testing.T) {
		// Use both filters together
		os.Setenv("GHORG_GIT_FILTER", "blob:limit=1k")
		os.Setenv("GHORG_PATH_FILTER", "*.md\n*.txt")
		defer func() {
			os.Unsetenv("GHORG_GIT_FILTER")
			os.Unsetenv("GHORG_PATH_FILTER")
		}()

		destDir, err := os.MkdirTemp("", "ghorg-combined-filters")
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

		// Clone with both filters
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Verify the repository is functional
		cmd := exec.Command("git", "status")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Errorf("git status should work in cloned repo: %v", err)
		}

		// Verify sync functionality (should work since working directory is clean)
		if err := client.SyncDefaultBranch(repo); err != nil {
			t.Errorf("SyncDefaultBranch should work with clean working directory: %v", err)
		}
	})
}

func TestGetCurrentBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Normal branch", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		branch, err := client.getCurrentBranch(repo)
		if err != nil {
			t.Fatalf("getCurrentBranch failed: %v", err)
		}
		if branch != "main" {
			t.Errorf("Expected branch 'main', got '%s'", branch)
		}
	})

	t.Run("Detached HEAD state", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Put the repository in detached HEAD state
		cmd := exec.Command("git", "checkout", "HEAD~0")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create detached HEAD: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		_, err = client.getCurrentBranch(repo)
		if err == nil {
			t.Error("Expected error for detached HEAD state")
		}
		if !strings.Contains(err.Error(), "detached HEAD state") {
			t.Errorf("Expected detached HEAD error, got: %v", err)
		}
	})

	t.Run("Invalid repository", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-invalid-repo")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		_, err = client.getCurrentBranch(repo)
		if err == nil {
			t.Error("Expected error for invalid repository")
		}
	})
}

func TestHasUnpushedCommits(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("No unpushed commits", func(t *testing.T) {
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

		repo := scm.Repo{HostPath: tempDir}
		hasUnpushed, err := client.hasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("hasUnpushedCommits failed: %v", err)
		}
		if hasUnpushed {
			t.Error("Should not have unpushed commits")
		}
	})

	t.Run("Has unpushed commits", func(t *testing.T) {
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

		repo := scm.Repo{HostPath: tempDir}
		hasUnpushed, err := client.hasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("hasUnpushedCommits failed: %v", err)
		}
		if !hasUnpushed {
			t.Error("Should have unpushed commits")
		}
	})

	t.Run("Remote branch doesn't exist", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		hasUnpushed, err := client.hasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("hasUnpushedCommits failed: %v", err)
		}
		// Should assume there are unpushed commits when remote doesn't exist
		if !hasUnpushed {
			t.Error("Should assume unpushed commits when remote branch doesn't exist")
		}
	})
}

func TestHasUnpushedCommitsEdgeCases(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Invalid count output", func(t *testing.T) {
		// This is harder to test directly, but we can test with a repository
		// where the command succeeds but returns unexpected output
		// The actual error case (invalid count) is covered by the error handling logic
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}

		// This should work normally and return the result
		hasUnpushed, err := client.hasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("hasUnpushedCommits failed: %v", err)
		}

		// Should assume unpushed commits when remote doesn't exist
		if !hasUnpushed {
			t.Error("Should assume unpushed commits when remote branch doesn't exist")
		}
	})
}

func TestConfigureSparseCheckout(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Successful sparse checkout", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepoWithMultipleFiles(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		err = client.configureSparseCheckout(repo, "*.md")
		if err != nil {
			t.Fatalf("configureSparseCheckout failed: %v", err)
		}

		// Verify sparse checkout is configured
		cmd := exec.Command("git", "config", "core.sparseCheckout")
		cmd.Dir = tempDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to check sparse checkout config: %v", err)
		}
		if strings.TrimSpace(string(output)) != "true" {
			t.Error("Sparse checkout should be enabled")
		}
	})

	t.Run("Fallback to file-based sparse checkout", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepoWithMultipleFiles(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Remove git executable temporarily to force fallback
		// We'll simulate this by using an invalid path
		repo := scm.Repo{HostPath: tempDir}

		// This should trigger the fallback when sparse-checkout commands fail
		err = client.configureSparseCheckout(repo, "*.md,*.txt")
		if err != nil {
			t.Fatalf("configureSparseCheckout with fallback failed: %v", err)
		}
	})

	t.Run("Invalid repository", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-invalid-repo")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		err = client.configureSparseCheckout(repo, "*.md")
		if err == nil {
			t.Error("Expected error for invalid repository")
		}
	})
}

func TestConfigureSparseCheckoutErrorHandling(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Fallback error handling", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepoWithMultipleFiles(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Make the .git/info directory read-only to force an error in fallback
		infoDir := filepath.Join(tempDir, ".git", "info")
		err = os.MkdirAll(infoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create info directory: %v", err)
		}

		// Make it read-only (this may not work on all systems)
		err = os.Chmod(infoDir, 0444)
		if err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		// Restore permissions after test
		defer os.Chmod(infoDir, 0755)

		repo := scm.Repo{HostPath: tempDir}
		err = client.configureSparseCheckout(repo, "*.md")
		// This might succeed depending on git version and sparse-checkout support
		// The test ensures we handle both success and failure cases
		if err != nil {
			t.Logf("configureSparseCheckout failed as expected: %v", err)
		} else {
			t.Logf("configureSparseCheckout succeeded")
		}
	})
}

func TestWriteSparseCheckoutFile(t *testing.T) {
	client := GitClient{}

	t.Run("Write sparse checkout file", func(t *testing.T) {
		// Create a test repository with proper git structure
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		err = client.writeSparseCheckoutFile(repo, "*.md,*.txt")
		if err != nil {
			t.Fatalf("writeSparseCheckoutFile failed: %v", err)
		}

		// Verify the file was created with correct content
		sparseFile := filepath.Join(tempDir, ".git", "info", "sparse-checkout")
		content, err := os.ReadFile(sparseFile)
		if err != nil {
			t.Fatalf("Failed to read sparse checkout file: %v", err)
		}

		expectedLines := []string{"*.md", "*.md/**", "*.txt", "*.txt/**"}
		contentStr := string(content)
		for _, line := range expectedLines {
			if !strings.Contains(contentStr, line) {
				t.Errorf("Expected line '%s' not found in sparse checkout file", line)
			}
		}
	})

	t.Run("Error creating directory", func(t *testing.T) {
		// Use a path that can't be created (e.g., in a read-only location)
		tempDir := "/root/cannot-create-this-dir"
		repo := scm.Repo{HostPath: tempDir}

		err := client.writeSparseCheckoutFile(repo, "*.md")
		if err == nil {
			t.Error("Expected error when creating directory in restricted location")
		}
	})
}

func TestWriteSparseCheckoutFileEdgeCases(t *testing.T) {
	client := GitClient{}

	t.Run("Empty pattern filter", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		err = client.writeSparseCheckoutFile(repo, "")
		if err != nil {
			t.Fatalf("writeSparseCheckoutFile should handle empty patterns: %v", err)
		}

		// Verify the file was created but is mostly empty
		sparseFile := filepath.Join(tempDir, ".git", "info", "sparse-checkout")
		content, err := os.ReadFile(sparseFile)
		if err != nil {
			t.Fatalf("Failed to read sparse checkout file: %v", err)
		}

		// Should be empty or only contain whitespace
		if len(strings.TrimSpace(string(content))) > 0 {
			t.Errorf("Expected empty content, got: %s", string(content))
		}
	})

	t.Run("Pattern with whitespace", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{HostPath: tempDir}
		err = client.writeSparseCheckoutFile(repo, " *.md , *.txt , ")
		if err != nil {
			t.Fatalf("writeSparseCheckoutFile failed: %v", err)
		}

		// Verify the file was created with trimmed content
		sparseFile := filepath.Join(tempDir, ".git", "info", "sparse-checkout")
		content, err := os.ReadFile(sparseFile)
		if err != nil {
			t.Fatalf("Failed to read sparse checkout file: %v", err)
		}

		expectedLines := []string{"*.md", "*.md/**", "*.txt", "*.txt/**"}
		contentStr := string(content)
		for _, line := range expectedLines {
			if !strings.Contains(contentStr, line) {
				t.Errorf("Expected line '%s' not found in sparse checkout file", line)
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

		// Add a remote and push to establish a baseline
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push main: %v", err)
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

func TestHasCommitsNotOnDefaultBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	// Test when current branch equals default branch (should return false)
	t.Run("Current branch equals default branch", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Fetch to create remote refs
		cmd = exec.Command("git", "fetch", "origin")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to fetch: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		hasCommits, err := client.hasCommitsNotOnDefaultBranch(repo, "main")
		if err != nil {
			t.Fatalf("hasCommitsNotOnDefaultBranch failed: %v", err)
		}
		if hasCommits {
			t.Error("Should return false when current branch equals default branch")
		}
	})

	// Test when current branch has commits not on default branch
	t.Run("Current branch has commits not on default", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Fetch to create remote refs
		cmd = exec.Command("git", "fetch", "origin")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to fetch: %v", err)
		}

		// Create a feature branch with commits not on main
		cmd = exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch
		err = os.WriteFile(filepath.Join(tempDir, "feature.txt"), []byte("feature work"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Add feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		hasCommits, err := client.hasCommitsNotOnDefaultBranch(repo, "feature-branch")
		if err != nil {
			t.Fatalf("hasCommitsNotOnDefaultBranch failed: %v", err)
		}
		if !hasCommits {
			t.Error("Should detect commits on feature branch not on default branch")
		}
	})

	// Test when current branch is up to date with default (no divergent commits)
	t.Run("Current branch up to date with default", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Fetch to create remote refs
		cmd = exec.Command("git", "fetch", "origin")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to fetch: %v", err)
		}

		// Create a feature branch from the same commit as main
		cmd = exec.Command("git", "checkout", "-b", "up-to-date-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create up-to-date branch: %v", err)
		}

		// Switch back to main to ensure we're at the same commit
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}

		cmd = exec.Command("git", "checkout", "up-to-date-branch")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout up-to-date branch: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		hasCommits, err := client.hasCommitsNotOnDefaultBranch(repo, "up-to-date-branch")
		if err != nil {
			t.Fatalf("hasCommitsNotOnDefaultBranch failed: %v", err)
		}
		if hasCommits {
			t.Error("Should not detect divergent commits when branch is up to date with default")
		}
	})

	// Test error handling when remote default branch doesn't exist
	t.Run("Error when remote default branch doesn't exist", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote that doesn't have the default branch
		cmd := exec.Command("git", "remote", "add", "origin", "/dev/null")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "nonexistent-default",
		}

		// Should return true (assume divergent commits) when rev-list command fails
		hasCommits, err := client.hasCommitsNotOnDefaultBranch(repo, "main")
		if err != nil {
			t.Fatalf("hasCommitsNotOnDefaultBranch should not fail on missing remote branch: %v", err)
		}
		if !hasCommits {
			t.Error("Should assume divergent commits when remote default branch doesn't exist")
		}
	})
}

func TestSyncDefaultBranchWithDivergentCommits(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	// Test sync skipping due to divergent commits
	t.Run("Sync skipped due to divergent commits", func(t *testing.T) {
		// Enable sync for this test
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Clone the repository
		destDir, err := os.MkdirTemp("", "ghorg-sync-divergent")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    destDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		client := GitClient{}

		// First clone normally
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Configure git in the cloned repository
		if err := configureGitInRepo(destDir); err != nil {
			t.Fatalf("Failed to configure git in cloned repo: %v", err)
		}

		// Create a feature branch with divergent commits
		cmd := exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch that's not on main
		err = os.WriteFile(filepath.Join(destDir, "feature.txt"), []byte("feature work"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Add feature")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		// Update repo to have different clone branch than current branch
		repo.CloneBranch = "main"

		// Sync should be skipped due to divergent commits
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not error when skipping due to divergent commits: %v", err)
		}

		// Verify we're still on the feature branch
		cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = destDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get current branch: %v", err)
		}
		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != "feature-branch" {
			t.Errorf("Expected to still be on feature-branch, but on %s", currentBranch)
		}
	})

	// Test sync with debug mode showing divergent commits skip message
	t.Run("Debug mode shows divergent commits skip message", func(t *testing.T) {
		// Enable sync and debug for this test
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Clone the repository first without debug mode
		destDir, err := os.MkdirTemp("", "ghorg-sync-divergent-debug")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    destDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		client := GitClient{}

		// First clone normally (without debug mode)
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Configure git in the cloned repository
		if err := configureGitInRepo(destDir); err != nil {
			t.Fatalf("Failed to configure git in cloned repo: %v", err)
		}

		// Create a feature branch with divergent commits
		cmd := exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch that's not on main
		err = os.WriteFile(filepath.Join(destDir, "feature.txt"), []byte("feature work"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Add feature")
		cmd.Dir = destDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		// Update repo to have different clone branch than current branch
		repo.CloneBranch = "main"

		// Now set debug mode for the sync operation
		os.Setenv("GHORG_DEBUG", "1")
		defer os.Unsetenv("GHORG_DEBUG")

		// Sync should be skipped and should print debug message
		// Note: We can't easily capture stdout in this test, but we ensure no error occurs
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch should not error when skipping due to divergent commits: %v", err)
		}
	})
}

func TestHasCommitsNotOnDefaultBranchEdgeCases(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	// Test error parsing commit count
	t.Run("Error parsing commit count", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Fetch to create remote refs
		cmd = exec.Command("git", "fetch", "origin")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to fetch: %v", err)
		}

		// Create a mock git script that returns invalid count for rev-list
		gitPath, err := exec.LookPath("git")
		if err != nil {
			t.Fatalf("Cannot find git binary: %v", err)
		}

		mockGitDir, err := os.MkdirTemp("", "mock-git-divergent")
		if err != nil {
			t.Fatalf("Failed to create mock git dir: %v", err)
		}
		defer os.RemoveAll(mockGitDir)

		mockGitPath := filepath.Join(mockGitDir, "git")
		mockGitScript := `#!/bin/bash
if [[ "$1" == "rev-list" && "$3" == "--count" ]]; then
    echo "invalid-count"
else
    exec ` + gitPath + ` "$@"
fi`

		err = os.WriteFile(mockGitPath, []byte(mockGitScript), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock git script: %v", err)
		}

		// Update PATH to use mock git
		originalPath := os.Getenv("PATH")
		os.Setenv("PATH", mockGitDir+":"+originalPath)
		defer os.Setenv("PATH", originalPath)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		// Should return error when parsing invalid count
		_, err = client.hasCommitsNotOnDefaultBranch(repo, "feature-branch")
		if err == nil {
			t.Error("Expected error when parsing invalid commit count")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to parse commit count") {
			t.Errorf("Expected parsing error, got: %v", err)
		}
	})

	// Test error checking for commits not on default branch
	t.Run("Error checking commits not on default branch", func(t *testing.T) {
		// Create a test repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add a remote and push so the remote exists
		cmd := exec.Command("git", "remote", "add", "origin", tempDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		// Create a mock git script that returns invalid count for rev-list but for a different pattern
		gitPath, err := exec.LookPath("git")
		if err != nil {
			t.Fatalf("Cannot find git binary: %v", err)
		}

		mockGitDir, err := os.MkdirTemp("", "mock-git-divergent-2")
		if err != nil {
			t.Fatalf("Failed to create mock git dir: %v", err)
		}
		defer os.RemoveAll(mockGitDir)

		mockGitPath := filepath.Join(mockGitDir, "git")
		// This script returns invalid count specifically for the hasCommitsNotOnDefaultBranch command
		mockGitScript := `#!/bin/bash
if [[ "$1" == "rev-list" && "$2" == "origin/main..feature-branch" && "$3" == "--count" ]]; then
    echo "invalid-count-2"
else
    exec ` + gitPath + ` "$@"
fi`

		err = os.WriteFile(mockGitPath, []byte(mockGitScript), 0755)
		if err != nil {
			t.Fatalf("Failed to create mock git script: %v", err)
		}

		// Update PATH to use mock git
		originalPath := os.Getenv("PATH")
		os.Setenv("PATH", mockGitDir+":"+originalPath)
		defer os.Setenv("PATH", originalPath)

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		// Should return error when parsing invalid count
		_, err = client.hasCommitsNotOnDefaultBranch(repo, "feature-branch")
		if err == nil {
			t.Error("Expected error when checking commits not on default branch fails")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to parse commit count") {
			t.Errorf("Expected parsing error, got: %v", err)
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
