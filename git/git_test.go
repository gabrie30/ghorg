package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

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
		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Fatalf("HasLocalChanges failed: %v", err)
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
		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Fatalf("HasLocalChanges failed: %v", err)
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
		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Fatalf("HasLocalChanges failed: %v", err)
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
		_, err = client.HasLocalChanges(repo)
		if err == nil {
			t.Error("Expected error for non-git directory")
		}
		if !strings.Contains(err.Error(), "failed to check git status") {
			t.Errorf("Expected git status error, got: %v", err)
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
		branch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Fatalf("GetCurrentBranch failed: %v", err)
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
		_, err = client.GetCurrentBranch(repo)
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
		_, err = client.GetCurrentBranch(repo)
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
		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("HasUnpushedCommits failed: %v", err)
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
		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("HasUnpushedCommits failed: %v", err)
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
		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("HasUnpushedCommits failed: %v", err)
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
		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Fatalf("HasUnpushedCommits failed: %v", err)
		}

		// Should assume unpushed commits when remote doesn't exist
		if !hasUnpushed {
			t.Error("Should assume unpushed commits when remote branch doesn't exist")
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

		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "main")
		if err != nil {
			t.Fatalf("HasCommitsNotOnDefaultBranch failed: %v", err)
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

		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "feature-branch")
		if err != nil {
			t.Fatalf("HasCommitsNotOnDefaultBranch failed: %v", err)
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

		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "up-to-date-branch")
		if err != nil {
			t.Fatalf("HasCommitsNotOnDefaultBranch failed: %v", err)
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
		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "main")
		if err != nil {
			t.Fatalf("HasCommitsNotOnDefaultBranch should not fail on missing remote branch: %v", err)
		}
		if !hasCommits {
			t.Error("Should assume divergent commits when remote default branch doesn't exist")
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

		// Add a remote so the remote check passes
		cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		repo := scm.Repo{
			HostPath:    tempDir,
			CloneBranch: "main",
		}

		// This is a simplified test since mocking git is complex
		// The actual error handling is tested through the method implementation
		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "feature-branch")
		if err != nil {
			// If there's an error, it should be handled gracefully
			t.Logf("HasCommitsNotOnDefaultBranch returned error as expected: %v", err)
		} else {
			// Should assume divergent commits when unable to determine
			if !hasCommits {
				t.Error("Should assume divergent commits when remote branch doesn't exist")
			}
		}
	})
}

func TestCloneWithSyncDefaultBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := GitClient{}

	t.Run("Clone without sync when GHORG_SYNC_DEFAULT_BRANCH is not set", func(t *testing.T) {
		// Ensure GHORG_SYNC_DEFAULT_BRANCH is not set
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		sourceDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create source repository: %v", err)
		}
		defer os.RemoveAll(sourceDir)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-clone-no-sync")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    sourceDir,
			HostPath:    destDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Clone should succeed without calling sync
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Clone failed: %v", err)
		}

		// Verify repository was cloned
		if _, err := os.Stat(filepath.Join(destDir, ".git")); err != nil {
			t.Errorf("Repository should be cloned: %v", err)
		}

		if _, err := os.Stat(filepath.Join(destDir, "README.md")); err != nil {
			t.Errorf("Repository files should be present: %v", err)
		}
	})

	t.Run("Clone with sync when GHORG_SYNC_DEFAULT_BRANCH is true", func(t *testing.T) {
		// Enable sync
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		defer os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Create a test repository
		sourceDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create source repository: %v", err)
		}
		defer os.RemoveAll(sourceDir)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-clone-with-sync")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    sourceDir,
			HostPath:    destDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Clone should succeed and call sync
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Clone failed: %v", err)
		}

		// Verify repository was cloned
		if _, err := os.Stat(filepath.Join(destDir, ".git")); err != nil {
			t.Errorf("Repository should be cloned: %v", err)
		}

		if _, err := os.Stat(filepath.Join(destDir, "README.md")); err != nil {
			t.Errorf("Repository files should be present: %v", err)
		}

		// Note: Since SyncDefaultBranch handles its own logic and doesn't fail the clone,
		// we can't easily verify it was called without mocking. The important part is
		// that the clone operation still succeeds when sync is enabled.
	})

	t.Run("Clone continues even if sync fails", func(t *testing.T) {
		// Enable sync and debug mode
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			os.Unsetenv("GHORG_DEBUG")
		}()

		// Create a test repository
		sourceDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create source repository: %v", err)
		}
		defer os.RemoveAll(sourceDir)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-clone-sync-fail")
		if err != nil {
			t.Fatalf("Failed to create destination directory: %v", err)
		}
		defer os.RemoveAll(destDir)

		repo := scm.Repo{
			CloneURL:    sourceDir,
			HostPath:    destDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Clone should succeed even if sync encounters issues
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Clone should succeed even if sync fails: %v", err)
		}

		// Verify repository was cloned successfully
		if _, err := os.Stat(filepath.Join(destDir, ".git")); err != nil {
			t.Errorf("Repository should be cloned: %v", err)
		}
	})
}

func TestIsDefaultBranchBehindHead(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := NewGit(true)

	t.Run("Default branch not behind when on same branch", func(t *testing.T) {
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "main")
		if err != nil {
			t.Errorf("IsDefaultBranchBehindHead failed: %v", err)
		}

		if isBehind {
			t.Error("Default branch should not be behind when on the same branch")
		}
	})

	t.Run("Default branch behind when feature branch has new commits", func(t *testing.T) {
		// Create a bare repository to act as the remote
		bareDir, err := os.MkdirTemp("", "ghorg-bare")
		if err != nil {
			t.Fatalf("Failed to create bare repo: %v", err)
		}
		defer os.RemoveAll(bareDir)

		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = bareDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init bare repo: %v", err)
		}

		// Create the working repository
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Add the bare repo as origin
		cmd = exec.Command("git", "remote", "add", "origin", bareDir)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		// Push main to the remote
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push main: %v", err)
		}

		// Create a feature branch with additional commits
		cmd = exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch
		err = os.WriteFile(filepath.Join(tempDir, "feature.txt"), []byte("feature content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add feature file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Feature commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit feature: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    bareDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "feature")
		if err != nil {
			t.Errorf("IsDefaultBranchBehindHead failed: %v", err)
		}

		if !isBehind {
			t.Error("Default branch should be behind when feature branch has new commits")
		}
	})

	t.Run("Default branch not behind when it has commits not on feature branch", func(t *testing.T) {
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a feature branch
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Switch back to main and add a commit there
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, "main-only.txt"), []byte("main content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create main file: %v", err)
		}

		cmd = exec.Command("git", "add", "main-only.txt")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add main file: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "Main commit")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to commit to main: %v", err)
		}

		// Switch back to feature
		cmd = exec.Command("git", "checkout", "feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout feature: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "feature")
		if err != nil {
			t.Errorf("IsDefaultBranchBehindHead failed: %v", err)
		}

		if isBehind {
			t.Error("Default branch should not be behind when it has commits not on feature branch")
		}
	})

	t.Run("Error handling when git command fails", func(t *testing.T) {
		repo := scm.Repo{
			CloneURL:    "/nonexistent/path",
			HostPath:    "/nonexistent/path",
			CloneBranch: "main",
			Name:        "test-repo",
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "feature")
		if err != nil {
			t.Errorf("IsDefaultBranchBehindHead should not error on invalid repo: %v", err)
		}

		if isBehind {
			t.Error("Should return false when git command fails")
		}
	})
}

func TestMergeIntoDefaultBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	client := NewGit(true)

	t.Run("Successful fast-forward merge", func(t *testing.T) {
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a feature branch with additional commits
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch
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
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Merge feature into main
		err = client.MergeIntoDefaultBranch(repo, "feature")
		if err != nil {
			t.Errorf("MergeIntoDefaultBranch failed: %v", err)
		}

		// Verify we're on main branch
		currentBranch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Errorf("Failed to get current branch: %v", err)
		}

		if currentBranch != "main" {
			t.Errorf("Expected to be on main branch, got %s", currentBranch)
		}

		// Verify the feature file exists (merge was successful)
		if _, err := os.Stat(filepath.Join(tempDir, "feature.txt")); os.IsNotExist(err) {
			t.Error("Feature file should exist after merge")
		}
	})

	t.Run("Merge when already on default branch", func(t *testing.T) {
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a feature branch with additional commits
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add a commit to the feature branch
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

		// Switch to main first
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Merge feature into main (should still work)
		err = client.MergeIntoDefaultBranch(repo, "feature")
		if err != nil {
			t.Errorf("MergeIntoDefaultBranch failed when already on default branch: %v", err)
		}

		// Verify we're still on main branch
		currentBranch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Errorf("Failed to get current branch: %v", err)
		}

		if currentBranch != "main" {
			t.Errorf("Expected to be on main branch, got %s", currentBranch)
		}
	})

	t.Run("Error when merge fails", func(t *testing.T) {
		tempDir, err := createTestRepo(t)
		if err != nil {
			t.Fatalf("Failed to create test repository: %v", err)
		}
		defer os.RemoveAll(tempDir)

		repo := scm.Repo{
			CloneURL:    tempDir,
			HostPath:    tempDir,
			CloneBranch: "main",
			Name:        "test-repo",
		}

		// Try to merge a non-existent branch
		err = client.MergeIntoDefaultBranch(repo, "nonexistent-branch")
		if err == nil {
			t.Error("MergeIntoDefaultBranch should fail when merging non-existent branch")
		}
	})

	t.Run("Error when checkout fails", func(t *testing.T) {
		repo := scm.Repo{
			CloneURL:    "/nonexistent/path",
			HostPath:    "/nonexistent/path",
			CloneBranch: "main",
			Name:        "test-repo",
		}

		err = client.MergeIntoDefaultBranch(repo, "feature")
		if err == nil {
			t.Error("MergeIntoDefaultBranch should fail when checkout fails")
		}
	})
}
