package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/davecgh/go-spew/spew"
)

type Gitter interface {
	Clone(scm.Repo) error
	Reset(scm.Repo) error
	Pull(scm.Repo) error
	SetOrigin(scm.Repo) error
	SetOriginWithCredentials(scm.Repo) error
	Clean(scm.Repo) error
	Checkout(scm.Repo) error
	RevListCompare(scm.Repo, string, string) (string, error)
	ShortStatus(scm.Repo) (string, error)
	Branch(scm.Repo) (string, error)
	UpdateRemote(scm.Repo) error
	FetchAll(scm.Repo) error
	FetchCloneBranch(scm.Repo) error
	RepoCommitCount(scm.Repo) (int, error)
	HasRemoteHeads(scm.Repo) (bool, error)
}

type GitClient struct{}

func NewGit() GitClient {
	return GitClient{}
}

func printDebugCmd(cmd *exec.Cmd, repo scm.Repo) error {
	fmt.Println("------------- GIT DEBUG -------------")
	fmt.Printf("GHORG_OUTPUT_DIR=%v\n", os.Getenv("GHORG_OUTPUT_DIR"))
	fmt.Printf("GHORG_ABSOLUTE_PATH_TO_CLONE_TO=%v\n", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
	fmt.Print("Repo Data: ")
	spew.Dump(repo)
	fmt.Print("Command Ran: ")
	spew.Dump(*cmd)
	fmt.Println("")
	output, err := cmd.CombinedOutput()
	fmt.Printf("Command Output: %s\n", string(output))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	return err
}

func getCloneDepth() int {
	cloneDepthStr := os.Getenv("GHORG_CLONE_DEPTH")
	if cloneDepthStr != "" {
		if depth, err := strconv.Atoi(cloneDepthStr); err == nil && depth > 0 {
			return depth
		}
	}
	return 1 // Default depth
}

func (g GitClient) HasRemoteHeads(repo scm.Repo) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", "--quiet", "--exit-code")
	cmd.Dir = repo.HostPath

	err := cmd.Run()
	if err == nil {
		// successfully listed the remote heads
		return true, nil
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		// error, but no exit code, return err
		return false, err
	}

	exitCode := exitError.ExitCode()
	if exitCode == 0 {
		// ls-remote did successfully list the remote heads
		return true, nil
	} else if exitCode == 2 {
		// repository is empty
		return false, nil
	} else {
		// another exit code, simply return err
		return false, err
	}

}

func (g GitClient) Clone(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"
	cloneDepth := getCloneDepth()
	// gitFilter := os.Getenv("GHORG_GIT_FILTER")
	isMirror := os.Getenv("GHORG_BACKUP") == "true"

	// Prepare clone options
	cloneOptions := &git.CloneOptions{
		URL:               repo.CloneURL,
		Depth:             cloneDepth,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	// Set submodule recursion if enabled
	if recurseSubmodules {
		cloneOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	} else {
		cloneOptions.RecurseSubmodules = 0
	}

	// // Set filter if specified
	// if gitFilter != "" {
	// 	cloneOptions.Filter = gitFilter
	// }

	// Set mirror option if enabled
	if isMirror {
		cloneOptions.Mirror = true
	}

	// Perform the clone
	_, err := git.PlainClone(repo.HostPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (g GitClient) SetOriginWithCredentials(repo scm.Repo) error {
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

func (g GitClient) SetOrigin(repo scm.Repo) error {
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

func (g GitClient) Checkout(repo scm.Repo) error {
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

func (g GitClient) Clean(repo scm.Repo) error {
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

func (g GitClient) UpdateRemote(repo scm.Repo) error {
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

func (g GitClient) Pull(repo scm.Repo) error {
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

func (g GitClient) Reset(repo scm.Repo) error {
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

func (g GitClient) FetchAll(repo scm.Repo) error {
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

func (g GitClient) Branch(repo scm.Repo) (string, error) {
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

// RevListCompare returns the list of commits in the local branch that are not in the remote branch.
func (g GitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
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

func (g GitClient) FetchCloneBranch(repo scm.Repo) error {
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

func (g GitClient) ShortStatus(repo scm.Repo) (string, error) {
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

func (g GitClient) RepoCommitCount(repo scm.Repo) (int, error) {
	if os.Getenv("GHORG_USE_GIT_CLI") == "true" {
		return g.repoCommitCountCLI(repo)
	}

	return g.repoCommitCountGoGit(repo)
}

// repoCommitCountCLI implements RepoCommitCount using git CLI
func (g GitClient) repoCommitCountCLI(repo scm.Repo) (int, error) {
	args := []string{"rev-list", "--count", repo.CloneBranch, "--"}
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		err := printDebugCmd(cmd, repo)
		if err != nil {
			return 0, err
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return count, nil
}

// repoCommitCountGoGit implements RepoCommitCount using go-git
func (g GitClient) repoCommitCountGoGit(repo scm.Repo) (int, error) {
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
