package scm

import (
	"context"
	"net/url"
	"os"
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
func (c Github) GetOrgRepos(targetOrg string) ([]Repo, error) {
	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u := configs.EnsureTrailingSlash(os.Getenv("GHORG_SCM_BASE_URL"))
		c.BaseURL, _ = url.Parse(u)
	}

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: c.perPage},
	}

	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

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

	return c.filter(allRepos, envTopics), nil
}

// GetUserRepos gets user repos
func (c Github) GetUserRepos(targetUser string) ([]Repo, error) {
	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u := configs.EnsureTrailingSlash(os.Getenv("GHORG_SCM_BASE_URL"))
		c.BaseURL, _ = url.Parse(u)
	}

	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: c.perPage},
	}

	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

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

	return c.filter(allRepos, envTopics), nil
}

// NewClient create new github scm client
func (_ Github) NewClient() (Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
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

func (c Github) filter(allRepos []*github.Repository, envTopics []string) []Repo {
	var repoData []Repo

	for _, ghRepo := range allRepos {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *ghRepo.Archived == true {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if *ghRepo.Fork == true {
				continue
			}
		}

		// If user defined a list of topics, check if any match with this repo
		if os.Getenv("GHORG_TOPICS") != "" {
			foundTopic := false
			for _, topic := range ghRepo.Topics {
				for _, envTopic := range envTopics {
					if topic == envTopic {
						foundTopic = true
						continue
					}
				}
			}
			if foundTopic == false {
				continue
			}
		}

		if os.Getenv("GHORG_MATCH_PREFIX") != "" {
			repoName := strings.ToLower(*ghRepo.Name)
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
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = c.addTokenToHTTPSCloneURL(*ghRepo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
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
