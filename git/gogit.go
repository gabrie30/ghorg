// This file contains the go-git library implementation of the Gitter interface.
// It provides Git operations using the go-git library instead of the Git CLI.
//
// The GoGitClient implements all Gitter interface methods using go-git,
// allowing for Git operations without requiring the git command line tool
// to be installed on the system.
//
// Usage:
//
//	gogit := git.NewGoGit()
//	err := gogit.Clone(repo, false) // useGitCLI is ignored for GoGitClient
package git

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GoGitClient implements the Gitter interface using go-git library
type GoGitClient struct{}

// NewGoGit creates a new GoGitClient
func NewGoGit() GoGitClient {
	return GoGitClient{}
}

// HasRemoteHeads implements the Gitter interface using go-git
func (g GoGitClient) HasRemoteHeads(repo scm.Repo) (bool, error) {
	return g.hasRemoteHeadsWithGo(repo)
}

// Clone implements the Gitter interface using go-git
func (g GoGitClient) Clone(repo scm.Repo) error {
	return g.cloneWithGo(repo)
}

// Reset implements the Gitter interface using go-git
func (g GoGitClient) Reset(repo scm.Repo) error {
	return g.resetWithGo(repo)
}

// Pull implements the Gitter interface using go-git
func (g GoGitClient) Pull(repo scm.Repo) error {
	return g.pullWithGo(repo)
}

// SetOrigin implements the Gitter interface using go-git
func (g GoGitClient) SetOrigin(repo scm.Repo) error {
	return g.setOriginWithGo(repo)
}

// SetOriginWithCredentials implements the Gitter interface using go-git
func (g GoGitClient) SetOriginWithCredentials(repo scm.Repo) error {
	return g.setOriginWithCredentialsWithGo(repo)
}

// Clean implements the Gitter interface using go-git
func (g GoGitClient) Clean(repo scm.Repo) error {
	return g.cleanWithGo(repo)
}

// Checkout implements the Gitter interface using go-git
func (g GoGitClient) Checkout(repo scm.Repo) error {
	return g.checkoutWithGo(repo)
}

// RevListCompare implements the Gitter interface using go-git
func (g GoGitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
	return g.revListCompareWithGo(repo, localBranch, remoteBranch)
}

// ShortStatus implements the Gitter interface using go-git
func (g GoGitClient) ShortStatus(repo scm.Repo) (string, error) {
	return g.shortStatusWithGo(repo)
}

// Branch implements the Gitter interface using go-git
func (g GoGitClient) Branch(repo scm.Repo) (string, error) {
	return g.branchWithGo(repo)
}

// UpdateRemote implements the Gitter interface using go-git
func (g GoGitClient) UpdateRemote(repo scm.Repo) error {
	return g.updateRemoteWithGo(repo)
}

// FetchAll implements the Gitter interface using go-git
func (g GoGitClient) FetchAll(repo scm.Repo) error {
	return g.fetchAllWithGo(repo)
}

// FetchCloneBranch implements the Gitter interface using go-git
func (g GoGitClient) FetchCloneBranch(repo scm.Repo) error {
	return g.fetchCloneBranchWithGo(repo)
}

// RepoCommitCount implements the Gitter interface using go-git
func (g GoGitClient) RepoCommitCount(repo scm.Repo) (int, error) {
	return g.repoCommitCountWithGo(repo)
}

// getCloneDepth returns the clone depth from the environment variable GHORG_CLONE_DEPTH
func getCloneDepth() int {
	cloneDepthStr := os.Getenv("GHORG_CLONE_DEPTH")
	if cloneDepthStr != "" {
		if depth, err := strconv.Atoi(cloneDepthStr); err == nil && depth > 0 {
			return depth
		}
	}
	return 1 // Default depth
}

// hasRemoteHeadsWithGo implements HasRemoteHeads using go-git
func (g GoGitClient) hasRemoteHeadsWithGo(repo scm.Repo) (bool, error) {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := r.Remote("origin")
	if err != nil {
		return false, fmt.Errorf("failed to get remote: %w", err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list remote references: %w", err)
	}

	for _, ref := range refs {
		if ref.Name().IsBranch() {
			return true, nil
		}
	}

	return false, nil
}

// cloneWithGo implements the clone operation using go-git library
func (g GoGitClient) cloneWithGo(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"
	cloneDepth := getCloneDepth()
	gitFilter := os.Getenv("GHORG_GIT_FILTER")
	isMirror := os.Getenv("GHORG_BACKUP") == "true"
	singleBranch := os.Getenv("GHORG_SINGLE_BRANCH") == "true"
	branch := repo.CloneBranch

	// Important: go-git doesn't fully support Git's --filter option for partial clones
	// If a filter is specified, we can try to configure it post-clone, but this has limitations
	// For full filter support, consider using the CLI implementation instead
	if gitFilter != "" {
		// Log that filter support is limited in go-git
		fmt.Printf("Warning: git filter '%s' has limited support in go-git implementation\n", gitFilter)
	}

	// Prepare clone options
	cloneOptions := &git.CloneOptions{
		URL:               repo.CloneURL,
		Depth:             cloneDepth,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          nil, // No progress output
	}

	// Set the branch if specified
	if branch != "" {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(branch)
		cloneOptions.SingleBranch = singleBranch
	}

	// Set submodule recursion if enabled
	if recurseSubmodules {
		cloneOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	} else {
		cloneOptions.RecurseSubmodules = 0
	}

	// Set mirror option if enabled
	if isMirror {
		cloneOptions.Mirror = true
	}

	// Perform the clone
	r, err := git.PlainClone(repo.HostPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// If git filter was specified, apply it as best we can to the repository config
	// for future fetch operations. Note: This is a best-effort approximation.
	if gitFilter != "" {
		if err := setGoGitFilterConfig(r, gitFilter); err != nil {
			// Don't fail the clone if filter config fails, just warn
			fmt.Printf("Warning: failed to configure git filter '%s': %v\n", gitFilter, err)
		}
	}

	return nil
}

// setOriginWithCredentialsWithGo implements SetOriginWithCredentials using go-git
func (g GoGitClient) setOriginWithCredentialsWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the repository configuration
	cfg, err := r.Config()
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	// Update the remote URL for "origin"
	cfg.Remotes[git.DefaultRemoteName] = &config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repo.CloneURL},
	}

	// Save the updated configuration
	err = r.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to set repository config: %w", err)
	}

	return nil
}

// setOriginWithGo implements SetOrigin using go-git
func (g GoGitClient) setOriginWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the repository configuration
	cfg, err := r.Config()
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	// Update the remote URL for "origin"
	cfg.Remotes[git.DefaultRemoteName] = &config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repo.URL},
	}

	// Save the updated configuration
	err = r.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to set repository config: %w", err)
	}

	return nil
}

// checkoutWithGo implements Checkout using go-git
func (g GoGitClient) checkoutWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout the specified branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(repo.CloneBranch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch '%s': %w", repo.CloneBranch, err)
	}

	return nil
}

// cleanWithGo implements Clean using go-git
func (g GoGitClient) cleanWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the status of the worktree
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("failed to get worktree status: %w", err)
	}

	// Iterate over the status to find untracked files
	for file, fileStatus := range status {
		if fileStatus.Worktree == git.Untracked {
			// Remove untracked files and directories
			err := os.RemoveAll(fmt.Sprintf("%s/%s", repo.HostPath, file))
			if err != nil {
				return fmt.Errorf("failed to remove untracked file or directory '%s': %w", file, err)
			}
		}
	}

	return nil
}

// updateRemoteWithGo implements UpdateRemote using go-git
func (g GoGitClient) updateRemoteWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Fetch updates for all remotes
	err = r.Fetch(&git.FetchOptions{
		RemoteName: "origin", // Update the default remote
		Force:      true,     // Force fetch to ensure updates
		Tags:       git.AllTags,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	return nil
}

// pullWithGo implements Pull using go-git
func (g GoGitClient) pullWithGo(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: git.DefaultRemoteName,
		Force:      true,
		Depth:      getCloneDepth(),
		RecurseSubmodules: git.SubmoduleRescursivity(func() int {
			if recurseSubmodules {
				return int(git.DefaultSubmoduleRecursionDepth)
			}
			return 0
		}()),
	})

	return err
}

// resetWithGo implements Reset using go-git
func (g GoGitClient) resetWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})

	return err
}

// fetchAllWithGo implements FetchAll using go-git
func (g GoGitClient) fetchAllWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return err
	}

	err = r.Fetch(&git.FetchOptions{
		RemoteName: git.DefaultRemoteName,
		RemoteURL:  repo.URL,
		Depth:      getCloneDepth(),
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

// branchWithGo implements Branch using go-git
func (g GoGitClient) branchWithGo(repo scm.Repo) (string, error) {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the list of references
	refs, err := r.References()
	if err != nil {
		return "", fmt.Errorf("failed to get references: %w", err)
	}

	var branches []string
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Filter for branch references
		if ref.Type() == plumbing.HashReference && strings.HasPrefix(ref.Name().String(), "refs/heads/") {
			branches = append(branches, strings.TrimPrefix(ref.Name().String(), "refs/heads/"))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate over references: %w", err)
	}

	// Join the branch names into a single string
	return strings.Join(branches, "\n"), nil
}

// revListCompareWithGo implements RevListCompare using go-git
func (g GoGitClient) revListCompareWithGo(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", err
	}

	// Get the local branch reference
	localRef, err := r.Reference(plumbing.NewBranchReferenceName(localBranch), true)
	if err != nil {
		return "", fmt.Errorf("failed to get local branch reference: %w", err)
	}

	// Get the remote branch reference
	remoteRef, err := r.Reference(plumbing.NewRemoteReferenceName("origin", remoteBranch), true)
	if err != nil {
		return "", fmt.Errorf("failed to get remote branch reference: %w", err)
	}

	// Get the commit objects for the local and remote branches
	localCommit, err := r.CommitObject(localRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get local commit: %w", err)
	}

	remoteCommit, err := r.CommitObject(remoteRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get remote commit: %w", err)
	}

	// Find the commits in the local branch that are not in the remote branch
	commitIter := object.NewCommitPreorderIter(localCommit, nil, nil)
	var commits []string
	err = commitIter.ForEach(func(c *object.Commit) error {
		isAncestor, err := c.IsAncestor(remoteCommit)
		if err != nil {
			return err
		}
		if !isAncestor {
			commits = append(commits, c.Hash.String())
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate over commits: %w", err)
	}

	return strings.Join(commits, "\n"), nil
}

// fetchCloneBranchWithGo implements FetchCloneBranch using go-git
func (g GoGitClient) fetchCloneBranchWithGo(repo scm.Repo) error {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Prepare fetch options
	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", repo.CloneBranch, repo.CloneBranch))},
		Depth:      getCloneDepth(),
	}

	// Perform the fetch
	err = r.Fetch(fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch branch: %w", err)
	}

	return nil
}

// shortStatusWithGo implements ShortStatus using go-git
func (g GoGitClient) shortStatusWithGo(repo scm.Repo) (string, error) {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the status of the worktree
	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree status: %w", err)
	}

	// Convert the status to a short format string
	var statusLines []string
	for file, fileStatus := range status {
		statusLines = append(statusLines, fmt.Sprintf("%s %s", string(fileStatus.Worktree), file))
	}

	return strings.Join(statusLines, "\n"), nil
}

// repoCommitCountWithGo implements RepoCommitCount using go-git
func (g GoGitClient) repoCommitCountWithGo(repo scm.Repo) (int, error) {
	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the reference for the specified branch
	ref, err := r.Reference(plumbing.NewBranchReferenceName(repo.CloneBranch), true)
	if err != nil {
		return 0, fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get the commit object for the branch
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return 0, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Iterate through the commit history and count the commits
	count := 0
	commitIter := object.NewCommitIterCTime(commit, nil, nil)
	err = commitIter.ForEach(func(*object.Commit) error {
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to iterate over commits: %w", err)
	}

	return count, nil
}

// setGoGitFilterConfig configures the Git repository to use the specified filter for partial clones
func setGoGitFilterConfig(repo *git.Repository, filterSpec string) error {
	// Get the repository configuration
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	// Make sure we have the origin remote
	remote, ok := cfg.Remotes["origin"]
	if !ok || remote == nil {
		return fmt.Errorf("origin remote not found in repository")
	}

	// Note: go-git doesn't have native support for partial clone filters like --filter=blob:none
	// These filters require Git's object filtering capabilities which are not implemented in go-git
	//
	// What Git CLI does with --filter=<spec>:
	// 1. Sets up a partial clone with the specified filter
	// 2. Configures remote.origin.promisor=true
	// 3. Configures remote.origin.partialclonefilter=<spec>
	// 4. Only downloads objects that pass the filter
	//
	// What we can do in go-git (limited approximation):
	// 1. Set the fetch refspecs to be more specific
	// 2. Configure the remote for future operations
	// 3. Add a comment about the limitation

	// Configure fetch refspecs (this doesn't provide the same filtering as Git CLI)
	fetchRefSpec := []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"}
	remote.Fetch = fetchRefSpec

	// Save the updated configuration
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to set repository config for filter: %w", err)
	}

	// Log the limitation
	fmt.Printf("Note: git filter '%s' configured for remote, but go-git has limited partial clone support\n", filterSpec)
	fmt.Printf("For full filter functionality (like blob:none), consider using CLI implementation\n")

	return nil
}
