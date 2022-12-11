package scm

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	_       Client = Gitlab{}
	perPage        = 100
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

// GetOrgRepos fetches repo data from a specific group
func (c Gitlab) GetOrgRepos(targetOrg string) ([]Repo, error) {
	allGroups := []string{}
	repoData := []Repo{}
	longFetch := false

	if targetOrg == "all-users" {
		colorlog.PrintErrorAndExit("When using the 'all-users' keyword the '--clone-type=user' flag should be set")
	}

	if targetOrg == "all-groups" {
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

	return repoData, nil
}

// GetTopLevelGroups all top level org groups
func (c Gitlab) GetTopLevelGroups() ([]string, error) {
	allGroups := []string{}

	opt := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
		TopLevelOnly: gitlab.Bool(true),
		AllAvailable: gitlab.Bool(true),
	}

	for {

		groups, resp, err := c.Client.Groups.ListGroups(opt)

		if err != nil {
			return allGroups, err
		}

		for _, g := range groups {
			allGroups = append(allGroups, g.Path)
		}

		// Exit the loop when we've seen all pages.
		if resp.NextPage == 0 {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return allGroups, nil
}

// GetGroupRepos fetches repo data from a specific group
func (c Gitlab) GetGroupRepos(targetGroup string) ([]Repo, error) {

	repoData := []Repo{}

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
		IncludeSubGroups: gitlab.Bool(true),
	}

	for {
		// Get the first page with projects.
		ps, resp, err := c.Groups.ListGroupProjects(targetGroup, opt)

		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				return nil, fmt.Errorf("group '%s' does not exist", targetGroup)
			}
			return []Repo{}, err
		}

		// filter from all the projects we've found so far.
		repoData = append(repoData, c.filter(targetGroup, ps)...)

		// Exit the loop when we've seen all pages.
		if resp.NextPage == 0 {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}

	return repoData, nil
}

// GetUserRepos gets all of a users gitlab repos
func (c Gitlab) GetUserRepos(targetUsername string) ([]Repo, error) {
	cloneData := []Repo{}
	targetUsers := []string{}

	projectOpts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
	}

	userOpts := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
	}

	if targetUsername == "all-users" {
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
				fmt.Printf("Error getting repo for user: %v", targetUser)
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

		r := Repo{}

		r.Name = p.Name

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

		// For GitLab Cloud, the PathWithNamespace includes the org/group name
		// https://github.com/gabrie30/ghorg/issues/228
		if os.Getenv("GHORG_SCM_BASE_URL") == "" {
			path = strings.TrimPrefix(path, group)
		}

		r.Path = path
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
			wiki.IsWiki = true
			wiki.CloneURL = strings.Replace(r.CloneURL, ".git", ".wiki.git", 1)
			wiki.URL = strings.Replace(r.URL, ".git", ".wiki.git", 1)
			wiki.CloneBranch = "master"
			wiki.Path = fmt.Sprintf("%s%s", p.PathWithNamespace, ".wiki")
			repoData = append(repoData, wiki)
		}
	}
	return repoData
}

func filterGitlabGroupByExcludeMatchRegex(groups []string) []string {
	filteredGroups := []string{}
	regex := fmt.Sprint(os.Getenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX"))

	for i, grp := range groups {
		fmt.Println(grp)
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
