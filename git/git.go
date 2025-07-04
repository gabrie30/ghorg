package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Gitter interface {
	Clone(scm.Repo, bool) error
	Reset(scm.Repo, bool) error
	Pull(scm.Repo, bool) error
	SetOrigin(scm.Repo, bool) error
	SetOriginWithCredentials(scm.Repo, bool) error
	Clean(scm.Repo, bool) error
	Checkout(scm.Repo, bool) error
	RevListCompare(scm.Repo, string, string, bool) (string, error)
	ShortStatus(scm.Repo, bool) (string, error)
	Branch(scm.Repo, bool) (string, error)
	UpdateRemote(scm.Repo, bool) error
	FetchAll(scm.Repo, bool) error
	FetchCloneBranch(scm.Repo, bool) error
	RepoCommitCount(scm.Repo, bool) (int, error)
	HasRemoteHeads(scm.Repo, bool) (bool, error)
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

func (g GitClient) HasRemoteHeads(repo scm.Repo, useGitCLI bool) (bool, error) {
	if useGitCLI {
		cmd := exec.Command("git", "ls-remote", "--heads", "--quiet", "--exit-code")
		cmd.Dir = repo.HostPath

		err := cmd.Run()
		if err == nil {
			return true, nil
		}

		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return false, err
		}

		exitCode := exitError.ExitCode()
		if exitCode == 0 {
			return true, nil
		} else if exitCode == 2 {
			return false, nil
		} else {
			return false, err
		}
	}

	return g.hasRemoteHeadsWithGo(repo)
}

func (g GitClient) Clone(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"clone", repo.CloneURL, repo.HostPath}

		if os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = "--recursive"
		}

		if os.Getenv("GHORG_CLONE_DEPTH") != "" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
		}

		if os.Getenv("GHORG_GIT_FILTER") != "" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = fmt.Sprintf("--filter=%v", os.Getenv("GHORG_GIT_FILTER"))
		}

		if os.Getenv("GHORG_BACKUP") == "true" {
			args = append(args, "--mirror")
		}

		cmd := exec.Command("git", args...)

		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}

		return cmd.Run()
	}

	return g.cloneWithGo(repo)
}

func (g GitClient) SetOriginWithCredentials(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"remote", "set-url", "origin", repo.CloneURL}
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.setOriginWithCredentialsWithGo(repo)
}

func (g GitClient) SetOrigin(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"remote", "set-url", "origin", repo.URL}
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.setOriginWithGo(repo)
}

func (g GitClient) Checkout(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		cmd := exec.Command("git", "checkout", repo.CloneBranch)
		cmd.Dir = repo.HostPath

		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}

		return cmd.Run()
	}

	return g.checkoutWithGo(repo)
}

func (g GitClient) Clean(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		cmd := exec.Command("git", "clean", "-f", "-d")
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.cleanWithGo(repo)
}

func (g GitClient) UpdateRemote(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		cmd := exec.Command("git", "remote", "update")
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.updateRemoteWithGo(repo)
}

func (g GitClient) Pull(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"pull", "origin", repo.CloneBranch}

		if os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = "--recurse-submodules"
		}

		if os.Getenv("GHORG_CLONE_DEPTH") != "" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
		}

		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath

		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}

		return cmd.Run()
	}

	return g.pullWithGo(repo)
}

func (g GitClient) Reset(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		cmd := exec.Command("git", "reset", "--hard", "origin/"+repo.CloneBranch)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.resetWithGo(repo)
}

func (g GitClient) FetchAll(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"fetch", "--all"}

		if os.Getenv("GHORG_CLONE_DEPTH") != "" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
		}
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.fetchAllWithGo(repo)
}

func (g GitClient) Branch(repo scm.Repo, useGitCLI bool) (string, error) {
	if useGitCLI {
		args := []string{"branch"}

		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			if err := printDebugCmd(cmd, repo); err != nil {
				return "", err
			}
		}

		output, err := cmd.Output()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(output)), nil
	}

	return g.branchWithGo(repo)
}

func (g GitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string, useGitCLI bool) (string, error) {
	if useGitCLI {
		cmd := exec.Command("git", "-C", repo.HostPath, "rev-list", localBranch, "^"+remoteBranch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}

	return g.revListCompareWithGo(repo, localBranch, remoteBranch)
}

func (g GitClient) FetchCloneBranch(repo scm.Repo, useGitCLI bool) error {
	if useGitCLI {
		args := []string{"fetch", "origin", repo.CloneBranch}

		if os.Getenv("GHORG_CLONE_DEPTH") != "" {
			index := 1
			args = append(args[:index+1], args[index:]...)
			args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
		}
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}
		return cmd.Run()
	}

	return g.fetchCloneBranchWithGo(repo)
}

func (g GitClient) ShortStatus(repo scm.Repo, useGitCLI bool) (string, error) {
	if useGitCLI {
		args := []string{"status", "--short"}

		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath
		if os.Getenv("GHORG_DEBUG") != "" {
			if err := printDebugCmd(cmd, repo); err != nil {
				return "", err
			}
		}

		output, err := cmd.Output()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(output)), nil
	}

	return g.shortStatusWithGo(repo)
}

func (g GitClient) RepoCommitCount(repo scm.Repo, useGitCLI bool) (int, error) {
	if useGitCLI {
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

	return g.repoCommitCountWithGo(repo)
}

// hasRemoteHeadsWithGo implements HasRemoteHeads using go-git
func (g GitClient) hasRemoteHeadsWithGo(repo scm.Repo) (bool, error) {
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
func (g GitClient) cloneWithGo(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"
	cloneDepth := getCloneDepth()
	gitFilter := os.Getenv("GHORG_GIT_FILTER")
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
		cloneOptions.RecurseSubmodules = 0
	}

	// Set mirror option if enabled
	if isMirror {
		cloneOptions.Mirror = true
	}

	// Note about Git filter: go-git doesn't support the equivalent of --filter directly
	// This is handled through CLI args when useGitCLI is true
	// For go-git implementation, we clone normally

	// Perform the clone
	r, err := git.PlainClone(repo.HostPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// If git filter was specified, apply it as best we can to the repository config
	// for future fetch operations
	if gitFilter != "" {
		// This is a best-effort approximation as go-git doesn't support
		// Git filter specifications fully
		_ = setGitFilterConfig(r, gitFilter)
	}

	return nil
}

// setOriginWithCredentialsWithGo implements SetOriginWithCredentials using go-git
func (g GitClient) setOriginWithCredentialsWithGo(repo scm.Repo) error {
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
func (g GitClient) setOriginWithGo(repo scm.Repo) error {
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
func (g GitClient) checkoutWithGo(repo scm.Repo) error {
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
func (g GitClient) cleanWithGo(repo scm.Repo) error {
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
func (g GitClient) updateRemoteWithGo(repo scm.Repo) error {
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
func (g GitClient) pullWithGo(repo scm.Repo) error {
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
func (g GitClient) resetWithGo(repo scm.Repo) error {
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
func (g GitClient) fetchAllWithGo(repo scm.Repo) error {
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
func (g GitClient) branchWithGo(repo scm.Repo) (string, error) {
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
func (g GitClient) revListCompareWithGo(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
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
func (g GitClient) fetchCloneBranchWithGo(repo scm.Repo) error {
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
func (g GitClient) shortStatusWithGo(repo scm.Repo) (string, error) {
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
func (g GitClient) repoCommitCountWithGo(repo scm.Repo) (int, error) {
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

// setGitFilterConfig configures the Git repository to use the specified filter for partial clones
func setGitFilterConfig(repo *git.Repository, filterSpec string) error {
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

	// Configure fetch options for the remote to support partial clones
	// Set up fetch options with the filter
	fetchRefSpec := []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"}

	// Update the remote with the new filter settings
	// Unfortunately go-git doesn't have direct API support for partial clone filters
	// In a real git repo, we'd set:
	// git config remote.origin.promisor true
	// git config remote.origin.partialCloneFilter <filterSpec>

	// Best approximation we can do for now is to set the fetch refspecs
	remote.Fetch = fetchRefSpec

	// Save the updated configuration
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to set repository config for filter: %w", err)
	}

	// Note: This is a partial implementation as go-git doesn't fully support Git filter specs
	// For complete support, we'd need to modify the Git config files directly after cloning

	return nil
}
