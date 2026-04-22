package scm

import (
	"context"
	"os"
	"sync"

	"github.com/google/go-github/v72/github"
)

// fetchOrgReposParallel fetches remaining pages of org repos with bounded
// concurrency to avoid GitHub secondary (abuse) rate limits.
func (c Github) fetchOrgReposParallel(targetOrg string, firstPageRepos []*github.Repository, lastPage int) ([]Repo, error) {
	allRepos := make([]*github.Repository, 0, len(firstPageRepos)*lastPage)
	allRepos = append(allRepos, firstPageRepos...)

	type pageResult struct {
		repos []*github.Repository
		err   error
		page  int
	}
	extra := lastPage - 1
	if extra < 1 {
		return c.filter(allRepos), nil
	}

	workers := githubListWorkerCount(extra)
	resultChan := make(chan pageResult, extra)
	jobs := make(chan int, extra)
	for p := 2; p <= lastPage; p++ {
		jobs <- p
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pageNum := range jobs {
				repos, err := fetchGitHubRepoPageWithRetry(context.Background(), func(ctx context.Context) ([]*github.Repository, error) {
					opt := &github.RepositoryListByOrgOptions{
						Type:        "all",
						ListOptions: github.ListOptions{PerPage: c.perPage, Page: pageNum},
					}
					r, _, e := c.Repositories.ListByOrg(ctx, targetOrg, opt)
					return r, e
				})
				resultChan <- pageResult{repos: repos, err: err, page: pageNum}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	pageResults := make(map[int][]*github.Repository, extra)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	for page := 2; page <= lastPage; page++ {
		if repos, ok := pageResults[page]; ok {
			allRepos = append(allRepos, repos...)
		}
	}

	return c.filter(allRepos), nil
}

// fetchUserReposParallel fetches remaining pages of user repos with bounded
// concurrency to avoid GitHub secondary (abuse) rate limits.
func (c Github) fetchUserReposParallel(targetUser string, firstPageRepos []*github.Repository, lastPage int) ([]Repo, error) {
	allRepos := make([]*github.Repository, 0, len(firstPageRepos)*lastPage)
	allRepos = append(allRepos, firstPageRepos...)

	type pageResult struct {
		repos []*github.Repository
		err   error
		page  int
	}
	extra := lastPage - 1
	if extra < 1 {
		return c.filter(allRepos), nil
	}

	workers := githubListWorkerCount(extra)
	resultChan := make(chan pageResult, extra)
	jobs := make(chan int, extra)
	for p := 2; p <= lastPage; p++ {
		jobs <- p
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pageNum := range jobs {
				repos, err := fetchGitHubRepoPageWithRetry(context.Background(), func(ctx context.Context) ([]*github.Repository, error) {
					opt := &github.ListOptions{PerPage: c.perPage, Page: pageNum}

					var r []*github.Repository
					var e error

					if targetUser == tokenUsername {
						authOpt := &github.RepositoryListByAuthenticatedUserOptions{
							Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
							ListOptions: *opt,
						}
						r, _, e = c.Repositories.ListByAuthenticatedUser(ctx, authOpt)
					} else {
						userOpt := &github.RepositoryListByUserOptions{
							Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
							ListOptions: *opt,
						}
						r, _, e = c.Repositories.ListByUser(ctx, targetUser, userOpt)
					}
					if e != nil {
						return nil, e
					}

					if targetUser != tokenUsername {
						userRepos := []*github.Repository{}
						for _, repo := range r {
							if repo.Owner != nil && repo.Owner.Type != nil && *repo.Owner.Type == "User" {
								userRepos = append(userRepos, repo)
							}
						}
						r = userRepos
					}
					return r, nil
				})
				resultChan <- pageResult{repos: repos, err: err, page: pageNum}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	pageResults := make(map[int][]*github.Repository, extra)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	for page := 2; page <= lastPage; page++ {
		if repos, ok := pageResults[page]; ok {
			allRepos = append(allRepos, repos...)
		}
	}

	return c.filter(allRepos), nil
}
