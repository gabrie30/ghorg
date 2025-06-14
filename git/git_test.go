package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestSyncDefaultBranch(t *testing.T) {
	// Skip if git CLI is not available
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git CLI not available, skipping test")
	}

	// Create a test repository
	tempDir, err := createTestRepo(t)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	defer os.RemoveAll(tempDir)
	// Test sync with clean working directory (should sync)
	t.Run("Sync with clean working directory", func(t *testing.T) {
		// Clone the repository
		destDir, err := os.MkdirTemp("", "ghorg-sync-dest")
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
		
		// First clone normally
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// SyncDefaultBranch should work since working directory is clean
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch failed: %v", err)
		}

		// Verify the file still exists
		_, err = os.Stat(filepath.Join(destDir, "README.md"))
		if err != nil {
			t.Errorf("README.md should exist: %v", err)
		}
	})
	// Test sync with local changes (should not sync)
	t.Run("Sync with local changes", func(t *testing.T) {
		// Clone the repository
		destDir, err := os.MkdirTemp("", "ghorg-sync-dest-dirty")
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
		
		// First clone normally
		err = client.Clone(repo)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Make some local changes to make the working directory dirty
		err = os.WriteFile(filepath.Join(destDir, "new-file.txt"), []byte("local changes"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Now sync should NOT work since there are local changes
		err = client.SyncDefaultBranch(repo)
		if err != nil {
			t.Fatalf("SyncDefaultBranch failed: %v", err)
		}

		// Verify that the .git directory still exists
		_, err = os.Stat(filepath.Join(destDir, ".git"))
		if err != nil {
			t.Errorf(".git directory should exist: %v", err)
		}
	})
}

func TestHasCheckedOutFiles(t *testing.T) {
	client := GitClient{}

	// Test with files present
	t.Run("Directory with files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-files-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create .git directory and a regular file
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasFiles, err := client.hasCheckedOutFiles(repo)
		if err != nil {
			t.Fatalf("hasCheckedOutFiles failed: %v", err)
		}
		if !hasFiles {
			t.Error("Should detect files are present")
		}
	})

	// Test with only .git directory
	t.Run("Directory with only .git", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "ghorg-nogitfiles-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create only .git directory
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		repo := scm.Repo{HostPath: tempDir}
		hasFiles, err := client.hasCheckedOutFiles(repo)
		if err != nil {
			t.Fatalf("hasCheckedOutFiles failed: %v", err)
		}
		if hasFiles {
			t.Error("Should detect no files are present")
		}
	})
}

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
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
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
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
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
		hasChanges, err := client.hasLocalChanges(repo)
		if err != nil {
			t.Fatalf("hasLocalChanges failed: %v", err)
		}
		if !hasChanges {
			t.Error("Repository with untracked files should have local changes")
		}
	})
}

// createTestRepo creates a simple test repository using git CLI
func createTestRepo(t *testing.T) (string, error) {
	tempDir, err := os.MkdirTemp("", "ghorg-test-repo")
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

	// Create a test file
	filename := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(filename, []byte("# Test Repository for Sync"), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	// Add and commit the file
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
