package scm

import (
	"fmt"
	"net/http"
	"sync"

	"code.gitea.io/sdk/gitea"
)

// fetchOrgReposParallel fetches remaining pages of org repos concurrently
// Gitea doesn't provide total page count, so we continue until we get a page with fewer than perPage items
func (c Gitea) fetchOrgReposParallel(targetOrg string, firstPageRepos []*gitea.Repository) ([]Repo, error) {
	// Filter first page
	repoData, err := c.filter(firstPageRepos)
	if err != nil {
		return nil, err
	}

	// If first page was not full, we're done
	if len(firstPageRepos) < c.perPage {
		return repoData, nil
	}

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []*gitea.Repository
		resp  *gitea.Response
		err   error
		page  int
	}
	resultChan := make(chan pageResult, 10) // Buffer for up to 10 concurrent pages

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Start fetching pages concurrently
	page := 2
	done := false

	for !done {
		// Fetch next batch of pages (up to 10 at a time to avoid overwhelming the server)
		batchSize := 10
		for i := 0; i < batchSize; i++ {
			wg.Add(1)
			currentPage := page
			page++

			go func(pageNum int) {
				defer wg.Done()

				rps, resp, err := c.ListOrgRepos(targetOrg, gitea.ListOrgReposOptions{ListOptions: gitea.ListOptions{
					Page:     pageNum,
					PageSize: c.perPage,
				}})

				resultChan <- pageResult{repos: rps, resp: resp, err: err, page: pageNum}
			}(currentPage)
		}

		// Wait for this batch to complete
		wg.Wait()
		close(resultChan)

		// Process results from this batch
		results := make(map[int]*pageResult)
		for result := range resultChan {
			if result.err != nil {
				if result.resp != nil && result.resp.StatusCode == http.StatusNotFound {
					return nil, fmt.Errorf("org \"%s\" not found", targetOrg)
				}
				return nil, result.err
			}
			results[result.page] = &result
		}

		// Append results in order
		for p := page - batchSize; p < page; p++ {
			if result, ok := results[p]; ok {
				filtered, err := c.filter(result.repos)
				if err != nil {
					return nil, err
				}
				repoData = append(repoData, filtered...)

				// Check if this was the last page
				if len(result.repos) < c.perPage {
					done = true
					break
				}
			}
		}

		// Reset channel for next batch
		if !done {
			resultChan = make(chan pageResult, 10)
		}
	}

	return repoData, nil
}

// fetchUserReposParallel fetches remaining pages of user repos concurrently
func (c Gitea) fetchUserReposParallel(targetUser string, firstPageRepos []*gitea.Repository) ([]Repo, error) {
	// Filter first page
	repoData, err := c.filter(firstPageRepos)
	if err != nil {
		return nil, err
	}

	// If first page was not full, we're done
	if len(firstPageRepos) < c.perPage {
		return repoData, nil
	}

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []*gitea.Repository
		resp  *gitea.Response
		err   error
		page  int
	}
	resultChan := make(chan pageResult, 10) // Buffer for up to 10 concurrent pages

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Start fetching pages concurrently
	page := 2
	done := false

	for !done {
		// Fetch next batch of pages (up to 10 at a time)
		batchSize := 10
		for i := 0; i < batchSize; i++ {
			wg.Add(1)
			currentPage := page
			page++

			go func(pageNum int) {
				defer wg.Done()

				rps, resp, err := c.ListUserRepos(targetUser, gitea.ListReposOptions{ListOptions: gitea.ListOptions{
					Page:     pageNum,
					PageSize: c.perPage,
				}})

				resultChan <- pageResult{repos: rps, resp: resp, err: err, page: pageNum}
			}(currentPage)
		}

		// Wait for this batch to complete
		wg.Wait()
		close(resultChan)

		// Process results from this batch
		results := make(map[int]*pageResult)
		for result := range resultChan {
			if result.err != nil {
				if result.resp != nil && result.resp.StatusCode == http.StatusNotFound {
					return nil, fmt.Errorf("user \"%s\" not found", targetUser)
				}
				return nil, result.err
			}
			results[result.page] = &result
		}

		// Append results in order
		for p := page - batchSize; p < page; p++ {
			if result, ok := results[p]; ok {
				filtered, err := c.filter(result.repos)
				if err != nil {
					return nil, err
				}
				repoData = append(repoData, filtered...)

				// Check if this was the last page
				if len(result.repos) < c.perPage {
					done = true
					break
				}
			}
		}

		// Reset channel for next batch
		if !done {
			resultChan = make(chan pageResult, 10)
		}
	}

	return repoData, nil
}
