package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test repository for CLI testing
func setupTestRepoForCLI(t *testing.T) (string, scm.Repo) {
	// Create a temporary directory for the test repository
	tempDir, err := os.MkdirTemp("", "ghorg-cli-test")
	assert.NoError(t, err)

	// Initialize a git repository using go-git for test setup
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(testFile, []byte("# Test Repository\nThis is a test repository for CLI testing."), 0644)
	assert.NoError(t, err)

	// Add and commit the file
	w, err := r.Worktree()
	assert.NoError(t, err)

	_, err = w.Add("README.md")
	assert.NoError(t, err)

	commit, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Create a main branch reference
	headRef, err := r.Head()
	assert.NoError(t, err)
	mainRef := plumbing.NewBranchReferenceName("main")
	err = r.Storer.SetReference(plumbing.NewHashReference(mainRef, headRef.Hash()))
	assert.NoError(t, err)

	repo := scm.Repo{
		CloneBranch: "main",
		HostPath:    tempDir,
		URL:         "https://example.com/test/repo.git",
		CloneURL:    "https://example.com/test/repo.git",
		Name:        "test-repo",
	}

	// Set up environment for CLI testing
	os.Setenv("GHORG_USE_GIT_CLI", "true")

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
		os.Unsetenv("GHORG_USE_GIT_CLI")
	})

	// Verify commit was created
	_, err = r.CommitObject(commit)
	assert.NoError(t, err)

	return tempDir, repo
}

func TestGitClient_NewGit(t *testing.T) {
	client := NewGit()
	assert.IsType(t, GitClient{}, client)
}

func TestGitClient_Branch(t *testing.T) {
	_, repo := setupTestRepoForCLI(t)

	client := NewGit()
	branches, err := client.Branch(repo)

	// The branch command might return different formats depending on git version
	// We just verify it doesn't error and returns some content
	assert.NoError(t, err)
	assert.NotEmpty(t, branches)
}

func TestGitClient_ShortStatus(t *testing.T) {
	_, repo := setupTestRepoForCLI(t)

	client := NewGit()
	status, err := client.ShortStatus(repo)

	// Clean repository should have empty status
	assert.NoError(t, err)
	assert.Empty(t, status, "Status should be empty for clean repository")
}

func TestGitClient_RepoCommitCount(t *testing.T) {
	_, repo := setupTestRepoForCLI(t)

	client := NewGit()
	count, err := client.RepoCommitCount(repo)

	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Should have exactly 1 commit")
}

func TestGitClient_HasRemoteHeads(t *testing.T) {
	_, repo := setupTestRepoForCLI(t)

	client := NewGit()

	// This will likely fail since we don't have a real remote
	// but we're testing that the method can be called
	hasHeads, err := client.HasRemoteHeads(repo)

	// We expect this to fail for a local test repo without remote
	// But the method should execute without panicking
	_ = hasHeads
	_ = err
	// Don't assert on the result since it depends on remote availability
}

func TestNewGitClient_Factory(t *testing.T) {
	t.Run("Returns GitClient when useGitCLI is true", func(t *testing.T) {
		client := NewGitClient(true)
		assert.IsType(t, GitClient{}, client)
	})

	t.Run("Returns GoGitClient when useGitCLI is false", func(t *testing.T) {
		client := NewGitClient(false)
		assert.IsType(t, GoGitClient{}, client)
	})
}

func TestImplementationCompatibility(t *testing.T) {
	t.Run("NewGitClient factory returns correct types", func(t *testing.T) {
		// Test that factory function returns the expected types
		cliClient := NewGitClient(true)
		goGitClient := NewGitClient(false)

		// Verify types
		assert.IsType(t, GitClient{}, cliClient)
		assert.IsType(t, GoGitClient{}, goGitClient)

		// Verify both implement Gitter interface
		var _ Gitter = cliClient
		var _ Gitter = goGitClient
	})

	t.Run("Both implementations have identical interface", func(t *testing.T) {
		// This test ensures both implementations satisfy the same interface
		// by attempting to use them interchangeably

		var clients []Gitter
		clients = append(clients, NewGit())
		clients = append(clients, NewGoGit())
		clients = append(clients, NewGitClient(true))
		clients = append(clients, NewGitClient(false))

		// All should be valid Gitter implementations
		assert.Len(t, clients, 4)
		for _, client := range clients {
			assert.NotNil(t, client)
		}
	})
}

func TestGitClient_AllMethodsExist(t *testing.T) {
	// Test that GitClient implements all Gitter interface methods
	var client Gitter = NewGit()
	_, repo := setupTestRepoForCLI(t)

	// Test all interface methods exist and can be called
	// (Results may vary based on setup, but methods should not panic)

	t.Run("HasRemoteHeads method exists", func(t *testing.T) {
		_, err := client.HasRemoteHeads(repo)
		// May error due to no remote, but should not panic
		_ = err
	})

	t.Run("Clone method exists", func(t *testing.T) {
		testRepo := scm.Repo{
			CloneURL:    "https://example.com/test.git",
			HostPath:    "/tmp/nonexistent",
			CloneBranch: "main",
		}
		err := client.Clone(testRepo)
		// Expected to error, but should not panic
		assert.Error(t, err)
	})

	t.Run("Reset method exists", func(t *testing.T) {
		err := client.Reset(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("Pull method exists", func(t *testing.T) {
		err := client.Pull(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("SetOrigin method exists", func(t *testing.T) {
		err := client.SetOrigin(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("SetOriginWithCredentials method exists", func(t *testing.T) {
		err := client.SetOriginWithCredentials(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("Clean method exists", func(t *testing.T) {
		err := client.Clean(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("Checkout method exists", func(t *testing.T) {
		err := client.Checkout(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("RevListCompare method exists", func(t *testing.T) {
		_, err := client.RevListCompare(repo, "main", "origin/main")
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("UpdateRemote method exists", func(t *testing.T) {
		err := client.UpdateRemote(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("FetchAll method exists", func(t *testing.T) {
		err := client.FetchAll(repo)
		// May succeed or fail, but should not panic
		_ = err
	})

	t.Run("FetchCloneBranch method exists", func(t *testing.T) {
		err := client.FetchCloneBranch(repo)
		// May succeed or fail, but should not panic
		_ = err
	})
}

func TestGitClient_InputValidation(t *testing.T) {
	client := NewGit()

	t.Run("Empty repository path validation", func(t *testing.T) {
		emptyRepo := scm.Repo{
			CloneBranch: "main",
			HostPath:    "",
			URL:         "https://example.com/test/repo.git",
		}

		// ShortStatus should validate empty path
		_, err := client.ShortStatus(emptyRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository path cannot be empty")
	})
}
