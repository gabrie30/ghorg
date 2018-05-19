package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

func createDirIfNotExist() {
	clonePath := os.Getenv("ABSOLUTE_PATH_TO_CLONE_TO")
	if _, err := os.Stat(clonePath + os.Args[1] + "_ghorg"); os.IsNotExist(err) {
		err = os.MkdirAll(clonePath, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func repoExistsLocally(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func getAppNameFromURL(url string) string {
	withGit := strings.Split(url, "/")
	appName := withGit[len(withGit)-1]
	split := strings.Split(appName, ".")
	return strings.Join(split[0:len(split)-1], ".")
}

// CloneAllReposByOrg clones all repos for a given org
func CloneAllReposByOrg() {
	resc, errc := make(chan string), make(chan error)
	createDirIfNotExist()
	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		fmt.Print("Problem fetching org repo urls to clone")
	}

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)
		go func(repoUrl string) {
			repoDir := os.Getenv("ABSOLUTE_PATH_TO_CLONE_TO") + os.Args[1] + "_ghorg" + "/" + appName

			if repoExistsLocally(repoDir) == true {
				cmd := exec.Command("git", "checkout", "master")
				cmd.Dir = repoDir
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error trying to checkout master from", repoDir)
					errc <- err
					return
				}

				cmd = exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					fmt.Println("Error trying to fetch all", repoDir)
					errc <- err
					return
				}

				cmd = exec.Command("git", "reset", "--hard", "origin/master")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					fmt.Println("Error trying to pull master from", repoDir)
					errc <- err
					return
				}
			} else {
				cmd := exec.Command("git", "clone", "-b master", repoUrl, repoDir)
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error trying to clone", repoUrl)
					errc <- err
					return
				}
			}

			resc <- repoUrl
		}(target)
	}

	for i := 0; i < len(cloneTargets); i++ {
		select {
		case res := <-resc:
			fmt.Println(res)
		case err := <-errc:
			fmt.Println(err)
		}
	}
	fmt.Println("Finished!")
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
