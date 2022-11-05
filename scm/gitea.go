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
	// extend the gitea client
	*gitea.Client
	// perPage contain the pagination item limit
	perPage int
}

func (_ Gitea) GetType() string {
	return "gitea"
}

// GetOrgRepos fetches repo data from a specific group
func (c Gitea) GetOrgRepos(targetOrg string) ([]Repo, error) {
	repoData := []Repo{}

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

// NewClient create new gitea scm client
func (_ Gitea) NewClient() (Client, error) {
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	token := os.Getenv("GHORG_GITEA_TOKEN")

	if baseURL == "" {
		baseURL = "https://gitea.com"
	}

	isHTTP := strings.HasPrefix(baseURL, "http://")

	if isHTTP && (os.Getenv("GHORG_INSECURE_GITEA_CLIENT") != "true") {
		colorlog.PrintErrorAndExit("You are attempting clone from an insecure Gitea instance. You must set the (--insecure-gitea-client) flag to proceed.")
	}

	c, err := gitea.NewClient(baseURL, gitea.SetToken(token))
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

func (_ Gitea) addTokenToCloneURL(url string, token string) string {
	isHTTP := strings.HasPrefix(url, "http://")

	if isHTTP {
		if os.Getenv("GHORG_INSECURE_GITEA_CLIENT") == "true" {
			splitURL := strings.Split(url, "http://")
			return "http://" + token + "@" + splitURL[1]
		}
		colorlog.PrintErrorAndExit("You are attempting clone from an insecure Gitea instance. You must set the (--insecure-gitea-client) flag to proceed.")
	}

	splitURL := strings.Split(url, "https://")
	return "https://" + token + "@" + splitURL[1]
}

func (c Gitea) filter(rps []*gitea.Repository) (repoData []Repo, err error) {
	for _, rp := range rps {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if rp.Archived {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if rp.Fork {
				continue
			}
		}

		if os.Getenv("GHORG_TOPICS") != "" {
			rpTopics, _, err := c.ListRepoTopics(rp.Owner.UserName, rp.Name, gitea.ListRepoTopicsOptions{})
			if err != nil {
				return []Repo{}, err
			}
			if !hasMatchingTopic(rpTopics) {
				continue
			}
		}

		r := Repo{}
		r.Path = rp.FullName
		r.Name = rp.Name

		if os.Getenv("GHORG_BRANCH") == "" {
			defaultBranch := rp.DefaultBranch
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = os.Getenv("GHORG_BRANCH")
		}

		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			cloneURL := rp.CloneURL
			if rp.Private {
				cloneURL = c.addTokenToCloneURL(cloneURL, os.Getenv("GHORG_GITEA_TOKEN"))
			}
			r.CloneURL = cloneURL
			r.URL = cloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = rp.SSHURL
			r.URL = rp.SSHURL
			repoData = append(repoData, r)
		}

		if rp.HasWiki && os.Getenv("GHORG_CLONE_WIKI") == "true" {
			wiki := Repo{}
			wiki.IsWiki = true
			wiki.CloneURL = strings.Replace(r.CloneURL, ".git", ".wiki.git", 1)
			wiki.URL = strings.Replace(r.URL, ".git", ".wiki.git", 1)
			wiki.CloneBranch = "master"
			wiki.Path = fmt.Sprintf("%s%s", r.Name, ".wiki")
			repoData = append(repoData, wiki)
		}
	}
	return repoData, nil
}
