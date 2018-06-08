// Package cmd holds functions associated with cloning all of a given orgs repos
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/config"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getToken() string {
	if len(config.GitHubToken) != 40 {
		colorlog.PrintInfo("Note: GHORG_GITHUB_TOKEN not set in $HOME/.ghorg, defaulting to keychain")
		fmt.Println()
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token := strings.TrimSuffix(string(out), "\n")

		if len(token) != 40 {
			log.Fatal("Could not find a GitHub token in keychain, create token, set GITHUB_TOKEN in your $HOME/.ghorg")
		}

		return token
	}

	return config.GitHubToken
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
	if _, err := os.Stat(config.AbsolutePathToCloneTo + os.Args[1] + "_ghorg"); os.IsNotExist(err) {
		err = os.MkdirAll(config.AbsolutePathToCloneTo, 0700)
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

func printRemainingMessages(infoMessages []error, errors []error) {
	if len(infoMessages) > 0 {
		fmt.Println()
		colorlog.PrintInfo("============ Info ============")
		fmt.Println()
		for _, i := range infoMessages {
			colorlog.PrintInfo(i)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Println()
		colorlog.PrintError("============ Issues ============")
		fmt.Println()
		for _, e := range errors {
			colorlog.PrintError(e)
		}
		fmt.Println()
	}
}

// CloneAllReposByOrg clones all repos for a given org
func CloneAllReposByOrg() {
	resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	createDirIfNotExist()

	if config.GhorgBranch != "master" {
		colorlog.PrintSubtleInfo("***********************************************************")
		colorlog.PrintSubtleInfo("* Ghorg will be running on branch: " + config.GhorgBranch)
		colorlog.PrintSubtleInfo("* To change back to master run $ export GHORG_BRANCH=master")
		colorlog.PrintSubtleInfo("***********************************************************")
		fmt.Println()
	}

	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		colorlog.PrintError(err)
	} else {
		colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos")
		fmt.Println()
	}

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)

		go func(repoUrl string, branch string) {
			repoDir := config.AbsolutePathToCloneTo + os.Args[1] + "_ghorg" + "/" + appName

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
		}(target, config.GhorgBranch)
	}

	errors := []error{}
	infoMessages := []error{}

	for i := 0; i < len(cloneTargets); i++ {
		select {
		case res := <-resc:
			colorlog.PrintSuccess("Success " + res)
		case err := <-errc:
			errors = append(errors, err)
		case info := <-infoc:
			infoMessages = append(infoMessages, info)
		}
	}

	printRemainingMessages(infoMessages, errors)

	colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg", config.AbsolutePathToCloneTo, os.Args[1]))
}

// TODO: Clone via http or ssh flag

// Could clone all repos on a user
// orgs, _, err := client.Organizations.List(context.Background(), "willnorris", nil)
