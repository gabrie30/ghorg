package scm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// fetchServerProjectReposParallel fetches remaining pages of Bitbucket Server repos concurrently
func (c Bitbucket) fetchServerProjectReposParallel(projectKey string, firstPageRepos []ServerRepository, totalSize int, limit int) ([]Repo, error) {
	// Calculate total number of pages
	totalPages := (totalSize + limit - 1) / limit // Ceiling division

	// Filter first page
	repoData := make([]Repo, 0, totalSize)
	repoData = append(repoData, c.filterServerRepos(firstPageRepos)...)

	// If only one page, return immediately
	if totalPages <= 1 {
		return repoData, nil
	}

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []ServerRepository
		err   error
		page  int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch remaining pages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			start := (pageNum - 1) * limit
			apiURL := strings.TrimSuffix(c.serverURL, "/") + fmt.Sprintf("/rest/api/1.0/projects/%s/repos?start=%d&limit=%d", projectKey, start, limit)

			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			req.SetBasicAuth(c.username, c.password)
			resp, err := c.httpClient.Do(req)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resultChan <- pageResult{err: fmt.Errorf("failed to fetch repos for project %s: %s", projectKey, string(body)), page: pageNum}
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			var projectResp ServerProjectResponse
			if err := json.Unmarshal(body, &projectResp); err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			resultChan <- pageResult{repos: projectResp.Values, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]ServerRepository, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if repos, ok := pageResults[page]; ok {
			repoData = append(repoData, c.filterServerRepos(repos)...)
		}
	}

	return repoData, nil
}

// fetchServerUserReposParallel fetches remaining pages of Bitbucket Server user repos concurrently
func (c Bitbucket) fetchServerUserReposParallel(username string, firstPageRepos []ServerRepository, totalSize int, limit int) ([]Repo, error) {
	// Calculate total number of pages
	totalPages := (totalSize + limit - 1) / limit // Ceiling division

	// Filter first page
	repoData := make([]Repo, 0, totalSize)
	repoData = append(repoData, c.filterServerRepos(firstPageRepos)...)

	// If only one page, return immediately
	if totalPages <= 1 {
		return repoData, nil
	}

	// Channel to collect results from parallel fetches
	type pageResult struct {
		repos []ServerRepository
		err   error
		page  int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch remaining pages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			start := (pageNum - 1) * limit
			apiURL := strings.TrimSuffix(c.serverURL, "/") + fmt.Sprintf("/rest/api/1.0/repos?start=%d&limit=%d", start, limit)

			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			req.SetBasicAuth(c.username, c.password)
			resp, err := c.httpClient.Do(req)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resultChan <- pageResult{err: fmt.Errorf("failed to fetch user repos: %s", string(body)), page: pageNum}
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			var projectResp ServerProjectResponse
			if err := json.Unmarshal(body, &projectResp); err != nil {
				resultChan <- pageResult{err: err, page: pageNum}
				return
			}

			resultChan <- pageResult{repos: projectResp.Values, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]ServerRepository, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.repos
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if repos, ok := pageResults[page]; ok {
			repoData = append(repoData, c.filterServerRepos(repos)...)
		}
	}

	return repoData, nil
}
