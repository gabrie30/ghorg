package scm

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v72/github"
	"golang.org/x/oauth2"
)

var (
	_             Client = Github{}
	reposPerPage         = 100
	tokenUsername        = ""
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

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: c.perPage},
	}

	c.SetTokensUsername()

	spinningSpinner.Start()
	defer spinningSpinner.Stop()

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

	return c.filter(allRepos), nil
}

// GetUserRepos gets user repos
func (c Github) GetUserRepos(targetUser string) ([]Repo, error) {
	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		c.BaseURL, _ = url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
	}

	c.SetTokensUsername()

	spinningSpinner.Start()
	defer spinningSpinner.Stop()

	// get all pages of results
	var allRepos []*github.Repository
	opt := &github.ListOptions{PerPage: c.perPage, Page: 1}

	for {
		var repos []*github.Repository
		var resp *github.Response
		var err error

		if targetUser == tokenUsername {

			authOpt := &github.RepositoryListByAuthenticatedUserOptions{
				Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
				ListOptions: *opt,
			}
			// List repositories for the authenticated user
			repos, resp, err = c.Repositories.ListByAuthenticatedUser(context.Background(), authOpt)
		} else {
			userOpt := &github.RepositoryListByUserOptions{
				Type:        os.Getenv("GHORG_GITHUB_USER_OPTION"),
				ListOptions: *opt,
			}
			// List repositories for the specified user
			repos, resp, err = c.Repositories.ListByUser(context.Background(), targetUser, userOpt)
		}

		if err != nil {
			return nil, err
		}

		if targetUser != tokenUsername {
			userRepos := []*github.Repository{}

			for _, repo := range repos {
				if repo.Owner != nil && repo.Owner.Type != nil && *repo.Owner.Type == "User" {
					userRepos = append(userRepos, repo)
				}
			}

			repos = userRepos
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return c.filter(allRepos), nil
}

// NewClient create new github scm client
func (_ Github) NewClient() (Client, error) {
	ctx := context.Background()
	var tc *http.Client

	if os.Getenv("GHORG_GITHUB_TOKEN") != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	// Authenticate as a GitHub App
	// If the user has set GHORG_GITHUB_APP_PEM_PATH, we assume they want to use a GitHub App
	if os.Getenv("GHORG_GITHUB_APP_PEM_PATH") != "" {
		// If the user has set GHORG_GITHUB_APP_INSTALLATION_ID, we assume they want to use a GitHub App
		installID, err := strconv.ParseInt(os.Getenv("GHORG_GITHUB_APP_INSTALLATION_ID"), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("GHORG_GITHUB_APP_INSTALLATION_ID must be set if GHORG_GITHUB_APP_PEM_PATH is set")
		}

		appID, err := strconv.ParseInt(os.Getenv("GHORG_GITHUB_APP_ID"), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("GHORG_GITHUB_APP_ID must be set if GHORG_GITHUB_APP_PEM_PATH is set")
		}

		itr, err := ghinstallation.NewKeyFromFile(
			http.DefaultTransport,
			appID,
			installID,
			os.Getenv("GHORG_GITHUB_APP_PEM_PATH"),
		)
		if err != nil {
			return nil, err
		}
		tc = &http.Client{Transport: itr}
		// get the token from the itr and update the GHORT_GITHUB_TOKEN env var
		token, err := itr.Token(ctx)

		if err != nil {
			return nil, err
		}
		os.Setenv("GHORG_GITHUB_TOKEN", token)
	}

	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	var ghClient *github.Client

	if baseURL != "" {
		ghClient = github.NewClient(tc)
		ghClient, _ = ghClient.WithEnterpriseURLs(baseURL, baseURL)
	} else {
		ghClient = github.NewClient(tc)
	}

	client := Github{Client: ghClient, perPage: reposPerPage}

	return client, nil
}

func (_ Github) addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://" + tokenUsername + ":" + token + "@" + splitURL[1]
}

func (c Github) filter(allRepos []*github.Repository) []Repo {
	var repoData []Repo

	for _, ghRepo := range allRepos {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *ghRepo.Archived {
				continue
			}
		}

		if os.Getenv("GHORG_SKIP_FORKS") == "true" {
			if *ghRepo.Fork {
				continue
			}
		}

		if !hasMatchingTopic(ghRepo.Topics) {
			continue
		}

		// NOTE: for some reason forks do not always have a language field set so sometimes they get filtered out
		if os.Getenv("GHORG_GITHUB_FILTER_LANGUAGE") != "" {
			if ghRepo.Language != nil {
				ghLang := strings.ToLower(*ghRepo.Language)
				userLangs := strings.Split(strings.ToLower(os.Getenv("GHORG_GITHUB_FILTER_LANGUAGE")), ",")
				matched := false
				for _, userLang := range userLangs {
					if ghLang == userLang {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			} else {
				continue
			}
		}

		r := Repo{}

		r.Name = *ghRepo.Name
		r.Path = r.Name

		if os.Getenv("GHORG_BRANCH") == "" {
			defaultBranch := ghRepo.GetDefaultBranch()
			if defaultBranch == "" {
				defaultBranch = "master"
			}
			r.CloneBranch = defaultBranch
		} else {
			r.CloneBranch = os.Getenv("GHORG_BRANCH")
		}

		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" || os.Getenv("GHORG_GITHUB_APP_PEM_PATH") != "" {
			r.CloneURL = c.addTokenToHTTPSCloneURL(*ghRepo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
			r.URL = *ghRepo.CloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = *ghRepo.SSHURL
			r.URL = *ghRepo.SSHURL
			repoData = append(repoData, r)
		}

		if ghRepo.GetHasWiki() && os.Getenv("GHORG_CLONE_WIKI") == "true" {
			wiki := Repo{}
			wiki.IsWiki = true
			wiki.CloneURL = strings.Replace(r.CloneURL, ".git", ".wiki.git", 1)
			wiki.URL = strings.Replace(r.URL, ".git", ".wiki.git", 1)
			wiki.CloneBranch = "master"
			wiki.Path = fmt.Sprintf("%s%s", r.Name, ".wiki")
			repoData = append(repoData, wiki)
		}
	}

	return repoData
}

// Sets the GitHub username tied to the github token to the package variable tokenUsername
// Then if https clone method is used the clone url will be https://username:token@github.com/org/repo.git
// The username is now needed when using the new fine-grained tokens for github
func (c Github) SetTokensUsername() {
	if os.Getenv("GHORG_GITHUB_TOKEN_FROM_GITHUB_APP") == "true" || os.Getenv("GHORG_GITHUB_APP_PEM_PATH") != "" {
		tokenUsername = "x-access-token"
		return
	}
	userToken, _, _ := c.Users.Get(context.Background(), "")
	tokenUsername = userToken.GetLogin()
}
