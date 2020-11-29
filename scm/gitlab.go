package scm

import (
	"fmt"
	"os"
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
)

var (
	_       Client = Gitlab{}
	perPage        = 50
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
	repoData := []Repo{}

	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
		IncludeSubgroups: gitlab.Bool(true),
	}

	for {
		// Get the first page with projects.
		ps, resp, err := c.Groups.ListGroupProjects(targetOrg, opt)

		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				return nil, fmt.Errorf("group '%s' does not exist", targetOrg)
			}
			return []Repo{}, err
		}

		// filter from all the projects we've found so far.
		repoData = append(repoData, c.filter(ps)...)

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
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

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: perPage,
			Page:    1,
		},
	}

	for {
		// Get the first page with projects.
		ps, resp, err := c.Projects.ListUserProjects(targetUsername, opt)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				return nil, fmt.Errorf("user '%s' does not exist", targetUsername)
			}
			return []Repo{}, err
		}

		// filter from all the projects we've found so far.
		cloneData = append(cloneData, c.filter(ps)...)

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
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
		c, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	} else {
		c, err = gitlab.NewClient(token)
	}
	return Gitlab{c}, err
}

func (_ Gitlab) addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://oauth2:" + token + "@" + splitURL[1]
}

func (c Gitlab) filter(ps []*gitlab.Project) []Repo {
	var repoData []Repo
	for _, p := range ps {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if p.Archived == true {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if p.ForkedFromProject != nil {
				continue
			}
		}

		if os.Getenv("GHORG_MATCH_PREFIX") != "" {
			repoName := strings.ToLower(p.Name)
			foundPrefix := false
			pfs := strings.Split(os.Getenv("GHORG_MATCH_PREFIX"), ",")
			for _, p := range pfs {
				if strings.HasPrefix(repoName, strings.ToLower(p)) {
					foundPrefix = true
				}
			}
			if foundPrefix == false {
				continue
			}
		}

		r := Repo{}

		if os.Getenv("GHORG_BRANCH") == "" {
			defaultBranch := p.DefaultBranch
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = os.Getenv("GHORG_BRANCH")
		}

		r.Path = p.PathWithNamespace
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = c.addTokenToHTTPSCloneURL(p.HTTPURLToRepo, os.Getenv("GHORG_GITLAB_TOKEN"))
			r.URL = p.HTTPURLToRepo
			repoData = append(repoData, r)
		} else {
			r.CloneURL = p.SSHURLToRepo
			r.URL = p.SSHURLToRepo
			repoData = append(repoData, r)
		}
	}
	return repoData
}
