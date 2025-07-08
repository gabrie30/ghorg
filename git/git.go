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
	SyncDefaultBranch(scm.Repo) error
	// Additional methods for sync functionality
	GetCurrentBranch(scm.Repo) (string, error)
	HasLocalChanges(scm.Repo) (bool, error)
	HasUnpushedCommits(scm.Repo) (bool, error)
	HasCommitsNotOnDefaultBranch(scm.Repo, string) (bool, error)
	UpdateRef(scm.Repo, string, string) error
	GetRemoteURL(scm.Repo, string) (string, error)
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
	switch exitCode {
	case 0:
		// ls-remote did successfully list the remote heads
		return true, nil
	case 2:
		// repository is empty
		return false, nil
	default:
		// another exit code, simply return err
		return false, err
	}

}

func (g GitClient) Clone(repo scm.Repo) error {
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

	var err error
	if os.Getenv("GHORG_DEBUG") != "" {
		err = printDebugCmd(cmd, repo)
	} else {
		err = cmd.Run()
	}

	if err != nil {
		return err
	}

	// If GHORG_SYNC_DEFAULT_BRANCH is enabled, sync the default branch after cloning
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") == "true" {
		if err := g.SyncDefaultBranch(repo); err != nil {
			// Log the sync error but don't fail the clone operation
			if os.Getenv("GHORG_DEBUG") != "" {
				fmt.Printf("Warning: Failed to sync default branch after clone: %v\n", err)
			}
		}
	}

	return nil
}

func (g GitClient) SetOriginWithCredentials(repo scm.Repo) error {
	args := []string{"remote", "set-url", "origin", repo.CloneURL}
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}
	return cmd.Run()
}

func (g GitClient) SetOrigin(repo scm.Repo) error {
	args := []string{"remote", "set-url", "origin", repo.URL}
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}
	return cmd.Run()
}

func (g GitClient) Checkout(repo scm.Repo) error {
	cmd := exec.Command("git", "checkout", repo.CloneBranch)
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}

	return cmd.Run()
}

func (g GitClient) Clean(repo scm.Repo) error {
	cmd := exec.Command("git", "clean", "-f", "-d")
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}
	return cmd.Run()
}

func (g GitClient) UpdateRemote(repo scm.Repo) error {
	cmd := exec.Command("git", "remote", "update")
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}
	return cmd.Run()
}

func (g GitClient) Pull(repo scm.Repo) error {
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

func (g GitClient) Reset(repo scm.Repo) error {
	cmd := exec.Command("git", "reset", "--hard", "origin/"+repo.CloneBranch)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}
	return cmd.Run()
}

func (g GitClient) FetchAll(repo scm.Repo) error {
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

func (g GitClient) Branch(repo scm.Repo) (string, error) {
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

// RevListCompare returns the list of commits in the local branch that are not in the remote branch.
func (g GitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
	cmd := exec.Command("git", "-C", repo.HostPath, "rev-list", localBranch, "^"+remoteBranch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (g GitClient) FetchCloneBranch(repo scm.Repo) error {
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

func (g GitClient) ShortStatus(repo scm.Repo) (string, error) {
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

func (g GitClient) RepoCommitCount(repo scm.Repo) (int, error) {
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

// GetCurrentBranch returns the name of the currently checked out branch
func (g GitClient) GetCurrentBranch(repo scm.Repo) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return "", err
		}
	}

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

// HasLocalChanges checks if there are any uncommitted changes in the working directory
func (g GitClient) HasLocalChanges(repo scm.Repo) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return false, err
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// If output is empty, there are no changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// HasUnpushedCommits checks if the current branch has commits that haven't been pushed to the remote
func (g GitClient) HasUnpushedCommits(repo scm.Repo) (bool, error) {
	// Get the current branch name
	currentBranch, err := g.GetCurrentBranch(repo)
	if err != nil {
		return false, err
	}

	// Compare local branch with remote branch to see if there are unpushed commits
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("origin/%s..%s", currentBranch, currentBranch), "--count")
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			// If the command fails, it might be because the remote branch doesn't exist
			// In this case, assume there are unpushed commits to be safe
			return true, nil
		}
	}

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

// HasCommitsNotOnDefaultBranch checks if the current branch has commits that are not on the default branch
func (g GitClient) HasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error) {
	// Skip the check if we're already on the default branch
	if currentBranch == repo.CloneBranch {
		return false, nil
	}

	// Compare current branch with default branch to see if there are commits not on default
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("origin/%s..%s", repo.CloneBranch, currentBranch), "--count")
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			// If the command fails, it might be because the remote default branch doesn't exist
			// In this case, assume there are divergent commits to be safe
			return true, nil
		}
	}

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

// UpdateRef updates a git reference to point to a specific commit
func (g GitClient) UpdateRef(repo scm.Repo, refName string, commitRef string) error {
	cmd := exec.Command("git", "update-ref", refName, commitRef)
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		return printDebugCmd(cmd, repo)
	}

	return cmd.Run()
}

// GetRemoteURL retrieves the URL for the specified remote
func (g GitClient) GetRemoteURL(repo scm.Repo, remoteName string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	cmd.Dir = repo.HostPath

	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return "", err
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
