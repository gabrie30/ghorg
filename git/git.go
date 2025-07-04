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
	RepoCommitCount(scm.Repo) (int, error)
	HasRemoteHeads(scm.Repo, bool) (bool, error)
}

type GitClient struct{}

func NewGit() GitClient {
	return GitClient{}
}

// executeGitCLI executes a git CLI command if useGitCLI is true, otherwise calls the provided fallback function
func (g GitClient) executeGitCLI(useGitCLI bool, repo scm.Repo, args []string, fallbackFunc func() error) error {
	if useGitCLI {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo.HostPath

		if os.Getenv("GHORG_DEBUG") != "" {
			return printDebugCmd(cmd, repo)
		}

		return cmd.Run()
	}

	return fallbackFunc()
}

// executeGitCLIWithOutput executes a git CLI command that returns output if useGitCLI is true, otherwise calls the provided fallback function
func (g GitClient) executeGitCLIWithOutput(useGitCLI bool, repo scm.Repo, args []string, fallbackFunc func() (string, error)) (string, error) {
	if useGitCLI {
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

	return fallbackFunc()
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
		return g.hasRemoteHeadsCLI(repo)
	}

	return g.hasRemoteHeadsGoGit(repo)
}

// hasRemoteHeadsCLI implements HasRemoteHeads using git CLI
func (g GitClient) hasRemoteHeadsCLI(repo scm.Repo) (bool, error) {
	// Original implementation using the git CLI
	cmd := exec.Command("git", "ls-remote", "--heads", "--quiet", "--exit-code")
	cmd.Dir = repo.HostPath

	err := cmd.Run()
	if err == nil {
		// Successfully listed the remote heads
		return true, nil
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		// Error, but no exit code, return err
		return false, err
	}

	exitCode := exitError.ExitCode()
	if exitCode == 0 {
		// ls-remote successfully listed the remote heads
		return true, nil
	} else if exitCode == 2 {
		// Repository is empty
		return false, nil
	} else {
		// Another exit code, simply return err
		return false, err
	}
}

// hasRemoteHeadsGoGit implements HasRemoteHeads using go-git
func (g GitClient) hasRemoteHeadsGoGit(repo scm.Repo) (bool, error) {
	// New implementation using go-git
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

func (g GitClient) Clone(repo scm.Repo, useGitCLI bool) error {
	// First do the normal clone
	if err := g.executeGitCLI(useGitCLI, repo, g.buildCloneArgs(repo), func() error {
		return g.cloneGoGit(repo)
	}); err != nil {
		return err
	}

	// Then set up sparse checkout for CLI users
	// This is only needed for CLI users because the go-git version handles sparse checkout differently
	pathFilter := os.Getenv("GHORG_PATH_FILTER")
	if useGitCLI && pathFilter != "" {
		// Configure git sparse-checkout
		if err := g.executeGitCLI(true, repo, []string{"config", "core.sparseCheckout", "true"}, nil); err != nil {
			return fmt.Errorf("failed to configure sparse checkout: %w", err)
		}

		// Initialize sparse checkout
		if err := g.executeGitCLI(true, repo, []string{"sparse-checkout", "init", "--cone"}, nil); err != nil {
			// If 'sparse-checkout init' fails (older Git versions), try the alternative approach
			if err := g.executeGitCLI(true, repo, []string{"config", "core.sparseCheckout", "true"}, nil); err != nil {
				return fmt.Errorf("failed to configure sparse checkout: %w", err)
			}
		}

		// Set the sparse checkout patterns
		if err := g.executeGitCLI(true, repo, []string{"sparse-checkout", "set", pathFilter}, nil); err != nil {
			// If 'sparse-checkout set' fails (older Git versions), try writing to the sparse-checkout file directly
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
			if err := g.executeGitCLI(true, repo, []string{"read-tree", "-mu", "HEAD"}, nil); err != nil {
				return fmt.Errorf("failed to apply sparse checkout: %w", err)
			}
		}
	}

	return nil
}

// buildCloneArgs builds the arguments for git clone command
func (g GitClient) buildCloneArgs(repo scm.Repo) []string {
	args := []string{"clone", repo.CloneURL, repo.HostPath}
	index := 1

	if os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true" {
		args = append(args[:index+1], args[index:]...)
		args[index] = "--recursive"
		index++
	}

	if os.Getenv("GHORG_CLONE_DEPTH") != "" {
		args = append(args[:index+1], args[index:]...)
		args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
		index++
	}

	if os.Getenv("GHORG_GIT_FILTER") != "" {
		args = append(args[:index+1], args[index:]...)
		args[index] = fmt.Sprintf("--filter=%v", os.Getenv("GHORG_GIT_FILTER"))
		index++
	}

	if os.Getenv("GHORG_SINGLE_BRANCH") == "true" {
		args = append(args[:index+1], args[index:]...)
		args[index] = "--single-branch"
		index++
	}

	// Ensure the branch is specified if it's not empty
	if repo.CloneBranch != "" {
		args = append(args[:index+1], args[index:]...)
		args[index] = fmt.Sprintf("--branch=%v", repo.CloneBranch)
		index++
	}

	if os.Getenv("GHORG_BACKUP") == "true" {
		args = append(args, "--mirror")
	}

	// For path filtering, when using CLI we need to set up a sparse checkout
	// after the clone, which will be handled in a separate step

	return args
}

func (g GitClient) SetOriginWithCredentials(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"remote", "set-url", "origin", repo.CloneURL}, func() error {
		return g.setOriginWithCredentialsGoGit(repo)
	})
}

func (g GitClient) SetOrigin(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"remote", "set-url", "origin", repo.URL}, func() error {
		return g.setOriginGoGit(repo)
	})
}

func (g GitClient) Checkout(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"checkout", repo.CloneBranch}, func() error {
		return g.checkoutGoGit(repo)
	})
}

func (g GitClient) Clean(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"clean", "-f", "-d"}, func() error {
		return g.cleanGoGit(repo)
	})
}

func (g GitClient) UpdateRemote(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"remote", "update"}, func() error {
		return g.updateRemoteGoGit(repo)
	})
}

func (g GitClient) Pull(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, g.buildPullArgs(repo), func() error {
		return g.pullGoGit(repo)
	})
}

// buildPullArgs builds the arguments for git pull command
func (g GitClient) buildPullArgs(repo scm.Repo) []string {
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

	return args
}

func (g GitClient) Reset(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, []string{"reset", "--hard", "origin/" + repo.CloneBranch}, func() error {
		return g.resetGoGit(repo)
	})
}

func (g GitClient) FetchAll(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, g.buildFetchAllArgs(), func() error {
		return g.fetchAllGoGit(repo)
	})
}

// buildFetchAllArgs builds the arguments for git fetch --all command
func (g GitClient) buildFetchAllArgs() []string {
	args := []string{"fetch", "--all"}

	if os.Getenv("GHORG_CLONE_DEPTH") != "" {
		index := 1
		args = append(args[:index+1], args[index:]...)
		args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
	}

	return args
}

func (g GitClient) Branch(repo scm.Repo, useGitCLI bool) (string, error) {
	return g.executeGitCLIWithOutput(useGitCLI, repo, []string{"branch"}, func() (string, error) {
		return g.branchGoGit(repo)
	})
}

// RevListCompare returns the list of commits in the local branch that are not in the remote branch.
func (g GitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string, useGitCLI bool) (string, error) {
	if useGitCLI {
		cmd := exec.Command("git", "-C", repo.HostPath, "rev-list", localBranch, "^"+remoteBranch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}

	return g.revListCompareGoGit(repo, localBranch, remoteBranch)
}

// revListCompareGoGit implements RevListCompare using go-git
func (g GitClient) revListCompareGoGit(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
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

func (g GitClient) FetchCloneBranch(repo scm.Repo, useGitCLI bool) error {
	return g.executeGitCLI(useGitCLI, repo, g.buildFetchCloneBranchArgs(repo), func() error {
		return g.fetchCloneBranchGoGit(repo)
	})
}

// buildFetchCloneBranchArgs builds the arguments for git fetch origin <branch> command
func (g GitClient) buildFetchCloneBranchArgs(repo scm.Repo) []string {
	args := []string{"fetch", "origin", repo.CloneBranch}

	if os.Getenv("GHORG_CLONE_DEPTH") != "" {
		index := 1
		args = append(args[:index+1], args[index:]...)
		args[index] = fmt.Sprintf("--depth=%v", os.Getenv("GHORG_CLONE_DEPTH"))
	}

	return args
}

func (g GitClient) ShortStatus(repo scm.Repo, useGitCLI bool) (string, error) {
	return g.executeGitCLIWithOutput(useGitCLI, repo, []string{"status", "--short"}, func() (string, error) {
		return g.shortStatusGoGit(repo)
	})
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

// markParentPaths marks all parent directories of a path as "keep"
// This ensures that if we want to keep "src/main/java", we also keep "src" and "src/main"
func markParentPaths(pathsToKeep map[string]bool, path string) {
	parts := strings.Split(path, "/")
	currentPath := ""

	for i, part := range parts {
		if i > 0 {
			currentPath += "/"
		}
		currentPath += part
		pathsToKeep[currentPath] = true
	}
}

// Go-git implementation functions
// These functions contain the go-git specific logic separated from the CLI implementations

// cloneGoGit implements the clone operation using go-git library
func (g GitClient) cloneGoGit(repo scm.Repo) error {
	recurseSubmodules := os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true"
	cloneDepth := getCloneDepth()
	pathFilter := os.Getenv("GHORG_PATH_FILTER")
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
	// For go-git implementation, we clone normally then apply sparse checkout

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

	// If a path filter is specified, manually implement sparse checkout
	if pathFilter != "" {
		// Split the filter by commas and prepare the sparse checkout directories
		filterPaths := strings.Split(pathFilter, ",")
		normalizedPaths := make([]string, 0, len(filterPaths))
		for _, path := range filterPaths {
			// Normalize the path
			path = strings.TrimSpace(path)
			if path != "" {
				normalizedPaths = append(normalizedPaths, path)
			}
		}

		// Apply manual sparse checkout by removing files that don't match the filter
		if len(normalizedPaths) > 0 {
			// Get the root path
			rootPath := repo.HostPath

			// Create a map of paths to keep - faster lookups
			pathsToKeep := make(map[string]bool)

			// First scan to collect paths we want to keep
			err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip the .git directory
				if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") {
					return nil
				}

				// Get the relative path from the repo root
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					return err
				}

				// Skip the root directory
				if relPath == "." {
					return nil
				}

				// Check if this path matches our filters
				for _, filterPath := range normalizedPaths {
					// Case 1: Direct match
					if relPath == filterPath {
						pathsToKeep[relPath] = true
						// Mark all parent paths to keep
						markParentPaths(pathsToKeep, relPath)
						break
					}

					// Case 2: Path is within a filtered directory
					if strings.HasPrefix(relPath, filterPath+"/") {
						pathsToKeep[relPath] = true
						break
					}

					// Case 3: Path is a parent directory of a filter path
					if strings.HasPrefix(filterPath, relPath+"/") {
						pathsToKeep[relPath] = true
						// Also mark all parent paths
						markParentPaths(pathsToKeep, relPath)
						break
					}
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to identify paths for sparse checkout: %w", err)
			}

			// Second scan to remove unwanted paths
			err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip the .git directory
				if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") {
					return nil
				}

				// Get the relative path from the repo root
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					return err
				}

				// Skip the root directory
				if relPath == "." {
					return nil
				}

				// If this path is not in our keep list, remove it
				if !pathsToKeep[relPath] {
					if info.IsDir() {
						// Remove the entire directory
						if err := os.RemoveAll(path); err != nil {
							return fmt.Errorf("failed to remove directory %s: %w", path, err)
						}
						return filepath.SkipDir // Skip the removed directory
					} else {
						// Remove the file
						if err := os.Remove(path); err != nil {
							return fmt.Errorf("failed to remove file %s: %w", path, err)
						}
					}
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to set up manual sparse checkout: %w", err)
			}
		}
	}

	return nil
}

// setOriginWithCredentialsGoGit implements SetOriginWithCredentials using go-git
func (g GitClient) setOriginWithCredentialsGoGit(repo scm.Repo) error {
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

// setOriginGoGit implements SetOrigin using go-git
func (g GitClient) setOriginGoGit(repo scm.Repo) error {
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

// checkoutGoGit implements Checkout using go-git
func (g GitClient) checkoutGoGit(repo scm.Repo) error {
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

// cleanGoGit implements Clean using go-git
func (g GitClient) cleanGoGit(repo scm.Repo) error {
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

// updateRemoteGoGit implements UpdateRemote using go-git
func (g GitClient) updateRemoteGoGit(repo scm.Repo) error {
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

// pullGoGit implements Pull using go-git
func (g GitClient) pullGoGit(repo scm.Repo) error {
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

// resetGoGit implements Reset using go-git
func (g GitClient) resetGoGit(repo scm.Repo) error {
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

// fetchAllGoGit implements FetchAll using go-git
func (g GitClient) fetchAllGoGit(repo scm.Repo) error {
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

// branchGoGit implements Branch using go-git
func (g GitClient) branchGoGit(repo scm.Repo) (string, error) {
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

// fetchCloneBranchGoGit implements FetchCloneBranch using go-git
func (g GitClient) fetchCloneBranchGoGit(repo scm.Repo) error {
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

// shortStatusGoGit implements ShortStatus using go-git
func (g GitClient) shortStatusGoGit(repo scm.Repo) (string, error) {
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
