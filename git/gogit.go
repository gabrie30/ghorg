package git

// Go-git implementation with git CLI fallback for filtering
//
// This implementation uses the go-git library for most operations, but falls back
// to the git CLI when filtering is requested since go-git v5 does not support
// native filtering and filter approximation has been removed.
//
// Behavior:
// - When GHORG_GIT_FILTER is set: automatically falls back to git CLI for proper filter support
// - When GHORG_GIT_FILTER is not set: uses go-git for faster cloning without shell dependencies
//
// Future improvements:
// - When go-git v6 becomes stable with native Filter support, this fallback can be removed
// - Filter support can be re-enabled using CloneOptions.Filter field in v6

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/gabrie30/ghorg/scm"
)

func getCloneDepth() int {
	cloneDepthStr := os.Getenv("GHORG_CLONE_DEPTH")
	if cloneDepthStr != "" {
		if depth, err := strconv.Atoi(cloneDepthStr); err == nil && depth > 0 {
			return depth
		}
	}
	return 1 // Default depth
}

// sanitizeURL removes credentials from a URL for safe logging
func sanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid URL]"
	}
	if u.User != nil {
		u.User = url.UserPassword("***", "***")
	}
	return u.String()
}

// GoGitClient implements the Gitter interface using the go-git library
type GoGitClient struct{}

// printDebugGoGit mimics the debug output for go-git operations
func printDebugGoGit(operation string, repo scm.Repo, details string) {
	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Println("------------- GO-GIT DEBUG -------------")
		fmt.Printf("GHORG_OUTPUT_DIR=%v\n", os.Getenv("GHORG_OUTPUT_DIR"))
		fmt.Printf("GHORG_ABSOLUTE_PATH_TO_CLONE_TO=%v\n", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
		fmt.Printf("Operation: %s\n", operation)
		fmt.Printf("Repository Path: %s\n", repo.HostPath)
		fmt.Printf("Repository Name: %s\n", repo.Name)
		fmt.Printf("Clone Branch: %s\n", repo.CloneBranch)
		if details != "" {
			// Sanitize any URLs in details to prevent logging credentials
			sanitizedDetails := details
			if strings.Contains(details, "URL:") {
				// Extract and sanitize any URLs in the details
				parts := strings.SplitN(details, "URL:", 2)
				if len(parts) == 2 {
					urlPart := strings.TrimSpace(parts[1])
					sanitizedURL := sanitizeURL(urlPart)
					sanitizedDetails = parts[0] + "URL: " + sanitizedURL
				}
			}
			fmt.Printf("Details: %s\n", sanitizedDetails)
		}
		fmt.Println("-------------------------------------")
	}
}

func (g *GoGitClient) HasRemoteHeads(repo scm.Repo) (bool, error) {
	printDebugGoGit("HasRemoteHeads", repo, "")

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

func (g *GoGitClient) Clone(repo scm.Repo) error {
	printDebugGoGit("Clone", repo, fmt.Sprintf("URL: %s", sanitizeURL(repo.CloneURL)))

	gitFilter := os.Getenv("GHORG_GIT_FILTER")

	// Fall back to git CLI when filtering is requested
	// go-git v5 does not support native filtering, and filter approximation
	// has been removed in favor of using the git CLI for proper filter support
	if gitFilter != "" {
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Git filter specified (%s), falling back to git CLI for proper filter support\n", gitFilter)
		}

		// Use the git CLI implementation for filtering
		gitCLI := &GitClient{}
		return gitCLI.Clone(repo)
	}

	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"
	cloneDepth := getCloneDepth()
	isMirror := os.Getenv("GHORG_BACKUP") == "true"
	singleBranch := os.Getenv("GHORG_SINGLE_BRANCH") == "true"
	branch := repo.CloneBranch

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
		cloneOptions.RecurseSubmodules = git.NoRecurseSubmodules
	}

	// Set mirror option if enabled
	if isMirror {
		cloneOptions.Mirror = true
	}

	// Perform the clone using go-git
	_, err := git.PlainClone(repo.HostPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (g *GoGitClient) SetOriginWithCredentials(repo scm.Repo) error {
	printDebugGoGit("SetOriginWithCredentials", repo, fmt.Sprintf("URL: %s", sanitizeURL(repo.CloneURL)))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the repository configuration
	cfg, err := r.Config()
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	// Update or create the remote URL for "origin"
	if cfg.Remotes == nil {
		cfg.Remotes = make(map[string]*config.RemoteConfig)
	}

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

func (g *GoGitClient) SetOrigin(repo scm.Repo) error {
	printDebugGoGit("SetOrigin", repo, fmt.Sprintf("URL: %s", repo.URL))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the repository configuration
	cfg, err := r.Config()
	if err != nil {
		return fmt.Errorf("failed to get repository config: %w", err)
	}

	// Update or create the remote URL for "origin"
	if cfg.Remotes == nil {
		cfg.Remotes = make(map[string]*config.RemoteConfig)
	}

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

func (g *GoGitClient) Checkout(repo scm.Repo) error {
	printDebugGoGit("Checkout", repo, fmt.Sprintf("Branch: %s", repo.CloneBranch))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to checkout the local branch first
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(repo.CloneBranch),
	})
	if err != nil {
		// If local branch doesn't exist, check if there's a remote branch to track
		remoteRef := plumbing.NewRemoteReferenceName("origin", repo.CloneBranch)
		_, err = r.Reference(remoteRef, true)
		if err != nil {
			// Remote branch doesn't exist either, fail
			return fmt.Errorf("failed to checkout branch '%s': %w", repo.CloneBranch, err)
		}

		// Remote branch exists, create a local tracking branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(repo.CloneBranch),
			Create: true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch '%s': %w", repo.CloneBranch, err)
		}

		// Set up tracking to the remote branch
		cfg, err := r.Config()
		if err == nil {
			if cfg.Branches == nil {
				cfg.Branches = make(map[string]*config.Branch)
			}
			cfg.Branches[repo.CloneBranch] = &config.Branch{
				Name:   repo.CloneBranch,
				Remote: "origin",
				Merge:  plumbing.NewBranchReferenceName(repo.CloneBranch),
			}
			r.Storer.SetConfig(cfg)
		}
	}

	return nil
}

func (g *GoGitClient) Clean(repo scm.Repo) error {
	printDebugGoGit("Clean", repo, "Removing untracked files and directories")

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
			fullPath := filepath.Join(repo.HostPath, file)
			// Remove untracked files and directories
			err := os.RemoveAll(fullPath)
			if err != nil {
				return fmt.Errorf("failed to remove untracked file or directory '%s': %w", file, err)
			}
		}
	}

	return nil
}

func (g *GoGitClient) UpdateRemote(repo scm.Repo) error {
	printDebugGoGit("UpdateRemote", repo, "Fetching updates from all remotes")

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Fetch updates for all remotes
	err = r.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
		Tags:       git.AllTags,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to update remote: %w", err)
	}

	return nil
}

func (g *GoGitClient) Pull(repo scm.Repo) error {
	printDebugGoGit("Pull", repo, fmt.Sprintf("Branch: %s", repo.CloneBranch))

	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	pullOptions := &git.PullOptions{
		RemoteName:    git.DefaultRemoteName,
		ReferenceName: plumbing.NewBranchReferenceName(repo.CloneBranch),
		Force:         true,
		Depth:         getCloneDepth(),
	}

	if recurseSubmodules {
		pullOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	} else {
		pullOptions.RecurseSubmodules = git.NoRecurseSubmodules
	}

	err = w.Pull(pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

func (g *GoGitClient) Reset(repo scm.Repo) error {
	printDebugGoGit("Reset", repo, fmt.Sprintf("Hard reset to origin/%s", repo.CloneBranch))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// If no clone branch is specified, get the current branch
	branchName := repo.CloneBranch
	if branchName == "" {
		head, err := r.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}
		if head.Name().IsBranch() {
			branchName = head.Name().Short()
		} else {
			// In detached HEAD state, just reset to HEAD
			err = w.Reset(&git.ResetOptions{
				Mode: git.HardReset,
			})
			if err != nil {
				return fmt.Errorf("failed to reset: %w", err)
			}
			return nil
		}
	}

	// Get the remote branch reference to reset to
	remoteRef, err := r.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true)
	if err != nil {
		// If remote reference doesn't exist, just do a hard reset to HEAD
		err = w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			return fmt.Errorf("failed to reset: %w", err)
		}
		return nil
	}

	err = w.Reset(&git.ResetOptions{
		Commit: remoteRef.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	return nil
}

func (g *GoGitClient) FetchAll(repo scm.Repo) error {
	printDebugGoGit("FetchAll", repo, "Fetching all branches")

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	fetchOptions := &git.FetchOptions{
		RemoteName: git.DefaultRemoteName,
		RefSpecs:   []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		Depth:      getCloneDepth(),
		Force:      true,
		Tags:       git.AllTags,
	}

	err = r.Fetch(fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch all: %w", err)
	}

	// Attempt to sync the default branch after fetch
	if err := g.SyncDefaultBranch(repo); err != nil {
		// Log the sync error but don't fail the fetch operation
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Printf("Warning: Failed to sync default branch after fetch: %v\n", err)
		}
	}

	return nil
}

func (g *GoGitClient) FetchCloneBranch(repo scm.Repo) error {
	printDebugGoGit("FetchCloneBranch", repo, fmt.Sprintf("Branch: %s", repo.CloneBranch))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Prepare fetch options for specific branch
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

func (g *GoGitClient) RepoCommitCount(repo scm.Repo) (int, error) {
	printDebugGoGit("RepoCommitCount", repo, fmt.Sprintf("Branch: %s", repo.CloneBranch))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the reference for the specified branch
	ref, err := r.Reference(plumbing.NewBranchReferenceName(repo.CloneBranch), true)
	if err != nil {
		// Try remote reference if local doesn't exist
		ref, err = r.Reference(plumbing.NewRemoteReferenceName("origin", repo.CloneBranch), true)
		if err != nil {
			return 0, fmt.Errorf("failed to get branch reference: %w", err)
		}
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

func (g *GoGitClient) Branch(repo scm.Repo) (string, error) {
	printDebugGoGit("Branch", repo, "Listing branches")

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
		// Filter for branch references, mimicking git branch output
		if ref.Type() == plumbing.HashReference && strings.HasPrefix(ref.Name().String(), "refs/heads/") {
			branchName := strings.TrimPrefix(ref.Name().String(), "refs/heads/")

			// Check if this is the current branch
			head, err := r.Head()
			if err == nil && head.Name().String() == ref.Name().String() {
				branches = append(branches, "* "+branchName)
			} else {
				branches = append(branches, "  "+branchName)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate over references: %w", err)
	}

	// Join the branch names into a single string
	return strings.Join(branches, "\n"), nil
}

func (g *GoGitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
	printDebugGoGit("RevListCompare", repo, fmt.Sprintf("Local: %s, Remote: %s", localBranch, remoteBranch))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
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
		// Check if this commit is an ancestor of the remote commit
		isAncestor, err := c.IsAncestor(remoteCommit)
		if err != nil {
			return err
		}
		if !isAncestor && c.Hash != remoteCommit.Hash {
			commits = append(commits, c.Hash.String())
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate over commits: %w", err)
	}

	return strings.Join(commits, "\n"), nil
}

func (g *GoGitClient) ShortStatus(repo scm.Repo) (string, error) {
	printDebugGoGit("ShortStatus", repo, "Getting short status")

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

	// Convert the status to a short format string, mimicking git status --short
	var statusLines []string
	for file, fileStatus := range status {
		var statusChars string

		// Format like git status --short: XY filename
		switch fileStatus.Staging {
		case git.Added:
			statusChars += "A"
		case git.Modified:
			statusChars += "M"
		case git.Deleted:
			statusChars += "D"
		case git.Renamed:
			statusChars += "R"
		case git.Copied:
			statusChars += "C"
		default:
			statusChars += " "
		}

		switch fileStatus.Worktree {
		case git.Modified:
			statusChars += "M"
		case git.Deleted:
			statusChars += "D"
		case git.Untracked:
			statusChars = "??"
		default:
			if len(statusChars) == 1 {
				statusChars += " "
			}
		}

		statusLines = append(statusLines, fmt.Sprintf("%s %s", statusChars, file))
	}

	return strings.Join(statusLines, "\n"), nil
}

// GetCurrentBranch returns the name of the currently checked out branch using go-git
func (g *GoGitClient) GetCurrentBranch(repo scm.Repo) (string, error) {
	printDebugGoGit("GetCurrentBranch", repo, "")

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if we're in detached HEAD state
	if !head.Name().IsBranch() {
		return "", fmt.Errorf("repository is in detached HEAD state")
	}

	// Extract branch name from reference
	branchName := head.Name().Short()
	return branchName, nil
}

// HasLocalChanges checks if there are any uncommitted changes in the working directory using go-git
func (g *GoGitClient) HasLocalChanges(repo scm.Repo) (bool, error) {
	printDebugGoGit("HasLocalChanges", repo, "")

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := r.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	// If status is clean, there are no changes
	return !status.IsClean(), nil
}

// HasUnpushedCommits checks if the current branch has commits that haven't been pushed to the remote using go-git
func (g *GoGitClient) HasUnpushedCommits(repo scm.Repo) (bool, error) {
	printDebugGoGit("HasUnpushedCommits", repo, "")

	// Get current branch
	currentBranch, err := g.GetCurrentBranch(repo)
	if err != nil {
		return false, err
	}

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		// If the command fails, assume there are unpushed commits to be safe (matches CLI behavior)
		return true, nil
	}

	// Get local branch reference
	localRef, err := r.Reference(plumbing.NewBranchReferenceName(currentBranch), true)
	if err != nil {
		// If the branch reference fails, assume there are unpushed commits to be safe (matches CLI behavior)
		return true, nil
	}

	// Get remote branch reference
	remoteRefName := plumbing.NewRemoteReferenceName("origin", currentBranch)
	remoteRef, err := r.Reference(remoteRefName, true)
	if err != nil {
		// If remote branch doesn't exist, assume there are unpushed commits
		return true, nil
	}

	// Compare commit hashes
	return localRef.Hash() != remoteRef.Hash(), nil
}

// HasCommitsNotOnDefaultBranch checks if the current branch has commits that are not on the default branch using go-git
func (g *GoGitClient) HasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error) {
	printDebugGoGit("HasCommitsNotOnDefaultBranch", repo, fmt.Sprintf("Current: %s, Default: %s", currentBranch, repo.CloneBranch))

	// Skip the check if we're already on the default branch
	if currentBranch == repo.CloneBranch {
		return false, nil
	}

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	// Get current branch reference
	currentRef, err := r.Reference(plumbing.NewBranchReferenceName(currentBranch), true)
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	// Get remote default branch reference
	defaultRemoteRefName := plumbing.NewRemoteReferenceName("origin", repo.CloneBranch)
	defaultRef, err := r.Reference(defaultRemoteRefName, true)
	if err != nil {
		// If remote default branch doesn't exist, assume there are divergent commits
		return true, nil
	}

	// Check if current branch has commits beyond the default branch
	defaultCommit, err := r.CommitObject(defaultRef.Hash())
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	currentCommit, err := r.CommitObject(currentRef.Hash())
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	// Check if current commit is an ancestor of default branch commit
	isAncestor, err := currentCommit.IsAncestor(defaultCommit)
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	// If current commit is an ancestor or same as default, then no divergent commits
	if isAncestor || currentCommit.Hash == defaultCommit.Hash {
		return false, nil
	}

	// Check if default commit is an ancestor of current commit
	isDefaultAncestor, err := defaultCommit.IsAncestor(currentCommit)
	if err != nil {
		// If the command fails, assume there are divergent commits to be safe (matches CLI behavior)
		return true, nil
	}

	// If default is ancestor of current, then current has commits not on default
	return isDefaultAncestor, nil
}

// UpdateRef updates a git reference to point to a specific commit using go-git
func (g *GoGitClient) UpdateRef(repo scm.Repo, refName string, commitRef string) error {
	printDebugGoGit("UpdateRef", repo, fmt.Sprintf("Ref: %s, Commit: %s", refName, commitRef))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Resolve the commit reference (could be a symbolic ref)
	var hash plumbing.Hash
	if strings.HasPrefix(commitRef, "refs/") {
		// It's a symbolic reference, resolve it
		ref, err := r.Reference(plumbing.ReferenceName(commitRef), true)
		if err != nil {
			return fmt.Errorf("failed to resolve reference %s: %w", commitRef, err)
		}
		hash = ref.Hash()
	} else {
		// It's already a commit hash
		hash = plumbing.NewHash(commitRef)
	}

	// Create or update the reference
	ref := plumbing.NewHashReference(plumbing.ReferenceName(refName), hash)
	err = r.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to update reference: %w", err)
	}

	return nil
}

// GetRemoteURL retrieves the URL for the specified remote using go-git
func (g *GoGitClient) GetRemoteURL(repo scm.Repo, remoteName string) (string, error) {
	printDebugGoGit("GetRemoteURL", repo, fmt.Sprintf("Remote: %s", remoteName))

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := r.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	if len(remote.Config().URLs) == 0 {
		return "", fmt.Errorf("no URLs configured for remote %s", remoteName)
	}

	return remote.Config().URLs[0], nil
}

// IsDefaultBranchBehindHead checks if the default branch is behind HEAD using go-git
func (g *GoGitClient) IsDefaultBranchBehindHead(repo scm.Repo, currentBranch string) (bool, error) {
	printDebugGoGit("IsDefaultBranchBehindHead", repo, fmt.Sprintf("Current: %s, Default: %s", currentBranch, repo.CloneBranch))

	if currentBranch == repo.CloneBranch {
		// If we're already on the default branch, it can't be behind itself
		return false, nil
	}

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		// If there's an error opening the repo, assume not behind to be safe (matches CLI behavior)
		return false, nil
	}

	// Get current branch reference
	currentRef, err := r.Reference(plumbing.NewBranchReferenceName(currentBranch), true)
	if err != nil {
		// If there's an error, assume not behind to be safe (matches CLI behavior)
		return false, nil
	}

	// Get remote default branch reference
	defaultRemoteRefName := plumbing.NewRemoteReferenceName("origin", repo.CloneBranch)
	defaultRef, err := r.Reference(defaultRemoteRefName, true)
	if err != nil {
		// If remote default branch doesn't exist, assume not behind
		return false, nil
	}

	// Get commit objects
	currentCommit, err := r.CommitObject(currentRef.Hash())
	if err != nil {
		// If there's an error, assume not behind to be safe (matches CLI behavior)
		return false, nil
	}

	defaultCommit, err := r.CommitObject(defaultRef.Hash())
	if err != nil {
		// If there's an error, assume not behind to be safe (matches CLI behavior)
		return false, nil
	}

	// Check if default branch has commits that current branch doesn't have
	defaultHasUniqueCommits, err := defaultCommit.IsAncestor(currentCommit)
	if err != nil {
		// If there's an error, assume not behind to be safe (matches CLI behavior)
		return false, nil
	}

	// If default is ancestor of current, then default doesn't have unique commits
	if defaultHasUniqueCommits {
		// Check if current has commits that default doesn't have
		currentHasUniqueCommits, err := currentCommit.IsAncestor(defaultCommit)
		if err != nil {
			// If there's an error, assume not behind to be safe (matches CLI behavior)
			return false, nil
		}

		// If current is not ancestor of default, then current has unique commits
		// and default is behind
		return !currentHasUniqueCommits, nil
	}

	// Default has unique commits, so it's ahead, not behind
	return false, nil
}

// MergeIntoDefaultBranch merges the current branch into the default branch using go-git
func (g *GoGitClient) MergeIntoDefaultBranch(repo scm.Repo, sourceBranch string) error {
	printDebugGoGit("MergeIntoDefaultBranch", repo, fmt.Sprintf("Source: %s, Target: %s", sourceBranch, repo.CloneBranch))

	// Get current branch
	currentBranch, err := g.GetCurrentBranch(repo)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Switch to the default branch if we're not already on it
	if currentBranch != repo.CloneBranch {
		err := g.Checkout(repo)
		if err != nil {
			return fmt.Errorf("failed to checkout default branch %s: %w", repo.CloneBranch, err)
		}
	}

	r, err := git.PlainOpen(repo.HostPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the source branch reference
	sourceRef, err := r.Reference(plumbing.NewBranchReferenceName(sourceBranch), true)
	if err != nil {
		return fmt.Errorf("failed to get source branch reference: %w", err)
	}

	// Get the source commit
	sourceCommit, err := r.CommitObject(sourceRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get source commit: %w", err)
	}

	// Get current HEAD
	head, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if this would be a fast-forward merge
	headCommit, err := r.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	isAncestor, err := headCommit.IsAncestor(sourceCommit)
	if err != nil {
		return fmt.Errorf("failed to check if HEAD is ancestor of source: %w", err)
	}

	if !isAncestor {
		return fmt.Errorf("cannot perform fast-forward merge: HEAD is not an ancestor of %s", sourceBranch)
	}

	// Update the default branch reference to point to the source commit (fast-forward)
	defaultRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(repo.CloneBranch), sourceRef.Hash())
	err = r.Storer.SetReference(defaultRef)
	if err != nil {
		return fmt.Errorf("failed to update default branch reference: %w", err)
	}

	// Update the worktree to match the new HEAD
	err = w.Reset(&git.ResetOptions{
		Commit: sourceRef.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset worktree after merge: %w", err)
	}

	return nil
}

// SyncDefaultBranch implements sync logic using go-git
func (g *GoGitClient) SyncDefaultBranch(repo scm.Repo) error {
	printDebugGoGit("SyncDefaultBranch", repo, "")

	// Check if sync is enabled via environment variable
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "true" {
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
