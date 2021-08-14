// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/scm"
	"github.com/korovkin/limiter"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket",
	Long:  `Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket. See $HOME/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.`,
	Run:   cloneFunc,
}

func cloneFunc(cmd *cobra.Command, argz []string) {

	if cmd.Flags().Changed("path") {
		absolutePath := configs.EnsureTrailingSlash((cmd.Flag("path").Value.String()))
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

	if cmd.Flags().Changed("match-regex") {
		regex := cmd.Flag("match-regex").Value.String()
		os.Setenv("GHORG_MATCH_REGEX", regex)
	}

	if cmd.Flags().Changed("skip-archived") {
		os.Setenv("GHORG_SKIP_ARCHIVED", "true")
	}

	if cmd.Flags().Changed("skip-forks") {
		os.Setenv("GHORG_SKIP_FORKS", "true")
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

	if len(argz) < 1 {
		if os.Getenv("GHORG_SCM_TYPE") == "github" && os.Getenv("GHORG_CLONE_TYPE") == "user" {
			argz = append(argz, "")
		} else {
			colorlog.PrintError("You must provide an org or user to clone")
			os.Exit(1)
		}
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

func getAllOrgCloneUrls() ([]scm.Repo, error) {
	return getCloneUrls(true)
}

func getAllUserCloneUrls() ([]scm.Repo, error) {
	return getCloneUrls(false)
}

func getCloneUrls(isOrg bool) ([]scm.Repo, error) {
	asciiTime()
	PrintConfigs()
	scmType := strings.ToLower(os.Getenv("GHORG_SCM_TYPE"))
	if len(scmType) == 0 {
		colorlog.PrintError("GHORG_SCM_TYPE not set")
		os.Exit(1)
	}
	client, err := scm.GetClient(scmType)
	if err != nil {
		colorlog.PrintError(err)
		os.Exit(1)
	}

	if isOrg {
		return client.GetOrgRepos(targetCloneSource)
	}
	return client.GetUserRepos(targetCloneSource)
}

func createDirIfNotExist() {
	if _, err := os.Stat(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + parentFolder); os.IsNotExist(err) {
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

func filterByRegex(repos []scm.Repo) []scm.Repo {
	filteredRepos := []scm.Repo{}
	regex := fmt.Sprintf(`%s`, os.Getenv("GHORG_MATCH_REGEX"))

	for i, r := range repos {
		re := regexp.MustCompile(regex)
		match := re.FindString(getAppNameFromURL(r.URL))
		if match != "" {
			filteredRepos = append(filteredRepos, repos[i])
		}
	}

	return filteredRepos
}

// CloneAllRepos clones all repos
func CloneAllRepos() {
	// resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	var cloneTargets []scm.Repo
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

	if os.Getenv("GHORG_MATCH_REGEX") != "" {
		colorlog.PrintInfo("Filtering repos down by regex that match the provided...")
		fmt.Println("")
		cloneTargets = filterByRegex(cloneTargets)
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

		filteredCloneTargets := []scm.Repo{}
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
		branch := target.CloneBranch
		repo := target

		limit.Execute(func() {

			path := appName
			if repo.Path != "" && os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
				path = repo.Path
			}

			repoDir := filepath.Join(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder, configs.GetCorrectFilePathSeparator(), path)

			if os.Getenv("GHORG_BACKUP") == "true" {
				repoDir = filepath.Join(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder+"_backup", configs.GetCorrectFilePathSeparator(), path)
			}

			action := "cloning"

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

					action = "updating"
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

				if os.Getenv("GHORG_BRANCH") != "" {
					cmd = exec.Command("git", "checkout", branch)
					cmd.Dir = repoDir
					err = cmd.Run()
					if err != nil {
						e := fmt.Sprintf("Could not checkout out %s, branch may not exist, no changes made Repo: %s Error: %v", branch, repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}
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

			colorlog.PrintSuccess("Success " + action + " repo: " + repo.URL + " -> branch: " + branch)
		})

	}

	limit.Wait()

	printRemainingMessages()

	// TODO: fix all these if else checks with ghorg_backups
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s_backup", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("Finished! %s%s", os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), parentFolder))
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
	colorlog.PrintInfo("* Location      : " + os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"))
	colorlog.PrintInfo("* Concurrency   : " + os.Getenv("GHORG_CONCURRENCY"))

	if os.Getenv("GHORG_BRANCH") != "" {
		colorlog.PrintInfo("* Branch        : " + getGhorgBranch())
	}

	if os.Getenv("GHORG_SCM_BASE_URL") != "" {
		colorlog.PrintInfo("* Base URL      : " + os.Getenv("GHORG_SCM_BASE_URL"))
	}
	if os.Getenv("GHORG_SKIP_ARCHIVED") == "true" {
		colorlog.PrintInfo("* Skip Archived : " + os.Getenv("GHORG_SKIP_ARCHIVED"))
	}
	if os.Getenv("GHORG_SKIP_FORKS") == "true" {
		colorlog.PrintInfo("* Skip Forks    : " + os.Getenv("GHORG_SKIP_FORKS"))
	}
	if os.Getenv("GHORG_BACKUP") == "true" {
		colorlog.PrintInfo("* Backup        : " + os.Getenv("GHORG_BACKUP"))
	}
	if configs.GhorgIgnoreDetected() == true {
		colorlog.PrintInfo("* Ghorgignore   : true")
	}
	if os.Getenv("GHORG_MATCH_REGEX") != "" {
		colorlog.PrintInfo("* Regex Match   : " + os.Getenv("GHORG_MATCH_REGEX"))
	}
	if os.Getenv("GHORG_OUTPUT_DIR") != "" {
		colorlog.PrintInfo("* Output Dir    : " + parentFolder)
	}
	colorlog.PrintInfo("* Config Used   : " + os.Getenv("GHORG_CONF"))

	colorlog.PrintInfo("*************************************")
	fmt.Println("")
}

func getGhorgBranch() string {
	if os.Getenv("GHORG_BRANCH") == "" {
		return "default branch"
	}

	return os.Getenv("GHORG_BRANCH")
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
		parentFolder = os.Getenv("GHORG_OUTPUT_DIR")
		return
	}

	parentFolder = strings.ToLower(argz[0])
}
