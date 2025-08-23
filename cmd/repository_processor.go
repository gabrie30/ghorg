package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/git"
	"github.com/gabrie30/ghorg/scm"
)

// RepositoryProcessor handles the processing of individual repositories
type RepositoryProcessor struct {
	git            git.Gitter
	stats          *CloneStats
	mutex          *sync.RWMutex
	untouchedRepos []string
}

// CloneStats tracks statistics during clone operations
type CloneStats struct {
	CloneCount        int
	PulledCount       int
	UpdateRemoteCount int
	NewCommits        int
	UntouchedPrunes   int
	CloneInfos        []string
	CloneErrors       []string
}

// NewRepositoryProcessor creates a new repository processor
func NewRepositoryProcessor(git git.Gitter) *RepositoryProcessor {
	return &RepositoryProcessor{
		git:   git,
		stats: &CloneStats{},
		mutex: &sync.RWMutex{},
	}
}

// ProcessRepository handles the cloning or updating of a single repository
func (rp *RepositoryProcessor) ProcessRepository(repo *scm.Repo, repoNameWithCollisions map[string]bool, hasCollisions bool, repoSlug string, index int) {
	// Update repo slug for collisions if needed
	finalRepoSlug := rp.handleNameCollisions(*repo, repoNameWithCollisions, hasCollisions, repoSlug, index)

	// Set the final host path
	repo.HostPath = rp.buildHostPath(*repo, finalRepoSlug)

	// Handle prune untouched logic
	if rp.shouldPruneUntouched(*repo) {
		return
	}

	// Skip if prune untouched is active (only prune, don't clone)
	if os.Getenv("GHORG_PRUNE_UNTOUCHED") == "true" {
		return
	}

	// Process the repository (clone or update)
	if repoExistsLocally(*repo) {
		rp.handleExistingRepository(*repo)
	} else {
		rp.handleNewRepository(*repo)
	}
}

// handleNameCollisions manages repository name collisions
func (rp *RepositoryProcessor) handleNameCollisions(repo scm.Repo, repoNameWithCollisions map[string]bool, hasCollisions bool, repoSlug string, index int) string {
	if !hasCollisions {
		return rp.addSuffixesIfNeeded(repo, repoSlug)
	}

	rp.mutex.Lock()
	var inHash bool
	if repo.IsGitLabSnippet && !repo.IsGitLabRootLevelSnippet {
		inHash = repoNameWithCollisions[repo.GitLabSnippetInfo.NameOfRepo]
	} else {
		inHash = repoNameWithCollisions[repo.Name]
	}
	rp.mutex.Unlock()

	if inHash {
		repoSlug = trimCollisionFilename(strings.Replace(repo.Path, string(os.PathSeparator), "_", -1))
		repoSlug = rp.addSuffixesIfNeeded(repo, repoSlug)

		rp.mutex.Lock()
		slugCollision := repoNameWithCollisions[repoSlug]
		rp.mutex.Unlock()

		if slugCollision {
			repoSlug = fmt.Sprintf("_%v_%v", strconv.Itoa(index), repoSlug)
		} else {
			rp.mutex.Lock()
			repoNameWithCollisions[repoSlug] = true
			rp.mutex.Unlock()
		}
	}

	return rp.addSuffixesIfNeeded(repo, repoSlug)
}

// addSuffixesIfNeeded adds appropriate suffixes for wikis and snippets
func (rp *RepositoryProcessor) addSuffixesIfNeeded(repo scm.Repo, repoSlug string) string {
	if repo.IsWiki && !strings.HasSuffix(repoSlug, ".wiki") {
		repoSlug = repoSlug + ".wiki"
	}

	if repo.IsGitLabSnippet && !repo.IsGitLabRootLevelSnippet && !strings.HasSuffix(repoSlug, ".snippets") {
		repoSlug = repoSlug + ".snippets"
	}

	return repoSlug
}

// buildHostPath constructs the final host path for the repository
func (rp *RepositoryProcessor) buildHostPath(repo scm.Repo, repoSlug string) string {
	if repo.IsGitLabRootLevelSnippet {
		return filepath.Join(outputDirAbsolutePath, "_ghorg_root_level_snippets", repo.GitLabSnippetInfo.Title+"-"+repo.GitLabSnippetInfo.ID)
	}

	if repo.IsGitLabSnippet {
		return filepath.Join(outputDirAbsolutePath, repoSlug, repo.GitLabSnippetInfo.Title+"-"+repo.GitLabSnippetInfo.ID)
	}

	return filepath.Join(outputDirAbsolutePath, repoSlug)
}

// shouldPruneUntouched determines if a repository should be pruned as untouched
func (rp *RepositoryProcessor) shouldPruneUntouched(repo scm.Repo) bool {
	if os.Getenv("GHORG_PRUNE_UNTOUCHED") != "true" || !repoExistsLocally(repo) {
		return false
	}

	// Fetch and check branches
	rp.git.FetchCloneBranch(repo)

	branches, err := rp.git.Branch(repo)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to list local branches for repository %s: %v", repo.Name, err))
		return false
	}

	// Delete if it has no branches
	if branches == "" {
		rp.untouchedRepos = append(rp.untouchedRepos, repo.HostPath)
		return true
	}

	// Skip if multiple branches
	if len(strings.Split(strings.TrimSpace(branches), "\n")) > 1 {
		return false
	}

	// Check for modified changes
	status, err := rp.git.ShortStatus(repo)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to get short status for repository %s: %v", repo.Name, err))
		return false
	}

	if status != "" {
		return false
	}

	// Check for new commits on the branch that exist locally but not on the remote
	commits, err := rp.git.RevListCompare(repo, "HEAD", "@{u}")
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to get commit differences for repository %s. The repository may be empty or does not have a .git directory. Error: %v", repo.Name, err))
		return false
	}

	if commits != "" {
		return false
	}

	rp.untouchedRepos = append(rp.untouchedRepos, repo.HostPath)
	return true
}

// handleExistingRepository processes repositories that already exist locally
func (rp *RepositoryProcessor) handleExistingRepository(repo scm.Repo) {
	action := "pulling"

	// Set origin with credentials
	err := rp.git.SetOriginWithCredentials(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem setting remote with credentials on: %s Error: %v", repo.Name, err))
		return
	}

	if os.Getenv("GHORG_BACKUP") == "true" {
		rp.handleBackupMode(repo)
		return
	}

	if os.Getenv("GHORG_NO_CLEAN") == "true" {
		rp.handleNoCleanMode(repo)
		return
	}

	// Standard pull mode
	rp.handleStandardPull(repo)

	// Reset origin
	err = rp.git.SetOrigin(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem resetting remote: %s Error: %v", repo.Name, err))
		return
	}

	rp.mutex.Lock()
	rp.stats.PulledCount++
	rp.mutex.Unlock()

	if repo.Commits.CountDiff > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("Success %s %s, branch: %s, new commits: %d", action, repo.URL, repo.CloneBranch, repo.Commits.CountDiff))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("Success %s %s, branch: %s", action, repo.URL, repo.CloneBranch))
	}
}

// handleNewRepository processes repositories that don't exist locally
func (rp *RepositoryProcessor) handleNewRepository(repo scm.Repo) {
	err := rp.git.Clone(repo)

	// Handle wiki clone attempts that might fail
	if err != nil && repo.IsWiki {
		rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone: %s Error: %v", repo.URL, err))
		return
	}

	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to clone: %s Error: %v", repo.URL, err))
		return
	}

	// Checkout specific branch if specified
	if os.Getenv("GHORG_BRANCH") != "" {
		err := rp.git.Checkout(repo)
		if err != nil {
			rp.addInfo(fmt.Sprintf("Could not checkout out %s, branch may not exist or may not have any contents/commits, no changes to: %s Error: %v", repo.CloneBranch, repo.URL, err))
			return
		}
	}

	rp.mutex.Lock()
	rp.stats.CloneCount++
	rp.mutex.Unlock()

	// Set origin to remove credentials from URL
	err = rp.git.SetOrigin(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to set remote: %s Error: %v", repo.URL, err))
		return
	}

	// Fetch all if enabled
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		err = rp.git.FetchAll(repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, err))
			return
		}
	}

	colorlog.PrintSuccess(fmt.Sprintf("Success cloning %s, branch: %s", repo.URL, repo.CloneBranch))
}

// handleBackupMode processes repositories in backup mode
func (rp *RepositoryProcessor) handleBackupMode(repo scm.Repo) {
	err := rp.git.UpdateRemote(repo)

	if err != nil && repo.IsWiki {
		rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone on: %s Error: %v", repo.URL, err))
		return
	}

	if err != nil {
		rp.addError(fmt.Sprintf("Could not update remotes: %s Error: %v", repo.URL, err))
		return
	}

	rp.mutex.Lock()
	rp.stats.UpdateRemoteCount++
	rp.mutex.Unlock()
}

// handleNoCleanMode processes repositories in no-clean mode
func (rp *RepositoryProcessor) handleNoCleanMode(repo scm.Repo) {
	err := rp.git.FetchAll(repo)

	if err != nil && repo.IsWiki {
		rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone on: %s Error: %v", repo.URL, err))
		return
	}

	if err != nil {
		rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, err))
		return
	}

	// Increment pulled count for no-clean mode
	rp.mutex.Lock()
	rp.stats.PulledCount++
	rp.mutex.Unlock()
}

// handleStandardPull processes repositories in standard pull mode
func (rp *RepositoryProcessor) handleStandardPull(repo scm.Repo) {
	// Fetch all if enabled
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		err := rp.git.FetchAll(repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, err))
			return
		}
	}

	// Checkout branch
	err := rp.git.Checkout(repo)
	if err != nil {
		rp.git.FetchCloneBranch(repo)

		// Retry checkout
		errRetry := rp.git.Checkout(repo)
		if errRetry != nil {
			hasRemoteHeads, errHasRemoteHeads := rp.git.HasRemoteHeads(repo)
			if errHasRemoteHeads != nil {
				rp.addError(fmt.Sprintf("Could not checkout %s, branch may not exist or may not have any contents/commits, no changes made on: %s Errors: %v %v", repo.CloneBranch, repo.URL, errRetry, errHasRemoteHeads))
				return
			}
			if hasRemoteHeads {
				rp.addError(fmt.Sprintf("Could not checkout %s, branch may not exist or may not have any contents/commits, no changes made on: %s Error: %v", repo.CloneBranch, repo.URL, errRetry))
				return
			} else {
				rp.addInfo(fmt.Sprintf("Could not checkout %s due to repository being empty, no changes made on: %s", repo.CloneBranch, repo.URL))
				return
			}
		}
	}

	// Get pre-pull commit count
	count, err := rp.git.RepoCommitCount(repo)
	if err != nil {
		rp.addInfo(fmt.Sprintf("Problem trying to get pre pull commit count for on repo: %s", repo.URL))
	}
	repo.Commits.CountPrePull = count

	// Clean
	err = rp.git.Clean(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem running git clean: %s Error: %v", repo.URL, err))
		return
	}

	// Reset
	err = rp.git.Reset(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem resetting branch: %s for: %s Error: %v", repo.CloneBranch, repo.URL, err))
		return
	}

	// Pull
	err = rp.git.Pull(repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to pull branch: %v for: %s Error: %v", repo.CloneBranch, repo.URL, err))
		return
	}

	// Get post-pull commit count
	count, err = rp.git.RepoCommitCount(repo)
	if err != nil {
		rp.addInfo(fmt.Sprintf("Problem trying to get post pull commit count for on repo: %s", repo.URL))
	}

	repo.Commits.CountPostPull = count
	repo.Commits.CountDiff = (repo.Commits.CountPostPull - repo.Commits.CountPrePull)

	rp.mutex.Lock()
	rp.stats.NewCommits += repo.Commits.CountDiff
	rp.mutex.Unlock()
}

// addError adds an error to the stats in a thread-safe manner
func (rp *RepositoryProcessor) addError(msg string) {
	rp.mutex.Lock()
	rp.stats.CloneErrors = append(rp.stats.CloneErrors, msg)
	rp.mutex.Unlock()
}

// addInfo adds an info message to the stats in a thread-safe manner
func (rp *RepositoryProcessor) addInfo(msg string) {
	rp.mutex.Lock()
	rp.stats.CloneInfos = append(rp.stats.CloneInfos, msg)
	rp.mutex.Unlock()
}

// GetStats returns a copy of the current statistics
func (rp *RepositoryProcessor) GetStats() CloneStats {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	return CloneStats{
		CloneCount:        rp.stats.CloneCount,
		PulledCount:       rp.stats.PulledCount,
		UpdateRemoteCount: rp.stats.UpdateRemoteCount,
		NewCommits:        rp.stats.NewCommits,
		UntouchedPrunes:   rp.stats.UntouchedPrunes,
		CloneInfos:        append([]string(nil), rp.stats.CloneInfos...),
		CloneErrors:       append([]string(nil), rp.stats.CloneErrors...),
	}
}

// GetUntouchedRepos returns the list of untouched repositories
func (rp *RepositoryProcessor) GetUntouchedRepos() []string {
	return rp.untouchedRepos
}
