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
}

type GitClient struct{}

func NewGit() GitClient {
	return GitClient{}
}

// NewGitClient returns the appropriate Git client based on the useGitCLI flag
func NewGitClient(useGitCLI bool) Gitter {
	if !useGitCLI {
		return GoGitClient{}
	}
	return GitClient{}
}

// NewGitClientForFilter returns the recommended Git client when git filters are needed.
// This function always returns the CLI implementation because go-git has limited filter support.
// Use this when GHORG_GIT_FILTER is important for your use case.
func NewGitClientForFilter() Gitter {
	return GitClient{} // Always use CLI for filter support
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
	// Don't capture output here as it conflicts with main execution
	// The actual output will be handled by the calling function
	return nil
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

	return cmd.Run()
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
	// Validate that the repository path is not empty
	if repo.HostPath == "" {
		return "", fmt.Errorf("repository path cannot be empty")
	}

	args := []string{"status", "--short"}

	cmd := exec.Command("git", args...)
	cmd.Dir = repo.HostPath
	if os.Getenv("GHORG_DEBUG") != "" {
		if err := printDebugCmd(cmd, repo); err != nil {
			return "", err
		}
	}

	output, err := cmd.Output()
	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Printf("Command Output: %s\n", string(output))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
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
