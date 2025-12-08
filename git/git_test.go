package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// createTestGitRepo creates a test git repository with a single commit
func createTestGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "ghorg-git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		cleanup()
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Rename to main branch
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to rename branch to main: %v", err)
	}

	return tempDir, cleanup
}

func TestGetRemoteURL(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("No remote configured", func(t *testing.T) {
		_, err := client.GetRemoteURL(repo, "origin")
		if err == nil {
			t.Error("Expected error when no remote is configured")
		}
	})

	t.Run("Remote exists", func(t *testing.T) {
		// Add a remote
		expectedURL := "https://github.com/test/repo.git"
		cmd := exec.Command("git", "remote", "add", "origin", expectedURL)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		url, err := client.GetRemoteURL(repo, "origin")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Git may return URL in different formats (https or ssh)
		// Just verify we got a non-empty URL
		if url == "" {
			t.Error("Expected non-empty URL")
		}

		// Verify it contains the repo identifier
		if !strings.Contains(url, "test/repo") && !strings.Contains(url, "test:repo") {
			t.Errorf("Expected URL to contain 'test/repo' or 'test:repo', got '%s'", url)
		}
	})

	t.Run("Non-existent remote", func(t *testing.T) {
		_, err := client.GetRemoteURL(repo, "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent remote")
		}
	})
}

func TestHasLocalChanges(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Clean working directory", func(t *testing.T) {
		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if hasChanges {
			t.Error("Expected no local changes in clean working directory")
		}
	})

	t.Run("Modified file", func(t *testing.T) {
		// Modify file
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
			t.Fatalf("Failed to modify file: %v", err)
		}

		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !hasChanges {
			t.Error("Expected local changes after modifying file")
		}

		// Clean up
		cmd := exec.Command("git", "checkout", "test.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to restore file: %v", err)
		}
	})

	t.Run("Untracked file", func(t *testing.T) {
		// Create untracked file
		newFile := filepath.Join(repoPath, "untracked.txt")
		if err := os.WriteFile(newFile, []byte("untracked"), 0644); err != nil {
			t.Fatalf("Failed to create untracked file: %v", err)
		}

		hasChanges, err := client.HasLocalChanges(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !hasChanges {
			t.Error("Expected local changes with untracked file")
		}

		// Clean up
		os.Remove(newFile)
	})
}

func TestHasUnpushedCommits(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("No upstream configured", func(t *testing.T) {
		_, err := client.HasUnpushedCommits(repo)
		if err == nil {
			t.Error("Expected error when no upstream is configured")
		}
	})

	t.Run("With upstream and no unpushed commits", func(t *testing.T) {
		// Create bare remote
		bareDir, err := os.MkdirTemp("", "ghorg-bare-")
		if err != nil {
			t.Fatalf("Failed to create bare directory: %v", err)
		}
		defer os.RemoveAll(bareDir)

		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = bareDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create bare repo: %v", err)
		}

		// Add remote and push
		cmd = exec.Command("git", "remote", "add", "origin", bareDir)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add remote: %v", err)
		}

		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to push: %v", err)
		}

		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if hasUnpushed {
			t.Error("Expected no unpushed commits after push")
		}
	})

	t.Run("With unpushed commits", func(t *testing.T) {
		// Create a new commit
		newFile := filepath.Join(repoPath, "new.txt")
		if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
			t.Fatalf("Failed to create new file: %v", err)
		}

		cmd := exec.Command("git", "add", "new.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "New commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		hasUnpushed, err := client.HasUnpushedCommits(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !hasUnpushed {
			t.Error("Expected unpushed commits after local commit")
		}
	})
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Get main branch", func(t *testing.T) {
		branch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("Expected branch 'main', got '%s'", branch)
		}
	})

	t.Run("Get feature branch", func(t *testing.T) {
		// Create and checkout feature branch
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		branch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if branch != "feature" {
			t.Errorf("Expected branch 'feature', got '%s'", branch)
		}

		// Switch back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}
	})
}

func TestHasCommitsNotOnDefaultBranch(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Main branch has no extra commits", func(t *testing.T) {
		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "main")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if hasCommits {
			t.Error("Expected no commits not on default branch")
		}
	})

	t.Run("Feature branch with commits", func(t *testing.T) {
		// Create feature branch
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add commit on feature branch
		newFile := filepath.Join(repoPath, "feature.txt")
		if err := os.WriteFile(newFile, []byte("feature content"), 0644); err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Feature commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		hasCommits, err := client.HasCommitsNotOnDefaultBranch(repo, "feature")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !hasCommits {
			t.Error("Expected commits not on default branch")
		}

		// Switch back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}
	})
}

func TestIsDefaultBranchBehindHead(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Default branch is current", func(t *testing.T) {
		isBehind, err := client.IsDefaultBranchBehindHead(repo, "main")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !isBehind {
			t.Error("Expected default branch to be ancestor of itself (true)")
		}
	})

	t.Run("Default branch behind feature branch", func(t *testing.T) {
		// Create feature branch with commits
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add commit
		newFile := filepath.Join(repoPath, "feature.txt")
		if err := os.WriteFile(newFile, []byte("feature content"), 0644); err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Feature commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "feature")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !isBehind {
			t.Error("Expected default branch to be behind feature branch")
		}

		// Switch back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}
	})

	t.Run("Divergent branches", func(t *testing.T) {
		// Create divergent branch
		cmd := exec.Command("git", "checkout", "-b", "divergent")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create divergent branch: %v", err)
		}

		// Add commit on divergent
		file1 := filepath.Join(repoPath, "divergent.txt")
		if err := os.WriteFile(file1, []byte("divergent content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		cmd = exec.Command("git", "add", "divergent.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Divergent commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		// Switch to main and add different commit
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}

		file2 := filepath.Join(repoPath, "main.txt")
		if err := os.WriteFile(file2, []byte("main content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		cmd = exec.Command("git", "add", "main.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Main commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		isBehind, err := client.IsDefaultBranchBehindHead(repo, "divergent")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if isBehind {
			t.Error("Expected default branch NOT to be behind divergent branch")
		}
	})
}

func TestMergeIntoDefaultBranch(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Fast-forward merge", func(t *testing.T) {
		// Create feature branch
		cmd := exec.Command("git", "checkout", "-b", "feature")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create feature branch: %v", err)
		}

		// Add commit
		newFile := filepath.Join(repoPath, "feature.txt")
		if err := os.WriteFile(newFile, []byte("feature content"), 0644); err != nil {
			t.Fatalf("Failed to create feature file: %v", err)
		}

		cmd = exec.Command("git", "add", "feature.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Feature commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		// Merge into main
		err := client.MergeIntoDefaultBranch(repo, "feature")
		if err != nil {
			t.Errorf("Unexpected error during merge: %v", err)
		}

		// Verify we're on main
		currentBranch, err := client.GetCurrentBranch(repo)
		if err != nil {
			t.Errorf("Failed to get current branch: %v", err)
		}
		if currentBranch != "main" {
			t.Errorf("Expected to be on main branch, got '%s'", currentBranch)
		}

		// Verify file exists on main
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			t.Error("Expected merged file to exist on main branch")
		}
	})

	t.Run("Merge non-existent branch fails", func(t *testing.T) {
		err := client.MergeIntoDefaultBranch(repo, "nonexistent")
		if err == nil {
			t.Error("Expected error when merging non-existent branch")
		}
	})
}

func TestUpdateRef(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()
	repo := scm.Repo{
		HostPath:    repoPath,
		CloneBranch: "main",
	}

	t.Run("Update ref to HEAD", func(t *testing.T) {
		// Create a new branch
		cmd := exec.Command("git", "branch", "test-branch")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create test branch: %v", err)
		}

		// Add a commit on main
		newFile := filepath.Join(repoPath, "update-ref.txt")
		if err := os.WriteFile(newFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		cmd = exec.Command("git", "add", "update-ref.txt")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		cmd = exec.Command("git", "-c", "commit.gpgsign=false", "commit", "-m", "Update ref commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		// Update test-branch ref to point to main
		err := client.UpdateRef(repo, "refs/heads/test-branch", "refs/heads/main")
		if err != nil {
			t.Errorf("Unexpected error updating ref: %v", err)
		}

		// Verify test-branch now points to same commit as main
		cmd = exec.Command("git", "rev-parse", "main")
		cmd.Dir = repoPath
		mainSHA, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get main SHA: %v", err)
		}

		cmd = exec.Command("git", "rev-parse", "test-branch")
		cmd.Dir = repoPath
		testSHA, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get test-branch SHA: %v", err)
		}

		if string(mainSHA) != string(testSHA) {
			t.Error("Expected test-branch to point to same commit as main after UpdateRef")
		}
	})

	t.Run("Update ref with invalid commitRef fails", func(t *testing.T) {
		err := client.UpdateRef(repo, "refs/heads/test", "nonexistent-ref")
		if err == nil {
			t.Error("Expected error when updating ref with invalid commitRef")
		}
	})
}

// TestErrorHandlingEdgeCases tests additional error handling scenarios
func TestErrorHandlingEdgeCases(t *testing.T) {
	repoPath, cleanup := createTestGitRepo(t)
	defer cleanup()

	client := NewGit()

	t.Run("GetRemoteURL with invalid repo path", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    "/nonexistent/path",
			CloneBranch: "main",
		}
		_, err := client.GetRemoteURL(repo, "origin")
		if err == nil {
			t.Error("Expected error with invalid repo path")
		}
	})

	t.Run("GetCurrentBranch with invalid repo path", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    "/nonexistent/path",
			CloneBranch: "main",
		}
		_, err := client.GetCurrentBranch(repo)
		if err == nil {
			t.Error("Expected error with invalid repo path")
		}
	})

	t.Run("HasCommitsNotOnDefaultBranch with invalid branch", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    repoPath,
			CloneBranch: "main",
		}
		_, err := client.HasCommitsNotOnDefaultBranch(repo, "nonexistent-branch")
		if err == nil {
			t.Error("Expected error with nonexistent branch")
		}
	})

	t.Run("IsDefaultBranchBehindHead with invalid branch", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    repoPath,
			CloneBranch: "main",
		}
		_, err := client.IsDefaultBranchBehindHead(repo, "nonexistent-branch")
		if err == nil {
			t.Error("Expected error with nonexistent branch")
		}
	})

	t.Run("MergeIntoDefaultBranch with invalid default branch", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    repoPath,
			CloneBranch: "nonexistent-default",
		}
		err := client.MergeIntoDefaultBranch(repo, "main")
		if err == nil {
			t.Error("Expected error when checking out nonexistent default branch")
		}
	})

	t.Run("UpdateRef with invalid ref name", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    repoPath,
			CloneBranch: "main",
		}
		err := client.UpdateRef(repo, "invalid..ref", "HEAD")
		if err == nil {
			t.Error("Expected error with invalid ref name")
		}
	})
}
