// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/scm"
	"github.com/korovkin/limiter"
	"github.com/spf13/cobra"
)

func cloneCmd() *cobra.Command {
	cloneCmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket",
		Long:  `Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket. See $HOME/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.`,
		Run:   cloneFunc,
	}
	cloneCmd.Flags().String("protocol", "https", "GHORG_PROTOCOL - protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringP("path", "p", configs.HomeDir()+"/Desktop/ghorg/", "GHORG_PATH - absolute path the ghorg_* directory will be created. Must end with / (default $HOME/Desktop/ghorg)")
	cloneCmd.Flags().StringP("branch", "b", "", "GHORG_BRANCH - branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringP("token", "t", "", "GHORG_TOKEN - scm token to clone with")
	cloneCmd.Flags().StringP("bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringP("scm", "s", "github", "GHORG_SCM - type of scm used, github, gitlab, gitea or bitbucket (default github)")
	cloneCmd.Flags().StringP("clone-type", "c", "org", "GHORG_CLONE_TYPE - clone target type, user or org (default org)")
	cloneCmd.Flags().Bool("skip-archived", false, "GHORG_SKIP_ARCHIVED - skips archived repos, github/gitlab/gitea only")
	cloneCmd.Flags().Bool("skip-forks", false, "GHORG_SKIP_FORKS - skips repo if its a fork, github/gitlab/gitea only")
	cloneCmd.Flags().Bool("preserve-dir", false, "GHORG_PRESERVE_DIR - clones repos in a directory structure that matches gitlab namespaces eg company/unit/subunit/app would clone into *_ghorg/unit/subunit/app, gitlab only")
	cloneCmd.Flags().Bool("backup", false, "GHORG_BACKUP - backup mode, clone as mirror, no working copy (ignores branch parameter)")
	cloneCmd.Flags().String("base-url", "", "GHORG_BASE_URL - change SCM base url, for on self hosted instances (currently gitlab, gitea and github (use format of https://git.mydomain.com/api/v3))")
	cloneCmd.Flags().Int("concurrency", 25, "GHORG_CONCURRENCY - max goroutines to spin up while cloning (default 25)")
	cloneCmd.Flags().StringSlice("topics", nil, "GHORG_TOPICS - comma separated list of github/gitea topics to filter for")
	cloneCmd.Flags().String("output-dir", "", "GHORG_OUTPUT_DIR - name of directory repos will be cloned into, will force underscores and always append _ghorg (default {org/repo being cloned}_ghorg)")
	cloneCmd.Flags().String("match-prefix", "", "GHORG_MATCH_PREFIX - only clone repos with matching prefix, can be a comma separated list (default \"\")")

	_ = viper.BindPFlag("protocol", cloneCmd.Flags().Lookup("protocol"))
	_ = viper.BindPFlag("path", cloneCmd.Flags().Lookup("path"))
	_ = viper.BindPFlag("branch", cloneCmd.Flags().Lookup("branch"))
	_ = viper.BindPFlag("token", cloneCmd.Flags().Lookup("token"))
	_ = viper.BindPFlag("bitbucket-username", cloneCmd.Flags().Lookup("bitbucket-username"))
	_ = viper.BindPFlag("scm", cloneCmd.Flags().Lookup("scm"))
	_ = viper.BindPFlag("clone-type", cloneCmd.Flags().Lookup("clone-type"))
	_ = viper.BindPFlag("skip-archived", cloneCmd.Flags().Lookup("skip-archived"))
	_ = viper.BindPFlag("skip-forks", cloneCmd.Flags().Lookup("skip-forks"))
	_ = viper.BindPFlag("preserve-dir", cloneCmd.Flags().Lookup("preserve-dir"))
	_ = viper.BindPFlag("backup", cloneCmd.Flags().Lookup("backup"))
	_ = viper.BindPFlag("base-url", cloneCmd.Flags().Lookup("base-url"))
	_ = viper.BindPFlag("concurrency", cloneCmd.Flags().Lookup("concurrency"))
	_ = viper.BindPFlag("topics", cloneCmd.Flags().Lookup("topics"))
	_ = viper.BindPFlag("output-dir", cloneCmd.Flags().Lookup("output-dir"))
	_ = viper.BindPFlag("match-prefix", cloneCmd.Flags().Lookup("match-prefix"))
	return cloneCmd
}

func cloneFunc(cmd *cobra.Command, argz []string) {
	config, err := configs.Load(argz)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Loading config failed: %s", err))
		os.Exit(1)
	}

	if len(argz) != 1 {
		colorlog.PrintError("You must provide an org or user to clone")
		os.Exit(1)
	}

	err = scm.VerifyScmType(config)
	if err != nil {
		colorlog.PrintError(err)
		os.Exit(1)
	}

	CloneAllRepos(argz[0], config)
}

func getCloneUrls(targetCloneSource string, config *configs.Config, isOrg bool) ([]scm.Repo, error) {
	asciiTime()
	config.Print()

	client, err := scm.GetClient(config, config.ScmType)
	if err != nil {
		colorlog.PrintError(err)
		os.Exit(1)
	}

	if isOrg {
		return client.GetOrgRepos(config, targetCloneSource)
	}
	return client.GetUserRepos(config, targetCloneSource)
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

func printRemainingMessages(cloneInfos, cloneErrors []string) {
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
func CloneAllRepos(targetCloneSource string, config *configs.Config) {
	var cloneTargets []scm.Repo
	var err error

	if config.CloneType == "org" {
		cloneTargets, err = getCloneUrls(targetCloneSource, config, true)
	} else if config.CloneType == "user" {
		cloneTargets, err = getCloneUrls(targetCloneSource, config, false)
	}

	if err != nil {
		colorlog.PrintError("Encountered an error, aborting")
		fmt.Println(err)
		os.Exit(1)
	}

	if len(cloneTargets) == 0 {
		colorlog.PrintInfo("No repos found for " + config.ScmType + " " + config.CloneType + ": " + targetCloneSource + ", check spelling and verify clone-type (user/org) is set correctly e.g. -c=user")
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

		var filteredCloneTargets []scm.Repo
		var flag bool
		for _, cloned := range cloneTargets {
			flag = false
			for _, ignore := range toIgnore {
				if strings.Contains(cloned.URL, ignore) {
					flag = true
				}
			}

			if !flag {
				filteredCloneTargets = append(filteredCloneTargets, cloned)
			}
		}

		cloneTargets = filteredCloneTargets

	}

	colorlog.PrintInfo(strconv.Itoa(len(cloneTargets)) + " repos found in " + targetCloneSource)
	fmt.Println()

	os.MkdirAll(config.Path, 0700)

	var cloneInfos, cloneErrors []string

	limit := limiter.NewConcurrencyLimiter(config.Concurrency)
	for _, target := range cloneTargets {
		appName := getAppNameFromURL(target.URL)
		branch := target.CloneBranch
		repo := target

		limit.Execute(func() {

			path := appName
			if repo.Path != "" && config.PreserveDirectoryStructure {
				path = repo.Path
			}

			repoDir := config.Path + "/" + path

			if config.Backup {
				repoDir = config.Path + "/" + path
			}

			if repoExistsLocally(repoDir) {
				if config.Backup {
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
				if config.Backup {
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

			colorlog.PrintSuccess("Success cloning repo: " + repo.URL + " -> branch: " + branch)
		})
	}

	limit.Wait()

	printRemainingMessages(cloneInfos, cloneErrors)

	colorlog.PrintSuccess(fmt.Sprintf("Finished! %s", config.Path))
}

func asciiTime() {
	colorlog.PrintInfo(
		`
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
 |T|I|M|E| |T|O| |G|H|O|R|G|
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
`)
}
