package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/scm"
)

// RepositoryFilter handles filtering of repositories based on various criteria
type RepositoryFilter struct{}

// NewRepositoryFilter creates a new repository filter
func NewRepositoryFilter() *RepositoryFilter {
	return &RepositoryFilter{}
}

// ApplyAllFilters applies all configured filters to the repository list
func (rf *RepositoryFilter) ApplyAllFilters(cloneTargets []scm.Repo) []scm.Repo {
	// Apply regex match filter
	if os.Getenv("GHORG_MATCH_REGEX") != "" {
		colorlog.PrintInfo("Filtering repos down by including regex matches...")
		cloneTargets = rf.FilterByRegexMatch(cloneTargets)
	}

	// Apply exclude regex match filter
	if os.Getenv("GHORG_EXCLUDE_MATCH_REGEX") != "" {
		colorlog.PrintInfo("Filtering repos down by excluding regex matches...")
		cloneTargets = rf.FilterByExcludeRegexMatch(cloneTargets)
	}

	// Apply prefix match filter
	if os.Getenv("GHORG_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("Filtering repos down by including prefix matches...")
		cloneTargets = rf.FilterByMatchPrefix(cloneTargets)
	}

	// Apply exclude prefix match filter
	if os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("Filtering repos down by excluding prefix matches...")
		cloneTargets = rf.FilterByExcludeMatchPrefix(cloneTargets)
	}

	// Apply target repos path filter
	if os.Getenv("GHORG_TARGET_REPOS_PATH") != "" {
		colorlog.PrintInfo("Filtering repos down by target repos path...")
		cloneTargets = rf.FilterByTargetReposPath(cloneTargets)
	}

	// Apply ghorgonly filter (must be applied before ghorgignore)
	cloneTargets = rf.FilterByGhorgonly(cloneTargets)

	// Apply ghorgignore filter
	cloneTargets = rf.FilterByGhorgignore(cloneTargets)

	return cloneTargets
}

// FilterByRegexMatch filters repositories that match the regex pattern
func (rf *RepositoryFilter) FilterByRegexMatch(repos []scm.Repo) []scm.Repo {
	regex := os.Getenv("GHORG_MATCH_REGEX")
	if regex == "" {
		return repos
	}

	filteredRepos := []scm.Repo{}
	re := regexp.MustCompile(regex)

	for _, repo := range repos {
		if re.FindString(repo.Name) != "" {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos
}

// FilterByExcludeRegexMatch filters out repositories that match the regex pattern
func (rf *RepositoryFilter) FilterByExcludeRegexMatch(repos []scm.Repo) []scm.Repo {
	regex := os.Getenv("GHORG_EXCLUDE_MATCH_REGEX")
	if regex == "" {
		return repos
	}

	filteredRepos := []scm.Repo{}
	re := regexp.MustCompile(regex)

	for _, repo := range repos {
		if re.FindString(repo.Name) == "" {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos
}

// FilterByMatchPrefix filters repositories that start with the specified prefix(es)
func (rf *RepositoryFilter) FilterByMatchPrefix(repos []scm.Repo) []scm.Repo {
	prefixes := os.Getenv("GHORG_MATCH_PREFIX")
	if prefixes == "" {
		return repos
	}

	filteredRepos := []scm.Repo{}
	prefixList := strings.Split(prefixes, ",")

	for _, repo := range repos {
		for _, prefix := range prefixList {
			if strings.HasPrefix(strings.ToLower(repo.Name), strings.ToLower(prefix)) {
				filteredRepos = append(filteredRepos, repo)
				break
			}
		}
	}

	return filteredRepos
}

// FilterByExcludeMatchPrefix filters out repositories that start with the specified prefix(es)
func (rf *RepositoryFilter) FilterByExcludeMatchPrefix(repos []scm.Repo) []scm.Repo {
	prefixes := os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX")
	if prefixes == "" {
		return repos
	}

	filteredRepos := []scm.Repo{}
	prefixList := strings.Split(prefixes, ",")

	for _, repo := range repos {
		exclude := false
		for _, prefix := range prefixList {
			if strings.HasPrefix(strings.ToLower(repo.Name), strings.ToLower(prefix)) {
				exclude = true
				break
			}
		}
		if !exclude {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos
}

// FilterByTargetReposPath filters repositories based on a file containing target repo names
func (rf *RepositoryFilter) FilterByTargetReposPath(cloneTargets []scm.Repo) []scm.Repo {
	targetReposPath := os.Getenv("GHORG_TARGET_REPOS_PATH")
	if targetReposPath == "" {
		return cloneTargets
	}

	_, err := os.Stat(targetReposPath)
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error finding your GHORG_TARGET_REPOS_PATH file, error: %v", err))
	}

	// Read target repos from file
	toTarget, err := readTargetReposFile()
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error parsing your GHORG_TARGET_REPOS_PATH file, error: %v", err))
	}

	colorlog.PrintInfo("Using GHORG_TARGET_REPOS_PATH, filtering repos down...")

	filteredCloneTargets := []scm.Repo{}
	targetRepoSeenOnOrg := make(map[string]bool)

	for _, cloneTarget := range cloneTargets {
		found := false
		for _, targetRepo := range toTarget {
			if _, ok := targetRepoSeenOnOrg[targetRepo]; !ok {
				targetRepoSeenOnOrg[targetRepo] = false
			}

			clonedRepoName := strings.TrimSuffix(filepath.Base(cloneTarget.URL), ".git")
			if strings.EqualFold(clonedRepoName, targetRepo) {
				found = true
				targetRepoSeenOnOrg[targetRepo] = true
			}

			// Handle wiki matching
			if os.Getenv("GHORG_CLONE_WIKI") == "true" {
				targetRepoWiki := targetRepo + ".wiki"
				if strings.EqualFold(targetRepoWiki, clonedRepoName) {
					found = true
					targetRepoSeenOnOrg[targetRepo] = true
				}
			}

			// Handle snippet matching
			if os.Getenv("GHORG_CLONE_SNIPPETS") == "true" && cloneTarget.IsGitLabSnippet {
				targetSnippetOriginalRepo := strings.TrimSuffix(filepath.Base(cloneTarget.GitLabSnippetInfo.URLOfRepo), ".git")
				if strings.EqualFold(targetSnippetOriginalRepo, targetRepo) {
					found = true
					targetRepoSeenOnOrg[targetRepo] = true
				}
			}
		}

		if found {
			filteredCloneTargets = append(filteredCloneTargets, cloneTarget)
		}
	}

	// Print repos from the file that were not found in the org
	for targetRepo, seen := range targetRepoSeenOnOrg {
		if !seen {
			msg := fmt.Sprintf("Target in GHORG_TARGET_REPOS_PATH was not found in the org, repo: %v", targetRepo)
			cloneInfos = append(cloneInfos, msg)
		}
	}

	return filteredCloneTargets
}

// FilterByGhorgonly filters repositories to only include those matching patterns in ghorgonly file
func (rf *RepositoryFilter) FilterByGhorgonly(cloneTargets []scm.Repo) []scm.Repo {
	onlyLocation := os.Getenv("GHORG_ONLY_PATH")
	if onlyLocation != "" {
		_, err := os.Stat(onlyLocation)
		if os.IsNotExist(err) {
			return cloneTargets
		}
	} else {
		// Use default location
		defaultOnlyPath := filepath.Join(os.Getenv("HOME"), ".config", "ghorg", "ghorgonly")
		_, err := os.Stat(defaultOnlyPath)
		if os.IsNotExist(err) {
			return cloneTargets
		}
	}

	// Read ghorgonly patterns
	toInclude, err := readGhorgOnly()
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error parsing your ghorgonly, error: %v", err))
	}

	colorlog.PrintInfo("Using ghorgonly, filtering repos down...")

	filteredCloneTargets := []scm.Repo{}
	for _, repo := range cloneTargets {
		included := false
		for _, includePattern := range toInclude {
			if strings.Contains(repo.URL, includePattern) {
				included = true
				break
			}
		}
		if included {
			filteredCloneTargets = append(filteredCloneTargets, repo)
		}
	}

	return filteredCloneTargets
}

// FilterByGhorgignore filters out repositories listed in the ghorgignore file
func (rf *RepositoryFilter) FilterByGhorgignore(cloneTargets []scm.Repo) []scm.Repo {
	ignoreLocation := os.Getenv("GHORG_IGNORE_PATH")
	if ignoreLocation != "" {
		_, err := os.Stat(ignoreLocation)
		if os.IsNotExist(err) {
			return cloneTargets
		}
	} else {
		// Use default location
		defaultIgnorePath := filepath.Join(os.Getenv("HOME"), ".config", "ghorg", "ghorgignore")
		_, err := os.Stat(defaultIgnorePath)
		if os.IsNotExist(err) {
			return cloneTargets
		}
	}

	// Read ghorgignore patterns
	toIgnore, err := readGhorgIgnore()
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error parsing your ghorgignore, error: %v", err))
	}

	colorlog.PrintInfo("Using ghorgignore, filtering repos down...")

	filteredCloneTargets := []scm.Repo{}
	for _, repo := range cloneTargets {
		ignored := false
		for _, ignorePattern := range toIgnore {
			if strings.Contains(repo.URL, ignorePattern) {
				ignored = true
				break
			}
		}
		if !ignored {
			filteredCloneTargets = append(filteredCloneTargets, repo)
		}
	}

	return filteredCloneTargets
}
