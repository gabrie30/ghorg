package cmd

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getGitHubOrgCloneUrls() ([]Repo, error) {

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
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), args[0], opt)

		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	cloneData := []Repo{}

	for _, repo := range allRepos {
		r := Repo{}
		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *repo.Archived == true {
				continue
			}
		}

		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = addTokenToHTTPSCloneURL(*repo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
			r.URL = *repo.CloneURL
			cloneData = append(cloneData, r)
		} else {
			r.CloneURL = *repo.SSHURL
			r.URL = *repo.SSHURL
			cloneData = append(cloneData, r)
		}
	}

	return cloneData, nil
}

// TODO: refactor with getAllOrgCloneUrls
func getGitHubUserCloneUrls() ([]Repo, error) {
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
		repos, resp, err := client.Repositories.List(context.Background(), args[0], opt)

		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	repoData := []Repo{}

	for _, repo := range allRepos {

		if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
			if *repo.Archived == true {
				continue
			}
		}
		r := Repo{}
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			r.CloneURL = addTokenToHTTPSCloneURL(*repo.CloneURL, os.Getenv("GHORG_GITHUB_TOKEN"))
			r.URL = *repo.CloneURL
			repoData = append(repoData, r)
		} else {
			r.CloneURL = *repo.SSHURL
			r.URL = *repo.SSHURL
			repoData = append(repoData, r)
		}
	}

	return repoData, nil
}
