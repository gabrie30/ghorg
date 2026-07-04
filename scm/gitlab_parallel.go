package scm

import (
	"os"
	"strconv"
	"sync"

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
					PerPage: int64(perPage),
					Page:    int64(pageNum),
				},
				TopLevelOnly: &[]bool{true}[0],
				AllAvailable: &[]bool{true}[0],
			}

			groups, _, err := c.Groups.ListGroups(opt)
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
					PerPage: int64(perPage),
					Page:    int64(pageNum),
				},
				IncludeSubGroups: gitlab.Ptr(true),
				WithShared:       gitlab.Ptr(os.Getenv("GHORG_GITLAB_INCLUDE_SHARED_PROJECTS") != "false"),
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
