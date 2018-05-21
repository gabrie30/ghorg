package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getToken() string {
	if len(os.Getenv("GITHUB_TOKEN")) != 40 {
		color.New(color.FgYellow).Println("Note: GITHUB_TOKEN not set in .env, defaulting to keychain")
		fmt.Println()
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

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getToken()},
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
	resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	createDirIfNotExist()

	if os.Getenv("GHORG_BRANCH") != "master" {
		color.New(color.FgHiMagenta).Println("***********************************************************")
		color.New(color.FgHiMagenta).Println("* Ghorg will be running on branch: " + os.Getenv("GHORG_BRANCH"))
		color.New(color.FgHiMagenta).Println("* To change back to master run $ export GHORG_BRANCH=master")
		color.New(color.FgHiMagenta).Println("***********************************************************")
		fmt.Println()
	}

	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		color.New(color.FgRed).Println(err)
	} else {
		color.New(color.FgYellow).Println(strconv.Itoa(len(cloneTargets)) + " repos")
		fmt.Println()
	}

	branch := os.Getenv("GHORG_BRANCH")

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)

		go func(repoUrl string, branch string) {
			repoDir := os.Getenv("ABSOLUTE_PATH_TO_CLONE_TO") + os.Args[1] + "_ghorg" + "/" + appName

			if repoExistsLocally(repoDir) == true {
				cmd := exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch all Repo: "+repoUrl+" Error: %v", err)
					return
				}

				cmd = exec.Command("git", "checkout", branch)
				cmd.Dir = repoDir
				err := cmd.Run()
				if err != nil {
					infoc <- fmt.Errorf("Could not checkout out "+branch+", no changes made."+" Repo: "+repoUrl+" Error: %v", err)
					return
				}

				cmd = exec.Command("git", "reset", "--hard", "origin/"+branch)
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to pull "+branch+" Repo: "+repoUrl+" Error: %v", err)
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

				cmd = exec.Command("git", "checkout", branch)
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					infoc <- fmt.Errorf("Repo cloned but could not checkout "+branch+" Repo: "+repoUrl+" Error: %v", err)
					return
				}
			}

			resc <- repoUrl
		}(target, branch)
	}

	errors := []error{}
	infoMessages := []error{}

	for i := 0; i < len(cloneTargets); i++ {
		select {
		case res := <-resc:
			color.New(color.FgGreen).Println("Success " + res)
		case err := <-errc:
			errors = append(errors, err)
		case info := <-infoc:
			infoMessages = append(infoMessages, info)
		}
	}

	if len(infoMessages) > 0 {
		fmt.Println()
		color.New(color.FgYellow).Println("============ Info ============")
		fmt.Println()
		for _, i := range infoMessages {
			color.New(color.FgYellow).Println(i)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Println()
		color.New(color.FgRed).Println("============ Issues ============")
		fmt.Println()
		for _, e := range errors {
			color.New(color.FgRed).Println(e)
		}
		fmt.Println()
	}

	color.New(color.FgYellow).Println("Finished!")
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
