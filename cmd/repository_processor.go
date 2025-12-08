package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/git"
	"github.com/gabrie30/ghorg/scm"
)

// Helper function to apply clone delay if configured
func applyCloneDelay(repoURL string) {
	delaySeconds, hasDelay := getCloneDelaySeconds()
	if !hasDelay {
		return
	}

	if os.Getenv("GHORG_DEBUG") != "" {
		colorlog.PrintInfo(fmt.Sprintf("Applying %d second delay before processing %s", delaySeconds, repoURL))
	}
	time.Sleep(time.Duration(delaySeconds) * time.Second)
}

// RepositoryProcessor handles the processing of individual repositories
type RepositoryProcessor struct {
	git            git.Gitter
	stats          *CloneStats
	mutex          *sync.RWMutex
	untouchedRepos []string
}

// CloneStats tracks statistics during clone operations
type CloneStats struct {
	CloneCount           int
	PulledCount          int
	UpdateRemoteCount    int
	NewCommits           int
	UntouchedPrunes      int
	TotalDurationSeconds int
	CloneInfos           []string
	CloneErrors          []string
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
	if rp.shouldPruneUntouched(repo) {
		return
	}

	// Skip if prune untouched is active (only prune, don't clone)
	if os.Getenv("GHORG_PRUNE_UNTOUCHED") == "true" {
		return
	}

	// Apply clone delay if configured (before any repository operations)
	applyCloneDelay(repo.URL)

	// Determine if this repo exists locally
	repoWillBePulled := repoExistsLocally(*repo)
	var action string

	// Process the repository (clone or update)
	if repoWillBePulled {
		success := rp.handleExistingRepository(repo, &action)
		if !success {
			return
		}
	} else {
		success := rp.handleNewRepository(repo, &action)
		if !success {
			return
		}
	}

	// Print unified success message (matching original behavior)
	if repoWillBePulled && repo.Commits.CountDiff > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("Success %s %s, branch: %s, new commits: %d", action, repo.URL, repo.CloneBranch, repo.Commits.CountDiff))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("Success %s %s, branch: %s", action, repo.URL, repo.CloneBranch))
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
		// Replace both forward slashes and backslashes with underscores for cross-platform compatibility
		pathWithUnderscores := strings.ReplaceAll(repo.Path, "/", "_")
		pathWithUnderscores = strings.ReplaceAll(pathWithUnderscores, "\\", "_")
		repoSlug = trimCollisionFilename(pathWithUnderscores)
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
func (rp *RepositoryProcessor) shouldPruneUntouched(repo *scm.Repo) bool {
	if os.Getenv("GHORG_PRUNE_UNTOUCHED") != "true" || !repoExistsLocally(*repo) {
		return false
	}

	// Fetch and check branches
	rp.git.FetchCloneBranch(*repo)

	branches, err := rp.git.Branch(*repo)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to list local branches for repository %s: %v", repo.Name, err))
		return false
	}

	// Delete if it has no branches
	if branches == "" {
		rp.mutex.Lock()
		rp.untouchedRepos = append(rp.untouchedRepos, repo.HostPath)
		rp.mutex.Unlock()
		return true
	}

	// Skip if multiple branches
	if len(strings.Split(strings.TrimSpace(branches), "\n")) > 1 {
		return false
	}

	// Check for modified changes
	status, err := rp.git.ShortStatus(*repo)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to get short status for repository %s: %v", repo.Name, err))
		return false
	}

	if status != "" {
		return false
	}

	// Check for new commits on the branch that exist locally but not on the remote
	commits, err := rp.git.RevListCompare(*repo, "HEAD", "@{u}")
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Failed to get commit differences for repository %s. The repository may be empty or does not have a .git directory. Error: %v", repo.Name, err))
		return false
	}

	if commits != "" {
		return false
	}

	rp.mutex.Lock()
	rp.untouchedRepos = append(rp.untouchedRepos, repo.HostPath)
	rp.mutex.Unlock()
	return true
}

// handleExistingRepository processes repositories that already exist locally
func (rp *RepositoryProcessor) handleExistingRepository(repo *scm.Repo, action *string) bool {
	*action = "pulling"

	// Set origin with credentials
	err := rp.git.SetOriginWithCredentials(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem setting remote with credentials on: %s Error: %v", repo.Name, err))
		return false
	}

	var success bool
	if os.Getenv("GHORG_BACKUP") == "true" {
		*action = "updating remote"
		success = rp.handleBackupMode(repo)
	} else if os.Getenv("GHORG_NO_CLEAN") == "true" {
		*action = "fetching"
		success = rp.handleNoCleanMode(repo)
	} else {
		// Standard pull mode
		success = rp.handleStandardPull(repo)
	}

	// Always reset origin to remove credentials, even if processing failed
	err = rp.git.SetOrigin(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem resetting remote: %s Error: %v", repo.Name, err))
		return false
	}

	// Return success after ensuring tokens are stripped
	if !success {
		return false
	}

	rp.mutex.Lock()
	rp.stats.PulledCount++
	rp.mutex.Unlock()

	return true
}

// handleNewRepository processes repositories that don't exist locally
func (rp *RepositoryProcessor) handleNewRepository(repo *scm.Repo, action *string) bool {
	*action = "cloning"

	err := rp.git.Clone(*repo)

	// Handle wiki clone attempts that might fail
	if err != nil && repo.IsWiki {
		// Create an empty directory for wikis with no content to maintain directory structure consistency
		if mkdirErr := os.MkdirAll(repo.HostPath, os.ModePerm); mkdirErr != nil {
			rp.addError(fmt.Sprintf("Failed to create directory for empty wiki: %s Error: %v", repo.HostPath, mkdirErr))
			return false
		}
		rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone: %s Error: %v", repo.URL, err))
		// Return true to indicate we've handled this successfully (directory created)
		// Skip the rest of the processing since there's no actual repository
		return true
	}

	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to clone: %s Error: %v", repo.URL, err))
		return false
	}

	// Checkout specific branch if specified
	if os.Getenv("GHORG_BRANCH") != "" {
		err := rp.git.Checkout(*repo)
		if err != nil {
			rp.addInfo(fmt.Sprintf("Could not checkout out %s, branch may not exist or may not have any contents/commits, no changes to: %s Error: %v", repo.CloneBranch, repo.URL, err))
			return false
		}
	}

	rp.mutex.Lock()
	rp.stats.CloneCount++
	rp.mutex.Unlock()

	// Set origin to remove credentials from URL
	err = rp.git.SetOrigin(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to set remote: %s Error: %v", repo.URL, err))
		return false
	}

	// Fetch all if enabled
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		// Temporarily restore credentials for fetch-all to work with private repos
		err = rp.git.SetOriginWithCredentials(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to set remote with credentials: %s Error: %v", repo.URL, err))
			return false
		}

		err = rp.git.FetchAll(*repo)
		fetchErr := err // Store fetch error for later reporting

		// Always strip credentials again for security, even if fetch failed
		err = rp.git.SetOrigin(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to reset remote after fetch: %s Error: %v", repo.URL, err))
			return false
		}

		// Report fetch error if it occurred
		if fetchErr != nil {
			rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, fetchErr))
			return false
		}
	}

	return true
}

// handleBackupMode processes repositories in backup mode
func (rp *RepositoryProcessor) handleBackupMode(repo *scm.Repo) bool {
	err := rp.git.UpdateRemote(*repo)

	if err != nil && repo.IsWiki {
		rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone on: %s Error: %v", repo.URL, err))
		return false
	}

	if err != nil {
		rp.addError(fmt.Sprintf("Could not update remotes: %s Error: %v", repo.URL, err))
		return false
	}

	rp.mutex.Lock()
	rp.stats.UpdateRemoteCount++
	rp.mutex.Unlock()

	return true
}

// handleNoCleanMode processes repositories in no-clean mode
func (rp *RepositoryProcessor) handleNoCleanMode(repo *scm.Repo) bool {
	// Fetch all if enabled
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		// Temporarily restore credentials for fetch-all to work with private repos
		err := rp.git.SetOriginWithCredentials(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to set remote with credentials: %s Error: %v", repo.URL, err))
			return false
		}

		err = rp.git.FetchAll(*repo)
		fetchErr := err // Store fetch error for later reporting

		// Always strip credentials again for security, even if fetch failed
		err = rp.git.SetOrigin(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to reset remote after fetch: %s Error: %v", repo.URL, err))
			return false
		}

		if fetchErr != nil && repo.IsWiki {
			rp.addInfo(fmt.Sprintf("Wiki may be enabled but there was no content to clone on: %s Error: %v", repo.URL, fetchErr))
			return false
		}

		if fetchErr != nil {
			rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, fetchErr))
			return false
		}
	}

	return true
}

// handleStandardPull processes repositories in standard pull mode
func (rp *RepositoryProcessor) handleStandardPull(repo *scm.Repo) bool {
	// Fetch all if enabled
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		// Temporarily restore credentials for fetch-all to work with private repos
		err := rp.git.SetOriginWithCredentials(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to set remote with credentials: %s Error: %v", repo.URL, err))
			return false
		}

		err = rp.git.FetchAll(*repo)
		fetchErr := err // Store fetch error for later reporting

		// Always strip credentials again for security, even if fetch failed
		err = rp.git.SetOrigin(*repo)
		if err != nil {
			rp.addError(fmt.Sprintf("Problem trying to reset remote after fetch: %s Error: %v", repo.URL, err))
			return false
		}

		// Report fetch error if it occurred
		if fetchErr != nil {
			rp.addError(fmt.Sprintf("Could not fetch remotes: %s Error: %v", repo.URL, fetchErr))
			return false
		}
	}

	// Checkout branch
	err := rp.git.Checkout(*repo)
	if err != nil {
		rp.git.FetchCloneBranch(*repo)

		// Retry checkout
		errRetry := rp.git.Checkout(*repo)
		if errRetry != nil {
			hasRemoteHeads, errHasRemoteHeads := rp.git.HasRemoteHeads(*repo)
			if errHasRemoteHeads != nil {
				rp.addError(fmt.Sprintf("Could not checkout %s, branch may not exist or may not have any contents/commits, no changes made on: %s Errors: %v %v", repo.CloneBranch, repo.URL, errRetry, errHasRemoteHeads))
				return false
			}
			if hasRemoteHeads {
				rp.addError(fmt.Sprintf("Could not checkout %s, branch may not exist or may not have any contents/commits, no changes made on: %s Error: %v", repo.CloneBranch, repo.URL, errRetry))
				return false
			} else {
				rp.addInfo(fmt.Sprintf("Could not checkout %s due to repository being empty, no changes made on: %s", repo.CloneBranch, repo.URL))
				return false
			}
		}
	}

	// Get pre-pull commit count
	count, err := rp.git.RepoCommitCount(*repo)
	if err != nil {
		rp.addInfo(fmt.Sprintf("Problem trying to get pre pull commit count for on repo: %s", repo.URL))
	}
	repo.Commits.CountPrePull = count

	// Clean
	err = rp.git.Clean(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem running git clean: %s Error: %v", repo.URL, err))
		return false
	}

	// Reset
	err = rp.git.Reset(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem resetting branch: %s for: %s Error: %v", repo.CloneBranch, repo.URL, err))
		return false
	}

	// Pull
	err = rp.git.Pull(*repo)
	if err != nil {
		rp.addError(fmt.Sprintf("Problem trying to pull branch: %v for: %s Error: %v", repo.CloneBranch, repo.URL, err))
		return false
	}

	// Get post-pull commit count
	count, err = rp.git.RepoCommitCount(*repo)
	if err != nil {
		rp.addInfo(fmt.Sprintf("Problem trying to get post pull commit count for on repo: %s", repo.URL))
	}

	repo.Commits.CountPostPull = count
	repo.Commits.CountDiff = (repo.Commits.CountPostPull - repo.Commits.CountPrePull)

	rp.mutex.Lock()
	rp.stats.NewCommits += repo.Commits.CountDiff
	rp.mutex.Unlock()

	return true
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
		CloneCount:           rp.stats.CloneCount,
		PulledCount:          rp.stats.PulledCount,
		UpdateRemoteCount:    rp.stats.UpdateRemoteCount,
		NewCommits:           rp.stats.NewCommits,
		UntouchedPrunes:      rp.stats.UntouchedPrunes,
		TotalDurationSeconds: rp.stats.TotalDurationSeconds,
		CloneInfos:           append([]string(nil), rp.stats.CloneInfos...),
		CloneErrors:          append([]string(nil), rp.stats.CloneErrors...),
	}
}

// GetUntouchedRepos returns the list of untouched repositories
func (rp *RepositoryProcessor) GetUntouchedRepos() []string {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()
	// Return a copy to prevent external modifications
	return append([]string(nil), rp.untouchedRepos...)
}

// SetTotalDuration sets the total duration in seconds for the clone operation
func (rp *RepositoryProcessor) SetTotalDuration(durationSeconds int) {
	rp.mutex.Lock()
	rp.stats.TotalDurationSeconds = durationSeconds
	rp.mutex.Unlock()
}
