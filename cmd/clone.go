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
		printInfo("Note: GITHUB_TOKEN not set in .env, defaulting to keychain")
		fmt.Println()
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			printError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token := strings.TrimSuffix(string(out), "\n")

		if len(token) != 40 {
			log.Fatal("Could not find a GitHub token in keychain, create token, set GITHUB_TOKEN in your .env, and then make install")
		}

		return token
	}

	return os.Getenv("GITHUB_TOKEN")
}

func printInfo(msg ...interface{}) {
	color.New(color.FgYellow).Println(msg)
}

func printSuccess(msg ...interface{}) {
	color.New(color.FgGreen).Println(msg)
}

func printError(msg ...interface{}) {
	color.New(color.FgRed).Println(msg)
}

func printSubtle(msg ...interface{}) {
	color.New(color.FgHiMagenta).Println(msg)
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
		printSubtle("***********************************************************")
		printSubtle("* Ghorg will be running on branch: " + os.Getenv("GHORG_BRANCH"))
		printSubtle("* To change back to master run $ export GHORG_BRANCH=master")
		printSubtle("***********************************************************")
		fmt.Println()
	}

	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		printError(err)
	} else {
		printInfo(strconv.Itoa(len(cloneTargets)) + " repos")
		fmt.Println()
	}

	branch := os.Getenv("GHORG_BRANCH")

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)

		go func(repoUrl string, branch string) {
			repoDir := os.Getenv("ABSOLUTE_PATH_TO_CLONE_TO") + os.Args[1] + "_ghorg" + "/" + appName

			if repoExistsLocally(repoDir) == true {

				cmd := exec.Command("git", "checkout", branch)
				cmd.Dir = repoDir
				err := cmd.Run()
				if err != nil {
					infoc <- fmt.Errorf("Could not checkout out %s, no changes made Repo: %s Error: %v", branch, repoUrl, err)
					return
				}

				cmd = exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch all Repo: %s Error: %v", repoUrl, err)
					return
				}

				cmd = exec.Command("git", "reset", "--hard", "origin/"+branch)
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem resetting %s Repo: %s Error: %v", branch, repoUrl, err)
					return
				}
			} else {
				cmd := exec.Command("git", "clone", repoUrl, repoDir)
				err := cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to clone Repo: %s Error: %v", repoUrl, err)
					return
				}

				cmd = exec.Command("git", "fetch", "--all")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch all Repo: %s Error: %v", repoUrl, err)
					return
				}

				cmd = exec.Command("git", "checkout", branch)
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					infoc <- fmt.Errorf("Repo cloned but could not checkout %s Repo: %s Error: %v", branch, repoUrl, err)
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
			printSuccess("Success " + res)
		case err := <-errc:
			errors = append(errors, err)
		case info := <-infoc:
			infoMessages = append(infoMessages, info)
		}
	}

	if len(infoMessages) > 0 {
		fmt.Println()
		printInfo("============ Info ============")
		fmt.Println()
		for _, i := range infoMessages {
			printInfo(i)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Println()
		printError("============ Issues ============")
		fmt.Println()
		for _, e := range errors {
			printError(e)
		}
		fmt.Println()
	}

	printInfo("Finished!")
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
