// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/internal/bitbucket"
	"github.com/gabrie30/ghorg/internal/gitea"
	"github.com/gabrie30/ghorg/internal/github"
	"github.com/gabrie30/ghorg/internal/gitlab"
	"github.com/gabrie30/ghorg/internal/repo"
	"github.com/korovkin/limiter"
	"github.com/spf13/cobra"
)

var (
	protocol          string
	path              string
	parentFolder      string
	branch            string
	token             string
	cloneType         string
	scmType           string
	bitbucketUsername string
	namespace         string
	color             string
	baseURL           string
	concurrency       string
	outputDir         string
	topics            string
	skipArchived      bool
	backup            bool
	args              []string
	cloneErrors       []string
	cloneInfos        []string
	targetCloneSource string
	matchPrefix       string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&color, "color", "", "", "GHORG_COLOR - toggles colorful output on/off (default on)")
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - absolute path the ghorg_* directory will be created. Must end with / (default $HOME/Desktop/ghorg)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_GITEA_TOKEN/GHORG_BITBUCKET_APP_PASSWORD - scm token to clone with")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - type of scm used, github, gitlab, gitea or bitbucket (default github)")
	// TODO: make gitlab terminology make sense https://about.gitlab.com/2016/01/27/comparing-terms-gitlab-github-bitbucket/
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - clone target type, user or org (default org)")
	cloneCmd.Flags().BoolVar(&skipArchived, "skip-archived", false, "GHORG_SKIP_ARCHIVED - skips archived repos, github/gitlab only")
	cloneCmd.Flags().BoolVar(&skipArchived, "preserve-dir", false, "GHORG_PRESERVE_DIRECTORY_STRUCTURE - clones repos in a directory structure that matches gitlab namespaces eg company/unit/subunit/app would clone into *_ghorg/unit/subunit/app, gitlab only")
	cloneCmd.Flags().BoolVar(&backup, "backup", false, "GHORG_BACKUP - backup mode, clone as mirror, no working copy (ignores branch parameter)")
	cloneCmd.Flags().StringVarP(&baseURL, "base-url", "", "", "GHORG_SCM_BASE_URL - change SCM base url, for on self hosted instances (currently gitlab/github only, use format of https://git.mydomain.com/api/v3)")
	cloneCmd.Flags().StringVarP(&concurrency, "concurrency", "", "", "GHORG_CONCURRENCY - max goroutines to spin up while cloning (default 25)")
	cloneCmd.Flags().StringVarP(&topics, "topics", "", "", "GHORG_TOPICS - comma seperated list of github topics to filter for")
	cloneCmd.Flags().StringVarP(&outputDir, "output-dir", "", "", "GHORG_OUTPUT_DIR - name of directory repos will be cloned into, will force underscores and always append _ghorg (default {org/repo being cloned}_ghorg)")
	cloneCmd.Flags().StringVarP(&matchPrefix, "match-prefix", "", "", "GHORG_MATCH_PREFIX - only clone repos with matching prefix, can be a comma separated list (default \"\")")
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket",
	Long:  `Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket. See $HOME/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.`,
	Run:   cloneFunc,
}

func cloneFunc(cmd *cobra.Command, argz []string) {

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

	if cmd.Flags().Changed("concurrency") {
		g := cmd.Flag("concurrency").Value.String()
		os.Setenv("GHORG_CONCURRENCY", g)
	}

	if cmd.Flags().Changed("topics") {
		topics := cmd.Flag("topics").Value.String()
		os.Setenv("GHORG_TOPICS", topics)
	}

	if cmd.Flags().Changed("match-prefix") {
		prefix := cmd.Flag("match-prefix").Value.String()
		os.Setenv("GHORG_MATCH_PREFIX", prefix)
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

	if cmd.Flags().Changed("output-dir") {
		d := cmd.Flag("output-dir").Value.String()
		os.Setenv("GHORG_OUTPUT_DIR", d)
	}

	configs.GetOrSetToken()

	if cmd.Flags().Changed("token") {
		if os.Getenv("GHORG_SCM_TYPE") == "github" {
			os.Setenv("GHORG_GITHUB_TOKEN", cmd.Flag("token").Value.String())
		} else if os.Getenv("GHORG_SCM_TYPE") == "gitlab" {
			os.Setenv("GHORG_GITLAB_TOKEN", cmd.Flag("token").Value.String())
		} else if os.Getenv("GHORG_SCM_TYPE") == "bitbucket" {
			os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", cmd.Flag("token").Value.String())
		} else if os.Getenv("GHORG_SCM_TYPE") == "gitea" {
			os.Setenv("GHORG_GITEA_TOKEN", cmd.Flag("token").Value.String())
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

	parseParentFolder(argz)
	args = argz
	targetCloneSource = argz[0]

	CloneAllRepos()
}

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]repo.Data, error) {
	asciiTime()
	PrintConfigs()
	var repos []repo.Data
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		ghc := github.NewGitHubClient()
		repos, err = github.GetOrgRepos(ghc, targetCloneSource)
	case "gitlab":
		repos, err = gitlab.GetOrgRepos(targetCloneSource)
	case "gitea":
		repos, err = gitea.GetOrgRepos(targetCloneSource)
	case "bitbucket":
		repos, err = bitbucket.GetOrgRepos(targetCloneSource)
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return repos, err
}

// TODO: Figure out how to use go channels for this
func getAllUserCloneUrls() ([]repo.Data, error) {
	asciiTime()
	PrintConfigs()
	var repos []repo.Data
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		repos, err = github.GetUserRepos(targetCloneSource)
	case "gitlab":
		repos, err = gitlab.GetUserRepos(targetCloneSource)
	case "gitea":
		repos, err = gitea.GetUserRepos(targetCloneSource)
	case "bitbucket":
		repos, err = bitbucket.GetUserRepos(targetCloneSource)
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return repos, err
}

func createDirIfNotExist() {
	if _, err := os.Stat(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + parentFolder + "_ghorg"); os.IsNotExist(err) {
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

func printRemainingMessages() {
	if len(cloneInfos) > 0 {
		fmt.Println()
		colorlog.PrintInfo("============ Info ============")
		fmt.Println()
		for _, i := range cloneInfos {
			colorlog.PrintInfo(i)
		}
		fmt.Println()
	}

	if len(cloneErrors) > 0 {
		fmt.Println()
		colorlog.PrintError("============ Issues ============")
		fmt.Println()
		for _, e := range cloneErrors {
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
		if scanner.Text() != "" {
			lines = append(lines, scanner.Text())
		}
	}
	return lines, scanner.Err()
}

// CloneAllRepos clones all repos
func CloneAllRepos() {
	// resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	var cloneTargets []repo.Data
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
		colorlog.PrintInfo("No repos found for " + os.Getenv("GHORG_SCM_TYPE") + " " + os.Getenv("GHORG_CLONE_TYPE") + ": " + targetCloneSource + ", check spelling and verify clone-type (user/org) is set correctly e.g. -c=user")
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
		fmt.Println("")

		filteredCloneTargets := []repo.Data{}
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

	colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos found in " + targetCloneSource)
	fmt.Println()

	createDirIfNotExist()

	l, err := strconv.Atoi(os.Getenv("GHORG_CONCURRENCY"))

	if err != nil {
		log.Fatal("Could not determine GHORG_CONCURRENCY")
	}

	limit := limiter.NewConcurrencyLimiter(l)
	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target.URL)
		branch := os.Getenv("GHORG_BRANCH")
		repo := target

		limit.Execute(func() {

			path := appName
			if repo.Path != "" && os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
				path = repo.Path
			}

			repoDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + parentFolder + "_ghorg" + "/" + path

			if os.Getenv("GHORG_BACKUP") == "true" {
				repoDir = os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + parentFolder + "_ghorg_backup" + "/" + path
			}

			if repoExistsLocally(repoDir) == true {
				if os.Getenv("GHORG_BACKUP") == "true" {
					cmd := exec.Command("git", "remote", "update")
					cmd.Dir = repoDir
					err := cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Could not update remotes in Repo: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}
				} else {

					cmd := exec.Command("git", "checkout", branch)
					cmd.Dir = repoDir
					err := cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Could not checkout out %s, branch may not exist, no changes made Repo: %s Error: %v", branch, repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}

					cmd = exec.Command("git", "clean", "-f", "-d")
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Problem running git clean: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

					cmd = exec.Command("git", "reset", "--hard", "origin/"+branch)
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Problem resetting %s Repo: %s Error: %v", branch, repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

					// TODO: handle case where repo was removed, should not give user an error
					cmd = exec.Command("git", "pull", "origin", branch)
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Problem trying to pull %v Repo: %s Error: %v", branch, repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}
				}
			} else {
				// if https clone and github/gitlab add personal access token to url

				args := []string{"clone", repo.CloneURL, repoDir}
				if os.Getenv("GHORG_BACKUP") == "true" {
					args = append(args, "--mirror")
				}

				cmd := exec.Command("git", args...)
				err := cmd.Run()

				if err != nil {
					e := fmt.Sprintf("Problem trying to clone Repo: %s Error: %v", repo.URL, err)
					cloneErrors = append(cloneErrors, e)
					return
				}

				// TODO: make configs around remote name
				// we clone with api-key in clone url
				args = []string{"remote", "set-url", "origin", repo.URL}
				cmd = exec.Command("git", args...)
				cmd.Dir = repoDir
				err = cmd.Run()

				if err != nil {
					e := fmt.Sprintf("Problem trying to set remote on Repo: %s Error: %v", repo.URL, err)
					cloneErrors = append(cloneErrors, e)
					return
				}
			}

			colorlog.PrintSuccess("Success " + repo.URL)
		})

	}

	limit.Wait()

	printRemainingMessages()

	// TODO: fix all these if else checks with ghorg_backups
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg_backup", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_ghorg", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder))
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
	colorlog.PrintInfo("* SCM           : " + os.Getenv("GHORG_SCM_TYPE"))
	colorlog.PrintInfo("* Type          : " + os.Getenv("GHORG_CLONE_TYPE"))
	colorlog.PrintInfo("* Protocol      : " + os.Getenv("GHORG_CLONE_PROTOCOL"))
	colorlog.PrintInfo("* Branch        : " + os.Getenv("GHORG_BRANCH"))
	colorlog.PrintInfo("* Location      : " + os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
	colorlog.PrintInfo("* Concurrency   : " + os.Getenv("GHORG_CONCURRENCY"))

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		colorlog.PrintInfo("* Base URL      : " + os.Getenv("GHORG_SCM_BASE_URL"))
	}
	if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
		colorlog.PrintInfo("* Skip Archived : " + os.Getenv("GHORG_SKIP_ARCHIVED"))
	}
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintInfo("* Backup        : " + os.Getenv("GHORG_BACKUP"))
	}
	if configs.GhorgIgnoreDetected() == true {
		colorlog.PrintInfo("* Ghorgignore   : true")
	}
	if os.Getenv("GHORG_OUTPUT_DIR") != "" {
		colorlog.PrintInfo("* Output Dir    : " + parentFolder + "_ghorg")
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

func addTokenToHTTPSCloneURL(url string, token string) string {
	splitURL := strings.Split(url, "https://")

	if os.Getenv("GHORG_SCM_TYPE") == "gitlab" {
		return "https://oauth2:" + token + "@" + splitURL[1]
	}

	return "https://" + token + "@" + splitURL[1]
}

func parseParentFolder(argz []string) {
	if os.Getenv("GHORG_OUTPUT_DIR") != "" {
		parentFolder = strings.ReplaceAll(os.Getenv("GHORG_OUTPUT_DIR"), "-", "_")
		return
	}

	pf := strings.ReplaceAll(argz[0], "-", "_")
	parentFolder = strings.ToLower(pf)
}
