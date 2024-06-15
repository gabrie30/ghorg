package scm

import (
	"fmt"
	"net/url"
	"strings"

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
	resp, err := c.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: targetOrg})
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp.Items)
}

// GetUserRepos gets user repos from bitbucket
func (c Bitbucket) GetUserRepos(targetUser string) ([]Repo, error) {
	resp, err := c.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: targetUser})
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(resp.Items)
}

// NewClient create new bitbucket scm client
func (_ Bitbucket) NewClient() (Client, error) {
	user := os.Getenv("GHORG_BITBUCKET_USERNAME")
	password := os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
	oAuth := os.Getenv("GHORG_BITBUCKET_OAUTH")
	var c *bitbucket.Client

	if oAuth != "" {
		c = bitbucket.NewOAuthbearerToken(oAuth)
	} else {
		c = bitbucket.NewBasicAuth(user, password)
	}

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u, _ := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
		c.SetApiBaseURL(*u)
	}

	return Bitbucket{c}, nil
}

func (_ Bitbucket) filter(resp []bitbucket.Repository) (repoData []Repo, err error) {
	cloneData := []Repo{}

	for _, a := range resp {
		links := a.Links["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]

			if os.Getenv("GHORG_TOPICS") != "" {
				colorlog.PrintError("WARNING: Filtering by topics is not supported for Bitbucket SCM")
			}

			r := Repo{}
			r.Name = a.Name
			r.Path = r.Name
			if os.Getenv("GHORG_BRANCH") == "" {
				r.CloneBranch = a.Mainbranch.Name
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
				if os.Getenv("GHORG_BITBUCKET_OAUTH") != "" {
					// TODO
				} else {
					r.CloneURL = insertAppPasswordCredentialsIntoURL(r.CloneURL)
				}
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}

func insertAppPasswordCredentialsIntoURL(url string) string {
	password := os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
	credentials := ":" + password + "@"
	urlWithCredentials := strings.Replace(url, "@", credentials, 1)

	return urlWithCredentials
}
