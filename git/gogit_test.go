package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

// Tests for GoGitClient implementation of the Gitter interface

// Mock in-memory repository for testing
func setupMockRepo() (*git.Repository, error) {
	storer := memory.NewStorage()
	return git.Init(storer, nil)
}

// setupTempRepo creates a real git repository on the filesystem with commits
// Don't forget to defer os.RemoveAll(path) in your test
func setupTempRepo(t *testing.T, commitCount int, branchName string) (string, error) {
	// Create a temporary directory for the repository
	path, err := os.MkdirTemp("", "ghorg-test-repo")
	if err != nil {
		return "", err
	}

	// Initialize a new repository
	r, err := git.PlainInit(path, false)
	if err != nil {
		return "", err
	}

	// Create a test file and commit it
	w, err := r.Worktree()
	if err != nil {
		return "", err
	}

	// Create a test file for initial commit
	filename := filepath.Join(path, "README.md")
	err = os.WriteFile(filename, []byte("# Test Repository"), 0644)
	if err != nil {
		return "", err
	}

	// Stage the file
	_, err = w.Add("README.md")
	if err != nil {
		return "", err
	}

	// Commit the file
	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", err
	}

	// Get the current HEAD reference (master branch by default)
	headRef, err := r.Head()
	if err != nil {
		return "", err
	}

	// Determine if we need to create and switch to a different branch
	if branchName == "main" || branchName == "" {
		// If we want "main" but Git creates "master" by default, rename the branch
		mainRef := plumbing.NewBranchReferenceName("main")
		err = r.Storer.SetReference(plumbing.NewHashReference(mainRef, headRef.Hash()))
		if err != nil {
			return "", err
		}

		// Checkout the main branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: mainRef,
			Create: false,
		})
		if err != nil {
			return "", err
		}

		// Branch has been set to "main"
	} else if branchName != "master" {
		// Create a new branch referencing the current HEAD
		refName := plumbing.NewBranchReferenceName(branchName)
		err = r.Storer.SetReference(plumbing.NewHashReference(refName, headRef.Hash()))
		if err != nil {
			return "", err
		}

		// Checkout the new branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: refName,
		})
		if err != nil {
			return "", err
		}
	}

	// Make additional commits
	for i := 1; i < commitCount; i++ {
		// Create a new file for each commit
		filename := filepath.Join(path, fmt.Sprintf("file%d.txt", i))
		err = os.WriteFile(filename, []byte(fmt.Sprintf("File content %d", i)), 0644)
		if err != nil {
			return "", err
		}

		// Stage the file
		_, err = w.Add(fmt.Sprintf("file%d.txt", i))
		if err != nil {
			return "", err
		}

		// Commit the file
		_, err = w.Commit(fmt.Sprintf("Commit %d", i), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now().Add(time.Duration(i) * time.Hour),
			},
		})
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func TestClone(t *testing.T) {
	repo := scm.Repo{
		CloneURL:    "https://github.com/example/repo.git",
		HostPath:    "/tmp/repo",
		CloneBranch: "main",
	}

	client := GoGitClient{}

	// Since Clone interacts with the filesystem, you can mock its behavior or test it in an isolated environment.
	err := client.Clone(repo)
	assert.NotNil(t, err, "Clone should fail because it requires a real repository")
}

func TestSetOrigin(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GoGitClient{}
	repo := scm.Repo{
		URL: "https://github.com/example/repo.git",
	}

	// Mock the repository's behavior
	err = client.SetOrigin(repo)
	assert.Error(t, err, "SetOrigin should fail because it requires a real repository")
}

func TestShortStatus(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GoGitClient{}
	repo := scm.Repo{
		HostPath: "/tmp/repo",
	}

	// Mock the repository's behavior
	status, err := client.ShortStatus(repo)
	assert.Error(t, err, "ShortStatus should fail because it requires a real repository")
	assert.Equal(t, "", status, "ShortStatus should return an empty string on failure")
}

func TestRepoCommitCount(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GoGitClient{}
	repo := scm.Repo{
		HostPath:    "/tmp/repo",
		CloneBranch: "main",
	}

	// Mock the repository's behavior
	count, err := client.RepoCommitCount(repo)
	assert.Error(t, err, "RepoCommitCount should fail because it requires a real repository")
	assert.Equal(t, 0, count, "RepoCommitCount should return 0 on failure")
}

func TestRepoCommitCountWithGo(t *testing.T) {
	goGitClient := GoGitClient{}

	// Test with an empty repository (1 commit)
	t.Run("Empty repository with only initial commit", func(t *testing.T) {
		// Setup a repository with 1 commit
		path, err := setupTempRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "main",
		}

		// Test the go-git implementation directly
		count, err := goGitClient.repoCommitCountWithGo(repo)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "Repository should have 1 commit")
	})

	// Test with multiple commits
	t.Run("Repository with multiple commits", func(t *testing.T) {
		// Setup a repository with 5 commits
		path, err := setupTempRepo(t, 5, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "main",
		}

		// Test the go-git implementation directly
		count, err := goGitClient.repoCommitCountWithGo(repo)
		assert.NoError(t, err)
		assert.Equal(t, 5, count, "Repository should have 5 commits")
	})

	// Test with a different branch
	t.Run("Repository with a different branch", func(t *testing.T) {
		// Setup a repository with 3 commits on a custom branch
		path, err := setupTempRepo(t, 3, "feature")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "feature",
		}

		// Test the go-git implementation directly
		count, err := goGitClient.repoCommitCountWithGo(repo)
		assert.NoError(t, err)
		assert.Equal(t, 3, count, "Repository should have 3 commits on feature branch") // Our setup copies commits to the new branch
	})

	// Test with non-existent branch
	t.Run("Repository with non-existent branch", func(t *testing.T) {
		// Setup a repository with commits on the main branch
		path, err := setupTempRepo(t, 2, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "non-existent",
		}

		// Test the go-git implementation with a non-existent branch
		count, err := goGitClient.repoCommitCountWithGo(repo)
		assert.Error(t, err, "Should return an error for non-existent branch")
		assert.Equal(t, 0, count, "Should return 0 for non-existent branch")
	})
}

func TestBranchWithGo(t *testing.T) {
	goGitClient := GoGitClient{}

	// Test with a main branch
	t.Run("Repository with main branch", func(t *testing.T) {
		// Setup a repository
		path, err := setupTempRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		// We need to set the environment variable to false to use go-git
		prevValue := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "false")
		defer os.Setenv("GHORG_USE_GIT_CLI", prevValue)

		// Test the branch method
		branch, err := goGitClient.Branch(repo)
		assert.NoError(t, err)
		assert.Contains(t, branch, "main", "Should detect main branch")
	})

	// Test with a custom branch
	t.Run("Repository with custom branch", func(t *testing.T) {
		// Setup a repository with a custom branch
		path, err := setupTempRepo(t, 1, "feature")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		// We need to set the environment variable to false to use go-git
		prevValue := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "false")
		defer os.Setenv("GHORG_USE_GIT_CLI", prevValue)

		// Test the branch method
		branch, err := goGitClient.Branch(repo)
		assert.NoError(t, err)
		assert.Contains(t, branch, "feature", "Should detect feature branch")
	})
}

func TestShortStatusWithGo(t *testing.T) {
	goGitClient := GoGitClient{}

	// Test with a clean repository
	t.Run("Clean repository status", func(t *testing.T) {
		// Setup a repository
		path, err := setupTempRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		// Test the ShortStatus method
		status, err := goGitClient.ShortStatus(repo)
		assert.NoError(t, err)
		assert.Empty(t, status, "Status should be empty for clean repo")
	})

	// Test with modified files
	t.Run("Repository with modifications", func(t *testing.T) {
		// Setup a repository
		path, err := setupTempRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		// Add an uncommitted change
		err = os.WriteFile(filepath.Join(path, "new_file.txt"), []byte("new content"), 0644)
		assert.NoError(t, err)

		repo := scm.Repo{
			HostPath: path,
		}

		// Test the ShortStatus method
		status, err := goGitClient.ShortStatus(repo)
		assert.NoError(t, err)
		assert.NotEmpty(t, status, "Status should not be empty for modified repo")
		assert.Contains(t, status, "?", "Status should contain ? for untracked files")
	})
}
func TestCloneWithGitFilter(t *testing.T) {
	// Create a test repository
	tempDir, err := os.MkdirTemp("", "ghorg-test-filter")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize a repository
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create a test file and commit it
	filename := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(filename, []byte("# Test Repository for Git Filter"), 0644)
	assert.NoError(t, err)

	_, err = w.Add("README.md")
	assert.NoError(t, err)

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Test with CLI - this should pass since Git CLI supports --filter
	t.Run("Clone with Git filter using CLI", func(t *testing.T) {
		// Set up environment
		originalFilter := os.Getenv("GHORG_GIT_FILTER")
		os.Setenv("GHORG_GIT_FILTER", "blob:none")
		defer os.Setenv("GHORG_GIT_FILTER", originalFilter)

		// Set CLI flag to true
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "true")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-filter")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Execute the clone
		client := GitClient{}
		err = client.Clone(repo)

		// This should work with the Git CLI which supports --filter
		assert.NoError(t, err)

		// Verify the clone worked by checking for README.md
		_, err = os.Stat(filepath.Join(destDir, "README.md"))
		assert.NoError(t, err, "README.md should exist in the cloned repository")
	})

	// Test with go-git - this should still work but without filter optimization
	t.Run("Clone with Git filter using go-git", func(t *testing.T) {
		// Set up environment
		originalFilter := os.Getenv("GHORG_GIT_FILTER")
		os.Setenv("GHORG_GIT_FILTER", "blob:none")
		defer os.Setenv("GHORG_GIT_FILTER", originalFilter)

		// Set CLI flag to false to force go-git usage
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "false")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-filter-withgo")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Execute the clone using GoGitClient
		goGitClient := GoGitClient{}
		err = goGitClient.Clone(repo)

		// This should work, even though go-git doesn't support --filter directly
		assert.NoError(t, err)

		// Verify the clone worked by checking for README.md
		_, err = os.Stat(filepath.Join(destDir, "README.md"))
		assert.NoError(t, err, "README.md should exist in the cloned repository")
	})
}

func TestCloneWithBothFilters(t *testing.T) {
	// Skip this entire test as we only support GHORG_GIT_FILTER, not GHORG_PATH_FILTER
	t.Skip("GHORG_PATH_FILTER is not supported - only GHORG_GIT_FILTER is supported")
}

// Additional tests for other go-git functions could be added here

func TestGoGitClient_AllMethodsExist(t *testing.T) {
	// Test that GoGitClient implements all Gitter interface methods
	var client Gitter = NewGoGit()

	// Create a basic test repo that might not work but shouldn't panic
	testRepo := scm.Repo{
		CloneBranch: "main",
		HostPath:    "/tmp/nonexistent-gogit-test",
		URL:         "https://example.com/test/repo.git",
		CloneURL:    "https://example.com/test/repo.git",
	}

	// Test all interface methods exist and can be called
	// (Results may vary based on setup, but methods should not panic)

	t.Run("HasRemoteHeads method exists", func(t *testing.T) {
		_, err := client.HasRemoteHeads(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("Clone method exists", func(t *testing.T) {
		err := client.Clone(testRepo)
		// Expected to error, but should not panic
		assert.Error(t, err)
	})

	t.Run("Reset method exists", func(t *testing.T) {
		err := client.Reset(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("Pull method exists", func(t *testing.T) {
		err := client.Pull(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("SetOrigin method exists", func(t *testing.T) {
		err := client.SetOrigin(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("SetOriginWithCredentials method exists", func(t *testing.T) {
		err := client.SetOriginWithCredentials(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("Clean method exists", func(t *testing.T) {
		err := client.Clean(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("Checkout method exists", func(t *testing.T) {
		err := client.Checkout(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("RevListCompare method exists", func(t *testing.T) {
		_, err := client.RevListCompare(testRepo, "main", "origin/main")
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("ShortStatus method exists", func(t *testing.T) {
		_, err := client.ShortStatus(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("Branch method exists", func(t *testing.T) {
		_, err := client.Branch(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("UpdateRemote method exists", func(t *testing.T) {
		err := client.UpdateRemote(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("FetchAll method exists", func(t *testing.T) {
		err := client.FetchAll(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("FetchCloneBranch method exists", func(t *testing.T) {
		err := client.FetchCloneBranch(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})

	t.Run("RepoCommitCount method exists", func(t *testing.T) {
		_, err := client.RepoCommitCount(testRepo)
		// Expected to error for non-existent repo, but should not panic
		assert.Error(t, err)
	})
}
