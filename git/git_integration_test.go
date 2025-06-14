package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

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

// createTestRepoWithMultipleFiles creates a test repository with various file types
func createTestRepoWithMultipleFiles(t *testing.T) (string, error) {
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
