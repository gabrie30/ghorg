package git

import (
	"fmt"
	"os"
	"os/exec"

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
	UpdateRemote(scm.Repo) error
	FetchAll(scm.Repo) error
}

type GitClient struct{}

func NewGit() GitClient {
	return GitClient{}
}

func printDebugCmd(cmd *exec.Cmd, repo scm.Repo) error {
	fmt.Println("------------- GIT DEBUG -------------")
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
