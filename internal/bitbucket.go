package internal

import (
	"os"
	"strings"

	"github.com/gabrie30/ghorg/internal/base"

	"github.com/ktrysmt/go-bitbucket"
)

var (
	_ base.Client = BitbucketClient{}
)

func init() {
	RegisterClient(BitbucketClient{})
}

type BitbucketClient struct{}

func (_ BitbucketClient) GetType() string {
	return "bitbucket"
}

// GetOrgRepos gets org repos
func (c BitbucketClient) GetOrgRepos(targetOrg string) ([]base.Repo, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))

	resp, err := client.Teams.Repositories(targetOrg)
	if err != nil {
		return []base.Repo{}, err
	}

	return c.filter(resp)
}

// GetUserRepos gets user repos from bitbucket
func (c BitbucketClient) GetUserRepos(targetUser string) ([]base.Repo, error) {

	client := bitbucket.NewBasicAuth(os.Getenv("GHORG_BITBUCKET_USERNAME"), os.Getenv("GHORG_BITBUCKET_APP_PASSWORD"))

	resp, err := client.Users.Repositories(targetUser)
	if err != nil {
		return []base.Repo{}, err
	}

	return c.filter(resp)
}

func (_ BitbucketClient) filter(resp interface{}) (repoData []base.Repo, err error) {
	cloneData := []base.Repo{}
	values := resp.(map[string]interface{})["values"].([]interface{})

	for _, a := range values {
		clone := a.(map[string]interface{})
		links := clone["links"].(map[string]interface{})["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]

			if os.Getenv("GHORG_MATCH_PREFIX") != "" {
				repoName := strings.ToLower(clone["name"].(string))
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
