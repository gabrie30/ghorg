// Package git provides Git repository synchronization functionality for ghorg.
//
// For comprehensive documentation on sync functionality, safety philosophy,
// configuration options, and troubleshooting, see README.md in this directory.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/scm"
)

// SyncDefaultBranch synchronizes the local default branch with the remote
// It checks for local changes and unpushed commits before performing the sync
func (g GitClient) SyncDefaultBranch(repo scm.Repo) error {
	// Check if sync is disabled via configuration
	// GHORG_SYNC_DEFAULT_BRANCH defaults to false (sync disabled by default)
	syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	if syncEnabled != "true" {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Skipping sync for %s: GHORG_SYNC_DEFAULT_BRANCH is not set to true\n", repo.Name)
		}
		return nil
	}

	// First check if the remote exists and is accessible
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repo.HostPath
	if err := cmd.Run(); err != nil {
		// Remote doesn't exist or isn't accessible, skip sync
		return nil
	}

	// Check if the working directory has any uncommitted changes
	hasWorkingDirChanges, err := g.hasLocalChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to check working directory status: %w", err)
	}

	// Check if the current branch has unpushed commits
	hasUnpushedCommits, err := g.hasUnpushedCommits(repo)
	if err != nil {
		return fmt.Errorf("failed to check for unpushed commits: %w", err)
	}

	// Check if we're on the correct branch
	currentBranch, err := g.getCurrentBranch(repo)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if current branch has commits not on the default branch (divergent development)
	hasCommitsNotOnDefault, err := g.hasCommitsNotOnDefaultBranch(repo, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to check for commits not on default branch: %w", err)
	}

	// Only sync if:
	// 1. Working directory is clean (no uncommitted changes)
	// 2. No unpushed commits on the current branch
	// 3. No commits on current branch that aren't on the default branch
	// 4. We're on the target branch or can safely switch to it
	if hasWorkingDirChanges {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Skipping sync for %s: working directory has uncommitted changes\n", repo.Name)
		}
		return nil
	}

	if hasUnpushedCommits {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Skipping sync for %s: branch has unpushed commits\n", repo.Name)
		}
		return nil
	}

	if hasCommitsNotOnDefault {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Skipping sync for %s: current branch has commits not on default branch\n", repo.Name)
		}
		return nil
	}

	// Switch to the target branch if we're not already on it
	if currentBranch != repo.CloneBranch {
		err := g.Checkout(repo)
		if err != nil {
			if os.Getenv("GHORG_DEBUG") != "" {
				fmt.Printf("Could not checkout %s for %s: %v\n", repo.CloneBranch, repo.Name, err)
			}
			return nil // Don't fail, just skip sync
		}
	}

	// Fetch the latest changes from the remote
	err = g.FetchCloneBranch(repo)
	if err != nil {
		return fmt.Errorf("failed to fetch default branch: %w", err)
	}

	// Update the local branch reference to match the remote
	args := []string{"update-ref", fmt.Sprintf("refs/heads/%s", repo.CloneBranch), fmt.Sprintf("refs/remotes/origin/%s", repo.CloneBranch)}
	cmd = exec.Command("git", args...)
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update branch reference: %w", err)
	}

	return nil
}

// hasCheckedOutFiles checks if there are any files checked out in the working directory
// Returns false if the working directory is empty (excluding .git directory)
func (g GitClient) hasCheckedOutFiles(repo scm.Repo) (bool, error) {
	entries, err := os.ReadDir(repo.HostPath)
	if err != nil {
		return false, err
	}

	// Check if there are any files/directories other than .git
	for _, entry := range entries {
		if entry.Name() != ".git" {
			return true, nil
		}
	}

	return false, nil
}

// configureSparseCheckout sets up sparse checkout for the repository
func (g GitClient) configureSparseCheckout(repo scm.Repo, pathFilter string) error {
	// Configure git sparse-checkout
	cmd := exec.Command("git", "config", "core.sparseCheckout", "true")
	cmd.Dir = repo.HostPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure sparse checkout: %w", err)
	}

	// Initialize sparse checkout
	cmd = exec.Command("git", "sparse-checkout", "init", "--cone")
	cmd.Dir = repo.HostPath
	if err := cmd.Run(); err != nil {
		// If 'sparse-checkout init' fails (older Git versions), try the alternative approach
		cmd = exec.Command("git", "config", "core.sparseCheckout", "true")
		cmd.Dir = repo.HostPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure sparse checkout: %w", err)
		}
	}

	// Set the sparse checkout patterns
	cmd = exec.Command("git", "sparse-checkout", "set", pathFilter)
	cmd.Dir = repo.HostPath
	if err := cmd.Run(); err != nil {
		// If 'sparse-checkout set' fails (older Git versions), try writing to the sparse-checkout file directly
		return g.writeSparseCheckoutFile(repo, pathFilter)
	}

	return nil
}

// writeSparseCheckoutFile writes sparse checkout patterns directly to the .git/info/sparse-checkout file
func (g GitClient) writeSparseCheckoutFile(repo scm.Repo, pathFilter string) error {
	sparseCheckoutFile := filepath.Join(repo.HostPath, ".git", "info", "sparse-checkout")
	patterns := strings.Split(pathFilter, ",")
	var content strings.Builder

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern != "" {
			content.WriteString(pattern)
			content.WriteString("\n")
			// Also include all files in subdirectories
			content.WriteString(pattern)
			content.WriteString("/**\n")
		}
	}

	if err := os.MkdirAll(filepath.Join(repo.HostPath, ".git", "info"), 0755); err != nil {
		return fmt.Errorf("failed to create sparse checkout directory: %w", err)
	}

	if err := os.WriteFile(sparseCheckoutFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write sparse checkout file: %w", err)
	}

	// Reset the index to apply sparse checkout
	cmd := exec.Command("git", "read-tree", "-mu", "HEAD")
	cmd.Dir = repo.HostPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply sparse checkout: %w", err)
	}

	return nil
}

// hasLocalChanges checks if there are any uncommitted changes in the working directory
// Returns true if there are modified, added, deleted, or untracked files
func (g GitClient) hasLocalChanges(repo scm.Repo) (bool, error) {
	// Use git status --porcelain to check for any changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repo.HostPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// If output is empty, there are no changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// hasUnpushedCommits checks if the current branch has commits that haven't been pushed to the remote
func (g GitClient) hasUnpushedCommits(repo scm.Repo) (bool, error) {
	// Get the current branch name
	currentBranch, err := g.getCurrentBranch(repo)
	if err != nil {
		return false, err
	}

	// Compare local branch with remote branch to see if there are unpushed commits
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("origin/%s..%s", currentBranch, currentBranch), "--count")
	cmd.Dir = repo.HostPath

	output, err := cmd.Output()
	if err != nil {
		// If the command fails, it might be because the remote branch doesn't exist
		// In this case, assume there are unpushed commits to be safe
		return true, nil
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return false, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count > 0, nil
}

// hasCommitsNotOnDefaultBranch checks if the current branch has commits that are not on the default branch
// Returns true if there are commits on the current branch that diverge from the default branch
func (g GitClient) hasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error) {
	// Skip the check if we're already on the default branch
	if currentBranch == repo.CloneBranch {
		return false, nil
	}

	// Compare current branch with default branch to see if there are commits not on default
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("origin/%s..%s", repo.CloneBranch, currentBranch), "--count")
	cmd.Dir = repo.HostPath

	output, err := cmd.Output()
	if err != nil {
		// If the command fails, it might be because the remote default branch doesn't exist
		// In this case, assume there are divergent commits to be safe
		return true, nil
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return false, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count > 0, nil
}

// getCurrentBranch returns the name of the currently checked out branch
func (g GitClient) getCurrentBranch(repo scm.Repo) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repo.HostPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))

	// Handle the case where we're in a detached HEAD state
	if branch == "HEAD" {
		return "", fmt.Errorf("repository is in detached HEAD state")
	}

	return branch, nil
}
