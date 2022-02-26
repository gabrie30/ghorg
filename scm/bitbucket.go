package scm

import (
	"net/url"
	"os"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/ktrysmt/go-bitbucket"
)

var (
	_ Client = Bitbucket{}
)

func init() {
	registerClient(Bitbucket{})
}

type Bitbucket struct {
	// extend the bitbucket client
	*bitbucket.Client
}

func (_ Bitbucket) GetType() string {
	return "bitbucket"
}

// GetOrgRepos gets org repos
func (c Bitbucket) GetOrgRepos(targetOrg string) ([]Repo, error) {
	resp, err := c.Teams.Repositories(targetOrg)
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp)
}

// GetUserRepos gets user repos from bitbucket
func (c Bitbucket) GetUserRepos(targetUser string) ([]Repo, error) {
	resp, err := c.Users.Repositories(targetUser)
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp)
}

// NewClient create new bitbucket scm client
func (_ Bitbucket) NewClient() (Client, error) {
	user := os.Getenv("GHORG_BITBUCKET_USERNAME")
	password := os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
	oAuth := os.Getenv("GHORG_BITBUCKET_OAUTH")
	var c *bitbucket.Client

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u, _ := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
		c.SetApiBaseURL(*u)
	}

	if oAuth != "" {
		c = bitbucket.NewOAuthbearerToken(oAuth)
	} else {
		c = bitbucket.NewBasicAuth(user, password)
	}
	return Bitbucket{c}, nil
}

func (_ Bitbucket) filter(resp interface{}) (repoData []Repo, err error) {
	cloneData := []Repo{}
	values := resp.(map[string]interface{})["values"].([]interface{})

	for _, a := range values {
		clone := a.(map[string]interface{})
		links := clone["links"].(map[string]interface{})["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]

			if os.Getenv("GHORG_TOPICS") != "" {
				colorlog.PrintError("WARNING: Filtering by topics is not supported for Bitbucket SCM")
			}

			r := Repo{}
			r.Name = clone["name"].(string)

			if os.Getenv("GHORG_BRANCH") == "" {
				var defaultBranch string
				if clone["mainbranch"] == nil {
					defaultBranch = "master"
				} else {
					defaultBranch = clone["mainbranch"].(map[string]interface{})["name"].(string)
				}

				r.CloneBranch = defaultBranch
			} else {
				r.CloneBranch = os.Getenv("GHORG_BRANCH")
			}

			if os.Getenv("GHORG_BRANCH") == "" {
				r.CloneBranch = "master"
			} else {
				r.CloneBranch = os.Getenv("GHORG_BRANCH")
			}

			if os.Getenv("GHORG_CLONE_PROTOCOL") == "ssh" && linkType == "ssh" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				cloneData = append(cloneData, r)
			} else if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" && linkType == "https" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}
