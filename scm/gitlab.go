package scm

import (
	"fmt"
	"strings"

	"github.com/gabrie30/ghorg/configs"
	"github.com/xanzy/go-gitlab"
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
func (c Gitlab) GetOrgRepos(config *configs.Config, targetOrg string) ([]Repo, error) {
	var repoData []Repo

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
		repoData = append(repoData, c.filter(config, ps)...)

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
func (c Gitlab) GetUserRepos(config *configs.Config, targetUsername string) ([]Repo, error) {
	var cloneData []Repo

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
		cloneData = append(cloneData, c.filter(config, ps)...)

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
func (_ Gitlab) NewClient(config *configs.Config) (Client, error) {
	var err error
	var c *gitlab.Client
	if config.ScmBaseUrl != "/" {
		c, err = gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.ScmBaseUrl))
	} else {
		c, err = gitlab.NewClient(config.Token)
	}
	return Gitlab{c}, err
}

func (_ Gitlab) addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://oauth2:" + token + "@" + splitURL[1]
}

func (c Gitlab) filter(config *configs.Config, ps []*gitlab.Project) []Repo {
	var repoData []Repo
	for _, p := range ps {

		if config.SkipArchived && p.Archived {
			continue
		}

		if config.SkipForks && p.ForkedFromProject != nil {
			continue
		}

		if config.MatchPrefix != "" {
			repoName := strings.ToLower(p.Name)
			foundPrefix := false
			pfs := strings.Split(config.MatchPrefix, ",")
			for _, p := range pfs {
				if strings.HasPrefix(repoName, strings.ToLower(p)) {
					foundPrefix = true
				}
			}
			if !foundPrefix {
				continue
			}
		}

		r := Repo{}

		if config.Branch == "" {
			defaultBranch := p.DefaultBranch
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = config.Branch
		}

		r.Path = p.PathWithNamespace
		if config.CloneProtocol == "https" {
			r.CloneURL = c.addTokenToHTTPSCloneURL(p.HTTPURLToRepo, config.Token)
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
