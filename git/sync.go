// Package git provides Git repository synchronization functionality for ghorg.
//
// For comprehensive documentation on sync functionality, safety philosophy,
// configuration options, and troubleshooting, see README.md in this directory.
package git

import (
	"fmt"
	"os"

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
	_, err := g.GetRemoteURL(repo, "origin")
	if err != nil {
		// Remote doesn't exist or isn't accessible, skip sync
		return nil
	}

	// Check if the working directory has any uncommitted changes
	hasWorkingDirChanges, err := g.HasLocalChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to check working directory status: %w", err)
	}

	// Check if the current branch has unpushed commits
	hasUnpushedCommits, err := g.HasUnpushedCommits(repo)
	if err != nil {
		return fmt.Errorf("failed to check for unpushed commits: %w", err)
	}

	// Check if we're on the correct branch
	currentBranch, err := g.GetCurrentBranch(repo)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if current branch has commits not on the default branch (divergent development)
	hasCommitsNotOnDefault, err := g.HasCommitsNotOnDefaultBranch(repo, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to check for commits not on default branch: %w", err)
	}

	// Check if the default branch is behind HEAD (missing commits from current branch)
	isDefaultBehindHead, err := g.IsDefaultBranchBehindHead(repo, currentBranch)
	if err != nil {
		return fmt.Errorf("failed to check if default branch is behind HEAD: %w", err)
	}

	// Only sync if:
	// 1. Working directory is clean (no uncommitted changes)
	// 2. No unpushed commits on the current branch
	// 3. Either:
	//    a. No commits on current branch that aren't on the default branch, OR
	//    b. Default branch is behind HEAD (we can fast-forward merge)
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

	// Allow sync if default branch is behind HEAD (can fast-forward merge)
	// or if there are no commits not on default branch
	if hasCommitsNotOnDefault && !isDefaultBehindHead {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Skipping sync for %s: current branch has commits not on default branch and default is not behind\n", repo.Name)
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

	// If the default branch is behind HEAD and we have commits to merge,
	// perform a fast-forward merge
	if isDefaultBehindHead && hasCommitsNotOnDefault {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Default branch is behind HEAD for %s, performing fast-forward merge\n", repo.Name)
		}

		err = g.MergeIntoDefaultBranch(repo, currentBranch)
		if err != nil {
			return fmt.Errorf("failed to merge into default branch: %w", err)
		}

		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Successfully updated default branch %s by merging %s for %s\n", repo.CloneBranch, currentBranch, repo.Name)
		}
		return nil
	}

	// Update the local branch reference to match the remote
	refName := fmt.Sprintf("refs/heads/%s", repo.CloneBranch)
	commitRef := fmt.Sprintf("refs/remotes/origin/%s", repo.CloneBranch)
	err = g.UpdateRef(repo, refName, commitRef)
	if err != nil {
		return fmt.Errorf("failed to update branch reference: %w", err)
	}

	// Reset the working directory to match the updated branch
	err = g.Reset(repo)
	if err != nil {
		return fmt.Errorf("failed to reset working directory to remote branch: %w", err)
	}

	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Printf("Successfully updated default branch %s for %s\n", repo.CloneBranch, repo.Name)
	}

	return nil
}
