package git

import (
	"fmt"
	"os"
	"os/exec"
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

	client := GitClient{}

	// Since Clone interacts with the filesystem, you can mock its behavior or test it in an isolated environment.
	err := client.Clone(repo, false)
	assert.NotNil(t, err, "Clone should fail because it requires a real repository")
}

func TestSetOrigin(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		URL: "https://github.com/example/repo.git",
	}

	// Mock the repository's behavior
	err = client.SetOrigin(repo, false)
	assert.Error(t, err, "SetOrigin should fail because it requires a real repository")
}

func TestShortStatus(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		HostPath: "/tmp/repo",
	}

	// Mock the repository's behavior
	status, err := client.ShortStatus(repo, false)
	assert.Error(t, err, "ShortStatus should fail because it requires a real repository")
	assert.Equal(t, "", status, "ShortStatus should return an empty string on failure")
}

func TestRepoCommitCount(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		HostPath:    "/tmp/repo",
		CloneBranch: "main",
	}

	// Mock the repository's behavior
	count, err := client.RepoCommitCount(repo)
	assert.Error(t, err, "RepoCommitCount should fail because it requires a real repository")
	assert.Equal(t, 0, count, "RepoCommitCount should return 0 on failure")
}

func TestRepoCommitCountGoGit(t *testing.T) {
	client := GitClient{}

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
		count, err := client.repoCommitCountGoGit(repo)
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
		count, err := client.repoCommitCountGoGit(repo)
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
		count, err := client.repoCommitCountGoGit(repo)
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
		count, err := client.repoCommitCountGoGit(repo)
		assert.Error(t, err, "Should return an error for non-existent branch")
		assert.Equal(t, 0, count, "Should return 0 for non-existent branch")
	})
}

func TestBranchGoGit(t *testing.T) {
	client := GitClient{}

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
		branch, err := client.Branch(repo, false)
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
		branch, err := client.Branch(repo, false)
		assert.NoError(t, err)
		assert.Contains(t, branch, "feature", "Should detect feature branch")
	})
}

func TestShortStatusGoGit(t *testing.T) {
	client := GitClient{}

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
		status, err := client.ShortStatus(repo, false)
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
		status, err := client.ShortStatus(repo, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, status, "Status should not be empty for modified repo")
		assert.Contains(t, status, "?", "Status should contain ? for untracked files")
	})
}

func TestCloneWithFiltering(t *testing.T) {
	// Create a test repository with multiple files
	tempDir, err := os.MkdirTemp("", "ghorg-source-repo")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize a repository and create multiple files
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create various directories and files
	dirs := []string{
		"src/main/java",
		"src/test/java",
		"docs",
		"configs",
	}

	for _, dir := range dirs {
		err = os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		assert.NoError(t, err)

		// Create a file in each directory
		err = os.WriteFile(filepath.Join(tempDir, dir, "file.txt"),
			[]byte(fmt.Sprintf("Content for %s", dir)), 0644)
		assert.NoError(t, err)

		// Stage the file
		_, err = w.Add(filepath.Join(dir, "file.txt"))
		assert.NoError(t, err)
	}

	// Make an initial commit
	_, err = w.Commit("Initial commit with multiple directories", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)
	// Test filtering with go-git (non-CLI)
	t.Run("Clone with filtering using go-git", func(t *testing.T) {
		// Set the filter environment variable
		originalFilter := os.Getenv("GHORG_PATH_FILTER")
		os.Setenv("GHORG_PATH_FILTER", "src/main")
		defer os.Setenv("GHORG_PATH_FILTER", originalFilter)

		// Set CLI flag to false
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "false")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create a destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-repo")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up the repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Perform the clone with filtering
		client := GitClient{}
		err = client.Clone(repo, false)
		assert.NoError(t, err)

		// Verify that only the selected directories are present
		_, err = os.Stat(filepath.Join(destDir, "src/main"))
		assert.NoError(t, err, "src/main directory should exist")

		// Verify src/main file exists
		_, err = os.Stat(filepath.Join(destDir, "src/main/java/file.txt"))
		assert.NoError(t, err, "file.txt in src/main/java should exist")

		// Verify that dirs not in the filter pattern don't exist
		_, err = os.Stat(filepath.Join(destDir, "docs"))
		assert.Error(t, err, "docs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "configs"))
		assert.Error(t, err, "configs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "src/test"))
		assert.Error(t, err, "src/test directory should not exist due to filtering")
	})
	// Test filtering with git CLI
	t.Run("Clone with filtering using git CLI", func(t *testing.T) {
		// Skip if git CLI is not available
		_, err := exec.LookPath("git")
		if err != nil {
			t.Skip("git CLI not available, skipping test")
		}

		// Set the filter environment variable
		originalFilter := os.Getenv("GHORG_PATH_FILTER")
		os.Setenv("GHORG_PATH_FILTER", "docs")
		defer os.Setenv("GHORG_PATH_FILTER", originalFilter)

		// Set CLI flag to true
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "true")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create a destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-cli-repo")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up the repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Perform the clone with filtering
		client := GitClient{}
		err = client.Clone(repo, true)
		assert.NoError(t, err)

		// Verify that only the selected directories are present
		_, err = os.Stat(filepath.Join(destDir, "docs"))
		assert.NoError(t, err, "docs directory should exist")

		// Verify docs file exists
		_, err = os.Stat(filepath.Join(destDir, "docs/file.txt"))
		assert.NoError(t, err, "file.txt in docs should exist")

		// Verify that dirs not in the filter pattern don't exist
		_, err = os.Stat(filepath.Join(destDir, "src"))
		assert.Error(t, err, "src directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "configs"))
		assert.Error(t, err, "configs directory should not exist due to filtering")
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
		err = client.Clone(repo, true)

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
		destDir, err := os.MkdirTemp("", "ghorg-dest-filter-gogit")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Execute the clone
		client := GitClient{}
		err = client.Clone(repo, false)

		// This should work, even though go-git doesn't support --filter directly
		assert.NoError(t, err)

		// Verify the clone worked by checking for README.md
		_, err = os.Stat(filepath.Join(destDir, "README.md"))
		assert.NoError(t, err, "README.md should exist in the cloned repository")
	})
}

func TestCloneWithBothFilters(t *testing.T) {
	// Create a test repository with multiple files
	tempDir, err := os.MkdirTemp("", "ghorg-both-filters-src")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize a repository
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create various directories and files
	dirs := []string{
		"src/main/java",
		"src/test/java",
		"docs",
		"configs",
	}

	for _, dir := range dirs {
		err = os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		assert.NoError(t, err)

		// Create a file in each directory
		err = os.WriteFile(filepath.Join(tempDir, dir, "file.txt"),
			[]byte(fmt.Sprintf("Content for %s", dir)), 0644)
		assert.NoError(t, err)

		// Stage the file
		_, err = w.Add(filepath.Join(dir, "file.txt"))
		assert.NoError(t, err)
	}

	// Make an initial commit
	_, err = w.Commit("Initial commit with multiple directories", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Test with both filters using CLI
	t.Run("Clone with both filters using CLI", func(t *testing.T) {
		// Skip if git CLI is not available
		_, err := exec.LookPath("git")
		if err != nil {
			t.Skip("git CLI not available, skipping test")
		}

		// Set up environment for both filters
		originalPathFilter := os.Getenv("GHORG_PATH_FILTER")
		originalGitFilter := os.Getenv("GHORG_GIT_FILTER")

		os.Setenv("GHORG_PATH_FILTER", "src/main")
		os.Setenv("GHORG_GIT_FILTER", "blob:none")

		defer func() {
			os.Setenv("GHORG_PATH_FILTER", originalPathFilter)
			os.Setenv("GHORG_GIT_FILTER", originalGitFilter)
		}()

		// Set CLI flag to true
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "true")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-both-filters")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Perform the clone with both filters
		client := GitClient{}
		err = client.Clone(repo, true)
		assert.NoError(t, err)

		// Verify that only the selected directories are present
		_, err = os.Stat(filepath.Join(destDir, "src/main"))
		assert.NoError(t, err, "src/main directory should exist")

		// Verify src/main file exists
		_, err = os.Stat(filepath.Join(destDir, "src/main/java/file.txt"))
		assert.NoError(t, err, "file.txt in src/main/java should exist")

		// Verify that dirs not in the filter pattern don't exist
		_, err = os.Stat(filepath.Join(destDir, "docs"))
		assert.Error(t, err, "docs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "configs"))
		assert.Error(t, err, "configs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "src/test"))
		assert.Error(t, err, "src/test directory should not exist due to filtering")
	})

	// Test with both filters using go-git
	t.Run("Clone with both filters using go-git", func(t *testing.T) {
		// Set up environment for both filters
		originalPathFilter := os.Getenv("GHORG_PATH_FILTER")
		originalGitFilter := os.Getenv("GHORG_GIT_FILTER")

		os.Setenv("GHORG_PATH_FILTER", "src/main")
		os.Setenv("GHORG_GIT_FILTER", "blob:none")

		defer func() {
			os.Setenv("GHORG_PATH_FILTER", originalPathFilter)
			os.Setenv("GHORG_GIT_FILTER", originalGitFilter)
		}()

		// Set CLI flag to false
		originalCliFlag := os.Getenv("GHORG_USE_GIT_CLI")
		os.Setenv("GHORG_USE_GIT_CLI", "false")
		defer os.Setenv("GHORG_USE_GIT_CLI", originalCliFlag)

		// Create destination directory
		destDir, err := os.MkdirTemp("", "ghorg-dest-both-filters-gogit")
		assert.NoError(t, err)
		defer os.RemoveAll(destDir)

		// Set up repo for cloning
		repo := scm.Repo{
			CloneURL: tempDir,
			HostPath: destDir,
		}

		// Perform the clone with both filters
		client := GitClient{}
		err = client.Clone(repo, false)
		assert.NoError(t, err)

		// Verify that only the selected directories are present
		_, err = os.Stat(filepath.Join(destDir, "src/main"))
		assert.NoError(t, err, "src/main directory should exist")

		// Verify src/main file exists
		_, err = os.Stat(filepath.Join(destDir, "src/main/java/file.txt"))
		assert.NoError(t, err, "file.txt in src/main/java should exist")

		// Verify that dirs not in the filter pattern don't exist
		_, err = os.Stat(filepath.Join(destDir, "docs"))
		assert.Error(t, err, "docs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "configs"))
		assert.Error(t, err, "configs directory should not exist due to filtering")

		_, err = os.Stat(filepath.Join(destDir, "src/test"))
		assert.Error(t, err, "src/test directory should not exist due to filtering")
	})
}

// Additional tests for other go-git functions could be added here
