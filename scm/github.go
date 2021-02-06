package scm

import (
	"context"
	"net/url"
	"strings"

	"github.com/gabrie30/ghorg/configs"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var (
	_ Client = Github{}
)

func init() {
	registerClient(Github{})
}

type Github struct {
	// extend the github client
	*github.Client
	// perPage contain the pagination item limit
	perPage int
}

func (_ Github) GetType() string {
	return "github"
}

// GetOrgRepos gets org repos
func (c Github) GetOrgRepos(config *configs.Config, targetOrg string) ([]Repo, error) {
	if config.ScmBaseUrl != "/" {
		c.BaseURL, _ = url.Parse(config.ScmBaseUrl)
	}

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: c.perPage},
	}

	// get all pages of results
	var allRepos []*github.Repository
	for {

		repos, resp, err := c.Repositories.ListByOrg(context.Background(), targetOrg, opt)

		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return c.filter(config, allRepos), nil
}

// GetUserRepos gets user repos
func (c Github) GetUserRepos(config *configs.Config, targetUser string) ([]Repo, error) {
	if config.ScmBaseUrl != "/" {
		c.BaseURL, _ = url.Parse(config.ScmBaseUrl)
	}

	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: c.perPage},
	}

	// get all pages of results
	var allRepos []*github.Repository
	for {

		repos, resp, err := c.Repositories.List(context.Background(), targetUser, opt)

		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return c.filter(config, allRepos), nil
}

// NewClient create new github scm client
func (_ Github) NewClient(config *configs.Config) (Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	c := github.NewClient(tc)

	client := Github{Client: c, perPage: 100}

	return client, nil
}

func (_ Github) addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://" + token + "@" + splitURL[1]
}

func (c Github) filter(config *configs.Config, allRepos []*github.Repository) []Repo {
	var repoData []Repo

	for _, ghRepo := range allRepos {
		if config.SkipArchived && *ghRepo.Archived {
			continue
		}

		if config.SkipForks && *ghRepo.Fork {
			continue
		}

		// If user defined a list of topics, check if any match with this repo
		if len(config.Topics) > 0 {
			foundTopic := false
			for _, topic := range ghRepo.Topics {
				for _, envTopic := range config.Topics {
					if topic == envTopic {
						foundTopic = true
						continue
					}
				}
			}
			if !foundTopic {
				continue
			}
		}

		if config.MatchPrefix != "" {
			repoName := strings.ToLower(*ghRepo.Name)
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
			defaultBranch := ghRepo.GetDefaultBranch()
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = config.Branch
		}

		if config.CloneProtocol == "https" {
			r.CloneURL = c.addTokenToHTTPSCloneURL(*ghRepo.CloneURL, config.Token)
			r.URL = *ghRepo.CloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = *ghRepo.SSHURL
			r.URL = *ghRepo.SSHURL
			repoData = append(repoData, r)
		}
	}

	return repoData
}
