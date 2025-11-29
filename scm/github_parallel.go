package scm

import (
	"context"
	"os"
	"sync"

	"github.com/google/go-github/v72/github"
)

// fetchOrgReposParallel fetches remaining pages of org repos concurrently
func (c Github) fetchOrgReposParallel(targetOrg string, firstPageRepos []*github.Repository, lastPage int) ([]Repo, error) {
	// Create slice to hold all repos with capacity for efficiency
	allRepos := make([]*github.Repository, 0, len(firstPageRepos)*lastPage)
	allRepos = append(allRepos, firstPageRepos...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []*github.Repository
		err   error
		page  int
	}
	resultChan := make(chan pageResult, lastPage-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through lastPage in parallel
	for page := 2; page <= lastPage; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &github.RepositoryListByOrgOptions{
				Type:        "all",
				ListOptions: github.ListOptions{PerPage: c.perPage, Page: pageNum},
			}

			repos, _, err := c.Repositories.ListByOrg(context.Background(), targetOrg, opt)
			resultChan <- pageResult{repos: repos, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*github.Repository, lastPage-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= lastPage; page++ {
		if repos, ok := pageResults[page]; ok {
			allRepos = append(allRepos, repos...)
		}
	}

	return c.filter(allRepos), nil
}

// fetchUserReposParallel fetches remaining pages of user repos concurrently
func (c Github) fetchUserReposParallel(targetUser string, firstPageRepos []*github.Repository, lastPage int) ([]Repo, error) {
	// Create slice to hold all repos with capacity for efficiency
	allRepos := make([]*github.Repository, 0, len(firstPageRepos)*lastPage)
	allRepos = append(allRepos, firstPageRepos...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []*github.Repository
		err   error
		page  int
	}
	resultChan := make(chan pageResult, lastPage-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through lastPage in parallel
	for page := 2; page <= lastPage; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &github.ListOptions{PerPage: c.perPage, Page: pageNum}

			var repos []*github.Repository
			var err error

			if targetUser == tokenUsername {
				authOpt := &github.RepositoryListByAuthenticatedUserOptions{
					Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
					ListOptions: *opt,
				}
				repos, _, err = c.Repositories.ListByAuthenticatedUser(context.Background(), authOpt)
			} else {
				userOpt := &github.RepositoryListByUserOptions{
					Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
					ListOptions: *opt,
				}
				repos, _, err = c.Repositories.ListByUser(context.Background(), targetUser, userOpt)
			}

			// Filter user repos if needed
			if targetUser != tokenUsername && err == nil {
				userRepos := []*github.Repository{}
				for _, repo := range repos {
					if repo.Owner != nil && repo.Owner.Type != nil && *repo.Owner.Type == "User" {
						userRepos = append(userRepos, repo)
					}
				}
				repos = userRepos
			}

			resultChan <- pageResult{repos: repos, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*github.Repository, lastPage-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= lastPage; page++ {
		if repos, ok := pageResults[page]; ok {
			allRepos = append(allRepos, repos...)
		}
	}

	return c.filter(allRepos), nil
}
