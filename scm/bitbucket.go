package scm

import (
	"strings"

	"github.com/gabrie30/ghorg/configs"
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
func (c Bitbucket) GetOrgRepos(config *configs.Config, targetOrg string) ([]Repo, error) {
	resp, err := c.Teams.Repositories(targetOrg)
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(config, resp)
}

// GetUserRepos gets user repos from bitbucket
func (c Bitbucket) GetUserRepos(config *configs.Config, targetUser string) ([]Repo, error) {
	resp, err := c.Users.Repositories(targetUser)
	if err != nil {
		return []Repo{}, err
	}

	return c.filter(config, resp)
}

// NewClient create new bitbucket scm client
func (_ Bitbucket) NewClient(config *configs.Config) (Client, error) {
	c := bitbucket.NewBasicAuth(config.BitbucketUsername, config.Token)
	return Bitbucket{c}, nil
}

func (_ Bitbucket) filter(config *configs.Config, resp interface{}) (repoData []Repo, err error) {
	var cloneData []Repo
	values := resp.(map[string]interface{})["values"].([]interface{})

	for _, a := range values {
		clone := a.(map[string]interface{})
		links := clone["links"].(map[string]interface{})["clone"].([]interface{})
		for _, l := range links {
			link := l.(map[string]interface{})["href"]
			linkType := l.(map[string]interface{})["name"]

			if config.MatchPrefix != "" {
				repoName := strings.ToLower(clone["name"].(string))
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
				var defaultBranch string
				if clone["mainbranch"] == nil {
					defaultBranch = "master"
				} else {
					defaultBranch = clone["mainbranch"].(map[string]interface{})["name"].(string)
				}

				r.CloneBranch = defaultBranch
			} else {
				r.CloneBranch = config.Branch
			}

			if config.Branch == "" {
				r.CloneBranch = "master"
			} else {
				r.CloneBranch = config.Branch
			}

			if config.CloneProtocol == "ssh" && linkType == "ssh" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				cloneData = append(cloneData, r)
			} else if config.CloneProtocol == "https" && linkType == "https" {
				r.URL = link.(string)
				r.CloneURL = link.(string)
				cloneData = append(cloneData, r)
			}
		}
	}

	return cloneData, nil
}
