package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// createTestRepo creates a test git repository using git CLI
func createTestRepo(t *testing.T) (string, error) {
	t.Helper()
	
	tempDir, err := os.MkdirTemp("", "ghorg-test-repo-")
	if err != nil {
		return "", err
	}

	// Initialize git repo
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Create initial file and commit
	readmePath := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

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

// Test GitClient (CLI) implementation
func TestGitClient(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping tests")
	}

	client := &GitClient{}
	
	testHasRemoteHeads(t, client)
	testClone(t, client)
	testSetOrigin(t, client)
	testBranch(t, client)
	testShortStatus(t, client)
	testCheckout(t, client)
	testPull(t, client)
	testReset(t, client)
	testClean(t, client)
	testUpdateRemote(t, client)
	testFetchAll(t, client)
	testFetchCloneBranch(t, client)
	testRepoCommitCount(t, client)
	testRevListCompare(t, client)
}

func testHasRemoteHeads(t *testing.T, git Gitter) {
	t.Run("HasRemoteHeads", func(t *testing.T) {
		t.Run("Valid repository", func(t *testing.T) {
			// Create a source repo and clone it so we have a proper remote
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "ghorg-hasheads-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL: sourceDir,
				HostPath: clonePath,
				Name:     "test-repo",
			}

			// Clone first to setup remote
			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			hasHeads, err := git.HasRemoteHeads(repo)
			if err != nil {
				t.Errorf("HasRemoteHeads failed: %v", err)
			}
			if !hasHeads {
				t.Error("Expected repository to have remote heads")
			}
		})

		t.Run("Empty repository URL", func(t *testing.T) {
			repo := scm.Repo{
				CloneURL: "",
			}

			_, err := git.HasRemoteHeads(repo)
			// GitClient doesn't validate URLs, so it may not return an error
			_ = err
		})
	})
}

func testClone(t *testing.T, git Gitter) {
	t.Run("Clone", func(t *testing.T) {
		t.Run("Clone from local path", func(t *testing.T) {
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "ghorg-clone-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL: sourceDir,
				HostPath: clonePath,
				Name:     "test-repo",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Errorf("Clone failed: %v", err)
			}

			// Verify clone was successful
			if _, err := os.Stat(filepath.Join(clonePath, ".git")); os.IsNotExist(err) {
				t.Error("Clone did not create .git directory")
			}

			if _, err := os.Stat(filepath.Join(clonePath, "README.md")); os.IsNotExist(err) {
				t.Error("Clone did not copy README.md file")
			}
		})
	})
}

func testSetOrigin(t *testing.T, git Gitter) {
	t.Run("SetOrigin", func(t *testing.T) {
		t.Run("Set origin URL", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			newURL := "https://github.com/test/repo.git"
			
			// First add a remote so we have one to change
			cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/old/repo.git")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to add initial remote: %v", err)
			}
			
			repo := scm.Repo{
				URL:      newURL,
				CloneURL: newURL,
				HostPath: tempDir,
			}

			err = git.SetOrigin(repo)
			if err != nil {
				t.Errorf("SetOrigin failed: %v", err)
			}

			// Verify origin was set (using git config)
			cmd = exec.Command("git", "config", "--get", "remote.origin.url")
			cmd.Dir = tempDir
			output, err := cmd.Output()
			if err != nil {
				t.Errorf("Failed to get remote URL: %v", err)
			}

			gotURL := strings.TrimSpace(string(output))
			if gotURL != newURL {
				t.Errorf("Expected origin URL %s, got %s", newURL, gotURL)
			}
		})
	})
}

func testBranch(t *testing.T, git Gitter) {
	t.Run("Branch", func(t *testing.T) {
		t.Run("Get current branch", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			repo := scm.Repo{
				HostPath: tempDir,
			}

			branch, err := git.Branch(repo)
			if err != nil {
				t.Errorf("Branch failed: %v", err)
			}

			// Branch() returns the output of "git branch" which includes "* " before current branch
			if !strings.Contains(branch, "main") {
				t.Errorf("Expected branch output to contain 'main', got '%s'", branch)
			}
		})
	})
}

func testShortStatus(t *testing.T, git Gitter) {
	t.Run("ShortStatus", func(t *testing.T) {
		t.Run("Clean repository", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			repo := scm.Repo{
				HostPath: tempDir,
			}

			status, err := git.ShortStatus(repo)
			if err != nil {
				t.Errorf("ShortStatus failed: %v", err)
			}

			if status != "" {
				t.Errorf("Expected empty status for clean repo, got: %s", status)
			}
		})

		t.Run("Modified file", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Modify README.md
			err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("Modified content\n"), 0644)
			if err != nil {
				t.Fatalf("Failed to modify file: %v", err)
			}

			repo := scm.Repo{
				HostPath: tempDir,
			}

			status, err := git.ShortStatus(repo)
			if err != nil {
				t.Errorf("ShortStatus failed: %v", err)
			}

			if status == "" {
				t.Error("Expected non-empty status for modified repo")
			}
		})
	})
}

func testCheckout(t *testing.T, git Gitter) {
	t.Run("Checkout", func(t *testing.T) {
		t.Run("Checkout branch", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create a new branch using git CLI
			cmd := exec.Command("git", "checkout", "-b", "feature")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to create feature branch: %v", err)
			}

			// Switch back to main
			cmd = exec.Command("git", "checkout", "main")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to checkout main manually: %v", err)
			}

			// Switch to feature
			cmd = exec.Command("git", "checkout", "feature")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to checkout feature manually: %v", err)
			}

			// Now test Checkout to go back to main
			repo := scm.Repo{
				HostPath:    tempDir,
				CloneBranch: "main",
			}

			err = git.Checkout(repo)
			if err != nil {
				t.Errorf("Checkout failed: %v", err)
			}

			// Verify we're on main (Branch returns "* main" format where * indicates current branch)
			branch, err := git.Branch(repo)
			if err != nil {
				t.Errorf("Failed to get branch: %v", err)
			}

			// Check that main has the * (is current) and feature doesn't
			if !strings.Contains(branch, "* main") {
				t.Errorf("Expected to be on main branch (with *), got %s", branch)
			}
		})
	})
}

func testPull(t *testing.T, git Gitter) {
	t.Run("Pull", func(t *testing.T) {
		t.Run("Pull from origin", func(t *testing.T) {
			// Create source repo
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			// Clone it
			destDir, err := os.MkdirTemp("", "ghorg-pull-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL:    sourceDir,
				HostPath:    clonePath,
				Name:        "test-repo",
				CloneBranch: "main",
			}

			// Clone first
			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			// Add a commit to source
			newFile := filepath.Join(sourceDir, "new-file.txt")
			err = os.WriteFile(newFile, []byte("New content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create new file: %v", err)
			}

			cmd := exec.Command("git", "add", "new-file.txt")
			cmd.Dir = sourceDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to add new file: %v", err)
			}

			cmd = exec.Command("git", "commit", "-m", "Add new file")
			cmd.Dir = sourceDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to commit new file: %v", err)
			}

			// Now pull
			err = git.Pull(repo)
			if err != nil {
				t.Errorf("Pull failed: %v", err)
			}

			// Verify new file exists
			if _, err := os.Stat(filepath.Join(clonePath, "new-file.txt")); os.IsNotExist(err) {
				t.Error("Pull did not fetch new file")
			}
		})
	})
}

func testReset(t *testing.T, git Gitter) {
	t.Run("Reset", func(t *testing.T) {
		t.Run("Reset to HEAD", func(t *testing.T) {
			// Create source repo
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			// Clone it so we have origin remote
			tempDir, err := os.MkdirTemp("", "git-reset-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			repo := scm.Repo{
				CloneURL:    "file://" + sourceDir,
				HostPath:    tempDir,
				CloneBranch: "main",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Failed to clone repository: %v", err)
			}

			// Modify a file and stage it
			err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("Modified"), 0644)
			if err != nil {
				t.Fatalf("Failed to modify file: %v", err)
			}

			cmd := exec.Command("git", "add", "README.md")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to stage file: %v", err)
			}

			repo = scm.Repo{
				HostPath:    tempDir,
				CloneBranch: "main",
			}

			err = git.Reset(repo)
			if err != nil {
				t.Errorf("Reset failed: %v", err)
			}

			// Verify file is no longer staged
			cmd = exec.Command("git", "diff", "--cached", "--name-only")
			cmd.Dir = tempDir
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Failed to check staged files: %v", err)
			}

			if len(output) > 0 {
				t.Error("Reset did not unstage files")
			}
		})
	})
}

func testClean(t *testing.T, git Gitter) {
	t.Run("Clean", func(t *testing.T) {
		t.Run("Remove untracked files", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create an untracked file
			untrackedPath := filepath.Join(tempDir, "untracked.txt")
			err = os.WriteFile(untrackedPath, []byte("Untracked"), 0644)
			if err != nil {
				t.Fatalf("Failed to create untracked file: %v", err)
			}

			repo := scm.Repo{
				HostPath: tempDir,
			}

			err = git.Clean(repo)
			if err != nil {
				t.Errorf("Clean failed: %v", err)
			}

			// Verify untracked file is gone
			if _, err := os.Stat(untrackedPath); !os.IsNotExist(err) {
				t.Error("Clean did not remove untracked file")
			}
		})
	})
}

func testUpdateRemote(t *testing.T, git Gitter) {
	t.Run("UpdateRemote", func(t *testing.T) {
		t.Run("Update remote refs", func(t *testing.T) {
			// Create source repo
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			// Clone it
			destDir, err := os.MkdirTemp("", "ghorg-update-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL: sourceDir,
				HostPath: clonePath,
				Name:     "test-repo",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			err = git.UpdateRemote(repo)
			if err != nil {
				t.Errorf("UpdateRemote failed: %v", err)
			}
		})
	})
}

func testFetchAll(t *testing.T, git Gitter) {
	t.Run("FetchAll", func(t *testing.T) {
		t.Run("Fetch all branches", func(t *testing.T) {
			// Create source repo
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			// Create a feature branch
			cmd := exec.Command("git", "checkout", "-b", "feature")
			cmd.Dir = sourceDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to create feature branch: %v", err)
			}

			cmd = exec.Command("git", "checkout", "main")
			cmd.Dir = sourceDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to checkout main: %v", err)
			}

			// Clone it
			destDir, err := os.MkdirTemp("", "ghorg-fetchall-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL: sourceDir,
				HostPath: clonePath,
				Name:     "test-repo",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			err = git.FetchAll(repo)
			if err != nil {
				t.Errorf("FetchAll failed: %v", err)
			}
		})
	})
}

func testFetchCloneBranch(t *testing.T, git Gitter) {
	t.Run("FetchCloneBranch", func(t *testing.T) {
		t.Run("Fetch specific branch", func(t *testing.T) {
			// Create source repo
			sourceDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			// Clone it
			destDir, err := os.MkdirTemp("", "ghorg-fetch-branch-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL:    sourceDir,
				HostPath:    clonePath,
				Name:        "test-repo",
				CloneBranch: "main",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			err = git.FetchCloneBranch(repo)
			if err != nil {
				t.Errorf("FetchCloneBranch failed: %v", err)
			}
		})
	})
}

func testRepoCommitCount(t *testing.T, git Gitter) {
	t.Run("RepoCommitCount", func(t *testing.T) {
		t.Run("Count commits", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Add another commit
			newFile := filepath.Join(tempDir, "file2.txt")
			err = os.WriteFile(newFile, []byte("Content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create new file: %v", err)
			}

			cmd := exec.Command("git", "add", "file2.txt")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to add file: %v", err)
			}

			cmd = exec.Command("git", "commit", "-m", "Second commit")
			cmd.Dir = tempDir
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to commit: %v", err)
			}

			repo := scm.Repo{
				HostPath:    tempDir,
				CloneBranch: "main",
			}

			count, err := git.RepoCommitCount(repo)
			if err != nil {
				t.Errorf("RepoCommitCount failed: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 commits, got %d", count)
			}
		})
	})
}

func testRevListCompare(t *testing.T, git Gitter) {
	t.Run("RevListCompare", func(t *testing.T) {
		t.Run("Compare branches", func(t *testing.T) {
			tempDir, err := createTestRepo(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// For RevListCompare to work properly with gogit, we need a cloned repo with remote
			sourceDir := tempDir
			destDir, err := os.MkdirTemp("", "ghorg-revlist-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL: sourceDir,
				HostPath: clonePath,
				Name:     "test-repo",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			// Compare main with itself (should be empty)
			// GitClient expects "origin/main" format
			result, err := git.RevListCompare(repo, "main", "origin/main")
			if err != nil {
				t.Errorf("RevListCompare failed: %v", err)
			}

			if result != "" {
				t.Errorf("Expected empty result when comparing branch with itself, got: %s", result)
			}
		})
	})
}
