package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// createTestRepoGoGit creates a test git repository using git CLI for GoGitClient tests
func createTestRepoGoGit(t *testing.T) (string, error) {
	t.Helper()
	
	tempDir, err := os.MkdirTemp("", "gogit-test-repo-")
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

// Test GoGitClient (go-git library) implementation
func TestGoGitClient(t *testing.T) {
	// Skip if git CLI is not available (needed for test setup)
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping tests")
	}

	client := &GoGitClient{}
	
	testHasRemoteHeadsGoGit(t, client)
	testCloneGoGit(t, client)
	testSetOriginGoGit(t, client)
	testBranchGoGit(t, client)
	testShortStatusGoGit(t, client)
	testCheckoutGoGit(t, client)
	testPullGoGit(t, client)
	testResetGoGit(t, client)
	testCleanGoGit(t, client)
	testUpdateRemoteGoGit(t, client)
	testFetchAllGoGit(t, client)
	testFetchCloneBranchGoGit(t, client)
	testRepoCommitCountGoGit(t, client)
	testRevListCompareGoGit(t, client)
}

func testHasRemoteHeadsGoGit(t *testing.T, git *GoGitClient) {
	t.Run("HasRemoteHeads", func(t *testing.T) {
		t.Run("Valid repository", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-hasheads-dest-")
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
			// GoGitClient validates the URL and should return an error
			if err == nil {
				t.Error("Expected error for empty repository URL")
			}
		})
	})
}

func testCloneGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Clone", func(t *testing.T) {
		t.Run("Clone from local path", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-clone-dest-")
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

			if _, err := os.Stat(filepath.Join(clonePath, ".git")); os.IsNotExist(err) {
				t.Error("Expected .git directory to exist")
			}

			if _, err := os.Stat(filepath.Join(clonePath, "README.md")); os.IsNotExist(err) {
				t.Error("Expected README.md to exist")
			}
		})
	})
}

func testSetOriginGoGit(t *testing.T, git *GoGitClient) {
	t.Run("SetOrigin", func(t *testing.T) {
		t.Run("Set origin URL", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
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

func testBranchGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Branch", func(t *testing.T) {
		t.Run("Get current branch", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
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

			if !strings.Contains(branch, "* main") {
				t.Errorf("Expected to be on main branch (with *), got %s", branch)
			}
		})
	})
}

func testShortStatusGoGit(t *testing.T, git *GoGitClient) {
	t.Run("ShortStatus", func(t *testing.T) {
		t.Run("Clean repository", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
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
				t.Errorf("Expected empty status for clean repository, got: %s", status)
			}
		})

		t.Run("Modified file", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Modify a file
			err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("Modified"), 0644)
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
				t.Error("Expected non-empty status for modified repository")
			}
		})
	})
}

func testCheckoutGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Checkout", func(t *testing.T) {
		t.Run("Checkout branch", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
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

			// Checkout back to main
			repo := scm.Repo{
				HostPath:    tempDir,
				CloneBranch: "main",
			}

			err = git.Checkout(repo)
			if err != nil {
				t.Errorf("Checkout failed: %v", err)
			}

			// Verify we're on main branch
			branch, err := git.Branch(repo)
			if err != nil {
				t.Fatalf("Failed to get branch: %v", err)
			}

			if !strings.Contains(branch, "* main") {
				t.Errorf("Expected to be on main branch (with *), got %s", branch)
			}
		})
	})
}

func testPullGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Pull", func(t *testing.T) {
		t.Run("Pull from origin", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-pull-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL:    sourceDir,
				HostPath:    clonePath,
				CloneBranch: "main",
				Name:        "test-repo",
			}

			err = git.Clone(repo)
			if err != nil {
				t.Fatalf("Clone failed: %v", err)
			}

			err = git.Pull(repo)
			if err != nil {
				t.Errorf("Pull failed: %v", err)
			}
		})
	})
}

func testResetGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Reset", func(t *testing.T) {
		t.Run("Reset to HEAD", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			tempDir, err := os.MkdirTemp("", "gogit-reset-test")
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

func testCleanGoGit(t *testing.T, git *GoGitClient) {
	t.Run("Clean", func(t *testing.T) {
		t.Run("Remove untracked files", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create an untracked file
			untrackedFile := filepath.Join(tempDir, "untracked.txt")
			err = os.WriteFile(untrackedFile, []byte("Untracked"), 0644)
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

			// Verify untracked file was removed
			if _, err := os.Stat(untrackedFile); !os.IsNotExist(err) {
				t.Error("Untracked file should have been removed")
			}
		})
	})
}

func testUpdateRemoteGoGit(t *testing.T, git *GoGitClient) {
	t.Run("UpdateRemote", func(t *testing.T) {
		t.Run("Update remote refs", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-update-dest-")
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

func testFetchAllGoGit(t *testing.T, git *GoGitClient) {
	t.Run("FetchAll", func(t *testing.T) {
		t.Run("Fetch all branches", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-fetch-dest-")
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

func testFetchCloneBranchGoGit(t *testing.T, git *GoGitClient) {
	t.Run("FetchCloneBranch", func(t *testing.T) {
		t.Run("Fetch specific branch", func(t *testing.T) {
			sourceDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create source repository: %v", err)
			}
			defer os.RemoveAll(sourceDir)

			destDir, err := os.MkdirTemp("", "gogit-fetchbranch-dest-")
			if err != nil {
				t.Fatalf("Failed to create destination directory: %v", err)
			}
			defer os.RemoveAll(destDir)

			clonePath := filepath.Join(destDir, "test-repo")

			repo := scm.Repo{
				CloneURL:    sourceDir,
				HostPath:    clonePath,
				CloneBranch: "main",
				Name:        "test-repo",
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

func testRepoCommitCountGoGit(t *testing.T, git *GoGitClient) {
	t.Run("RepoCommitCount", func(t *testing.T) {
		t.Run("Count commits", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
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

func testRevListCompareGoGit(t *testing.T, git *GoGitClient) {
	t.Run("RevListCompare", func(t *testing.T) {
		t.Run("Compare branches", func(t *testing.T) {
			tempDir, err := createTestRepoGoGit(t)
			if err != nil {
				t.Fatalf("Failed to create test repository: %v", err)
			}
			defer os.RemoveAll(tempDir)

			sourceDir := tempDir
			destDir, err := os.MkdirTemp("", "gogit-revlist-dest-")
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
			// GoGitClient expects just "main" (adds "origin/" automatically)
			result, err := git.RevListCompare(repo, "main", "main")
			if err != nil {
				t.Errorf("RevListCompare failed: %v", err)
			}

			if result != "" {
				t.Errorf("Expected empty result when comparing branch with itself, got: %s", result)
			}
		})
	})
}
