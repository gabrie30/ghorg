// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
)

var (
	protocol          string
	path              string
	branch            string
	token             string
	cloneType         string
	scmType           string
	bitbucketUsername string
	namespace         string
	color             string
	baseURL           string
	skipArchived      bool
	backup            bool
	args              []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&color, "color", "", "", "GHORG_COLOR - toggles colorful output on/off (default on)")
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - absolute path the ghorg_* directory will be created. Must end with / (default $HOME/Desktop/)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_BITBUCKET_APP_PASSWORD - scm token to clone with")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - type of scm used, github, gitlab or bitbucket (default github)")
	// TODO: make gitlab terminology make sense https://about.gitlab.com/2016/01/27/comparing-terms-gitlab-github-bitbucket/
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - clone target type, user or org (default org)")
	cloneCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "GHORG_GITLAB_DEFAULT_NAMESPACE - gitlab only: limits clone targets to a specific namespace e.g. --namespace=gitlab-org/security-products")

	cloneCmd.Flags().BoolVar(&skipArchived, "skip-archived", false, "GHORG_SKIP_ARCHIVED skips archived repos, github/gitlab only")
	cloneCmd.Flags().BoolVar(&skipArchived, "preserve-dir", false, "GHORG_PRESERVE_DIRECTORY_STRUCTURE clones repos in a directory structure that matches gitlab namespaces eg company/unit/subunit/app would clone into *_ghorg/unit/subunit/app, gitlab only")
	cloneCmd.Flags().BoolVar(&backup, "backup", false, "GHORG_BACKUP backup mode, clone as mirror, no working copy (ignores branch parameter)")

	cloneCmd.Flags().StringVarP(&baseURL, "base-url", "", "", "GHORG_SCM_BASE_URL change SCM base url, for on self hosted instances (currently gitlab only, use format of https://git.mydomain.com/api/v3)")
}

// Repo represents an SCM repo
type Repo struct {
	Name string
	Path string
	URL  string
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone user or org repos from GitHub, GitLab, or Bitbucket",
	Long:  `Clone user or org repos from GitHub, GitLab, or Bitbucket. See $HOME/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.`,
	Run: func(cmd *cobra.Command, argz []string) {

		if cmd.Flags().Changed("color") {
			colorToggle := cmd.Flag("color").Value.String()
			if colorToggle == "on" {
				os.Setenv("GHORG_COLOR", colorToggle)
			} else {
				os.Setenv("GHORG_COLOR", "off")
			}

		}

		if len(argz) < 1 {
			colorlog.PrintError("You must provide an org or user to clone")
			os.Exit(1)
		}

		if cmd.Flags().Changed("path") {
			absolutePath := ensureTrailingSlash(cmd.Flag("path").Value.String())
			os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", absolutePath)
		}

		if cmd.Flags().Changed("protocol") {
			protocol := cmd.Flag("protocol").Value.String()
			os.Setenv("GHORG_CLONE_PROTOCOL", protocol)
		}

		if cmd.Flags().Changed("branch") {
			os.Setenv("GHORG_BRANCH", cmd.Flag("branch").Value.String())
		}

		if cmd.Flags().Changed("bitbucket-username") {
			os.Setenv("GHORG_BITBUCKET_USERNAME", cmd.Flag("bitbucket-username").Value.String())
		}

		if cmd.Flags().Changed("namespace") {
			os.Setenv("GHORG_GITLAB_DEFAULT_NAMESPACE", cmd.Flag("namespace").Value.String())
		}

		if cmd.Flags().Changed("clone-type") {
			cloneType := strings.ToLower(cmd.Flag("clone-type").Value.String())
			os.Setenv("GHORG_CLONE_TYPE", cloneType)
		}

		if cmd.Flags().Changed("scm") {
			scmType := strings.ToLower(cmd.Flag("scm").Value.String())
			os.Setenv("GHORG_SCM_TYPE", scmType)
		}

		if cmd.Flags().Changed("base-url") {
			url := cmd.Flag("base-url").Value.String()
			os.Setenv("GHORG_SCM_BASE_URL", url)
		}

		if cmd.Flags().Changed("skip-archived") {
			os.Setenv("GHORG_SKIP_ARCHIVED", "true")
		}

		if cmd.Flags().Changed("preserve-dir") {
			os.Setenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE", "true")
		}

		if cmd.Flags().Changed("backup") {
			os.Setenv("GHORG_BACKUP", "true")
		}

		configs.GetOrSetToken()

		if cmd.Flags().Changed("token") {
			if os.Getenv("GHORG_SCM_TYPE") == "github" {
				os.Setenv("GHORG_GITHUB_TOKEN", cmd.Flag("token").Value.String())
			} else if os.Getenv("GHORG_SCM_TYPE") == "gitlab" {
				os.Setenv("GHORG_GITLAB_TOKEN", cmd.Flag("token").Value.String())
			} else if os.Getenv("GHORG_SCM_TYPE") == "bitbucket" {
				os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", cmd.Flag("token").Value.String())
			}
		}

		err := configs.VerifyTokenSet()
		if err != nil {
			colorlog.PrintError(err)
			os.Exit(1)
		}

		err = configs.VerifyConfigsSetCorrectly()
		if err != nil {
			colorlog.PrintError(err)
			os.Exit(1)
		}

		args = argz

		CloneAllRepos()
	},
}

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]Repo, error) {
	asciiTime()
	PrintConfigs()
	var repos []Repo
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		repos, err = getGitHubOrgCloneUrls()
	case "gitlab":
		repos, err = getGitLabOrgCloneUrls()
	case "bitbucket":
		repos, err = getBitBucketOrgCloneUrls()
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return repos, err
}

// TODO: Figure out how to use go channels for this
func getAllUserCloneUrls() ([]Repo, error) {
	asciiTime()
	PrintConfigs()
	var repos []Repo
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		repos, err = getGitHubUserCloneUrls()
	case "gitlab":
		repos, err = getGitLabUserCloneUrls()
	case "bitbucket":
		repos, err = getBitBucketUserCloneUrls()
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return repos, err
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

func readGhorgIgnore() ([]string, error) {
	file, err := os.Open(configs.GhorgIgnoreLocation())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// CloneAllRepos clones all repos
func CloneAllRepos() {
	resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	var cloneTargets []Repo
	var err error

	if os.Getenv("GHORG_CLONE_TYPE") == "org" {
		cloneTargets, err = getAllOrgCloneUrls()
	} else if os.Getenv("GHORG_CLONE_TYPE") == "user" {
		cloneTargets, err = getAllUserCloneUrls()
	} else {
		colorlog.PrintError("GHORG_CLONE_TYPE not set or unsupported")
		os.Exit(1)
	}

	if err != nil {
		colorlog.PrintError("Encountered an error, aborting")
		fmt.Println(err)
		os.Exit(1)
	}

	if len(cloneTargets) == 0 {
		colorlog.PrintInfo("No repos found for " + os.Getenv("GHORG_SCM_TYPE") + " " + os.Getenv("GHORG_CLONE_TYPE") + ": " + args[0] + ", check spelling and verify clone-type (user/org) is set correctly e.g. -c=user")
		os.Exit(0)
	}

	// filter repos down based on ghorgignore if one exists
	_, err = os.Stat(configs.GhorgIgnoreLocation())
	if !os.IsNotExist(err) {
		// Open the file parse each line and remove cloneTargets containing
		toIgnore, err := readGhorgIgnore()
		if err != nil {
			colorlog.PrintError("Error parsing your ghorgignore, aborting")
			fmt.Println(err)
			os.Exit(1)
		}

		colorlog.PrintInfo("Using ghorgignore, filtering repos down...")

		filteredCloneTargets := []Repo{}
		var flag bool
		for _, cloned := range cloneTargets {
			flag = false
			for _, ignore := range toIgnore {
				if strings.Contains(cloned.URL, ignore) {
					flag = true
				}
			}

			if flag == false {
				filteredCloneTargets = append(filteredCloneTargets, cloned)
			}
		}

		cloneTargets = filteredCloneTargets

	}

	colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos found in " + args[0])
	fmt.Println()

	createDirIfNotExist()

	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target.URL)

		go func(repo Repo, branch string) {

			path := appName
			if repo.Path != "" && os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
				path = repo.Path
			}

			repoDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + args[0] + "_ghorg" + "/" + path

			if os.Getenv("GHORG_BACKUP") == "true" {
				repoDir = os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + args[0] + "_ghorg_backup" + "/" + path
			}

			if repoExistsLocally(repoDir) == true {
				if os.Getenv("GHORG_BACKUP") == "true" {
					cmd := exec.Command("git", "remote", "update")
					cmd.Dir = repoDir
					err := cmd.Run()
					if err != nil {
						infoc <- fmt.Errorf("Could not update remotes in Repo: %s Error: %v", repo.URL, err)
						return
					}
				} else {
					cmd := exec.Command("git", "checkout", branch)
					cmd.Dir = repoDir
					err := cmd.Run()
					if err != nil {
						infoc <- fmt.Errorf("Could not checkout out %s, no changes made Repo: %s Error: %v", branch, repo.URL, err)
						return
					}

					cmd = exec.Command("git", "clean", "-f", "-d")
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						errc <- fmt.Errorf("Problem running git clean: %s Error: %v", repo.URL, err)
						return
					}

					cmd = exec.Command("git", "fetch", "-n", "origin", branch)
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						errc <- fmt.Errorf("Problem trying to fetch %v Repo: %s Error: %v", branch, repo.URL, err)
						return
					}

					cmd = exec.Command("git", "reset", "--hard", "origin/"+branch)
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						errc <- fmt.Errorf("Problem resetting %s Repo: %s Error: %v", branch, repo.URL, err)
						return
					}
				}
			} else {
				args := []string{"clone", repo.URL, repoDir}
				if os.Getenv("GHORG_BACKUP") == "true" {
					args = append(args, "--mirror")
				}
				cmd := exec.Command("git", args...)
				err := cmd.Run()
				if err != nil {
					errc <- fmt.Errorf("Problem trying to clone Repo: %s Error: %v", repo.URL, err)
					return
				}
			}

			resc <- repo.URL
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

	// TODO: fix all these if else checks with ghorg_backups
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg_backup", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), args[0]))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), args[0]))
	}
}

func asciiTime() {
	colorlog.PrintInfo(
		`
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
 |T|I|M|E| |T|O| |G|H|O|R|G|
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
`)
}

// PrintConfigs shows the user what is set before cloning
func PrintConfigs() {
	colorlog.PrintInfo("*************************************")
	colorlog.PrintInfo("* SCM      : " + os.Getenv("GHORG_SCM_TYPE"))
	colorlog.PrintInfo("* Type     : " + os.Getenv("GHORG_CLONE_TYPE"))
	colorlog.PrintInfo("* Protocol : " + os.Getenv("GHORG_CLONE_PROTOCOL"))
	colorlog.PrintInfo("* Branch   : " + os.Getenv("GHORG_BRANCH"))
	colorlog.PrintInfo("* Location : " + os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		colorlog.PrintInfo("* Base URL : " + os.Getenv("GHORG_SCM_BASE_URL"))
	}
	if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
		colorlog.PrintInfo("* Skip Archived : " + os.Getenv("GHORG_SKIP_ARCHIVED"))
	}
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintInfo("* Backup   : " + os.Getenv("GHORG_BACKUP"))
	}
	colorlog.PrintInfo("*************************************")
	fmt.Println("")
}

func ensureTrailingSlash(path string) string {
	if string(path[len(path)-1]) == "/" {
		return path
	}

	return path + "/"
}
