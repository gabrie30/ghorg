// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/git"
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
		absolutePath := configs.EnsureTrailingSlashOnFilePath((cmd.Flag("path").Value.String()))
		os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", absolutePath)
	}

	if cmd.Flags().Changed("protocol") {
		protocol := cmd.Flag("protocol").Value.String()
		os.Setenv("GHORG_CLONE_PROTOCOL", protocol)
	}

	if cmd.Flags().Changed("branch") {
		os.Setenv("GHORG_BRANCH", cmd.Flag("branch").Value.String())
	}

	if cmd.Flags().Changed("github-app-pem-path") {
		os.Setenv("GHORG_GITHUB_APP_PEM_PATH", cmd.Flag("github-app-pem-path").Value.String())
	}

	if cmd.Flags().Changed("github-app-installation-id") {
		os.Setenv("GHORG_GITHUB_APP_INSTALLATION_ID", cmd.Flag("github-app-installation-id").Value.String())
	}

	if cmd.Flags().Changed("github-app-id") {
		os.Setenv("GHORG_GITHUB_APP_ID", cmd.Flag("github-app-id").Value.String())
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
		f := cmd.Flag("concurrency").Value.String()
		os.Setenv("GHORG_CONCURRENCY", f)
	}

	if cmd.Flags().Changed("clone-depth") {
		f := cmd.Flag("clone-depth").Value.String()
		os.Setenv("GHORG_CLONE_DEPTH", f)
	}

	if cmd.Flags().Changed("exit-code-on-clone-infos") {
		f := cmd.Flag("exit-code-on-clone-infos").Value.String()
		os.Setenv("GHORG_EXIT_CODE_ON_CLONE_INFOS", f)
	}

	if cmd.Flags().Changed("exit-code-on-clone-issues") {
		f := cmd.Flag("exit-code-on-clone-issues").Value.String()
		os.Setenv("GHORG_EXIT_CODE_ON_CLONE_ISSUES", f)
	}

	if cmd.Flags().Changed("topics") {
		topics := cmd.Flag("topics").Value.String()
		os.Setenv("GHORG_TOPICS", topics)
	}

	if cmd.Flags().Changed("match-prefix") {
		prefix := cmd.Flag("match-prefix").Value.String()
		os.Setenv("GHORG_MATCH_PREFIX", prefix)
	}

	if cmd.Flags().Changed("exclude-match-prefix") {
		prefix := cmd.Flag("exclude-match-prefix").Value.String()
		os.Setenv("GHORG_EXCLUDE_MATCH_PREFIX", prefix)
	}

	if cmd.Flags().Changed("gitlab-group-exclude-match-regex") {
		prefix := cmd.Flag("gitlab-group-exclude-match-regex").Value.String()
		os.Setenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX", prefix)
	}

	if cmd.Flags().Changed("match-regex") {
		regex := cmd.Flag("match-regex").Value.String()
		os.Setenv("GHORG_MATCH_REGEX", regex)
	}

	if cmd.Flags().Changed("exclude-match-regex") {
		regex := cmd.Flag("exclude-match-regex").Value.String()
		os.Setenv("GHORG_EXCLUDE_MATCH_REGEX", regex)
	}

	if cmd.Flags().Changed("ghorgignore-path") {
		path := cmd.Flag("ghorgignore-path").Value.String()
		os.Setenv("GHORG_IGNORE_PATH", path)
	}

	if cmd.Flags().Changed("git-filter") {
		filter := cmd.Flag("git-filter").Value.String()
		os.Setenv("GHORG_GIT_FILTER", filter)
	}

	if cmd.Flags().Changed("skip-archived") {
		os.Setenv("GHORG_SKIP_ARCHIVED", "true")
	}

	if cmd.Flags().Changed("no-clean") {
		os.Setenv("GHORG_NO_CLEAN", "true")
	}

	if cmd.Flags().Changed("prune") {
		os.Setenv("GHORG_PRUNE", "true")
	}

	if cmd.Flags().Changed("prune-no-confirm") {
		os.Setenv("GHORG_PRUNE_NO_CONFIRM", "true")
	}

	if cmd.Flags().Changed("fetch-all") {
		os.Setenv("GHORG_FETCH_ALL", "true")
	}

	if cmd.Flags().Changed("include-submodules") {
		os.Setenv("GHORG_INCLUDE_SUBMODULES", "true")
	}

	if cmd.Flags().Changed("dry-run") {
		os.Setenv("GHORG_DRY_RUN", "true")
	}

	if cmd.Flags().Changed("clone-wiki") {
		os.Setenv("GHORG_CLONE_WIKI", "true")
	}

	if cmd.Flags().Changed("insecure-gitlab-client") {
		os.Setenv("GHORG_INSECURE_GITLAB_CLIENT", "true")
	}

	if cmd.Flags().Changed("insecure-gitea-client") {
		os.Setenv("GHORG_INSECURE_GITEA_CLIENT", "true")
	}

	if cmd.Flags().Changed("skip-forks") {
		os.Setenv("GHORG_SKIP_FORKS", "true")
	}

	if cmd.Flags().Changed("quiet") {
		os.Setenv("GHORG_QUIET", "true")
	}

	if cmd.Flags().Changed("no-token") {
		os.Setenv("GHORG_NO_TOKEN", "true")
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
			if cmd.Flags().Changed("bitbucket-username") {
				os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", cmd.Flag("token").Value.String())
			} else {
				os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", cmd.Flag("token").Value.String())
			}
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

	setOutputDirName(argz)
	setOuputDirAbsolutePath()
	args = argz
	targetCloneSource = argz[0]
	setupRepoClone()
}

func setupRepoClone() {
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
		colorlog.PrintInfo("No repos found for " + os.Getenv("GHORG_SCM_TYPE") + " " + os.Getenv("GHORG_CLONE_TYPE") + ": " + targetCloneSource + ", please verify you have sufficient permissions to clone target repos, double check spelling and try again.")
		os.Exit(0)
	}
	git := git.NewGit()
	CloneAllRepos(git, cloneTargets)
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
	if _, err := os.Stat(outputDirAbsolutePath); os.IsNotExist(err) {
		err = os.MkdirAll(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), 0o700)
		if err != nil {
			panic(err)
		}
	}
}

func repoExistsLocally(repo scm.Repo) bool {
	if _, err := os.Stat(repo.HostPath); os.IsNotExist(err) {
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
		colorlog.PrintInfo("\n============ Info ============\n")
		for _, i := range cloneInfos {
			colorlog.PrintInfo(i)
		}
	}

	if len(cloneErrors) > 0 {
		colorlog.PrintError("\n============ Issues ============\n")
		for _, e := range cloneErrors {
			colorlog.PrintError(e)
		}
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

func filterByRegexMatch(repos []scm.Repo) []scm.Repo {
	filteredRepos := []scm.Repo{}
	regex := fmt.Sprint(os.Getenv("GHORG_MATCH_REGEX"))

	for i, r := range repos {
		re := regexp.MustCompile(regex)
		match := re.FindString(r.Name)
		if match != "" {
			filteredRepos = append(filteredRepos, repos[i])
		}
	}

	return filteredRepos
}

func filterByExcludeRegexMatch(repos []scm.Repo) []scm.Repo {
	filteredRepos := []scm.Repo{}
	regex := fmt.Sprint(os.Getenv("GHORG_EXCLUDE_MATCH_REGEX"))

	for i, r := range repos {
		exclude := false
		re := regexp.MustCompile(regex)
		match := re.FindString(r.Name)
		if match != "" {
			exclude = true
		}

		if !exclude {
			filteredRepos = append(filteredRepos, repos[i])
		}
	}

	return filteredRepos
}

func filterByMatchPrefix(repos []scm.Repo) []scm.Repo {
	filteredRepos := []scm.Repo{}
	for i, r := range repos {
		pfs := strings.Split(os.Getenv("GHORG_MATCH_PREFIX"), ",")
		for _, p := range pfs {
			if strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(p)) {
				filteredRepos = append(filteredRepos, repos[i])
			}
		}
	}

	return filteredRepos
}

func filterByExcludeMatchPrefix(repos []scm.Repo) []scm.Repo {
	filteredRepos := []scm.Repo{}
	for i, r := range repos {
		var exclude bool
		pfs := strings.Split(os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX"), ",")
		for _, p := range pfs {
			if strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(p)) {
				exclude = true
			}
		}

		if !exclude {
			filteredRepos = append(filteredRepos, repos[i])
		}
	}

	return filteredRepos
}

// exclude wikis from repo count
func getRepoCountOnly(targets []scm.Repo) int {
	count := 0
	for _, t := range targets {
		if !t.IsWiki {
			count++
		}
	}

	return count
}

func hasRepoNameCollisions(repos []scm.Repo) (map[string]bool, bool) {

	repoNameWithCollisions := make(map[string]bool)

	if os.Getenv("GHORG_GITLAB_TOKEN") == "" {
		return repoNameWithCollisions, false
	}

	if os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
		return repoNameWithCollisions, false
	}

	hasCollisions := false

	for _, repo := range repos {
		if _, ok := repoNameWithCollisions[repo.Name]; ok {
			repoNameWithCollisions[repo.Name] = true
			hasCollisions = true
		} else {
			repoNameWithCollisions[repo.Name] = false
		}
	}

	return repoNameWithCollisions, hasCollisions
}

func printDryRun(repos []scm.Repo) {
	for _, repo := range repos {
		colorlog.PrintSubtleInfo(repo.URL + "\n")
	}
	count := len(repos)
	colorlog.PrintSuccess(fmt.Sprintf("%v repos to be cloned into: %s", count, outputDirAbsolutePath))

	if os.Getenv("GHORG_PRUNE") == "true" {

		if stat, err := os.Stat(outputDirAbsolutePath); err == nil && stat.IsDir() {
			// We check that the clone path exists, otherwise there would definitely be no pruning
			// to do.
			colorlog.PrintInfo("\nScanning for local clones that have been removed on remote...")

			files, err := ioutil.ReadDir(outputDirAbsolutePath)
			if err != nil {
				log.Fatal(err)
			}

			eligibleForPrune := 0
			for _, f := range files {
				// for each item in the org's clone directory, let's make sure we found a
				// corresponding repo on the remote.
				if !sliceContainsNamedRepo(repos, f.Name()) {
					eligibleForPrune++
					colorlog.PrintSubtleInfo(fmt.Sprintf("%s not found in remote.", f.Name()))
				}
			}
			colorlog.PrintSuccess(fmt.Sprintf("Local clones eligible for pruning: %d", eligibleForPrune))
		}
	}
}

func trimCollisionFilename(filename string) string {
	maxLen := 248
	if len(filename) > maxLen {
		return filename[:strings.LastIndex(filename[:maxLen], "_")]
	}

	return filename
}

// CloneAllRepos clones all repos
func CloneAllRepos(git git.Gitter, cloneTargets []scm.Repo) {
	// Filter repos that have attributes that don't need specific scm api calls
	if os.Getenv("GHORG_MATCH_REGEX") != "" {
		colorlog.PrintInfo("Filtering repos down by including regex matches...")
		cloneTargets = filterByRegexMatch(cloneTargets)
	}
	if os.Getenv("GHORG_EXCLUDE_MATCH_REGEX") != "" {
		colorlog.PrintInfo("Filtering repos down by excluding regex matches...")
		cloneTargets = filterByExcludeRegexMatch(cloneTargets)
	}
	if os.Getenv("GHORG_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("Filtering repos down by including prefix matches...")
		cloneTargets = filterByMatchPrefix(cloneTargets)
	}
	if os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("Filtering repos down by excluding prefix matches...")
		cloneTargets = filterByExcludeMatchPrefix(cloneTargets)
	}

	// filter repos down based on ghorgignore if one exists
	_, err := os.Stat(configs.GhorgIgnoreLocation())
	if !os.IsNotExist(err) {
		// Open the file parse each line and remove cloneTargets containing
		toIgnore, err := readGhorgIgnore()
		if err != nil {
			colorlog.PrintErrorAndExit(fmt.Sprintf("Error parsing your ghorgignore, error: %v", err))
		}

		colorlog.PrintInfo("Using ghorgignore, filtering repos down...")

		filteredCloneTargets := []scm.Repo{}
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

	repoCount := getRepoCountOnly(cloneTargets)

	if os.Getenv("GHORG_CLONE_WIKI") == "true" {
		wikiCount := strconv.Itoa(len(cloneTargets) - repoCount)
		colorlog.PrintInfo(strconv.Itoa(repoCount) + " repos found in " + targetCloneSource + ", including " + wikiCount + " enabled wikis\n")
	} else {
		colorlog.PrintInfo(strconv.Itoa(repoCount) + " repos found in " + targetCloneSource + "\n")
	}

	if os.Getenv("GHORG_DRY_RUN") == "true" {
		printDryRun(cloneTargets)
		return
	}

	createDirIfNotExist()

	// check for duplicate names will cause issues for some clone types on gitlab
	repoNameWithCollisions, hasCollisions := hasRepoNameCollisions(cloneTargets)

	l, err := strconv.Atoi(os.Getenv("GHORG_CONCURRENCY"))
	if err != nil {
		log.Fatal("Could not determine GHORG_CONCURRENCY")
	}

	limit := limiter.NewConcurrencyLimiter(l)

	var cloneCount, pulledCount, updateRemoteCount int

	// maps in go are not safe for concurrent use
	var mutex = &sync.RWMutex{}

	for i := range cloneTargets {
		repo := cloneTargets[i]
		repoSlug := getAppNameFromURL(repo.URL)
		limit.Execute(func() {
			if repo.Path != "" && os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
				repoSlug = repo.Path
			}
			mutex.Lock()
			inHash := repoNameWithCollisions[repo.Name]
			mutex.Unlock()
			// Only GitLab repos can have collisions due to groups and subgroups
			// If there are collisions and this is a repo with a naming collision change name to avoid collisions
			if hasCollisions && inHash {
				repoSlug = trimCollisionFilename(strings.Replace(repo.Path, "/", "_", -1))

				mutex.Lock()
				slugCollision := repoNameWithCollisions[repoSlug]
				mutex.Unlock()
				// If a collision has another collision with trimmed name append a number
				if ok := slugCollision; ok {
					repoSlug = fmt.Sprintf("_%v_%v", strconv.Itoa(i), repoSlug)
				} else {
					mutex.Lock()
					repoNameWithCollisions[repoSlug] = true
					mutex.Unlock()
				}
			}

			repo.HostPath = filepath.Join(outputDirAbsolutePath, repoSlug)

			if repo.IsWiki {
				if !strings.HasSuffix(repo.HostPath, ".wiki") {
					repo.HostPath = repo.HostPath + ".wiki"
				}
			}

			action := "cloning"
			if repoExistsLocally(repo) {
				// prevents git from asking for user for credentials, needs to be unset so creds aren't stored
				err := git.SetOriginWithCredentials(repo)
				if err != nil {
					e := fmt.Sprintf("Problem setting remote with credentials Repo: %s Error: %v", repo.Name, err)
					cloneErrors = append(cloneErrors, e)
					return
				}

				if os.Getenv("GHORG_BACKUP") == "true" {
					err := git.UpdateRemote(repo)
					action = "updating remote"
					// Theres no way to tell if a github repo has a wiki to clone
					if err != nil && repo.IsWiki {
						e := fmt.Sprintf("Wiki may be enabled but there was no content to clone on Repo: %s Error: %v", repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}

					if err != nil {
						e := fmt.Sprintf("Could not update remotes in Repo: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}
					updateRemoteCount++
				} else if os.Getenv("GHORG_NO_CLEAN") == "true" {
					action = "fetching"
					err := git.FetchAll(repo)

					// Theres no way to tell if a github repo has a wiki to clone
					if err != nil && repo.IsWiki {
						e := fmt.Sprintf("Wiki may be enabled but there was no content to clone on Repo: %s Error: %v", repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}

					if err != nil {
						e := fmt.Sprintf("Could not fetch remotes in Repo: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

				} else {
					err := git.Checkout(repo)
					if err != nil {
						e := fmt.Sprintf("Could not checkout out %s, branch may not exist or may not have any contents, no changes made Repo: %s Error: %v", repo.CloneBranch, repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}

					err = git.Clean(repo)

					if err != nil {
						e := fmt.Sprintf("Problem running git clean: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

					err = git.Reset(repo)

					if err != nil {
						e := fmt.Sprintf("Problem resetting %s Repo: %s Error: %v", repo.CloneBranch, repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

					err = git.Pull(repo)

					if err != nil {
						e := fmt.Sprintf("Problem trying to pull %v Repo: %s Error: %v", repo.CloneBranch, repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}

					action = "pulling"
					pulledCount++

					if os.Getenv("GHORG_FETCH_ALL") == "true" {
						err = git.FetchAll(repo)

						if err != nil {
							e := fmt.Sprintf("Could not fetch remotes in Repo: %s Error: %v", repo.URL, err)
							cloneErrors = append(cloneErrors, e)
							return
						}
					}
				}

				err = git.SetOrigin(repo)
				if err != nil {
					e := fmt.Sprintf("Problem resetting remote Repo: %s Error: %v", repo.Name, err)
					cloneErrors = append(cloneErrors, e)
					return
				}
			} else {
				// if https clone and github/gitlab add personal access token to url

				err = git.Clone(repo)

				// Theres no way to tell if a github repo has a wiki to clone
				if err != nil && repo.IsWiki {
					e := fmt.Sprintf("Wiki may be enabled but there was no content to clone on Repo: %s Error: %v", repo.URL, err)
					cloneInfos = append(cloneInfos, e)
					return
				}

				if err != nil {
					e := fmt.Sprintf("Problem trying to clone Repo: %s Error: %v", repo.URL, err)
					cloneErrors = append(cloneErrors, e)
					return
				}

				if os.Getenv("GHORG_BRANCH") != "" {
					err := git.Checkout(repo)
					if err != nil {
						e := fmt.Sprintf("Could not checkout out %s, branch may not exist or may not have any contents, no changes made Repo: %s Error: %v", repo.CloneBranch, repo.URL, err)
						cloneInfos = append(cloneInfos, e)
						return
					}
				}

				cloneCount++

				// TODO: make configs around remote name
				// we clone with api-key in clone url
				err = git.SetOrigin(repo)

				// if repo has wiki, but content does not exist this is going to error
				if err != nil {
					e := fmt.Sprintf("Problem trying to set remote on Repo: %s Error: %v", repo.URL, err)
					cloneErrors = append(cloneErrors, e)
					return
				}

				if os.Getenv("GHORG_FETCH_ALL") == "true" {
					err = git.FetchAll(repo)

					if err != nil {
						e := fmt.Sprintf("Could not fetch remotes in Repo: %s Error: %v", repo.URL, err)
						cloneErrors = append(cloneErrors, e)
						return
					}
				}
			}

			colorlog.PrintSuccess("Success " + action + " repo: " + repo.URL + " -> branch: " + repo.CloneBranch)
		})

	}

	limit.WaitAndClose()

	printRemainingMessages()
	printCloneStatsMessage(cloneCount, pulledCount, updateRemoteCount)

	if hasCollisions {
		fmt.Println("")
		colorlog.PrintInfo("ATTENTION: ghorg detected collisions in repo names from the groups that were cloned. This occurs when one or more groups share common repo names trying to be cloned to the same directory. The repos that would have collisions were renamed with the group/subgroup appended.")
	}

	// Now, clean up local repos that don't exist in remote, if prune flag is set
	if os.Getenv("GHORG_PRUNE") == "true" {
		pruneRepos(cloneTargets)
	}

	if os.Getenv("GHORG_QUIET") != "true" {
		colorlog.PrintSuccess(fmt.Sprintf("\nFinished! %s", outputDirAbsolutePath))
	}

	if os.Getenv("GHORG_EXIT_CODE_ON_CLONE_INFOS") != "0" && len(cloneInfos) > 0 {
		exitCode, err := strconv.Atoi(os.Getenv("GHORG_EXIT_CODE_ON_CLONE_INFOS"))
		if err != nil {
			colorlog.PrintError("Could not convert GHORG_EXIT_CODE_ON_CLONE_INFOS from string to integer")
			os.Exit(1)
		}

		os.Exit(exitCode)
	}

	if len(cloneErrors) > 0 {
		exitCode, err := strconv.Atoi(os.Getenv("GHORG_EXIT_CODE_ON_CLONE_ISSUES"))
		if err != nil {
			colorlog.PrintError("Could not convert GHORG_EXIT_CODE_ON_CLONE_ISSUES from string to integer")
			os.Exit(1)
		}

		os.Exit(exitCode)
	}

}

func pruneRepos(cloneTargets []scm.Repo) {
	colorlog.PrintInfo("\nScanning for local clones that have been removed on remote...")

	files, err := ioutil.ReadDir(outputDirAbsolutePath)
	if err != nil {
		log.Fatal(err)
	}

	// The first time around, we set userAgreesToDelete to true, otherwise we'd immediately
	// break out of the loop.
	userAgreesToDelete := true
	pruneNoConfirm := os.Getenv("GHORG_PRUNE_NO_CONFIRM") == "true"
	for _, f := range files {
		// For each item in the org's clone directory, let's make sure we found a corresponding
		// repo on the remote.  We check userAgreesToDelete here too, so that if the user says
		// "No" at any time, we stop trying to prune things altogether.
		if userAgreesToDelete && !sliceContainsNamedRepo(cloneTargets, f.Name()) {
			// If the user specified --prune-no-confirm, we needn't prompt interactively.
			userAgreesToDelete = pruneNoConfirm || interactiveYesNoPrompt(
				fmt.Sprintf("%s was not found in remote.  Do you want to prune it?", f.Name()))
			if userAgreesToDelete {
				colorlog.PrintSubtleInfo(
					fmt.Sprintf("Deleting %s", filepath.Join(outputDirAbsolutePath, f.Name())))
				err = os.RemoveAll(filepath.Join(outputDirAbsolutePath, f.Name()))
				if err != nil {
					log.Fatal(err)
				}
			} else {
				colorlog.PrintError("Pruning cancelled by user.  No more prunes will be considered.")
			}
		}
	}
}

func printCloneStatsMessage(cloneCount, pulledCount, updateRemoteCount int) {
	if updateRemoteCount > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("New repos cloned: %v, existing repos pulled: %v, remotes updated: %v", cloneCount, pulledCount, updateRemoteCount))
		return
	}

	colorlog.PrintSuccess(fmt.Sprintf("New repos cloned: %v, existing repos pulled: %v", cloneCount, pulledCount))
}

func interactiveYesNoPrompt(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(strings.TrimSpace(prompt) + " (y/N) ")
	s, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}

// There's probably a nicer way of finding whether any scm.Repo in the slice matches a given name.
// TODO, currently this does not work if user sets --preserve-dir see https://github.com/gabrie30/ghorg/issues/210 for more info
func sliceContainsNamedRepo(haystack []scm.Repo, needle string) bool {

	if os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
		colorlog.PrintError("GHORG_PRUNE (--prune) does not currently work in combination with GHORG_PRESERVE_DIRECTORY_STRUCTURE (--preserve-dir), this will come in later versions")
		os.Exit(1)
	}

	for _, repo := range haystack {
		basepath := filepath.Base(repo.Path)

		if basepath == needle {
			return true
		}
	}

	return false
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
	if os.Getenv("GHORG_QUIET") == "true" {
		return
	}

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
	if os.Getenv("GHORG_CLONE_WIKI") == "true" {
		colorlog.PrintInfo("* Wikis         : " + os.Getenv("GHORG_CLONE_WIKI"))
	}
	if configs.GhorgIgnoreDetected() {
		colorlog.PrintInfo("* Ghorgignore   : " + configs.GhorgIgnoreLocation())
	}
	if os.Getenv("GHORG_MATCH_REGEX") != "" {
		colorlog.PrintInfo("* Regex Match   : " + os.Getenv("GHORG_MATCH_REGEX"))
	}
	if os.Getenv("GHORG_EXCLUDE_MATCH_REGEX") != "" {
		colorlog.PrintInfo("* Exclude Regex : " + os.Getenv("GHORG_EXCLUDE_MATCH_REGEX"))
	}
	if os.Getenv("GHORG_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("* Prefix Match  : " + os.Getenv("GHORG_MATCH_PREFIX"))
	}
	if os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX") != "" {
		colorlog.PrintInfo("* Exclude Prefix: " + os.Getenv("GHORG_EXCLUDE_MATCH_PREFIX"))
	}
	if os.Getenv("GHORG_INCLUDE_SUBMODULES") == "true" {
		colorlog.PrintInfo("* Submodules    : " + os.Getenv("GHORG_INCLUDE_SUBMODULES"))
	}
	if os.Getenv("GHORG_GIT_FILTER") != "" {
		colorlog.PrintInfo("* Git --filter= : " + os.Getenv("GHORG_GIT_FILTER"))
	}
	if os.Getenv("GHORG_OUTPUT_DIR") != "" {
		colorlog.PrintInfo("* Output Dir    : " + outputDirName)
	}
	if os.Getenv("GHORG_NO_CLEAN") == "true" {
		colorlog.PrintInfo("* No Clean      : " + "true")
	}
	if os.Getenv("GHORG_PRUNE") == "true" {
		noConfirmText := ""
		if os.Getenv("GHORG_PRUNE_NO_CONFIRM") == "true" {
			noConfirmText = " (skipping confirmation)"
		}
		colorlog.PrintInfo("* Prune         : " + "true" + noConfirmText)
	}
	if os.Getenv("GHORG_FETCH_ALL") == "true" {
		colorlog.PrintInfo("* Fetch All     : " + "true")
	}
	if os.Getenv("GHORG_DRY_RUN") == "true" {
		colorlog.PrintInfo("* Dry Run       : " + "true")
	}

	if os.Getenv("GHORG_RECLONE_PATH") != "" && os.Getenv("GHORG_RECLONE_RUNNING") == "true" {
		colorlog.PrintInfo("* Reclone Conf  : " + os.Getenv("GHORG_RECLONE_PATH"))
	}

	if os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
		colorlog.PrintInfo("* Preserve Dir  : " + "true")
	}

	if os.Getenv("GHORG_GITHUB_APP_PEM_PATH") != "" {
		colorlog.PrintInfo("* GH App Auth   : " + "true")
	}

	if os.Getenv("GHORG_CLONE_DEPTH") != "" {
		colorlog.PrintInfo("* Clone Depth   : " + os.Getenv("GHORG_CLONE_DEPTH"))
	}

	colorlog.PrintInfo("* Config Used   : " + os.Getenv("GHORG_CONFIG"))
	colorlog.PrintInfo("* Ghorg version : " + GetVersion())

	colorlog.PrintInfo("*************************************")
}

func getGhorgBranch() string {
	if os.Getenv("GHORG_BRANCH") == "" {
		return "default branch"
	}

	return os.Getenv("GHORG_BRANCH")
}

func setOuputDirAbsolutePath() {
	outputDirAbsolutePath = filepath.Join(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), outputDirName)
}

func setOutputDirName(argz []string) {
	if os.Getenv("GHORG_OUTPUT_DIR") != "" {
		outputDirName = os.Getenv("GHORG_OUTPUT_DIR")
		return
	}

	outputDirName = strings.ToLower(argz[0])

	// If all-group is used set the parent folder to the name of the baseurl
	if argz[0] == "all-groups" && os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u, err := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
		if err != nil {
			return
		}
		outputDirName = strings.TrimSuffix(strings.TrimPrefix(u.Host, "www."), ".com")
	}

	if argz[0] == "all-users" && os.Getenv("GHORG_SCM_BASE_URL") != "" {
		u, err := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
		if err != nil {
			return
		}
		outputDirName = strings.TrimSuffix(strings.TrimPrefix(u.Host, "www."), ".com")
		outputDirName = outputDirName + "_users"
	}

	if os.Getenv("GHORG_BACKUP") == "true" {
		outputDirName = outputDirName + "_backup"
	}
}
