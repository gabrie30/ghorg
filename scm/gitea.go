package scm

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/gabrie30/ghorg/colorlog"
)

var (
	_ Client = Gitea{}
)

func init() {
	registerClient(Gitea{})
}

type Gitea struct {
	// client contain the gitea client
	client *gitea.Client
	// perPage contain the pagination item limit
	perPage int
}

func (_ Gitea) GetType() string {
	return "gitea"
}

// GetOrgRepos fetches repo data from a specific group
func (c Gitea) GetOrgRepos(targetOrg string) ([]Repo, error) {
	repoData := []Repo{}
	if c.client == nil {
		if err := c.determineClient(); err != nil {
			colorlog.PrintError(err)
		}
	}

	for i := 1; ; i++ {
		rps, resp, err := c.client.ListOrgRepos(targetOrg, gitea.ListOrgReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: c.perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				colorlog.PrintError(fmt.Errorf("org \"%s\" not found", targetOrg))
			}
			return []Repo{}, err
		}

		repoDataFiltered, err := c.filter(rps)
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
func (c Gitea) GetUserRepos(targetUsername string) ([]Repo, error) {
	repoData := []Repo{}
	if c.client == nil {
		if err := c.determineClient(); err != nil {
			colorlog.PrintError(err)
		}
	}

	for i := 1; ; i++ {
		rps, resp, err := c.client.ListUserRepos(targetUsername, gitea.ListReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: c.perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				colorlog.PrintError(fmt.Errorf("org \"%s\" not found", targetUsername))
			}
			return []Repo{}, err
		}

		repoDataFiltered, err := c.filter(rps)
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

func (c Gitea) determineClient() error {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_GITEA_TOKEN")

	if baseURL == "" {
		baseURL = "https://gitea.com"
	}

	client, err := gitea.NewClient(baseURL, gitea.SetToken(token))
	if err != nil {
		return err
	}
	c.client = client

	//set small limit so gitea most likely will have a bigger one
	c.perPage = 10
	if conf, _, err := client.GetGlobalAPISettings(); err == nil && conf != nil {
		// gitea >= 1.13 will tell us the limit it has
		c.perPage = conf.MaxResponseItems
	}

	return nil
}

func (c Gitea) filter(rps []*gitea.Repository) (repoData []Repo, err error) {
	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

	for _, rp := range rps {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if rp.Archived == true {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if rp.Fork == true {
				continue
			}
		}

		// If user defined a list of topics, check if any match with this repo
		if os.Getenv("GHORG_TOPICS") != "" && c.client != nil {
			rpTopics, _, err := c.client.ListRepoTopics(rp.Owner.UserName, rp.Name, gitea.ListRepoTopicsOptions{})
			if err != nil {
				return []Repo{}, err
			}
			foundTopic := false
			for _, topic := range rpTopics {
				for _, envTopic := range envTopics {
					if topic == envTopic {
						foundTopic = true
						break
					}
				}
				if foundTopic == true {
					break
				}
			}
			if foundTopic == false {
				continue
			}
		}

		if os.Getenv("GHORG_MATCH_PREFIX") != "" {
			repoName := strings.ToLower(rp.Name)
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
		r.Path = rp.FullName

		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			cloneURL := rp.CloneURL
			if rp.Private {
				cloneURL = "https://" + os.Getenv("GHORG_GITEA_TOKEN") + strings.TrimLeft(cloneURL, "https://")
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
