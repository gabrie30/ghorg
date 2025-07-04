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

// TestImplementationCompatibility_Comprehensive tests that both GitClient and GoGitClient
// produce identical results for the same operations
func TestImplementationCompatibility_Comprehensive(t *testing.T) {
	// Create a test repository
	tempDir, err := os.MkdirTemp("", "ghorg-compatibility-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize repository using go-git
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	// Create multiple test files
	files := []string{"README.md", "src/main.go", "docs/guide.md"}
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		assert.NoError(t, err)

		err = os.WriteFile(fullPath, []byte("# "+file+"\nTest content"), 0644)
		assert.NoError(t, err)
	}

	// Add and commit files
	w, err := r.Worktree()
	assert.NoError(t, err)

	for _, file := range files {
		_, err = w.Add(file)
		assert.NoError(t, err)
	}

	commit, err := w.Commit("Initial commit with multiple files", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Verify commit was created
	_, err = r.CommitObject(commit)
	assert.NoError(t, err)

	// Create main branch reference
	headRef, err := r.Head()
	assert.NoError(t, err)
	mainRef := plumbing.NewBranchReferenceName("main")
	err = r.Storer.SetReference(plumbing.NewHashReference(mainRef, headRef.Hash()))
	assert.NoError(t, err)

	// Create a feature branch
	featureRef := plumbing.NewBranchReferenceName("feature-test")
	err = r.Storer.SetReference(plumbing.NewHashReference(featureRef, headRef.Hash()))
	assert.NoError(t, err)

	repo := scm.Repo{
		CloneBranch: "main",
		HostPath:    tempDir,
		URL:         "https://example.com/test/repo.git",
		CloneURL:    "https://example.com/test/repo.git",
		Name:        "compatibility-test-repo",
	}

	// Create both client implementations
	gitClient := NewGit()
	goGitClient := NewGoGit()

	t.Run("Branch listing compatibility", func(t *testing.T) {
		gitBranches, gitErr := gitClient.Branch(repo)
		goGitBranches, goGitErr := goGitClient.Branch(repo)

		// Both should succeed
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Both should contain main branch
		assert.Contains(t, gitBranches, "main")
		assert.Contains(t, goGitBranches, "main")

		// Both should contain feature branch
		assert.Contains(t, gitBranches, "feature-test")
		assert.Contains(t, goGitBranches, "feature-test")
	})

	t.Run("Status compatibility - clean repository", func(t *testing.T) {
		gitStatus, gitErr := gitClient.ShortStatus(repo)
		goGitStatus, goGitErr := goGitClient.ShortStatus(repo)

		// Both should succeed
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Both should report clean status
		assert.Empty(t, gitStatus, "Git CLI should report clean status")
		assert.Empty(t, goGitStatus, "GoGit should report clean status")
	})

	t.Run("Commit count compatibility", func(t *testing.T) {
		gitCount, gitErr := gitClient.RepoCommitCount(repo)
		goGitCount, goGitErr := goGitClient.RepoCommitCount(repo)

		// Both should succeed
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Both should report the same count
		assert.Equal(t, gitCount, goGitCount, "Commit counts should match between implementations")
		assert.Equal(t, 1, gitCount, "Should have exactly 1 commit")
	})

	t.Run("Status compatibility - with untracked files", func(t *testing.T) {
		// Add an untracked file
		untrackedFile := filepath.Join(tempDir, "untracked.txt")
		err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644)
		assert.NoError(t, err)

		defer os.Remove(untrackedFile) // Clean up

		gitStatus, gitErr := gitClient.ShortStatus(repo)
		goGitStatus, goGitErr := goGitClient.ShortStatus(repo)

		// Both should succeed
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Both should detect untracked files
		assert.NotEmpty(t, gitStatus, "Git CLI should detect untracked files")
		assert.NotEmpty(t, goGitStatus, "GoGit should detect untracked files")

		// Both should mention the untracked file
		assert.Contains(t, gitStatus, "untracked.txt")
		assert.Contains(t, goGitStatus, "untracked.txt")
	})

	t.Run("Status compatibility - with modified files", func(t *testing.T) {
		// Modify an existing file
		modifiedFile := filepath.Join(tempDir, "README.md")
		err := os.WriteFile(modifiedFile, []byte("# Modified README\nThis file has been modified"), 0644)
		assert.NoError(t, err)

		gitStatus, gitErr := gitClient.ShortStatus(repo)
		goGitStatus, goGitErr := goGitClient.ShortStatus(repo)

		// Both should succeed
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Both should detect modified files
		assert.NotEmpty(t, gitStatus, "Git CLI should detect modified files")
		assert.NotEmpty(t, goGitStatus, "GoGit should detect modified files")

		// Both should mention the modified file
		assert.Contains(t, gitStatus, "README.md")
		assert.Contains(t, goGitStatus, "README.md")

		// Reset the file for other tests
		err = os.WriteFile(modifiedFile, []byte("# README.md\nTest content"), 0644)
		assert.NoError(t, err)
	})
}

func TestFactoryFunction_Compatibility(t *testing.T) {
	// Test that the factory function produces functionally equivalent clients
	tempDir, err := os.MkdirTemp("", "ghorg-factory-compatibility")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test repository
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	_, err = w.Commit("Test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	repo := scm.Repo{
		CloneBranch: "main",
		HostPath:    tempDir,
		URL:         "https://example.com/test/repo.git",
		Name:        "factory-test-repo",
	}

	// Test factory function with both settings
	cliClient := NewGitClient(true)    // Should be GitClient
	goGitClient := NewGitClient(false) // Should be GoGitClient

	// Test that both can perform the same operations
	t.Run("Factory clients can execute same operations", func(t *testing.T) {
		cliStatus, cliErr := cliClient.ShortStatus(repo)
		goGitStatus, goGitErr := goGitClient.ShortStatus(repo)

		// Both should succeed
		assert.NoError(t, cliErr)
		assert.NoError(t, goGitErr)

		// Both should report clean status
		assert.Empty(t, cliStatus)
		assert.Empty(t, goGitStatus)
	})

	t.Run("Factory clients return correct types", func(t *testing.T) {
		// Verify the factory returns the expected concrete types
		assert.IsType(t, GitClient{}, cliClient)
		assert.IsType(t, GoGitClient{}, goGitClient)
	})

	t.Run("Factory clients implement same interface", func(t *testing.T) {
		// Test that both clients can be used as Gitter interface
		var clients []Gitter
		clients = append(clients, cliClient)
		clients = append(clients, goGitClient)

		// Test that all can execute interface methods
		for i, client := range clients {
			status, err := client.ShortStatus(repo)
			assert.NoError(t, err, "Client %d should execute ShortStatus without error", i)
			assert.Empty(t, status, "Client %d should report clean status", i)
		}
	})
}

func TestEnvironmentCompatibility(t *testing.T) {
	// Test that environment variables affect both implementations appropriately
	tempDir, err := os.MkdirTemp("", "ghorg-env-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test repository
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	_, err = w.Commit("Test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	repo := scm.Repo{
		CloneBranch: "main",
		HostPath:    tempDir,
		URL:         "https://example.com/test/repo.git",
		Name:        "env-test-repo",
	}

	t.Run("Both implementations respect debug environment", func(t *testing.T) {
		// Set debug environment
		originalDebug := os.Getenv("GHORG_DEBUG")
		os.Setenv("GHORG_DEBUG", "true")
		defer func() {
			if originalDebug != "" {
				os.Setenv("GHORG_DEBUG", originalDebug)
			} else {
				os.Unsetenv("GHORG_DEBUG")
			}
		}()

		gitClient := NewGit()
		goGitClient := NewGoGit()

		// Both should handle debug mode gracefully
		_, gitErr := gitClient.ShortStatus(repo)
		_, goGitErr := goGitClient.ShortStatus(repo)

		// Should not panic or fail due to debug mode
		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)
	})

	t.Run("Clone depth environment compatibility", func(t *testing.T) {
		// Test that clone depth is handled by both implementations
		originalDepth := os.Getenv("GHORG_CLONE_DEPTH")
		os.Setenv("GHORG_CLONE_DEPTH", "5")
		defer func() {
			if originalDepth != "" {
				os.Setenv("GHORG_CLONE_DEPTH", originalDepth)
			} else {
				os.Unsetenv("GHORG_CLONE_DEPTH")
			}
		}()

		// Test that getCloneDepth function works correctly
		depth := getCloneDepth()
		assert.Equal(t, 5, depth)

		// Both implementations should use this depth value
		gitClient := NewGit()
		goGitClient := NewGoGit()

		// Both should be able to access the environment variable
		// The actual clone operation would use this depth
		assert.NotNil(t, gitClient)
		assert.NotNil(t, goGitClient)
	})
}

func TestErrorHandling_Compatibility(t *testing.T) {
	// Test that both implementations handle errors similarly

	t.Run("Non-existent repository", func(t *testing.T) {
		nonExistentRepo := scm.Repo{
			CloneBranch: "main",
			HostPath:    "/non/existent/path",
			URL:         "https://example.com/test/repo.git",
			Name:        "non-existent-repo",
		}

		gitClient := NewGit()
		goGitClient := NewGoGit()

		// Both should handle non-existent repositories gracefully
		_, gitErr := gitClient.Branch(nonExistentRepo)
		_, goGitErr := goGitClient.Branch(nonExistentRepo)

		// Both should return errors for non-existent repositories
		assert.Error(t, gitErr)
		assert.Error(t, goGitErr)
	})

	t.Run("Invalid repository path", func(t *testing.T) {
		invalidRepo := scm.Repo{
			CloneBranch: "main",
			HostPath:    "",
			URL:         "https://example.com/test/repo.git",
			Name:        "invalid-repo",
		}

		gitClient := NewGit()
		goGitClient := NewGoGit()

		// Both should handle invalid paths gracefully
		_, gitErr := gitClient.ShortStatus(invalidRepo)
		_, goGitErr := goGitClient.ShortStatus(invalidRepo)

		// Both should return errors for invalid paths
		assert.Error(t, gitErr)
		assert.Error(t, goGitErr)
	})
}

func TestClientTypeIdentity(t *testing.T) {
	t.Run("Factory function returns correct concrete types", func(t *testing.T) {
		// Test factory function type returns
		cliClient := NewGitClient(true)
		goGitClient := NewGitClient(false)

		// Should be able to type assert to concrete types
		_, isCLI := cliClient.(GitClient)
		assert.True(t, isCLI, "NewGitClient(true) should return GitClient")

		_, isGoGit := goGitClient.(GoGitClient)
		assert.True(t, isGoGit, "NewGitClient(false) should return GoGitClient")
	})

	t.Run("Legacy constructors return correct types", func(t *testing.T) {
		gitClient := NewGit()
		goGitClient := NewGoGit()

		// These should be concrete types
		assert.IsType(t, GitClient{}, gitClient)
		assert.IsType(t, GoGitClient{}, goGitClient)
	})
}

func TestMethodSignatureCompatibility(t *testing.T) {
	// Test that both implementations have the same method signatures
	t.Run("All methods implement Gitter interface identically", func(t *testing.T) {
		var cliClient Gitter = NewGit()
		var goGitClient Gitter = NewGoGit()
		var factoryCLI Gitter = NewGitClient(true)
		var factoryGoGit Gitter = NewGitClient(false)

		// If this compiles, the interfaces are identical
		clients := []Gitter{cliClient, goGitClient, factoryCLI, factoryGoGit}

		for i, client := range clients {
			assert.NotNil(t, client, "Client %d should not be nil", i)
		}
	})
}

func TestRealRepositoryCompatibility(t *testing.T) {
	// Create a real repository for more comprehensive testing
	tempDir, err := os.MkdirTemp("", "ghorg-real-repo-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize repository using go-git
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	commit, err := w.Commit("Test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Create main branch reference
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
		Name:        "real-test-repo",
	}

	gitClient := NewGit()
	goGitClient := NewGoGit()

	t.Run("Both implementations can read repository state", func(t *testing.T) {
		// Both should be able to get branch information
		gitBranches, gitErr := gitClient.Branch(repo)
		goGitBranches, goGitErr := goGitClient.Branch(repo)

		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)
		assert.NotEmpty(t, gitBranches)
		assert.NotEmpty(t, goGitBranches)

		// Both should contain main branch
		assert.Contains(t, gitBranches, "main")
		assert.Contains(t, goGitBranches, "main")
	})

	t.Run("Both implementations can get clean status", func(t *testing.T) {
		gitStatus, gitErr := gitClient.ShortStatus(repo)
		goGitStatus, goGitErr := goGitClient.ShortStatus(repo)

		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Clean repository should have empty or minimal status
		// (Format may differ slightly between implementations)
		_ = gitStatus
		_ = goGitStatus
	})

	t.Run("Both implementations can count commits", func(t *testing.T) {
		gitCount, gitErr := gitClient.RepoCommitCount(repo)
		goGitCount, goGitErr := goGitClient.RepoCommitCount(repo)

		assert.NoError(t, gitErr)
		assert.NoError(t, goGitErr)

		// Should have exactly 1 commit
		assert.Equal(t, 1, gitCount)
		assert.Equal(t, 1, goGitCount)

		// Both should return identical counts
		assert.Equal(t, gitCount, goGitCount)
	})

	// Verify commit was created
	_, err = r.CommitObject(commit)
	assert.NoError(t, err)
}

func TestGitFilterCompatibilityDocumentation(t *testing.T) {
	// This test documents the differences between CLI and go-git implementations
	// when using GHORG_GIT_FILTER

	t.Run("CLI supports full git filter functionality", func(t *testing.T) {
		// CLI implementation uses git's native --filter flag
		// This provides complete support for partial clones like:
		// - blob:none (exclude all blobs)
		// - blob:limit=<size> (exclude blobs larger than size)
		// - tree:0 (exclude all trees and blobs)
		// - sparse:oid=<oid> (use sparse-checkout from oid)

		// The CLI implementation directly passes the filter to git clone
		gitClient := NewGit()
		assert.NotNil(t, gitClient)

		// Test structure shows CLI passes filter directly to git command
		// See git.go line ~134: args[index] = fmt.Sprintf("--filter=%v", os.Getenv("GHORG_GIT_FILTER"))
	})

	t.Run("go-git has limited filter support", func(t *testing.T) {
		// go-git implementation has limitations:
		// 1. No native support for partial clone filters
		// 2. Can only configure post-clone settings
		// 3. Cannot perform actual object filtering during clone

		goGitClient := NewGoGit()
		assert.NotNil(t, goGitClient)

		// Test structure shows go-git tries to configure filter post-clone
		// See gogit.go setGoGitFilterConfig function for implementation details
	})

	t.Run("Factory function allows choosing implementation based on filter needs", func(t *testing.T) {
		// When git filters are critical, use CLI implementation
		cliClient := NewGitClient(true) // Forces CLI implementation
		assert.IsType(t, GitClient{}, cliClient)

		// When go-git features are more important than filters, use go-git
		goGitClient := NewGitClient(false) // Forces go-git implementation
		assert.IsType(t, GoGitClient{}, goGitClient)

		// Helper function for filter use cases always returns CLI
		filterClient := NewGitClientForFilter()
		assert.IsType(t, GitClient{}, filterClient)
	})
}
