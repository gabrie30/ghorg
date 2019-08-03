package cmd

import (
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
	args              []string
)

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - absolute path the ghorg_* directory will be created (default $HOME/Desktop)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_BITBUCKET_APP_PASSWORD - scm token to clone with")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - type of scm used, github, gitlab or bitbucket (default github)")
	// TODO: make gitlab terminology make sense https://about.gitlab.com/2016/01/27/comparing-terms-gitlab-github-bitbucket/
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - clone target type, user or org (default org)")
	cloneCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "GHORG_GITLAB_DEFAULT_NAMESPACE - gitlab only: limits clone targets to a specific namespace e.g. --namespace=gitlab-org/security-products")
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone user or org repos from GitHub, GitLab, or Bitbucket",
	Long:  `Clone user or org repos from GitHub, GitLab, or Bitbucket. See $HOME/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.`,
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

		configs.VerifyTokenSet()
		configs.VerifyConfigsSetCorrectly()

		args = argz

		CloneAllRepos()
	},
}

// TODO: Figure out how to use go channels for this
func getAllOrgCloneUrls() ([]string, error) {
	asciiTime()
	var urls []string
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		urls, err = getGitHubOrgCloneUrls()
	case "gitlab":
		urls, err = getGitLabOrgCloneUrls()
	case "bitbucket":
		urls, err = getBitBucketOrgCloneUrls()
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return urls, err
}

// TODO: Figure out how to use go channels for this
func getAllUserCloneUrls() ([]string, error) {
	asciiTime()
	var urls []string
	var err error
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		urls, err = getGitHubUserCloneUrls()
	case "gitlab":
		urls, err = getGitLabUserCloneUrls()
	case "bitbucket":
		urls, err = getBitBucketUserCloneUrls()
	default:
		colorlog.PrintError("GHORG_SCM_TYPE not set or unsupported, also make sure its all lowercase")
		os.Exit(1)
	}

	return urls, err
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

// CloneAllRepos clones all repos
func CloneAllRepos() {
	resc, errc, infoc := make(chan string), make(chan error), make(chan error)

	if os.Getenv("GHORG_BRANCH") != "master" {
		colorlog.PrintSubtleInfo("***********************************************************")
		colorlog.PrintSubtleInfo("* Ghorg will be running on branch: " + os.Getenv("GHORG_BRANCH"))
		colorlog.PrintSubtleInfo("* To change back to master update your $HOME/ghorg/conf.yaml or include -b=master flag")
		colorlog.PrintSubtleInfo("***********************************************************")
		fmt.Println()
	}

	var cloneTargets []string
	var err error

	if os.Getenv("GHORG_CLONE_TYPE") == "org" {
		cloneTargets, err = getAllOrgCloneUrls()
	} else if os.Getenv("GHORG_CLONE_TYPE") == "user" {
		cloneTargets, err = getAllUserCloneUrls()
	} else {
		colorlog.PrintError("GHORG_CLONE_TYPE not set or unsupported")
		os.Exit(1)
	}

	if len(cloneTargets) == 0 {
		colorlog.PrintInfo("No repos found for " + os.Getenv("GHORG_SCM_TYPE") + " " + os.Getenv("GHORG_CLONE_TYPE") + ": " + args[0] + ", check spelling and verify clone-type (user/org) is set correctly e.g. -c=user")
		os.Exit(0)
	}

	if err != nil {
		colorlog.PrintSubtleInfo("Did not find " + os.Getenv("GHORG_SCM_TYPE") + " " + os.Getenv("GHORG_CLONE_TYPE") + ": " + args[0] + ", check spelling and try again.")
		os.Exit(1)
	} else {
		colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos found in " + args[0])
		fmt.Println()
	}

	createDirIfNotExist()

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
