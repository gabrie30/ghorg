package github

import (
	"context"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/internal/repo"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// GetOrgRepos gets org repos
func GetOrgRepos(targetOrg string) ([]repo.Data, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100, Page: 0},
	}

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
	cloneData := []repo.Data{}

	for _, ghRepo := range allRepos {
		r := repo.Data{}
		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *ghRepo.Archived == true {
				continue
			}
		}

		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = addTokenToHTTPSCloneURL(*ghRepo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
			r.URL = *ghRepo.CloneURL
			cloneData = append(cloneData, r)
		} else {
			r.CloneURL = *ghRepo.SSHURL
			r.URL = *ghRepo.SSHURL
			cloneData = append(cloneData, r)
		}
	}

	return cloneData, nil
}

// GetUserRepos gets user repos
func GetUserRepos(targetUser string) ([]repo.Data, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GHORG_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100, Page: 0},
	}
	// get all pages of results
	var allRepos []*github.Repository
	for {

		repos, resp, err := client.Repositories.List(context.Background(), targetUser, opt)

		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	repoData := []repo.Data{}

	for _, ghRepo := range allRepos {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *ghRepo.Archived == true {
				continue
			}
		}
		r := repo.Data{}
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = addTokenToHTTPSCloneURL(*ghRepo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
			r.URL = *ghRepo.CloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = *ghRepo.SSHURL
			r.URL = *ghRepo.SSHURL
			repoData = append(repoData, r)
		}
	}

	return repoData, nil
}

func addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")
	return "https://" + token + "@" + splitURL[1]
}
