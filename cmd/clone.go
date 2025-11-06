// Package cmd encapsulates the logic for all cli commands
package cmd

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/gabrie30/ghorg/git"
	"github.com/gabrie30/ghorg/scm"
	"github.com/gabrie30/ghorg/utils"
	"github.com/korovkin/limiter"
	"github.com/spf13/cobra"
)

// Helper function to safely parse integer environment variables
func parseIntEnv(envVar string) (int, error) {
	value := os.Getenv(envVar)
	if value == "" {
		return 0, fmt.Errorf("environment variable %s not set", envVar)
	}
	return strconv.Atoi(value)
}

// Helper function to get clone delay configuration
func getCloneDelaySeconds() (int, bool) {
	delaySeconds, err := parseIntEnv("GHORG_CLONE_DELAY_SECONDS")
	if err != nil || delaySeconds <= 0 {
		return 0, false
	}
	return delaySeconds, true
}

// Helper function to check if concurrency should be auto-adjusted for delay
func shouldAutoAdjustConcurrency() (int, bool, bool) {
	delaySeconds, hasDelay := getCloneDelaySeconds()
	if !hasDelay {
		return 0, false, false
	}

	concurrency, err := parseIntEnv("GHORG_CONCURRENCY")
	if err != nil || concurrency <= 1 {
		return delaySeconds, false, false
	}

	return delaySeconds, true, true
}

var cloneCmd = &cobra.Command{
	Use:   "clone [org/user]",
	Short: "Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket",
	Long: `Clone user or org repos from GitHub, GitLab, Gitea or Bitbucket. See $HOME/.config/ghorg/conf.yaml for defaults, its likely you will need to update some of these values of use the flags to overwrite them. Values are set first by a default value, then based off what is set in $HOME/.config/ghorg/conf.yaml, finally the cli flags, which have the highest level of precedence.

For complete examples of how to clone repos from each SCM provider, run one of the following examples commands:
$ ghorg examples github
$ ghorg examples gitlab
$ ghorg examples bitbucket
$ ghorg examples gitea

Or see examples directory at https://github.com/gabrie30/ghorg/tree/master/examples
`,
	Run: cloneFunc,
}

var cachedDirSizeMB float64
var isDirSizeCached bool
var commandStartTime time.Time

func cloneFunc(cmd *cobra.Command, argz []string) {
	// Record start time for the entire command duration including SCM API calls
	commandStartTime = time.Now()

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

	if cmd.Flags().Changed("github-token-from-github-app") {
		os.Setenv("GHORG_GITHUB_TOKEN_FROM_GITHUB_APP", cmd.Flag("github-token-from-github-app").Value.String())
	}

	if cmd.Flags().Changed("github-app-pem-path") {
		os.Setenv("GHORG_GITHUB_APP_PEM_PATH", cmd.Flag("github-app-pem-path").Value.String())
	}

	if cmd.Flags().Changed("github-app-installation-id") {
		os.Setenv("GHORG_GITHUB_APP_INSTALLATION_ID", cmd.Flag("github-app-installation-id").Value.String())
	}

	if cmd.Flags().Changed("github-filter-language") {
		os.Setenv("GHORG_GITHUB_FILTER_LANGUAGE", cmd.Flag("github-filter-language").Value.String())
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

	if cmd.Flags().Changed("github-user-option") {
		opt := cmd.Flag("github-user-option").Value.String()
		os.Setenv("GHORG_GITHUB_USER_OPTION", opt)
	}

	if cmd.Flags().Changed("base-url") {
		url := cmd.Flag("base-url").Value.String()
		os.Setenv("GHORG_SCM_BASE_URL", url)
	}

	if cmd.Flags().Changed("concurrency") {
		f := cmd.Flag("concurrency").Value.String()
		os.Setenv("GHORG_CONCURRENCY", f)
	}

	if cmd.Flags().Changed("clone-delay-seconds") {
		f := cmd.Flag("clone-delay-seconds").Value.String()
		os.Setenv("GHORG_CLONE_DELAY_SECONDS", f)
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

	if cmd.Flags().Changed("ghorgonly-path") {
		path := cmd.Flag("ghorgonly-path").Value.String()
		os.Setenv("GHORG_ONLY_PATH", path)
	}

	if cmd.Flags().Changed("target-repos-path") {
		path := cmd.Flag("target-repos-path").Value.String()
		os.Setenv("GHORG_TARGET_REPOS_PATH", path)
	}

	if cmd.Flags().Changed("git-filter") {
		filter := cmd.Flag("git-filter").Value.String()
		os.Setenv("GHORG_GIT_FILTER", filter)
	}

	if cmd.Flags().Changed("preserve-scm-hostname") {
		os.Setenv("GHORG_PRESERVE_SCM_HOSTNAME", "true")
	}

	if cmd.Flags().Changed("skip-archived") {
		os.Setenv("GHORG_SKIP_ARCHIVED", "true")
	}

	if cmd.Flags().Changed("stats-enabled") {
		os.Setenv("GHORG_STATS_ENABLED", "true")
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

	if cmd.Flags().Changed("prune-untouched") {
		os.Setenv("GHORG_PRUNE_UNTOUCHED", "true")
	}

	if cmd.Flags().Changed("prune-untouched-no-confirm") {
		os.Setenv("GHORG_PRUNE_UNTOUCHED_NO_CONFIRM", "true")
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

	if cmd.Flags().Changed("clone-snippets") {
		os.Setenv("GHORG_CLONE_SNIPPETS", "true")
	}

	if cmd.Flags().Changed("insecure-gitlab-client") {
		os.Setenv("GHORG_INSECURE_GITLAB_CLIENT", "true")
	}

	if cmd.Flags().Changed("insecure-gitea-client") {
		os.Setenv("GHORG_INSECURE_GITEA_CLIENT", "true")
	}

	if cmd.Flags().Changed("insecure-bitbucket-client") {
		os.Setenv("GHORG_INSECURE_BITBUCKET_CLIENT", "true")
	}

	if cmd.Flags().Changed("insecure-sourcehut-client") {
		os.Setenv("GHORG_INSECURE_SOURCEHUT_CLIENT", "true")
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

	if cmd.Flags().Changed("no-dir-size") {
		os.Setenv("GHORG_NO_DIR_SIZE", "true")
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
		token := cmd.Flag("token").Value.String()
		if configs.IsFilePath(token) {
			token = configs.GetTokenFromFile(token)
		}
		if os.Getenv("GHORG_SCM_TYPE") == "github" {
			os.Setenv("GHORG_GITHUB_TOKEN", token)
		} else if os.Getenv("GHORG_SCM_TYPE") == "gitlab" {
			os.Setenv("GHORG_GITLAB_TOKEN", token)
		} else if os.Getenv("GHORG_SCM_TYPE") == "bitbucket" {
			if cmd.Flags().Changed("bitbucket-username") {
				os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", cmd.Flag("token").Value.String())
			} else {
				os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", cmd.Flag("token").Value.String())
			}
		} else if os.Getenv("GHORG_SCM_TYPE") == "gitea" {
			os.Setenv("GHORG_GITEA_TOKEN", token)
		} else if os.Getenv("GHORG_SCM_TYPE") == "sourcehut" {
			os.Setenv("GHORG_SOURCEHUT_TOKEN", token)
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

	if os.Getenv("GHORG_PRESERVE_SCM_HOSTNAME") == "true" {
		updateAbsolutePathToCloneToWithHostname()
	}

	setOutputDirName(argz)
	setOuputDirAbsolutePath()
	targetCloneSource = argz[0]

	// Auto-adjust concurrency for clone delay before setup (silently)
	if _, _, shouldAdjust := shouldAutoAdjustConcurrency(); shouldAdjust {
		os.Setenv("GHORG_CONCURRENCY", "1")
		os.Setenv("GHORG_CONCURRENCY_AUTO_ADJUSTED", "true")
	}

	setupRepoClone()
}

func setupRepoClone() {
	// Clear global slices and cached values at the start of each clone operation
	// to prevent memory leaks in long-running processes like reclone-server
	cloneErrors = nil
	cloneInfos = nil
	cachedDirSizeMB = 0
	isDirSizeCached = false

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

func readTargetReposFile() ([]string, error) {
	file, err := os.Open(os.Getenv("GHORG_TARGET_REPOS_PATH"))
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

func readGhorgOnly() ([]string, error) {
	file, err := os.Open(configs.GhorgOnlyLocation())
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

		// Snippets should never have collions because we append the snippet id to the directory name
		if repo.IsGitLabSnippet {
			continue
		}

		if repo.IsWiki {
			continue
		}

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

			repositories, err := getRelativePathRepositories(outputDirAbsolutePath)
			if err != nil {
				log.Fatal(err)
			}

			eligibleForPrune := 0
			for _, repository := range repositories {
				// for each item in the org's clone directory, let's make sure we found a
				// corresponding repo on the remote.
				if !sliceContainsNamedRepo(repos, repository) {
					eligibleForPrune++
					colorlog.PrintSubtleInfo(fmt.Sprintf("%s not found in remote.", repository))
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

func getCloneableInventory(allRepos []scm.Repo) (int, int, int, int) {
	var wikis, snippets, repos, total int
	for _, repo := range allRepos {
		if repo.IsGitLabSnippet {
			snippets++
		} else if repo.IsWiki {
			wikis++
		} else {
			repos++
		}
	}
	total = repos + snippets + wikis
	return total, repos, snippets, wikis
}

func isGitRepository(path string) bool {
	stat, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && stat.IsDir()
}

func getRelativePathRepositories(root string) ([]string, error) {
	var relativePaths []string
	err := filepath.WalkDir(root, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != outputDirAbsolutePath && file.IsDir() && isGitRepository(path) {
			rel, err := filepath.Rel(outputDirAbsolutePath, path)
			if err != nil {
				return err
			}
			relativePaths = append(relativePaths, rel)
		}
		return nil
	})
	return relativePaths, err
}

// CloneAllRepos clones all repos
func CloneAllRepos(git git.Gitter, cloneTargets []scm.Repo) {
	// Initialize filter and apply all filtering
	filter := NewRepositoryFilter()
	cloneTargets = filter.ApplyAllFilters(cloneTargets)

	totalResourcesToClone, reposToCloneCount, snippetToCloneCount, wikisToCloneCount := getCloneableInventory(cloneTargets)

	if os.Getenv("GHORG_CLONE_WIKI") == "true" && os.Getenv("GHORG_CLONE_SNIPPETS") == "true" {
		m := fmt.Sprintf("%v resources to clone found in %v, %v repos, %v snippets, and %v wikis\n", totalResourcesToClone, targetCloneSource, snippetToCloneCount, reposToCloneCount, wikisToCloneCount)
		colorlog.PrintInfo(m)
	} else if os.Getenv("GHORG_CLONE_WIKI") == "true" {
		m := fmt.Sprintf("%v resources to clone found in %v, %v repos and %v wikis\n", totalResourcesToClone, targetCloneSource, reposToCloneCount, wikisToCloneCount)
		colorlog.PrintInfo(m)
	} else if os.Getenv("GHORG_CLONE_SNIPPETS") == "true" {
		m := fmt.Sprintf("%v resources to clone found in %v, %v repos and %v snippets\n", totalResourcesToClone, targetCloneSource, reposToCloneCount, snippetToCloneCount)
		colorlog.PrintInfo(m)
	} else {
		colorlog.PrintInfo(strconv.Itoa(reposToCloneCount) + " repos found in " + targetCloneSource + "\n")
	}

	// Show concurrency adjustment message if it was auto-adjusted
	if os.Getenv("GHORG_CONCURRENCY_AUTO_ADJUSTED") == "true" {
		if delaySeconds, hasDelay := getCloneDelaySeconds(); hasDelay {
			colorlog.PrintInfo(fmt.Sprintf("GHORG_CLONE_DELAY_SECONDS is set to %d seconds. Automatically setting GHORG_CONCURRENCY to 1 for predictable rate limiting.", delaySeconds))
		}
		// Clear the tracking variable
		os.Unsetenv("GHORG_CONCURRENCY_AUTO_ADJUSTED")
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

	// Initialize repository processor
	processor := NewRepositoryProcessor(git)

	for i := range cloneTargets {
		repo := cloneTargets[i]

		// We use this because we dont want spaces in the final directory, using the web address makes it more file friendly
		// In the case of root level snippets we use the title which will have spaces in it, the url uses an ID so its not possible to use name from url
		// With snippets that originate on repos, we use that repo name
		var repoSlug string
		if os.Getenv("GHORG_SCM_TYPE") == "sourcehut" {
			// The URL handling in getAppNameFromURL makes strong presumptions that the URL will end in an
			// extension like '.git', but this is not the case for sourcehut (and possibly other forges).
			repoSlug = repo.Name
		} else if repo.IsGitLabSnippet && !repo.IsGitLabRootLevelSnippet {
			repoSlug = getAppNameFromURL(repo.GitLabSnippetInfo.URLOfRepo)
		} else if repo.IsGitLabRootLevelSnippet {
			repoSlug = repo.Name
		} else {
			repoSlug = getAppNameFromURL(repo.URL)
		}

		if !isPathSegmentSafe(repoSlug) {
			log.Fatal("Unsafe path segment found in SCM output")
		}

		limit.Execute(func() {
			if repo.Path != "" && os.Getenv("GHORG_PRESERVE_DIRECTORY_STRUCTURE") == "true" {
				repoSlug = repo.Path
			}

			processor.ProcessRepository(&repo, repoNameWithCollisions, hasCollisions, repoSlug, i)
		})

	}

	limit.WaitAndClose()

	// Calculate total duration from command start (including SCM API calls) and set it on the processor
	totalDuration := time.Since(commandStartTime)
	totalDurationSeconds := int(totalDuration.Seconds() + 0.5) // Round to nearest second
	processor.SetTotalDuration(totalDurationSeconds)

	// Get statistics and untouched repos from processor
	stats := processor.GetStats()
	untouchedReposToPrune := processor.GetUntouchedRepos()
	var untouchedPrunes int

	if os.Getenv("GHORG_PRUNE_UNTOUCHED") == "true" && len(untouchedReposToPrune) > 0 {
		if os.Getenv("GHORG_PRUNE_UNTOUCHED_NO_CONFIRM") != "true" {
			colorlog.PrintSuccess(fmt.Sprintf("PLEASE CONFIRM: The following %d untouched repositories will be deleted. Press enter to confirm: ", len(untouchedReposToPrune)))
			for _, repoPath := range untouchedReposToPrune {
				colorlog.PrintInfo(fmt.Sprintf("- %s", repoPath))
			}
			fmt.Scanln()
		}

		for _, repoPath := range untouchedReposToPrune {
			err := os.RemoveAll(repoPath)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Failed to prune repository at %s: %v", repoPath, err))
			} else {
				untouchedPrunes++
				colorlog.PrintSuccess(fmt.Sprintf("Successfully deleted %s", repoPath))
			}
		}
	}

	// Update global error/info arrays for backward compatibility
	cloneInfos = stats.CloneInfos
	cloneErrors = stats.CloneErrors

	printRemainingMessages()
	printCloneStatsMessage(stats.CloneCount, stats.PulledCount, stats.UpdateRemoteCount, stats.NewCommits, untouchedPrunes, stats.TotalDurationSeconds)

	if hasCollisions {
		fmt.Println("")
		colorlog.PrintInfo("ATTENTION: ghorg detected collisions in repo names from the groups that were cloned. This occurs when one or more groups share common repo names trying to be cloned to the same directory. The repos that would have collisions were renamed with the group/subgroup appended.")
		if os.Getenv("GHORG_DEBUG") != "" {
			fmt.Println("")
			colorlog.PrintInfo("Collisions Occured in the following repos...")
			for repoName, collision := range repoNameWithCollisions {
				if collision {
					colorlog.PrintInfo("- " + repoName)
				}
			}
		}
	}

	var pruneCount int
	cloneInfosCount := len(stats.CloneInfos)
	cloneErrorsCount := len(stats.CloneErrors)
	allReposToCloneCount := len(cloneTargets)
	// Now, clean up local repos that don't exist in remote, if prune flag is set
	if os.Getenv("GHORG_PRUNE") == "true" {
		pruneCount = pruneRepos(cloneTargets)
	}

	if os.Getenv("GHORG_QUIET") != "true" {
		if os.Getenv("GHORG_NO_DIR_SIZE") == "false" {
			printFinishedWithDirSize()
		} else {
			colorlog.PrintSuccess(fmt.Sprintf("\nFinished! %s", outputDirAbsolutePath))
		}
	}

	// This needs to be called after printFinishedWithDirSize()
	if os.Getenv("GHORG_STATS_ENABLED") == "true" {
		date := time.Now().Format("2006-01-02 15:04:05")
		writeGhorgStats(date, allReposToCloneCount, stats.CloneCount, stats.PulledCount, cloneInfosCount, cloneErrorsCount, stats.UpdateRemoteCount, stats.NewCommits, pruneCount, stats.TotalDurationSeconds, hasCollisions)
	}

	if os.Getenv("GHORG_DONT_EXIT_UNDER_TEST") != "true" {
		if os.Getenv("GHORG_EXIT_CODE_ON_CLONE_INFOS") != "0" && cloneInfosCount > 0 {
			exitCode, err := strconv.Atoi(os.Getenv("GHORG_EXIT_CODE_ON_CLONE_INFOS"))
			if err != nil {
				colorlog.PrintError("Could not convert GHORG_EXIT_CODE_ON_CLONE_INFOS from string to integer")
				os.Exit(1)
			}

			os.Exit(exitCode)
		}
	}

	if os.Getenv("GHORG_DONT_EXIT_UNDER_TEST") != "true" {
		if cloneErrorsCount > 0 {
			exitCode, err := strconv.Atoi(os.Getenv("GHORG_EXIT_CODE_ON_CLONE_ISSUES"))
			if err != nil {
				colorlog.PrintError("Could not convert GHORG_EXIT_CODE_ON_CLONE_ISSUES from string to integer")
				os.Exit(1)
			}
			os.Exit(exitCode)
		}
	} else {
		cloneErrorsCount = 0
	}

}

func getGhorgStatsFilePath() string {
	var statsFilePath string
	absolutePath := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	if os.Getenv("GHORG_PRESERVE_SCM_HOSTNAME") == "true" {
		originalAbsolutePath := os.Getenv("GHORG_ORIGINAL_ABSOLUTE_PATH_TO_CLONE_TO")
		statsFilePath = filepath.Join(originalAbsolutePath, "_ghorg_stats.csv")
	} else {
		statsFilePath = filepath.Join(absolutePath, "_ghorg_stats.csv")
	}

	return statsFilePath
}

func writeGhorgStats(date string, allReposToCloneCount, cloneCount, pulledCount, cloneInfosCount, cloneErrorsCount, updateRemoteCount, newCommits, pruneCount, totalDurationSeconds int, hasCollisions bool) error {

	statsFilePath := getGhorgStatsFilePath()
	fileExists := true

	if _, err := os.Stat(statsFilePath); os.IsNotExist(err) {
		fileExists = false
	}

	header := "datetime,clonePath,scm,cloneType,cloneTarget,totalCount,newClonesCount,existingResourcesPulledCount,dirSizeInMB,newCommits,cloneInfosCount,cloneErrorsCount,updateRemoteCount,pruneCount,hasCollisions,ghorgignore,ghorgonly,totalDurationSeconds,ghorgVersion\n"

	var file *os.File
	var err error

	if fileExists {
		// Read the existing header
		existingHeader, err := readFirstLine(statsFilePath)
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Error reading header from stats file: %v", err))
			return err
		}

		// Check if the existing header is different from the new header, need to add a newline
		if existingHeader+"\n" != header {
			hashedHeader := fmt.Sprintf("%x", sha256.Sum256([]byte(header)))
			newHeaderFilePath := filepath.Join(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), fmt.Sprintf("ghorg_stats_new_header_%s.csv", hashedHeader))
			// Create a new file with the new header
			file, err = os.OpenFile(newHeaderFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error creating new header stats file: %v", err))
				return err
			}
			if _, err := file.WriteString(header); err != nil {
				colorlog.PrintError(fmt.Sprintf("Error writing new header to GHORG_STATS file: %v", err))
				return err
			}
		} else {
			// Open the existing file in append mode
			file, err = os.OpenFile(statsFilePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error opening stats file for appending: %v", err))
				return err
			}
		}
	} else {
		// Create the file and write the header
		file, err = os.OpenFile(statsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Error creating stats file: %v", err))
			return err
		}
		if _, err := file.WriteString(header); err != nil {
			colorlog.PrintError(fmt.Sprintf("Error writing header to GHORG_STATS file: %v", err))
			return err
		}
	}
	defer file.Close()

	data := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%.2f,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
		date,
		outputDirAbsolutePath,
		os.Getenv("GHORG_SCM_TYPE"),
		os.Getenv("GHORG_CLONE_TYPE"),
		targetCloneSource,
		allReposToCloneCount,
		cloneCount,
		pulledCount,
		cachedDirSizeMB,
		newCommits,
		cloneInfosCount,
		cloneErrorsCount,
		updateRemoteCount,
		pruneCount,
		hasCollisions,
		configs.GhorgIgnoreDetected(),
		configs.GhorgOnlyDetected(),
		totalDurationSeconds,
		GetVersion())
	if _, err := file.WriteString(data); err != nil {
		colorlog.PrintError(fmt.Sprintf("Error writing data to GHORG_STATS file: %v", err))
		return err
	}

	return nil
}

func readFirstLine(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

func printFinishedWithDirSize() {
	dirSizeMB, err := getCachedOrCalculatedOutputDirSizeInMb()
	if err != nil {
		if os.Getenv("GHORG_DEBUG") == "true" {
			colorlog.PrintError(fmt.Sprintf("Error calculating directory size: %v", err))
		}
		colorlog.PrintSuccess(fmt.Sprintf("\nFinished! %s", outputDirAbsolutePath))
		return
	}

	if dirSizeMB > 1000 {
		dirSizeGB := dirSizeMB / 1000
		colorlog.PrintSuccess(fmt.Sprintf("\nFinished! %s (Size: %.2f GB)", outputDirAbsolutePath, dirSizeGB))
	} else {
		colorlog.PrintSuccess(fmt.Sprintf("\nFinished! %s (Size: %.2f MB)", outputDirAbsolutePath, dirSizeMB))
	}
}

func getCachedOrCalculatedOutputDirSizeInMb() (float64, error) {
	if !isDirSizeCached {
		dirSizeMB, err := utils.CalculateDirSizeInMb(outputDirAbsolutePath)
		if err != nil {
			return 0, err
		}
		cachedDirSizeMB = dirSizeMB
		isDirSizeCached = true
	}
	return cachedDirSizeMB, nil
}

func filterByTargetReposPath(cloneTargets []scm.Repo) []scm.Repo {

	_, err := os.Stat(os.Getenv("GHORG_TARGET_REPOS_PATH"))

	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("Error finding your GHORG_TARGET_REPOS_PATH file, error: %v", err))
	}

	if !os.IsNotExist(err) {
		// Open the file parse each line and remove cloneTargets containing
		toTarget, err := readTargetReposFile()
		if err != nil {
			colorlog.PrintErrorAndExit(fmt.Sprintf("Error parsing your GHORG_TARGET_REPOS_PATH file, error: %v", err))
		}

		colorlog.PrintInfo("Using GHORG_TARGET_REPOS_PATH, filtering repos down...")

		filteredCloneTargets := []scm.Repo{}
		var flag bool

		targetRepoSeenOnOrg := make(map[string]bool)

		for _, cloneTarget := range cloneTargets {
			flag = false
			for _, targetRepo := range toTarget {
				if _, ok := targetRepoSeenOnOrg[targetRepo]; !ok {
					targetRepoSeenOnOrg[targetRepo] = false
				}
				clonedRepoName := strings.TrimSuffix(filepath.Base(cloneTarget.URL), ".git")
				if strings.EqualFold(clonedRepoName, targetRepo) {
					flag = true
					targetRepoSeenOnOrg[targetRepo] = true
				}

				if os.Getenv("GHORG_CLONE_WIKI") == "true" {
					targetRepoWiki := targetRepo + ".wiki"
					if strings.EqualFold(targetRepoWiki, clonedRepoName) {
						flag = true
						targetRepoSeenOnOrg[targetRepo] = true
					}
				}

				if os.Getenv("GHORG_CLONE_SNIPPETS") == "true" {
					if cloneTarget.IsGitLabSnippet {
						targetSnippetOriginalRepo := strings.TrimSuffix(filepath.Base(cloneTarget.GitLabSnippetInfo.URLOfRepo), ".git")
						if strings.EqualFold(targetSnippetOriginalRepo, targetRepo) {
							flag = true
							targetRepoSeenOnOrg[targetRepo] = true
						}
					}
				}
			}

			if flag {
				filteredCloneTargets = append(filteredCloneTargets, cloneTarget)
			}
		}

		// Print all the repos in the file that were not in the org so users know the entry is not being cloned
		for targetRepo, seen := range targetRepoSeenOnOrg {
			if !seen {
				cloneInfos = append(cloneInfos, fmt.Sprintf("Target in GHORG_TARGET_REPOS_PATH was not found in the org, repo: %v", targetRepo))
			}
		}

		cloneTargets = filteredCloneTargets

	}

	return cloneTargets
}

func pruneRepos(cloneTargets []scm.Repo) int {
	count := 0
	colorlog.PrintInfo("\nScanning for local clones that have been removed on remote...")

	repositories, err := getRelativePathRepositories(outputDirAbsolutePath)
	if err != nil {
		log.Fatal(err)
	}

	// The first time around, we set userAgreesToDelete to true, otherwise we'd immediately
	// break out of the loop.
	userAgreesToDelete := true
	pruneNoConfirm := os.Getenv("GHORG_PRUNE_NO_CONFIRM") == "true"
	for _, repository := range repositories {
		absolutePathToDelete := filepath.Join(outputDirAbsolutePath, repository)

		// Safeguard: Ensure the path is within the expected base directory
		if !strings.HasPrefix(absolutePathToDelete, outputDirAbsolutePath) {
			colorlog.PrintErrorAndExit(fmt.Sprintf("DANGEROUS ACTION DETECTED! Preventing deletion of %s as it is outside the base directory this deletion is not expected, exiting.", absolutePathToDelete))
		}

		// For each item in the org's clone directory, let's make sure we found a corresponding
		// repo on the remote.  We check userAgreesToDelete here too, so that if the user says
		// "No" at any time, we stop trying to prune things altogether.
		if userAgreesToDelete && !sliceContainsNamedRepo(cloneTargets, repository) {
			// If the user specified --prune-no-confirm, we needn't prompt interactively.
			userAgreesToDelete = pruneNoConfirm || interactiveYesNoPrompt(
				fmt.Sprintf("%s was not found in remote.  Do you want to prune it? %s", repository, absolutePathToDelete))
			if userAgreesToDelete {
				colorlog.PrintSubtleInfo(
					fmt.Sprintf("Deleting %s", absolutePathToDelete))
				err = os.RemoveAll(absolutePathToDelete)
				count++
				if err != nil {
					log.Fatal(err)
				}
			} else {
				colorlog.PrintError("Pruning cancelled by user.  No more prunes will be considered.")
			}
		}
	}

	return count
}

// formatDurationText formats duration in seconds to a human-readable string
func formatDurationText(durationSeconds int) string {
	if durationSeconds >= 60 {
		minutes := durationSeconds / 60
		seconds := durationSeconds % 60
		if seconds > 0 {
			return fmt.Sprintf(" (completed in %dm%ds)", minutes, seconds)
		} else {
			return fmt.Sprintf(" (completed in %dm)", minutes)
		}
	} else {
		return fmt.Sprintf(" (completed in %ds)", durationSeconds)
	}
}

func printCloneStatsMessage(cloneCount, pulledCount, updateRemoteCount, newCommits, untouchedPrunes, durationSeconds int) {
	durationText := formatDurationText(durationSeconds)

	if updateRemoteCount > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("New clones: %v, existing resources pulled: %v, total new commits: %v, remotes updated: %v%s", cloneCount, pulledCount, newCommits, updateRemoteCount, durationText))
		return
	}

	if newCommits > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("New clones: %v, existing resources pulled: %v, total new commits: %v%s", cloneCount, pulledCount, newCommits, durationText))
		return
	}

	if untouchedPrunes > 0 {
		colorlog.PrintSuccess(fmt.Sprintf("New clones: %v, existing resources pulled: %v, total prunes: %v%s", cloneCount, pulledCount, untouchedPrunes, durationText))
		return
	}

	colorlog.PrintSuccess(fmt.Sprintf("New clones: %v, existing resources pulled: %v%s", cloneCount, pulledCount, durationText))
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
func sliceContainsNamedRepo(haystack []scm.Repo, needle string) bool {

	// GitLab Cloud vs GitLab on Prem seem to have different needle/repo.Paths when it comes to this,
	// so normalize to handle both
	// I'm not really sure whats going on here though, could be a bug with how this is set
	needle = strings.TrimPrefix(needle, "/")

	// Normalize path separators for cross-platform compatibility (Windows vs Unix)
	// Convert both needle and repo paths to use forward slashes for comparison
	// We need to handle both forward and back slashes regardless of OS
	needle = strings.ReplaceAll(needle, "\\", "/")
	needle = filepath.ToSlash(needle)

	for _, repo := range haystack {
		normalizedPath := strings.TrimPrefix(repo.Path, "/")
		// Convert repo path to forward slashes for comparison
		// We need to handle both forward and back slashes regardless of OS
		normalizedPath = strings.ReplaceAll(normalizedPath, "\\", "/")
		normalizedPath = filepath.ToSlash(normalizedPath)

		if normalizedPath == needle {
			if os.Getenv("GHORG_DEBUG") != "" {
				fmt.Printf("Debug: Match found for repo path: %s\n", repo.Path)
			}
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
	if delaySeconds, hasDelay := getCloneDelaySeconds(); hasDelay {
		colorlog.PrintInfo("* Clone Delay   : " + strconv.Itoa(delaySeconds) + " seconds")
	}

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
	if os.Getenv("GHORG_CLONE_SNIPPETS") == "true" {
		colorlog.PrintInfo("* Snippets      : " + os.Getenv("GHORG_CLONE_SNIPPETS"))
	}
	if configs.GhorgIgnoreDetected() {
		colorlog.PrintInfo("* Ghorgignore   : " + configs.GhorgIgnoreLocation())
	}
	if configs.GhorgOnlyDetected() {
		colorlog.PrintInfo("* Ghorgonly     : " + configs.GhorgOnlyLocation())
	}
	if os.Getenv("GHORG_TARGET_REPOS_PATH") != "" {
		colorlog.PrintInfo("* Target Repos  : " + os.Getenv("GHORG_TARGET_REPOS_PATH"))
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
	if os.Getenv("GHORG_STATS_ENABLED") == "true" {
		colorlog.PrintInfo("* Stats Enabled : " + os.Getenv("GHORG_STATS_ENABLED"))
	}
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

	// Strip ~ prefix for sourcehut usernames to avoid shell expansion issues
	if os.Getenv("GHORG_SCM_TYPE") == "sourcehut" {
		outputDirName = strings.TrimPrefix(outputDirName, "~")
	}

	if os.Getenv("GHORG_PRESERVE_SCM_HOSTNAME") != "true" {
		// If all-group is used set the parent folder to the name of the baseurl
		if argz[0] == "all-groups" && os.Getenv("GHORG_SCM_BASE_URL") != "" {
			u, err := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error parsing GHORG_SCM_BASE_URL, clone may be affected, error: %v", err))
			}
			outputDirName = u.Hostname()
		}

		if argz[0] == "all-users" && os.Getenv("GHORG_SCM_BASE_URL") != "" {
			u, err := url.Parse(os.Getenv("GHORG_SCM_BASE_URL"))
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error parsing GHORG_SCM_BASE_URL, clone may be affected, error: %v", err))
			}
			outputDirName = u.Hostname()
		}
	}

	if os.Getenv("GHORG_BACKUP") == "true" {
		outputDirName = outputDirName + "_backup"
	}
}

// filter repos down based on ghorgignore if one exists
func filterByGhorgignore(cloneTargets []scm.Repo) []scm.Repo {

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

	return cloneTargets
}

func isPathSegmentSafe(seg string) bool {
	return strings.IndexByte(seg, '/') < 0 && strings.IndexRune(seg, filepath.Separator) < 0
}
