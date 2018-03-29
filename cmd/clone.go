package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func authSetup() (*github.Client, *github.RepositoryListByOrgOptions) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 1000},
	}

	return client, opt
}

// CloneAllReposByOrg clones all repos for a given org
func CloneAllReposByOrg() {
	client, opt := authSetup()
	repos, _, err := client.Repositories.ListByOrg(context.Background(), os.Args[1], opt)
	if err != nil {
		fmt.Print("Oh no error city: ", err)
	}

	fmt.Println(repos)
}

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
