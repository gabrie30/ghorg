package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestNewGitClient(t *testing.T) {
	client := NewGit(true) // true for CLI client
	if client == nil {
		t.Error("NewGit(true) should return a non-nil GitClient")
	}

	// Test that we get the correct type
	if _, ok := client.(*GitClient); !ok {
		t.Error("NewGit(true) should return a *GitClient")
	}
}

func TestNewGit_StrategyPattern(t *testing.T) {
	tests := []struct {
		name         string
		useGitCLI    bool
		expectedType string
	}{
		{
			name:         "CLI strategy returns GitClient",
			useGitCLI:    true,
			expectedType: "*git.GitClient",
		},
		{
			name:         "Go-git strategy returns GoGitClient",
			useGitCLI:    false,
			expectedType: "*git.GoGitClient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGit(tt.useGitCLI)

			if client == nil {
				t.Errorf("NewGit(%t) should return a non-nil client", tt.useGitCLI)
				return
			}

			// Check the type
			clientType := ""
			switch client.(type) {
			case *GitClient:
				clientType = "*git.GitClient"
			case *GoGitClient:
				clientType = "*git.GoGitClient"
			default:
				clientType = "unknown"
			}

			if clientType != tt.expectedType {
				t.Errorf("NewGit(%t) returned %s, expected %s", tt.useGitCLI, clientType, tt.expectedType)
			}
		})
	}
}

func TestNewGit_ImplementsGitterInterface(t *testing.T) {
	tests := []struct {
		name      string
		useGitCLI bool
	}{
		{
			name:      "CLI client implements Gitter",
			useGitCLI: true,
		},
		{
			name:      "Go-git client implements Gitter",
			useGitCLI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGit(tt.useGitCLI)

			// Test that the client implements all Gitter interface methods
			// by calling them with a dummy repo
			dummyRepo := scm.Repo{
				HostPath: "/tmp/nonexistent",
				CloneURL: "https://github.com/nonexistent/repo.git",
			}

			// These calls should not panic, even if they return errors
			_, _ = client.HasRemoteHeads(dummyRepo)
			_ = client.Clone(dummyRepo)
			_ = client.SetOriginWithCredentials(dummyRepo)
			_ = client.SetOrigin(dummyRepo)
			_ = client.Checkout(dummyRepo)
			_ = client.Clean(dummyRepo)
			_ = client.UpdateRemote(dummyRepo)
			_ = client.Pull(dummyRepo)
			_ = client.Reset(dummyRepo)
			_ = client.FetchAll(dummyRepo)
			_ = client.FetchCloneBranch(dummyRepo)
			_, _ = client.RepoCommitCount(dummyRepo)
			_, _ = client.Branch(dummyRepo)
			_, _ = client.RevListCompare(dummyRepo, "main", "origin/main")
			_, _ = client.ShortStatus(dummyRepo)
		})
	}
}

func TestGitClient_Clone(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-clone")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func()
		cleanup func()
		wantErr bool
	}{
		{
			name: "Clone invalid repository",
			repo: scm.Repo{
				CloneURL: "https://github.com/nonexistent/nonexistent.git",
				HostPath: filepath.Join(tmpDir, "nonexistent"),
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := client.Clone(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Clone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_SetOriginWithCredentials(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-origin-creds")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Set origin with credentials - valid repo",
			repo: scm.Repo{
				CloneURL: "https://token@github.com/test/repo.git",
				HostPath: tmpDir,
			},
			setup: func() error {
				// Initialize a git repository with proper configuration
				return setupTestGitRepo(tmpDir)
			},
			wantErr: false,
		},
		{
			name: "Set origin with credentials - non-existent repo",
			repo: scm.Repo{
				CloneURL: "https://token@github.com/test/repo.git",
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.SetOriginWithCredentials(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.SetOriginWithCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_SetOrigin(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-origin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Set origin - valid repo",
			repo: scm.Repo{
				CloneURL: "https://github.com/test/repo.git",
				HostPath: tmpDir,
			},
			setup: func() error {
				// Initialize a git repository
				return setupTestGitRepo(tmpDir)
			},
			wantErr: false,
		},
		{
			name: "Set origin - non-existent repo",
			repo: scm.Repo{
				CloneURL: "https://github.com/test/repo.git",
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.SetOrigin(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.SetOrigin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_Checkout(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-checkout")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Checkout - non-existent repo",
			repo: scm.Repo{
				CloneBranch: "main",
				HostPath:    "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.Checkout(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Checkout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_Clean(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-clean")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Clean - valid repo",
			repo: scm.Repo{
				HostPath: tmpDir,
			},
			setup: func() error {
				// Initialize a git repository and add some untracked files
				if err := setupTestGitRepo(tmpDir); err != nil {
					return err
				}
				// Create an untracked file
				untrackedFile := filepath.Join(tmpDir, "untracked.txt")
				return os.WriteFile(untrackedFile, []byte("untracked content"), 0644)
			},
			wantErr: false,
		},
		{
			name: "Clean - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.Clean(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Clean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_UpdateRemote(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-update-remote")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Update remote - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.UpdateRemote(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.UpdateRemote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_Pull(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-pull")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Pull - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.Pull(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Pull() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_Reset(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-reset")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Reset - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.Reset(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_FetchAll(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-fetch-all")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Fetch all - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.FetchAll(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.FetchAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_FetchCloneBranch(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-fetch-branch")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Fetch clone branch - non-existent repo",
			repo: scm.Repo{
				CloneBranch: "main",
				HostPath:    "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := client.FetchCloneBranch(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.FetchCloneBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_RepoCommitCount(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-commit-count")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name       string
		repo       scm.Repo
		setup      func() error
		wantErr    bool
		wantResult string
	}{
		{
			name: "Commit count - non-existent repo",
			repo: scm.Repo{
				CloneBranch: "main",
				HostPath:    "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			result, err := client.RepoCommitCount(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.RepoCommitCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != 0 {
				// For non-existent repos, we expect an error, not a result
				t.Errorf("GitClient.RepoCommitCount() = %v, want error", result)
			}
		})
	}
}

func TestGitClient_Branch(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-branch")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Branch - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			_, err := client.Branch(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.Branch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_RevListCompare(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-revlist")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "RevList compare - non-existent repo",
			repo: scm.Repo{
				CloneBranch: "main",
				HostPath:    "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			_, err := client.RevListCompare(tt.repo, "main", "origin/main")
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.RevListCompare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_ShortStatus(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "test-cli-status")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		repo    scm.Repo
		setup   func() error
		wantErr bool
	}{
		{
			name: "Short status - valid repo",
			repo: scm.Repo{
				HostPath: tmpDir,
			},
			setup: func() error {
				// Initialize a git repository with proper config
				if err := setupTestGitRepo(tmpDir); err != nil {
					return err
				}
				// Create and commit a file
				testFile := filepath.Join(tmpDir, "test.txt")
				if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
					return err
				}
				if err := runGitCommand(tmpDir, "add", "test.txt"); err != nil {
					return err
				}
				return runGitCommand(tmpDir, "commit", "-m", "Initial commit")
			},
			wantErr: false,
		},
		{
			name: "Short status - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			_, err := client.ShortStatus(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.ShortStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitClient_HasRemoteHeads(t *testing.T) {
	tests := []struct {
		name    string
		repo    scm.Repo
		wantErr bool
	}{
		{
			name: "HasRemoteHeads - non-existent repo",
			repo: scm.Repo{
				HostPath: "/nonexistent/path",
			},
			wantErr: true,
		},
	}

	client := NewGit(true)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.HasRemoteHeads(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitClient.HasRemoteHeads() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to run git commands
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// setupTestGitRepo creates a properly configured git repository for testing
func setupTestGitRepo(dir string) error {
	// Initialize the repository
	if err := runGitCommand(dir, "init"); err != nil {
		return err
	}

	// Configure git user for this repo (required for commits)
	if err := runGitCommand(dir, "config", "user.name", "Test User"); err != nil {
		return err
	}

	if err := runGitCommand(dir, "config", "user.email", "test@example.com"); err != nil {
		return err
	}

	// Disable GPG signing for this repo
	if err := runGitCommand(dir, "config", "commit.gpgsign", "false"); err != nil {
		return err
	}

	return nil
}
