package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// GoGitClient implements the Gitter interface using the go-git library
type GoGitClient struct{}

func (g *GoGitClient) HasRemoteHeads(repo scm.Repo) (bool, error) {
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
	// This is handled through CLI args when using CLI implementation
	// For go-git implementation, we clone normally

	// Perform the clone
	r, err := git.PlainClone(repo.HostPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// If git filter was specified, apply it as best we can
	if gitFilter != "" {
		// Apply post-clone filtering as go-git doesn't support partial clone filters directly
		err = applyGitFilterApproximation(r, repo.HostPath, gitFilter)
		if err != nil {
			// Log the error but don't fail the clone since this is a best-effort approximation
			fmt.Printf("Warning: Failed to apply git filter approximation: %v\n", err)
		}
	}

	return nil
}

func (g *GoGitClient) SetOriginWithCredentials(repo scm.Repo) error {
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

func (g *GoGitClient) SetOrigin(repo scm.Repo) error {
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

func (g *GoGitClient) Checkout(repo scm.Repo) error {
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

func (g *GoGitClient) Clean(repo scm.Repo) error {
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

func (g *GoGitClient) UpdateRemote(repo scm.Repo) error {
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

func (g *GoGitClient) Pull(repo scm.Repo) error {
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

func (g *GoGitClient) Reset(repo scm.Repo) error {
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

func (g *GoGitClient) FetchAll(repo scm.Repo) error {
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

func (g *GoGitClient) FetchCloneBranch(repo scm.Repo) error {
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

func (g *GoGitClient) RepoCommitCount(repo scm.Repo) (int, error) {
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

func (g *GoGitClient) Branch(repo scm.Repo) (string, error) {
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

func (g *GoGitClient) RevListCompare(repo scm.Repo, localBranch string, remoteBranch string) (string, error) {
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

func (g *GoGitClient) ShortStatus(repo scm.Repo) (string, error) {
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

// applyGitFilterApproximation applies a best-effort approximation of Git's --filter option
// Since go-git doesn't support partial clone filters, we try to mimic the behavior post-clone
func applyGitFilterApproximation(repo *git.Repository, repoPath, filterSpec string) error {
	// Parse the filter specification
	filterType, filterValue := parseGitFilter(filterSpec)

	switch filterType {
	case "blob":
		return applyBlobFilter(repo, repoPath, filterValue)
	case "tree":
		return applyTreeFilter(repo, repoPath, filterValue)
	case "sparse":
		return applySparseFilter(repo, repoPath, filterValue)
	case "combine":
		return applyCombinedFilter(repo, repoPath, filterValue)
	default:
		// Unknown filter type, configure for future fetches only
		return setGitFilterConfig(repo, filterSpec)
	}
}

// parseGitFilter parses a Git filter specification and returns the type and value
func parseGitFilter(filterSpec string) (string, string) {
	parts := strings.SplitN(filterSpec, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return filterSpec, ""
}

// applyBlobFilter approximates Git's blob filter by removing large binary files
func applyBlobFilter(_ *git.Repository, repoPath, filterValue string) error {
	if filterValue == "none" {
		// blob:none - exclude all blobs (keep only tree and commit objects)
		return removeBinaryFilesFromWorkingTree(repoPath)
	}

	// Parse size limits like blob:limit=10M
	if strings.HasPrefix(filterValue, "limit=") {
		sizeSpec := strings.TrimPrefix(filterValue, "limit=")
		sizeLimit, err := parseSize(sizeSpec)
		if err != nil {
			return fmt.Errorf("invalid blob size limit: %w", err)
		}
		return removeLargeFilesFromWorkingTree(repoPath, sizeLimit)
	}

	return nil
}

// applyTreeFilter approximates Git's tree filter
func applyTreeFilter(_ *git.Repository, repoPath, filterValue string) error {
	if filterValue == "0" {
		// tree:0 - exclude all trees (shallow clone with no subdirectories)
		return flattenDirectoryStructure(repoPath)
	}

	// Parse depth limits like tree:1, tree:2, etc.
	if depth, err := parseDepth(filterValue); err == nil {
		return limitDirectoryDepth(repoPath, depth)
	}

	return nil
}

// applySparseFilter approximates Git's sparse filter using .git/info/sparse-checkout
func applySparseFilter(_ *git.Repository, repoPath, filterValue string) error {
	// Parse sparse filter specification
	patterns := strings.Split(filterValue, ",")
	return createSparseCheckout(repoPath, patterns)
}

// applyCombinedFilter handles combined filter specifications
func applyCombinedFilter(repo *git.Repository, repoPath, filterValue string) error {
	// Parse combined filters like combine:blob:none+tree:1
	filters := strings.Split(filterValue, "+")
	for _, filter := range filters {
		filterType, filterVal := parseGitFilter(filter)
		switch filterType {
		case "blob":
			if err := applyBlobFilter(repo, repoPath, filterVal); err != nil {
				return err
			}
		case "tree":
			if err := applyTreeFilter(repo, repoPath, filterVal); err != nil {
				return err
			}
		case "sparse":
			if err := applySparseFilter(repo, repoPath, filterVal); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeBinaryFilesFromWorkingTree removes binary files to approximate blob:none
func removeBinaryFilesFromWorkingTree(repoPath string) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git directory
		if strings.Contains(path, ".git/") {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is binary
		if isBinaryFile(path) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove binary file %s: %w", path, err)
			}
		}

		return nil
	})
}

// removeLargeFilesFromWorkingTree removes files larger than the specified size
func removeLargeFilesFromWorkingTree(repoPath string, sizeLimit int64) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git directory
		if strings.Contains(path, ".git/") {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Remove files larger than the size limit
		if info.Size() > sizeLimit {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove large file %s: %w", path, err)
			}
		}

		return nil
	})
}

// flattenDirectoryStructure moves all files to the root to approximate tree:0
func flattenDirectoryStructure(repoPath string) error {
	var filesToMove []string

	// First, collect all files
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git directory and root
		if strings.Contains(path, ".git/") || path == repoPath {
			return nil
		}

		// Skip directories, we only want files
		if !info.IsDir() {
			relPath, _ := filepath.Rel(repoPath, path)
			if strings.Contains(relPath, "/") {
				filesToMove = append(filesToMove, path)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Move files to root and remove empty directories
	for _, filePath := range filesToMove {
		fileName := filepath.Base(filePath)
		newPath := filepath.Join(repoPath, fileName)

		// Handle name conflicts
		counter := 1
		for {
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				break
			}
			ext := filepath.Ext(fileName)
			name := strings.TrimSuffix(fileName, ext)
			newPath = filepath.Join(repoPath, fmt.Sprintf("%s_%d%s", name, counter, ext))
			counter++
		}

		if err := os.Rename(filePath, newPath); err != nil {
			return fmt.Errorf("failed to move file %s to %s: %w", filePath, newPath, err)
		}
	}

	// Remove empty directories
	return removeEmptyDirectories(repoPath)
}

// limitDirectoryDepth removes directories beyond the specified depth
func limitDirectoryDepth(repoPath string, maxDepth int) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git directory
		if strings.Contains(path, ".git/") {
			return nil
		}

		// Calculate depth relative to repo root
		relPath, _ := filepath.Rel(repoPath, path)
		depth := strings.Count(relPath, "/")

		// Remove directories and files beyond max depth
		if depth > maxDepth {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove path %s: %w", path, err)
			}
			return filepath.SkipDir
		}

		return nil
	})
}

// createSparseCheckout creates a sparse-checkout configuration
func createSparseCheckout(repoPath string, patterns []string) error {
	gitDir := filepath.Join(repoPath, ".git")
	infoDir := filepath.Join(gitDir, "info")
	sparseFile := filepath.Join(infoDir, "sparse-checkout")

	// Create info directory if it doesn't exist
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return fmt.Errorf("failed to create info directory: %w", err)
	}

	// Write sparse-checkout patterns
	content := strings.Join(patterns, "\n") + "\n"
	if err := os.WriteFile(sparseFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Enable sparse-checkout (this would normally be done via git config)
	// Since we can't easily modify git config with go-git, we'll create a basic approximation

	return nil
}

// removeEmptyDirectories removes empty directories recursively
func removeEmptyDirectories(repoPath string) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git directory and root
		if strings.Contains(path, ".git/") || path == repoPath {
			return nil
		}

		if info.IsDir() {
			// Try to remove if empty
			if err := os.Remove(path); err != nil {
				// Directory not empty, that's fine
				if !os.IsExist(err) {
					return nil
				}
			}
		}

		return nil
	})
}

// isBinaryFile checks if a file is likely binary
func isBinaryFile(filePath string) bool {
	// Simple heuristic: check file extension and scan first 512 bytes for null bytes
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".o": true, ".a": true, ".lib": true, ".obj": true,
	}

	if binaryExts[ext] {
		return true
	}

	// Check file content for null bytes
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return false
	}

	// If we find null bytes in the first 512 bytes, consider it binary
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

// parseSize parses size specifications like "10M", "1G", "500K"
func parseSize(sizeSpec string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)([KMGT]?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(sizeSpec))
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size specification: %s", sizeSpec)
	}

	var base int64 = 1
	switch matches[2] {
	case "K":
		base = 1024
	case "M":
		base = 1024 * 1024
	case "G":
		base = 1024 * 1024 * 1024
	case "T":
		base = 1024 * 1024 * 1024 * 1024
	}

	size := parseInt64(matches[1])
	return size * base, nil
}

// parseDepth parses depth specifications
func parseDepth(depthSpec string) (int, error) {
	if depth := parseInt(depthSpec); depth >= 0 {
		return depth, nil
	}
	return 0, fmt.Errorf("invalid depth specification: %s", depthSpec)
}

// parseInt parses an integer from a string
func parseInt(s string) int {
	var result int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return -1
		}
	}
	return result
}

// parseInt64 parses an int64 from a string
func parseInt64(s string) int64 {
	var result int64
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int64(r-'0')
		} else {
			return 0
		}
	}
	return result
}

// setGitFilterConfig configures the Git repository to use the specified filter for partial clones
func setGitFilterConfig(repo *git.Repository, _ string) error {
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
