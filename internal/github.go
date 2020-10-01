package internal

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/internal/base"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var (
	_ base.Client = GithubClient{}
)

func init() {
	RegisterClient(GithubClient{})
}

type GithubClient struct {
	client *github.Client
}

func (_ GithubClient) GetType() string {
	return "github"
}

// GetOrgRepos gets org repos
func (c GithubClient) GetOrgRepos(targetOrg string) ([]base.Repo, error) {
	client := c.newGitHubClient()

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u := configs.EnsureTrailingSlash(os.Getenv("GHORG_SCM_BASE_URL"))
		client.BaseURL, _ = url.Parse(u)
	}

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100, Page: 0},
	}

	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

	// get all pages of results
	var allRepos []*github.Repository
	for {

		repos, resp, err := client.Repositories.ListByOrg(context.Background(), targetOrg, opt)

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
func (c GithubClient) GetUserRepos(targetUser string) ([]base.Repo, error) {
	if c.client == nil {
		c.client = c.newGitHubClient()
	}

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u := configs.EnsureTrailingSlash(os.Getenv("GHORG_SCM_BASE_URL"))
		c.client.BaseURL, _ = url.Parse(u)
	}

	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100, Page: 0},
	}

	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

	// get all pages of results
	var allRepos []*github.Repository
	for {

		repos, resp, err := c.client.Repositories.List(context.Background(), targetUser, opt)

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

func (_ GithubClient) addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://" + token + "@" + splitURL[1]
}

func (c GithubClient) filter(allRepos []*github.Repository, envTopics []string) []base.Repo {
	var repoData []base.Repo

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

		r := base.Repo{}
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

// newGitHubClient creates a github client
func (_ GithubClient) newGitHubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}
