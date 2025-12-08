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

	err := cmd.Run()
	return err
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

// GetRemoteURL returns the URL for the given remote name (e.g., "origin").
func (g GitClient) GetRemoteURL(repo scm.Repo, remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return "", err
		}
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// HasLocalChanges returns true if there are uncommitted changes in the working tree.
func (g GitClient) HasLocalChanges(repo scm.Repo) (bool, error) {
	status, err := g.ShortStatus(repo)
	if err != nil {
		return false, err
	}
	return status != "", nil
}

// HasUnpushedCommits returns true if there are commits present locally that are not pushed to upstream.
func (g GitClient) HasUnpushedCommits(repo scm.Repo) (bool, error) {
	cmd := exec.Command("git", "rev-list", "--count", "@{u}..HEAD")
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return false, err
		}
	}

	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	countStr := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCurrentBranch returns the currently checked-out branch name.
func (g GitClient) GetCurrentBranch(repo scm.Repo) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return "", err
		}
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// HasCommitsNotOnDefaultBranch returns true if currentBranch contains commits not present on the default branch.
func (g GitClient) HasCommitsNotOnDefaultBranch(repo scm.Repo, currentBranch string) (bool, error) {
	// Count commits reachable from currentBranch that are not on default branch
	cmd := exec.Command("git", "rev-list", "--count", currentBranch, "^refs/heads/"+repo.CloneBranch)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return false, err
		}
	}
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	countStr := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsDefaultBranchBehindHead returns true if the default branch is an ancestor of the current branch (i.e., can be fast-forwarded).
func (g GitClient) IsDefaultBranchBehindHead(repo scm.Repo, currentBranch string) (bool, error) {
	// git merge-base --is-ancestor <default> <current>
	cmd := exec.Command("git", "merge-base", "--is-ancestor", "refs/heads/"+repo.CloneBranch, currentBranch)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			// merge-base --is-ancestor exits with code 1 when not ancestor; treat as false
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				if exitError.ExitCode() == 1 {
					return false, nil
				}
			}
			return false, err
		}
	}

	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		if exitError.ExitCode() == 1 {
			return false, nil
		}
	}
	return false, err
}

// MergeIntoDefaultBranch attempts a fast-forward merge of currentBranch into the default branch locally.
func (g GitClient) MergeIntoDefaultBranch(repo scm.Repo, currentBranch string) error {
	// Checkout default branch
	cmd := exec.Command("git", "checkout", repo.CloneBranch)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return err
		}
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// Merge with --ff-only
	mergeCmd := exec.Command("git", "merge", "--ff-only", currentBranch)
	mergeCmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(mergeCmd, repo); err != nil {
			return err
		}
	}
	return mergeCmd.Run()
}

// UpdateRef updates a local ref to point to the given remote ref (by resolving the remote ref SHA first).
func (g GitClient) UpdateRef(repo scm.Repo, refName string, commitRef string) error {
	// Resolve commitRef to SHA
	revCmd := exec.Command("git", "rev-parse", commitRef)
	revCmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(revCmd, repo); err != nil {
			return err
		}
	}
	out, err := revCmd.Output()
	if err != nil {
		return err
	}
	sha := strings.TrimSpace(string(out))

	// Update the ref
	updCmd := exec.Command("git", "update-ref", refName, sha)
	updCmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(updCmd, repo); err != nil {
			return err
		}
	}
	return updCmd.Run()
}
