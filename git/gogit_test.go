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
	"github.com/stretchr/testify/assert"
)

// setupGoGitTestRepo creates a test repository using go-git for testing
func setupGoGitTestRepo(_ *testing.T, commitCount int, branchName string) (string, *git.Repository, error) {
	// Create a temporary directory for the repository
	path, err := os.MkdirTemp("", "gogit-test-repo")
	if err != nil {
		return "", nil, err
	}

	// Initialize a new repository
	r, err := git.PlainInit(path, false)
	if err != nil {
		return "", nil, err
	}

	// Create a test file and commit it
	w, err := r.Worktree()
	if err != nil {
		return "", nil, err
	}

	// Create a test file for initial commit
	filename := filepath.Join(path, "README.md")
	err = os.WriteFile(filename, []byte("# Test Repository"), 0644)
	if err != nil {
		return "", nil, err
	}

	// Stage the file
	_, err = w.Add("README.md")
	if err != nil {
		return "", nil, err
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
		return "", nil, err
	}

	// Handle branch creation/switching if needed
	if branchName != "" && branchName != "master" {
		// Get the current HEAD reference
		headRef, err := r.Head()
		if err != nil {
			return "", nil, err
		}

		// Create a new branch referencing the current HEAD
		refName := plumbing.NewBranchReferenceName(branchName)
		err = r.Storer.SetReference(plumbing.NewHashReference(refName, headRef.Hash()))
		if err != nil {
			return "", nil, err
		}

		// Checkout the new branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: refName,
		})
		if err != nil {
			return "", nil, err
		}
	}

	// Make additional commits
	for i := 1; i < commitCount; i++ {
		// Create a new file for each commit
		filename := filepath.Join(path, fmt.Sprintf("file%d.txt", i))
		err = os.WriteFile(filename, []byte(fmt.Sprintf("File content %d", i)), 0644)
		if err != nil {
			return "", nil, err
		}

		// Stage the file
		_, err = w.Add(fmt.Sprintf("file%d.txt", i))
		if err != nil {
			return "", nil, err
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
			return "", nil, err
		}
	}

	return path, r, nil
}

func TestNewGit_GoGitImplementation(t *testing.T) {
	client := NewGit(false)
	assert.NotNil(t, client)
	assert.IsType(t, &GoGitClient{}, client)
}

func TestGoGitClient_HasRemoteHeads(t *testing.T) {
	client := NewGit(false)

	t.Run("Repository without remote", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		hasHeads, err := client.HasRemoteHeads(repo)
		assert.Error(t, err)
		assert.False(t, hasHeads)
		assert.Contains(t, err.Error(), "failed to get remote")
	})

	t.Run("Non-existent repository", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		hasHeads, err := client.HasRemoteHeads(repo)
		assert.Error(t, err)
		assert.False(t, hasHeads)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_Clone(t *testing.T) {
	client := NewGit(false)

	t.Run("Basic clone", func(t *testing.T) {
		// Create source repository
		srcPath, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		// Create destination directory
		destPath, err := os.MkdirTemp("", "gogit-clone-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: srcPath,
			HostPath: destPath,
		}

		err = client.Clone(repo)
		assert.NoError(t, err)

		// Verify clone worked
		_, err = os.Stat(filepath.Join(destPath, "README.md"))
		assert.NoError(t, err)
	})

	t.Run("Clone with specific branch", func(t *testing.T) {
		// Create source repository with custom branch
		srcPath, _, err := setupGoGitTestRepo(t, 2, "feature")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-branch-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL:    srcPath,
			HostPath:    destPath,
			CloneBranch: "feature",
		}

		err = client.Clone(repo)
		assert.NoError(t, err)

		// Verify clone worked
		_, err = os.Stat(filepath.Join(destPath, "README.md"))
		assert.NoError(t, err)
	})

	t.Run("Clone with submodules", func(t *testing.T) {
		// Set environment variable
		originalSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES")
		os.Setenv("GHORG_INCLUDE_SUBMODULES", "true")
		defer os.Setenv("GHORG_INCLUDE_SUBMODULES", originalSubmodules)

		srcPath, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-submodules-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: srcPath,
			HostPath: destPath,
		}

		err = client.Clone(repo)
		assert.NoError(t, err)
	})

	t.Run("Clone with depth", func(t *testing.T) {
		originalDepth := os.Getenv("GHORG_CLONE_DEPTH")
		os.Setenv("GHORG_CLONE_DEPTH", "1")
		defer os.Setenv("GHORG_CLONE_DEPTH", originalDepth)

		srcPath, _, err := setupGoGitTestRepo(t, 3, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-depth-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: srcPath,
			HostPath: destPath,
		}

		err = client.Clone(repo)
		assert.NoError(t, err)
	})

	t.Run("Clone with mirror", func(t *testing.T) {
		originalBackup := os.Getenv("GHORG_BACKUP")
		os.Setenv("GHORG_BACKUP", "true")
		defer os.Setenv("GHORG_BACKUP", originalBackup)

		srcPath, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-mirror-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: srcPath,
			HostPath: destPath,
		}

		err = client.Clone(repo)
		assert.NoError(t, err)
	})

	t.Run("Clone with single branch", func(t *testing.T) {
		originalSingleBranch := os.Getenv("GHORG_SINGLE_BRANCH")
		os.Setenv("GHORG_SINGLE_BRANCH", "true")
		defer os.Setenv("GHORG_SINGLE_BRANCH", originalSingleBranch)

		srcPath, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-single-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL:    srcPath,
			HostPath:    destPath,
			CloneBranch: "main",
		}

		err = client.Clone(repo)
		assert.NoError(t, err)
	})

	t.Run("Clone with git filter falls back to CLI", func(t *testing.T) {
		originalFilter := os.Getenv("GHORG_GIT_FILTER")
		os.Setenv("GHORG_GIT_FILTER", "blob:none")
		defer os.Setenv("GHORG_GIT_FILTER", originalFilter)

		srcPath, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(srcPath)

		destPath, err := os.MkdirTemp("", "gogit-clone-filter-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: srcPath,
			HostPath: destPath,
		}

		// When git filter is set, it should fall back to git CLI
		err = client.Clone(repo)
		assert.NoError(t, err)

		// Verify that the repository was cloned successfully
		assert.DirExists(t, destPath)
		// .git can be either a directory or a file (worktree), so just check it exists
		gitPath := filepath.Join(destPath, ".git")
		_, err = os.Stat(gitPath)
		assert.NoError(t, err, "Git repository should be cloned successfully")
	})

	t.Run("Clone invalid repository", func(t *testing.T) {
		destPath, err := os.MkdirTemp("", "gogit-clone-invalid-dest")
		assert.NoError(t, err)
		defer os.RemoveAll(destPath)

		repo := scm.Repo{
			CloneURL: "/non/existent/repo",
			HostPath: destPath,
		}

		err = client.Clone(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to clone repository")
	})
}

func TestGoGitClient_SetOriginWithCredentials(t *testing.T) {
	client := NewGit(false)

	t.Run("Set origin with credentials", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
			CloneURL: "https://user:pass@github.com/example/repo.git",
		}

		err = client.SetOriginWithCredentials(repo)
		assert.NoError(t, err)
	})

	t.Run("Set origin with credentials - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
			CloneURL: "https://user:pass@github.com/example/repo.git",
		}

		err := client.SetOriginWithCredentials(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_SetOrigin(t *testing.T) {
	client := NewGit(false)

	t.Run("Set origin", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
			URL:      "https://github.com/example/repo.git",
		}

		err = client.SetOrigin(repo)
		assert.NoError(t, err)
	})

	t.Run("Set origin - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
			URL:      "https://github.com/example/repo.git",
		}

		err := client.SetOrigin(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_Checkout(t *testing.T) {
	client := NewGit(false)

	t.Run("Checkout existing branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "feature")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "feature",
		}

		err = client.Checkout(repo)
		assert.NoError(t, err)
	})

	t.Run("Checkout non-existent branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "non-existent",
		}

		err = client.Checkout(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to checkout branch")
	})

	t.Run("Checkout - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    "/non/existent/path",
			CloneBranch: "main",
		}

		err := client.Checkout(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_Clean(t *testing.T) {
	client := NewGit(false)

	t.Run("Clean repository with untracked files", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		// Add an untracked file
		untrackedFile := filepath.Join(path, "untracked.txt")
		err = os.WriteFile(untrackedFile, []byte("untracked content"), 0644)
		assert.NoError(t, err)

		repo := scm.Repo{
			HostPath: path,
		}

		err = client.Clean(repo)
		assert.NoError(t, err)

		// Verify untracked file was removed
		_, err = os.Stat(untrackedFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Clean repository - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		err := client.Clean(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_UpdateRemote(t *testing.T) {
	client := NewGit(false)

	t.Run("Update remote - no remote configured", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		err = client.UpdateRemote(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update remote")
	})

	t.Run("Update remote - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		err := client.UpdateRemote(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_Pull(t *testing.T) {
	client := NewGit(false)

	t.Run("Pull with submodules", func(t *testing.T) {
		originalSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES")
		os.Setenv("GHORG_INCLUDE_SUBMODULES", "true")
		defer os.Setenv("GHORG_INCLUDE_SUBMODULES", originalSubmodules)

		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		err = client.Pull(repo)
		assert.Error(t, err) // Expected to fail because no remote is configured
	})

	t.Run("Pull without submodules", func(t *testing.T) {
		originalSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES")
		os.Setenv("GHORG_INCLUDE_SUBMODULES", "false")
		defer os.Setenv("GHORG_INCLUDE_SUBMODULES", originalSubmodules)

		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		err = client.Pull(repo)
		assert.Error(t, err) // Expected to fail because no remote is configured
	})

	t.Run("Pull - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		err := client.Pull(repo)
		assert.Error(t, err)
	})
}

func TestGoGitClient_Reset(t *testing.T) {
	client := NewGit(false)

	t.Run("Reset repository", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		err = client.Reset(repo)
		assert.NoError(t, err)
	})

	t.Run("Reset - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		err := client.Reset(repo)
		assert.Error(t, err)
	})
}

func TestGoGitClient_FetchAll(t *testing.T) {
	client := NewGit(false)

	t.Run("Fetch all - no remote", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
			URL:      "https://github.com/example/repo.git",
		}

		err = client.FetchAll(repo)
		assert.Error(t, err)
	})

	t.Run("Fetch all - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
			URL:      "https://github.com/example/repo.git",
		}

		err := client.FetchAll(repo)
		assert.Error(t, err)
	})
}

func TestGoGitClient_FetchCloneBranch(t *testing.T) {
	client := NewGit(false)

	t.Run("Fetch clone branch - no remote", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "main",
		}

		err = client.FetchCloneBranch(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch branch")
	})

	t.Run("Fetch clone branch - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    "/non/existent/path",
			CloneBranch: "main",
		}

		err := client.FetchCloneBranch(repo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_RepoCommitCount(t *testing.T) {
	client := NewGit(false)

	t.Run("Count commits on main branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 3, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "main",
		}

		count, err := client.RepoCommitCount(repo)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("Count commits on custom branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 2, "feature")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "feature",
		}

		count, err := client.RepoCommitCount(repo)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("Count commits - non-existent branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath:    path,
			CloneBranch: "non-existent",
		}

		count, err := client.RepoCommitCount(repo)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get branch reference")
	})

	t.Run("Count commits - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath:    "/non/existent/path",
			CloneBranch: "main",
		}

		count, err := client.RepoCommitCount(repo)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_Branch(t *testing.T) {
	client := NewGit(false)

	t.Run("List branches", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		branches, err := client.Branch(repo)
		assert.NoError(t, err)
		assert.Contains(t, branches, "main")
	})

	t.Run("List branches - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		branches, err := client.Branch(repo)
		assert.Error(t, err)
		assert.Empty(t, branches)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}

func TestGoGitClient_RevListCompare(t *testing.T) {
	client := NewGit(false)

	t.Run("RevList compare - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		result, err := client.RevListCompare(repo, "main", "origin/main")
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("RevList compare - non-existent local branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		result, err := client.RevListCompare(repo, "non-existent", "main")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to get local branch reference")
	})

	t.Run("RevList compare - non-existent remote branch", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		result, err := client.RevListCompare(repo, "main", "non-existent")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to get remote branch reference")
	})
}

func TestGoGitClient_ShortStatus(t *testing.T) {
	client := NewGit(false)

	t.Run("Short status - clean repo", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		repo := scm.Repo{
			HostPath: path,
		}

		status, err := client.ShortStatus(repo)
		assert.NoError(t, err)
		assert.Empty(t, status)
	})

	t.Run("Short status - with untracked files", func(t *testing.T) {
		path, _, err := setupGoGitTestRepo(t, 1, "main")
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		// Add an untracked file
		err = os.WriteFile(filepath.Join(path, "untracked.txt"), []byte("content"), 0644)
		assert.NoError(t, err)

		repo := scm.Repo{
			HostPath: path,
		}

		status, err := client.ShortStatus(repo)
		assert.NoError(t, err)
		assert.Contains(t, status, "untracked.txt")
		assert.Contains(t, status, "?")
	})

	t.Run("Short status - non-existent repo", func(t *testing.T) {
		repo := scm.Repo{
			HostPath: "/non/existent/path",
		}

		status, err := client.ShortStatus(repo)
		assert.Error(t, err)
		assert.Empty(t, status)
		assert.Contains(t, err.Error(), "failed to open repository")
	})
}
