package scm

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	_Client         = Gitlab{}
	perPage         = 100
	gitLabAllGroups = false
	gitLabAllUsers  = false
)

func init() {
	registerClient(Gitlab{})
}

type Gitlab struct {
	// extend the gitlab client
	*gitlab.Client
}

func (_ Gitlab) GetType() string {
	return "gitlab"
}

func (_ Gitlab) rootLevelSnippet(url string) bool {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	if baseURL != "" {
		customSnippetPattern := regexp.MustCompile(`^` + baseURL + `/-/snippets/\d+$`)
		if customSnippetPattern.MatchString(url) {
			return true
		}
		return false
	} else {
		// cloud instances
		// Check if the URL follows the pattern of a root level snippet
		rootLevelSnippetPattern := regexp.MustCompile(`^https://gitlab\.com/-/snippets/\d+$`)
		if rootLevelSnippetPattern.MatchString(url) {
			return true
		}
		return false
	}
}

// GetOrgRepos fetches repo data from a specific group
func (c Gitlab) GetOrgRepos(targetOrg string) ([]Repo, error) {
	allGroups := []string{}
	repoData := []Repo{}
	longFetch := false

	if targetOrg == "all-users" {
		colorlog.PrintErrorAndExit("When using the 'all-users' keyword the '--clone-type=user' flag should be set")
	}

	spinningSpinner.Start()
	defer spinningSpinner.Stop()

	if targetOrg == "all-groups" {
		gitLabAllGroups = true
		longFetch = true

		grps, err := c.GetTopLevelGroups()
		if err != nil {
			return nil, fmt.Errorf("error getting groups error: %v", err)
		}

		allGroups = append(allGroups, grps...)

	} else {
		allGroups = append(allGroups, targetOrg)
	}

	if os.Getenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX") != "" {
		allGroups = filterGitlabGroupByExcludeMatchRegex(allGroups)
	}

	for i, group := range allGroups {
		if longFetch {
			spinningSpinner.Stop()
			if i == 0 {
				fmt.Println("")
			}
			msg := fmt.Sprintf("fetching repos for group: %v", group)
			colorlog.PrintInfo(msg)
		}
		repos, err := c.GetGroupRepos(group)
		if err != nil {
			return nil, fmt.Errorf("error fetching repos for group '%s', error: %v", group, err)
		}

		repoData = append(repoData, repos...)

	}

	snippets, err := c.GetSnippets(repoData, targetOrg)
	if err != nil {
		spinningSpinner.Stop()
		colorlog.PrintError(fmt.Sprintf("Error getting snippets, error: %v", err))
	}
	repoData = append(repoData, snippets...)

	return repoData, nil
}

// GetTopLevelGroups all top level org groups with parallel pagination
func (c Gitlab) GetTopLevelGroups() ([]string, error) {
	opt := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
		TopLevelOnly: &[]bool{true}[0],
		AllAvailable: &[]bool{true}[0],
	}

	// Fetch first page to discover total number of pages
	groups, resp, err := c.Client.Groups.ListGroups(opt)
	if err != nil {
		return nil, err
	}

	// If only one page, return immediately
	if resp.TotalPages <= 1 {
		allGroups := make([]string, 0, len(groups))
		for _, g := range groups {
			allGroups = append(allGroups, strconv.FormatInt(int64(g.ID), 10))
		}
		return allGroups, nil
	}

	// Multiple pages - fetch remaining pages in parallel
	return c.fetchTopLevelGroupsParallel(groups, int(resp.TotalPages))
}

// In this case take the cloneURL from the cloneTartet repo and just inject /snippets/:id before the .git
// cloud example
// http clone target url https://gitlab.com/ghorg-test-group/subgroup-2/foobar.git
// http snippet clone url https://gitlab.com/ghorg-test-group/subgroup-2/foobar/snippets/3711587.git
// ssh clone target url git@gitlab.com:ghorg-test-group/subgroup-2/foobar.git
// ssh snippet clone url git@gitlab.com:ghorg-test-group/subgroup-2/foobar/snippets/3711587.git
func (c Gitlab) createRepoSnippetCloneURL(cloneTargetURL string, snippetID string) string {

	// Split the cloneTargetURL into two parts at the ".git"
	parts := strings.Split(cloneTargetURL, ".git")
	// Insert the "/snippets/:id" before the ".git"
	cloneURL := parts[0] + "/snippets/" + snippetID + ".git"

	if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
		return cloneURL
	}

	// git@gitlab.example.com:local-gitlab-group3/subgroup-a/subgroup-b/subgroup_b_repo_1/snippets/12.git

	// http://gitlab.example.com/snippets/1.git
	if os.Getenv("GHORG_INSECURE_GITLAB_CLIENT") == "true" {
		cloneURL = strings.Replace(cloneURL, "http://", "git@", 1)
	} else {
		cloneURL = strings.Replace(cloneURL, "https://", "git@", 1)
	}
	// git@gitlab.example.com/snippets/1.git
	cloneURL = strings.Replace(cloneURL, "/", ":", 1)
	// git@gitlab.example.com:snippets/1.git
	return cloneURL
}

// hosted example
// root snippet ssh clone url git@gitlab.example.com:snippets/1.git
// root snippet http clone url http://gitlab.example.com/snippets/1.git
func (c Gitlab) createRootLevelSnippetCloneURL(snippetWebURL string) string {
	// Web URL example, http://gitlab.example.com/-/snippets/1
	// Both http and ssh clone urls do not have the /-/ in them so just remove it first and add the .git extention
	cloneURL := strings.Replace(snippetWebURL, "/-/", "/", -1) + ".git"
	if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
		return c.addTokenToCloneURL(cloneURL, os.Getenv("GHORG_GITLAB_TOKEN"))
	}

	if os.Getenv("GHORG_INSECURE_GITLAB_CLIENT") == "true" {
		cloneURL = strings.Replace(cloneURL, "http://", "git@", 1)
	} else {
		cloneURL = strings.Replace(cloneURL, "https://", "git@", 1)
	}
	// git@gitlab.example.com/snippets/1.git
	cloneURL = strings.Replace(cloneURL, "/", ":", 1)
	// git@gitlab.example.com:snippets/1.git
	return cloneURL
}

func (c Gitlab) getRepoSnippets(r Repo) []*gitlab.Snippet {
	var allSnippets []*gitlab.Snippet
	opt := &gitlab.ListProjectSnippetsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
	}

	for {
		snippets, resp, err := c.ProjectSnippets.ListSnippets(r.ID, opt)

		if resp.StatusCode == 403 {
			break
		}

		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Error fetching snippets for project %s: %v, ignoring error and proceeding to next project", r.Name, err))
			break
		}

		allSnippets = append(allSnippets, snippets...)

		// Exit the loop when we've seen all pages.
		if resp.NextPage == 0 {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return allSnippets
}

func (c Gitlab) getAllSnippets() []*gitlab.Snippet {
	var allSnippets []*gitlab.Snippet
	opt := &gitlab.ListAllSnippetsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
	}

	for {
		snippets, resp, err := c.Snippets.ListAllSnippets(opt)
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Issue fetching all snippets, not all snippets will be cloned error: %v", err))
			return allSnippets
		}

		allSnippets = append(allSnippets, snippets...)

		// Exit the loop when we've seen all pages.
		if resp.NextPage == 0 {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return allSnippets
}

func (c Gitlab) GetSnippets(cloneData []Repo, target string) ([]Repo, error) {

	if os.Getenv("GHORG_CLONE_SNIPPETS") != "true" {
		return []Repo{}, nil
	}

	var allSnippetsToClone []*gitlab.Snippet

	// Snippets are converted into Repos so we can clone them
	snippetsToClone := []Repo{}

	// If it is a cloud group clone iterate over each project and try to get its snippets. We have to do this because if you use the /snippets/all endpoint it will return every public snippet from the cloud.
	if os.Getenv("GHORG_CLONE_TYPE") != "user" && os.Getenv("GHORG_SCM_BASE_URL") == "" {
		// Iterate over all projects in the group. If it has snippets add them
		colorlog.PrintInfo("Note: only snippets you have access to will be cloned. This process may take a while depending on the size of group you are trying to clone, please be patient.")
		for _, repo := range cloneData {
			snippets := c.getRepoSnippets(repo)
			allSnippetsToClone = append(allSnippetsToClone, snippets...)
		}
	} else {
		allSnippets := c.getAllSnippets()

		// if its an all-user or all-group clone, for each repo get its snippets then also include all root level snippets
		if target == "all-users" || target == "all-groups" {
			for _, repo := range cloneData {
				repoSnippets := c.getRepoSnippets(repo)
				allSnippetsToClone = append(allSnippetsToClone, repoSnippets...)
			}

			for _, snippet := range allSnippets {
				if c.rootLevelSnippet(snippet.WebURL) {
					allSnippetsToClone = append(allSnippetsToClone, snippet)
				}
			}
		} else if os.Getenv("GHORG_CLONE_TYPE") != "user" {
			// Handle single group clones on hosted instances
			for _, repo := range cloneData {
				repoSnippets := c.getRepoSnippets(repo)
				allSnippetsToClone = append(allSnippetsToClone, repoSnippets...)
			}
		}

		if os.Getenv("GHORG_CLONE_TYPE") == "user" && os.Getenv("GHORG_SCM_BASE_URL") == "" {

		}

	}

	for _, snippet := range allSnippetsToClone {
		snippetID := strconv.FormatInt(snippet.ID, 10)
		snippetTitle := ToSlug(snippet.Title)
		s := Repo{}
		s.IsGitLabSnippet = true
		s.CloneBranch = "main"
		s.GitLabSnippetInfo.Title = snippetTitle
		s.Name = snippetTitle
		s.GitLabSnippetInfo.ID = snippetID
		s.URL = snippet.WebURL
		// If the snippet is not made on any repo its a root level snippet, this works for cloud
		if c.rootLevelSnippet(snippet.WebURL) {
			s.IsGitLabRootLevelSnippet = true
			s.CloneURL = c.createRootLevelSnippetCloneURL(snippet.WebURL)
			cloneData = append(cloneData, s)
		} else {
			// Since this isn't a root level repo we want to find which repo the snippet is coming from
			for _, cloneTarget := range cloneData {
				if cloneTarget.ID == strconv.FormatInt(snippet.ProjectID, 10) {
					s.CloneURL = c.createRepoSnippetCloneURL(cloneTarget.CloneURL, snippetID)
					s.Path = cloneTarget.Path
					s.GitLabSnippetInfo.URLOfRepo = cloneTarget.URL
					s.GitLabSnippetInfo.NameOfRepo = cloneTarget.Name
					cloneData = append(cloneData, s)
				}
			}
		}

		snippetsToClone = append(snippetsToClone, s)
	}

	return snippetsToClone, nil
}

// GetGroupRepos fetches repo data from a specific group with parallel pagination
func (c Gitlab) GetGroupRepos(targetGroup string) ([]Repo, error) {
	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
		IncludeSubGroups: gitlab.Ptr(true),
	}

	// Fetch first page to discover total number of pages
	ps, resp, err := c.Groups.ListGroupProjects(targetGroup, opt)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, fmt.Errorf("group '%s' does not exist", targetGroup)
		}
		return []Repo{}, err
	}

	// If only one page, return immediately
	if resp.TotalPages <= 1 {
		return c.filter(targetGroup, ps), nil
	}

	// Multiple pages - fetch remaining pages in parallel
	return c.fetchGroupReposParallel(targetGroup, ps, int(resp.TotalPages))
}

// GetUserRepos gets all of a users gitlab repos
func (c Gitlab) GetUserRepos(targetUsername string) ([]Repo, error) {
	cloneData := []Repo{}
	targetUsers := []string{}

	projectOpts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
	}

	userOpts := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: int64(perPage),
			Page:    1,
		},
	}

	spinningSpinner.Start()
	defer spinningSpinner.Stop()

	if targetUsername == "all-users" {
		gitLabAllUsers = true
		for {
			allUsers, resp, err := c.Users.ListUsers(userOpts)
			if err != nil {
				return nil, fmt.Errorf("error getting all users, err: %v", err)
			}

			for _, u := range allUsers {
				targetUsers = append(targetUsers, u.Username)
			}

			if resp.NextPage == 0 {
				break
			}

			// Update the page number to get the next page.
			userOpts.Page = resp.NextPage
		}
	} else {
		targetUsers = append(targetUsers, targetUsername)
	}

	for _, targetUser := range targetUsers {
		for {
			// Get the first page with projects.
			ps, resp, err := c.Projects.ListUserProjects(targetUser, projectOpts)
			if err != nil {
				spinningSpinner.Stop()
				colorlog.PrintError(fmt.Sprintf("Error getting repo for user: %v", targetUser))
				continue
			}

			// filter from all the projects we've found so far.
			cloneData = append(cloneData, c.filter(targetUser, ps)...)

			// Exit the loop when we've seen all pages.
			if resp.NextPage == 0 {
				break
			}

			// Update the page number to get the next page.
			userOpts.Page = resp.NextPage
		}
	}

	snippets, err := c.GetSnippets(cloneData, targetUsername)
	if err != nil {
		spinningSpinner.Stop()
		colorlog.PrintError(fmt.Sprintf("Error getting snippets, error: %v", err))
	}
	cloneData = append(cloneData, snippets...)
	return cloneData, nil
}

// NewClient create new gitlab scm client
func (_ Gitlab) NewClient() (Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_GITLAB_TOKEN")

	var err error
	var c *gitlab.Client
	if baseURL != "" {
		if os.Getenv("GHORG_INSECURE_GITLAB_CLIENT") == "true" {
			defaultTransport := http.DefaultTransport.(*http.Transport)
			// Create new Transport that ignores self-signed SSL
			customTransport := &http.Transport{
				Proxy:                 defaultTransport.Proxy,
				DialContext:           defaultTransport.DialContext,
				MaxIdleConns:          defaultTransport.MaxIdleConns,
				IdleConnTimeout:       defaultTransport.IdleConnTimeout,
				ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
				TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: customTransport}
			opt := gitlab.WithHTTPClient(client)
			c, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL), opt)
			colorlog.PrintError("WARNING: USING AN INSECURE GITLAB CLIENT")
		} else {
			c, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
		}

	} else {
		c, err = gitlab.NewClient(token)
	}
	return Gitlab{c}, err
}

func (_ Gitlab) addTokenToCloneURL(url string, token string) string {
	// allows for http and https for local testing
	splitURL := strings.Split(url, "://")
	return splitURL[0] + "://oauth2:" + token + "@" + splitURL[1]
}

func (c Gitlab) filter(group string, ps []*gitlab.Project) []Repo {
	var repoData []Repo

	isSubgroup := strings.Contains(group, "/")

	for _, p := range ps {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if p.Archived {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if p.ForkedFromProject != nil {
				continue
			}
		}

		if !hasMatchingTopic(p.Topics) {
			continue
		}

		// Apply GitLab group exclude regex to repository path
		if os.Getenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX") != "" {
			regex := os.Getenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX")
			re := regexp.MustCompile(regex)
			if re.FindString(p.PathWithNamespace) != "" {
				continue // Skip this repository as it matches the exclude pattern
			}
		}

		r := Repo{}

		r.Name = p.Name
		r.ID = strconv.FormatInt(int64(p.ID), 10)

		if os.Getenv("GHORG_BRANCH") == "" {
			defaultBranch := p.DefaultBranch
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = os.Getenv("GHORG_BRANCH")
		}

		path := p.PathWithNamespace

		// The PathWithNamespace includes the org/group name
		// https://github.com/gabrie30/ghorg/issues/228
		// https://github.com/gabrie30/ghorg/issues/267
		// https://github.com/gabrie30/ghorg/issues/271
		if !gitLabAllGroups && !gitLabAllUsers {
			if isSubgroup {
				if os.Getenv("GHORG_OUTPUT_DIR") == "" {
					path = strings.TrimPrefix(path, group)
				}
			} else {
				path = strings.TrimPrefix(path, group)
			}
		}

		r.Path = path
		r.ID = fmt.Sprint(p.ID)
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = c.addTokenToCloneURL(p.HTTPURLToRepo, os.Getenv("GHORG_GITLAB_TOKEN"))
			r.URL = p.HTTPURLToRepo
			repoData = append(repoData, r)
		} else {
			r.CloneURL = p.SSHURLToRepo
			r.URL = p.SSHURLToRepo
			repoData = append(repoData, r)
		}

		if p.WikiEnabled && os.Getenv("GHORG_CLONE_WIKI") == "true" {
			wiki := Repo{}
			// wiki needs name for gitlab name collisions
			wiki.Name = p.Name
			wiki.IsWiki = true
			wiki.CloneURL = strings.Replace(r.CloneURL, ".git", ".wiki.git", 1)
			wiki.URL = strings.Replace(r.URL, ".git", ".wiki.git", 1)
			wiki.CloneBranch = "master"
			wiki.Path = fmt.Sprintf("%s%s", path, ".wiki")
			repoData = append(repoData, wiki)
		}
	}
	return repoData
}

func filterGitlabGroupByExcludeMatchRegex(groups []string) []string {
	filteredGroups := []string{}
	regex := fmt.Sprint(os.Getenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX"))

	for i, grp := range groups {
		exclude := false
		re := regexp.MustCompile(regex)
		match := re.FindString(grp)
		if match != "" {
			exclude = true
		}

		if !exclude {
			filteredGroups = append(filteredGroups, groups[i])
		}
	}

	return filteredGroups
}

// ToSlug converts a title into a URL-friendly slug.
func ToSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[\s\p{P}]+`).ReplaceAllString(slug, "-")

	// Remove any non-alphanumeric characters except for hyphens
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "")

	// Trim any leading or trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
