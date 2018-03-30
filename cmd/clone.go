package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
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
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), os.Args[1], opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	cloneUrls := []string{}
	for _, repo := range allRepos {
		cloneUrls = append(cloneUrls, *repo.CloneURL)
	}

	return cloneUrls, nil
}

// CloneAllReposByOrg clones all repos for a given org
func CloneAllReposByOrg() error {
	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		return errors.New("Problem fetching org repo urls to clone")
	}

	for _, target := range cloneTargets {
		go func(repoUrl string) (string, error) {
			fmt.Println("Cloning!!!!!!", repoUrl)
			cmd := exec.Command("git", "clone", repoUrl)
			err := cmd.Run()
			if err != nil {
				fmt.Print("ERROR DETECTEDs")
				return repoUrl, err
			}

			return "Done", nil
		}(target)
	}

	fmt.Scanln("Press any key when things look done")
	return nil
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
