package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getToken() string {
	if len(os.Getenv("GITHUB_TOKEN")) != 40 {
		color.New(color.FgYellow).Println("GITHUB_TOKEN not set in .ghorg, defaulting to keychain")
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return color.New(color.FgRed).Sprintf("Failed to execute command: %s", cmd)
		}

		token := strings.TrimSuffix(string(out), "\n")

		if len(token) != 40 {
			log.Fatal("Could not find a GitHub token in keychain, create token and set GITHUB_TOKEN in .env")
		}

		return token
	}

	return os.Getenv("GITHUB_TOKEN")
}

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]string, error) {
	ctx := context.Background()
	githubToken := getToken()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(githubToken)},
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
		err = os.MkdirAll(clonePath, 0666)
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
		color.New(color.FgRed).Println(err)
	}

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)
		go func(repoUrl string) {
			repoDir := os.Getenv("ABSOLUTE_PATH_TO_CLONE_TO") + os.Args[1] + "_ghorg" + "/" + appName

			if repoExistsLocally(repoDir) == true {
				cmd := exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch all Repo: "+repoUrl+" Error: %v", err)
					return
				}

				cmd = exec.Command("git", "checkout", "master")
				cmd.Dir = repoDir
				err := cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem checking out master Repo: "+repoUrl+" Error: %v", err)
					return
				}

				cmd = exec.Command("git", "reset", "--hard", "origin/master")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to pull master Repo: "+repoUrl+" Error: %v", err)
					return
				}
			} else {
				cmd := exec.Command("git", "clone", repoUrl, repoDir)
				err := cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to clone Repo: "+repoUrl+" Error: %v", err)
					return
				}

				cmd = exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch all Repo: "+repoUrl+" Error: %v", err)
					return
				}
			}

			resc <- repoUrl
		}(target)
	}

	for i := 0; i < len(cloneTargets); i++ {
		select {
		case res := <-resc:
			color.New(color.FgGreen).Println("Success " + res)
		case err := <-errc:
			color.New(color.FgRed).Println(err)
		}
	}
	color.New(color.FgYellow).Println("Finished!")
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
