package internal

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/internal/base"

	"code.gitea.io/sdk/gitea"
)

var (
	_ base.Client = GiteaClient{}
)

func init() {
	RegisterClient(GiteaClient{})
}

type GiteaClient struct{}

func (_ GiteaClient) GetType() string {
	return "gitea"
}

// GetOrgRepos fetches repo data from a specific group
func (c GiteaClient) GetOrgRepos(targetOrg string) ([]base.Repo, error) {
	repoData := []base.Repo{}
	client, err := c.determineClient()

	if err != nil {
		colorlog.PrintError(err)
	}

	perPage := 10
	if conf, _, err := client.GetGlobalAPISettings(); err == nil && conf != nil {
		perPage = conf.MaxResponseItems
	}

	for i := 1; ; i++ {
		rps, resp, err := client.ListOrgRepos(targetOrg, gitea.ListOrgReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				colorlog.PrintError(fmt.Errorf("org \"%s\" not found", targetOrg))
			}
			return []base.Repo{}, err
		}

		repoDataFiltered, err := c.filter(client, rps)
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
func (c GiteaClient) GetUserRepos(targetUsername string) ([]base.Repo, error) {
	repoData := []base.Repo{}
	client, err := c.determineClient()

	if err != nil {
		colorlog.PrintError(err)
	}

	perPage := 10
	if conf, _, err := client.GetGlobalAPISettings(); err == nil && conf != nil {
		perPage = conf.MaxResponseItems
	}

	for i := 1; ; i++ {
		rps, resp, err := client.ListUserRepos(targetUsername, gitea.ListReposOptions{ListOptions: gitea.ListOptions{
			Page:     i,
			PageSize: perPage,
		}})

		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				colorlog.PrintError(fmt.Errorf("org \"%s\" not found", targetUsername))
			}
			return []base.Repo{}, err
		}

		repoDataFiltered, err := c.filter(client, rps)
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

func (c GiteaClient) determineClient() (*gitea.Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_GITEA_TOKEN")

	if baseURL == "" {
		baseURL = "https://gitea.com"
	}

	return gitea.NewClient(baseURL, gitea.SetToken(token))
}

func (c GiteaClient) filter(client *gitea.Client, rps []*gitea.Repository) (repoData []base.Repo, err error) {
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
		if os.Getenv("GHORG_TOPICS") != "" {
			rpTopics, _, err := client.ListRepoTopics(rp.Owner.UserName, rp.Name, gitea.ListRepoTopicsOptions{})
			if err != nil {
				return []base.Repo{}, err
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

		r := base.Repo{}
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
