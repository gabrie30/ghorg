package scm

import (
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/gabrie30/ghorg/configs"
)

var (
	_ Client = Gitea{}
)

func init() {
	registerClient(Gitea{})
}

type Gitea struct {
	// extend the gitea client
	*gitea.Client
	// perPage contain the pagination item limit
	perPage int
}

func (_ Gitea) GetType() string {
	return "gitea"
}

// GetOrgRepos fetches repo data from a specific group
func (c Gitea) GetOrgRepos(config *configs.Config, targetOrg string) ([]Repo, error) {
	var repoData []Repo

	for i := 1; ; i++ {
		rps, resp, err := c.ListOrgRepos(targetOrg, gitea.ListOrgReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: c.perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				err = fmt.Errorf("org \"%s\" not found", targetOrg)
			}
			return nil, err
		}

		repoDataFiltered, err := c.filter(config, rps)
		if err != nil {
			return nil, err
		}
		repoData = append(repoData, repoDataFiltered...)

		// Exit the loop when we've seen all pages.
		if len(rps) < perPage {
			break
		}
	}

	return repoData, nil
}

// GetUserRepos gets all of a users gitlab repos
func (c Gitea) GetUserRepos(config *configs.Config, targetUsername string) ([]Repo, error) {
	var repoData []Repo

	for i := 1; ; i++ {
		rps, resp, err := c.ListUserRepos(targetUsername, gitea.ListReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: c.perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				err = fmt.Errorf("org \"%s\" not found", targetUsername)
			}
			return nil, err
		}

		repoDataFiltered, err := c.filter(config, rps)
		if err != nil {
			return nil, err
		}
		repoData = append(repoData, repoDataFiltered...)

		// Exit the loop when we've seen all pages.
		if len(rps) < perPage {
			break
		}
	}

	return repoData, nil
}

// NewClient create new gitea scm client
func (_ Gitea) NewClient(config *configs.Config) (Client, error) {
	if config.ScmBaseUrl == "/" {
		config.ScmBaseUrl = "https://gitea.com"
	}

	c, err := gitea.NewClient(config.ScmBaseUrl, gitea.SetToken(config.Token))
	if err != nil {
		return nil, err
	}
	client := Gitea{Client: c}

	//set small limit so gitea most likely will have a bigger one
	client.perPage = 10
	if conf, _, err := client.GetGlobalAPISettings(); err == nil && conf != nil {
		// gitea >= 1.13 will tell us the limit it has
		client.perPage = conf.MaxResponseItems
	}

	return client, nil
}

func (c Gitea) filter(config *configs.Config, rps []*gitea.Repository) (repoData []Repo, err error) {
	for _, rp := range rps {
		if config.SkipArchived && rp.Archived {
			continue
		}

		if config.SkipForks && rp.Fork {
			continue
		}

		// If user defined a list of topics, check if any match with this repo
		if len(config.Topics) > 0 {
			rpTopics, _, err := c.ListRepoTopics(rp.Owner.UserName, rp.Name, gitea.ListRepoTopicsOptions{})
			if err != nil {
				return []Repo{}, err
			}
			foundTopic := false
			for _, topic := range rpTopics {
				for _, envTopic := range config.Topics {
					if topic == envTopic {
						foundTopic = true
						break
					}
				}
				if foundTopic {
					break
				}
			}
			if !foundTopic {
				continue
			}
		}

		if config.MatchPrefix != "" {
			repoName := strings.ToLower(rp.Name)
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
		r.Path = rp.FullName

		if config.Branch == "" {
			defaultBranch := rp.DefaultBranch
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = config.Branch
		}

		if config.CloneProtocol == "https" {
			cloneURL := rp.CloneURL
			if rp.Private {
				cloneURL = "https://" + config.Token + strings.Replace(cloneURL, "https://", "", 1)
			}
			r.CloneURL = cloneURL
			r.URL = cloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = rp.SSHURL
			r.URL = rp.SSHURL
			repoData = append(repoData, r)
		}
	}
	return repoData, nil
}
