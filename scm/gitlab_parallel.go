package scm

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/gabrie30/ghorg/colorlog"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// fetchTopLevelGroupsParallel fetches remaining pages of top-level groups concurrently
func (c Gitlab) fetchTopLevelGroupsParallel(firstPageGroups []*gitlab.Group, totalPages int) ([]string, error) {
	// Create slice to hold all group IDs
	allGroups := make([]string, 0, len(firstPageGroups)*totalPages)

	// Add first page groups
	for _, g := range firstPageGroups {
		allGroups = append(allGroups, strconv.FormatInt(int64(g.ID), 10))
	}

	// Channel to collect results from parallel fetches
	type pageResult struct {
		groups []*gitlab.Group
		err    error
		page   int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through totalPages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &gitlab.ListGroupsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    pageNum,
				},
				TopLevelOnly: &[]bool{true}[0],
				AllAvailable: &[]bool{true}[0],
			}

			groups, _, err := c.Client.Groups.ListGroups(opt)
			resultChan <- pageResult{groups: groups, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*gitlab.Group, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.groups
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if groups, ok := pageResults[page]; ok {
			for _, g := range groups {
				allGroups = append(allGroups, strconv.FormatInt(int64(g.ID), 10))
			}
		}
	}

	return allGroups, nil
}

// fetchGroupReposParallel fetches remaining pages of group projects concurrently
func (c Gitlab) fetchGroupReposParallel(targetGroup string, firstPageProjects []*gitlab.Project, totalPages int) ([]Repo, error) {
	// Create slice to hold all repos
	repoData := make([]Repo, 0, len(firstPageProjects)*totalPages)
	repoData = append(repoData, c.filter(targetGroup, firstPageProjects)...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		projects []*gitlab.Project
		err      error
		page     int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through totalPages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &gitlab.ListGroupProjectsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    pageNum,
				},
				IncludeSubGroups: gitlab.Ptr(true),
			}

			ps, _, err := c.Groups.ListGroupProjects(targetGroup, opt)
			resultChan <- pageResult{projects: ps, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*gitlab.Project, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.projects
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if projects, ok := pageResults[page]; ok {
			repoData = append(repoData, c.filter(targetGroup, projects)...)
		}
	}

	return repoData, nil
}

// fetchUserProjectsParallel fetches remaining pages of user projects concurrently
func (c Gitlab) fetchUserProjectsParallel(targetUser string, firstPageProjects []*gitlab.Project, totalPages int) ([]Repo, error) {
	// Create slice to hold all repos
	repoData := make([]Repo, 0, len(firstPageProjects)*totalPages)

	// Filter first page
	repoData = append(repoData, c.filter(targetUser, firstPageProjects)...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		projects []*gitlab.Project
		err      error
		page     int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through totalPages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			projectOpts := &gitlab.ListProjectsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    pageNum,
				},
			}

			projects, _, err := c.Projects.ListProjects(projectOpts, gitlab.WithContext(context.Background()))
			resultChan <- pageResult{projects: projects, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*gitlab.Project, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		pageResults[result.page] = result.projects
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if projects, ok := pageResults[page]; ok {
			repoData = append(repoData, c.filter(targetUser, projects)...)
		}
	}

	return repoData, nil
}

// fetchRepoSnippetsParallel fetches remaining pages of repository snippets concurrently
func (c Gitlab) fetchRepoSnippetsParallel(repoID string, firstPageSnippets []*gitlab.Snippet, totalPages int) []*gitlab.Snippet {
	// Create slice to hold all snippets
	allSnippets := make([]*gitlab.Snippet, 0, len(firstPageSnippets)*totalPages)
	allSnippets = append(allSnippets, firstPageSnippets...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		snippets []*gitlab.Snippet
		err      error
		page     int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through totalPages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &gitlab.ListProjectSnippetsOptions{
				PerPage: perPage,
				Page:    pageNum,
			}

			snippets, _, err := c.ProjectSnippets.ListSnippets(repoID, opt)
			resultChan <- pageResult{snippets: snippets, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*gitlab.Snippet, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			// For snippets, we log errors but continue
			colorlog.PrintError(fmt.Sprintf("Error fetching snippet page %d: %v", result.page, result.err))
			continue
		}
		pageResults[result.page] = result.snippets
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if snippets, ok := pageResults[page]; ok {
			allSnippets = append(allSnippets, snippets...)
		}
	}

	return allSnippets
}

// fetchAllSnippetsParallel fetches remaining pages of all snippets concurrently
func (c Gitlab) fetchAllSnippetsParallel(firstPageSnippets []*gitlab.Snippet, totalPages int) []*gitlab.Snippet {
	// Create slice to hold all snippets
	allSnippets := make([]*gitlab.Snippet, 0, len(firstPageSnippets)*totalPages)
	allSnippets = append(allSnippets, firstPageSnippets...)

	// Channel to collect results from parallel fetches
	type pageResult struct {
		snippets []*gitlab.Snippet
		err      error
		page     int
	}
	resultChan := make(chan pageResult, totalPages-1)

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Fetch pages 2 through totalPages in parallel
	for page := 2; page <= totalPages; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			opt := &gitlab.ListAllSnippetsOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    pageNum,
				},
			}

			snippets, _, err := c.Snippets.ListAllSnippets(opt)
			resultChan <- pageResult{snippets: snippets, err: err, page: pageNum}
		}(page)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, organized by page number for consistent ordering
	pageResults := make(map[int][]*gitlab.Snippet, totalPages-1)
	for result := range resultChan {
		if result.err != nil {
			colorlog.PrintError(fmt.Sprintf("Error fetching all snippets page %d: %v", result.page, result.err))
			continue
		}
		pageResults[result.page] = result.snippets
	}

	// Append results in page order to maintain consistency
	for page := 2; page <= totalPages; page++ {
		if snippets, ok := pageResults[page]; ok {
			allSnippets = append(allSnippets, snippets...)
		}
	}

	return allSnippets
}
