package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	protocol string
	path     string
	branch   string
	token    string
	args     []string
)

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "protocol to clone with, ssh or https, (defaults to https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "absolute path the ghorg_* directory will be created (defaults to Desktop)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "branch left checked out for each repo cloned (defaults to master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "github token to clone with")
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a user or org",
	Long:  `Clone a user or org you can specify which you want to clone with flags, otherwise it will default to org first than user`,
	Run: func(cmd *cobra.Command, argz []string) {

		if len(argz) < 1 {
			colorlog.PrintError("You must provide an org or user to clone")
			os.Exit(1)
		}

		if cmd.Flags().Changed("path") {
			absolutePath := ensureTrailingSlash(cmd.Flag("path").Value.String())
			os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", absolutePath)
		}

		if cmd.Flags().Changed("protocol") {
			path := cmd.Flag("protocol").Value.String()
			if path != "ssh" && path != "https" {
				colorlog.PrintError("Protocol must be one of https or ssh")
				os.Exit(1)
			}
			os.Setenv("GHORG_CLONE_PROTOCOL", path)
		}

		if cmd.Flags().Changed("branch") {
			os.Setenv("GHORG_BRANCH", cmd.Flag("branch").Value.String())
		}

		if cmd.Flags().Changed("token") {
			os.Setenv("GHORG_GITHUB_TOKEN", cmd.Flag("token").Value.String())
		}

		args = argz

		CloneAllReposByOrg()
	},
}

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]string, error) {
	asciiTime()
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
	cloneUrls := []string{}

	for _, repo := range allRepos {
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			cloneUrls = append(cloneUrls, *repo.CloneURL)
		} else {
			cloneUrls = append(cloneUrls, *repo.SSHURL)
		}
	}

	return cloneUrls, nil
}

// TODO: refactor with getAllOrgCloneUrls
func getAllUserCloneUrls() ([]string, error) {
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
	cloneUrls := []string{}

	for _, repo := range allRepos {
		if os.Getenv("GHORG_CLONE_PROTOCOL") == "https" {
			cloneUrls = append(cloneUrls, *repo.CloneURL)
		} else {
			cloneUrls = append(cloneUrls, *repo.SSHURL)
		}
	}

	return cloneUrls, nil
}

func createDirIfNotExist() {
	if _, err := os.Stat(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + args[0] + "_ghorg"); os.IsNotExist(err) {
		err = os.MkdirAll(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), 0700)
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

	if os.Getenv("GHORG_BRANCH") != "master" {
		colorlog.PrintSubtleInfo("***********************************************************")
		colorlog.PrintSubtleInfo("* Ghorg will be running on branch: " + os.Getenv("GHORG_BRANCH"))
		colorlog.PrintSubtleInfo("* To change back to master update your $HOME/ghorg/conf.yaml or include -b=master flag")
		colorlog.PrintSubtleInfo("***********************************************************")
		fmt.Println()
	}

	cloneTargets, err := getAllOrgCloneUrls()

	if err != nil {
		colorlog.PrintSubtleInfo("Change of Plans! Did not find GitHub Org " + args[0] + " -- Looking instead for a GitHub User: " + args[0])
		fmt.Println()
		cloneTargets, err = getAllUserCloneUrls()
	}

	if err != nil {
		colorlog.PrintError(err)
		os.Exit(1)
	} else {
		colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos found in " + args[0])
		fmt.Println()
	}

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target)

		go func(repoUrl string, branch string) {
			repoDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + args[0] + "_ghorg" + "/" + appName

			if repoExistsLocally(repoDir) == true {

				cmd := exec.Command("git", "checkout", branch)
				cmd.Dir = repoDir
				err := cmd.Run()
				if err != nil {
					infoc <- fmt.Errorf("Could not checkout out %s, no changes made Repo: %s Error: %v", branch, repoUrl, err)
					return
				}

				cmd = exec.Command("git", "clean", "-f", "-d")
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem running git clean: %s Error: %v", repoUrl, err)
					return
				}

				cmd = exec.Command("git", "fetch", "-n", "origin", branch)
				cmd.Dir = repoDir
				err = cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to fetch %v Repo: %s Error: %v", branch, repoUrl, err)
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
			}

			resc <- repoUrl
		}(target, os.Getenv("GHORG_BRANCH"))
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

	colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), args[0]))
}

func asciiTime() {
	colorlog.PrintInfo(
		`
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
 |T|I|M|E| |T|O| |G|H|O|R|G|
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
`)
}

func debug() {
	fmt.Println(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
	fmt.Println(os.Getenv("GHORG_CLONE_PROTOCOL"))
	fmt.Println(os.Getenv("GHORG_BRANCH"))
}

func ensureTrailingSlash(path string) string {
	if string(path[len(path)-1]) == "/" {
		return path
	}

	return path + "/"
}
