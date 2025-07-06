package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		err := printDebugCmd(cmd, repo)
		if err != nil {
			return err
		}
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	// Set up sparse checkout if GHORG_PATH_FILTER is specified
	pathFilter := os.Getenv("GHORG_PATH_FILTER")
	if pathFilter != "" {
		// Configure git sparse-checkout
		if err := g.configureSparseCheckout(repo, pathFilter); err != nil {
			return fmt.Errorf("failed to configure sparse checkout: %w", err)
		}
	}

	// Sync default branch if no files are checked out (due to filtering)
	if err := g.SyncDefaultBranch(repo); err != nil {
		return fmt.Errorf("failed to sync default branch: %w", err)
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

func (g GitClient) SyncDefaultBranch(repo scm.Repo) error {
	// Check if there are any uncommitted changes in the working directory
	// Only sync if the working directory is clean (no local changes)
	hasChanges, err := g.hasLocalChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to check working directory status: %w", err)
	}

	// If there are no local changes, fetch and update the default branch
	if !hasChanges {
		// First check if the remote exists and is accessible
		cmd := exec.Command("git", "remote", "get-url", "origin")
		cmd.Dir = repo.HostPath
		if err := cmd.Run(); err != nil {
			// Remote doesn't exist or isn't accessible, skip sync
			return nil
		}

		// Fetch the latest changes from the remote
		err := g.FetchCloneBranch(repo)
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
