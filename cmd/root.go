package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
)

var (
	protocol                     string
	path                         string
	outputDirName                string
	outputDirAbsolutePath        string
	branch                       string
	token                        string
	cloneType                    string
	scmType                      string
	bitbucketUsername            string
	bitbucketAPIEmail            string
	color                        string
	baseURL                      string
	concurrency                  string
	cloneDepth                   string
	exitCodeOnCloneInfos         string
	exitCodeOnCloneIssues        string
	outputDir                    string
	topics                       string
	gitFilter                    string
	targetCloneSource            string
	matchPrefix                  string
	excludeMatchPrefix           string
	matchRegex                   string
	excludeMatchRegex            string
	config                       string
	gitlabGroupExcludeMatchRegex string
	gitlabGroupMatchRegex        string
	ghorgIgnorePath              string
	ghorgOnlyPath                string
	targetReposPath              string
	ghorgReClonePath             string
	githubAppID                  string
	githubAppPemPath             string
	githubAppInstallationID      string
	githubUserOption             string
	githubFilterLanguage         string
	cronTimerMinutes             string
	recloneServerPort            string
	cloneDelaySeconds            string
	sshHostname                  string
	includeSubmodules            bool
	skipArchived                 bool
	skipForks                    bool
	backup                       bool
	noClean                      bool
	dryRun                       bool
	prune                        bool
	pruneNoConfirm               bool
	cloneWiki                    bool
	cloneSnippets                bool
	preserveDir                  bool
	insecureGitlabClient         bool
	insecureGiteaClient          bool
	insecureBitbucketClient      bool
	insecureSourcehutClient      bool
	fetchAll                     bool
	ghorgReCloneQuiet            bool
	ghorgReCloneList             bool
	ghorgReCloneEnvConfigOnly    bool
	githubTokenFromGithubApp     bool
	noToken                      bool
	quietMode                    bool
	noDirSize                    bool
	ghorgStatsEnabled            bool
	ghorgPreserveScmHostname     bool
	ghorgPruneUntouched          bool
	ghorgPruneUntouchedNoConfirm bool
	cloneErrors                  []string
	cloneInfos                   []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghorg",
	Short: "Ghorg is a fast way to clone multiple repos into a single directory",
	Long:  `Ghorg is a fast way to clone multiple repos into a single directory`,
	Run: func(cmd *cobra.Command, argz []string) {
		fmt.Println("For help run: ghorg clone --help")
	},
}

func getHostname() string {
	var hostname string
	baseURL := os.Getenv("GHORG_SCM_BASE_URL")
	if baseURL != "" {
		// Parse the URL to extract the hostname
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Error parsing GHORG_SCM_BASE_URL clone may be affected, error: %v", err))
		}
		// Append the hostname to the absolute path
		hostname = parsedURL.Hostname()
	} else {
		// Use the predefined hostname based on the SCM type
		hostname = configs.GetCloudScmTypeHostnames()
	}

	return hostname
}

// updateAbsolutePathToCloneToWithHostname modifies the absolute path by appending the hostname if the user has enabled it,
// supporting the GHORG_PRESERVE_SCM_HOSTNAME feature. It checks the GHORG_PRESERVE_SCM_HOSTNAME environment variable, and if set to "true",
// it uses the hostname from GHORG_SCM_BASE_URL if available, otherwise, it defaults to a predefined hostname based on the SCM type.
func updateAbsolutePathToCloneToWithHostname() {
	// Verify if GHORG_PRESERVE_SCM_HOSTNAME is set to "true"
	if os.Getenv("GHORG_PRESERVE_SCM_HOSTNAME") == "true" {
		// Retrieve the hostname from the environment variable
		hostname := getHostname()
		absolutePath := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
		os.Setenv("GHORG_ORIGINAL_ABSOLUTE_PATH_TO_CLONE_TO", absolutePath)
		absolutePath = filepath.Join(absolutePath, hostname)
		os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", configs.EnsureTrailingSlashOnFilePath(absolutePath))
	}
}

// reads in configuration file and updates anything not set to default
func getOrSetDefaults(envVar string) {
	if envVar == "GHORG_COLOR" {
		if color == "enabled" {
			os.Setenv("GHORG_COLOR", "enabled")
			return
		}

		if color == "disabled" {
			os.Setenv("GHORG_COLOR", "disabled")
			return
		}

		if viper.GetString(envVar) == "enabled" {
			os.Setenv("GHORG_COLOR", "enabled")
			return
		}
	}

	// When a user does not set value in $HOME/.config/ghorg/conf.yaml set the default values, else set env to what they have added to the file.
	if viper.GetString(envVar) == "" {
		switch envVar {
		case "GHORG_ABSOLUTE_PATH_TO_CLONE_TO":
			os.Setenv(envVar, configs.GetAbsolutePathToCloneTo())
		case "GHORG_IGNORE_PATH":
			os.Setenv(envVar, configs.GhorgIgnoreLocation())
		case "GHORG_RECLONE_PATH":
			os.Setenv(envVar, configs.GhorgReCloneLocation())
		case "GHORG_CLONE_PROTOCOL":
			os.Setenv(envVar, "https")
		case "GHORG_CLONE_TYPE":
			os.Setenv(envVar, "org")
		case "GHORG_SCM_TYPE":
			os.Setenv(envVar, "github")
		case "GHORG_SKIP_ARCHIVED":
			os.Setenv(envVar, "false")
		case "GHORG_INCLUDE_SUBMODULES":
			os.Setenv(envVar, "false")
		case "GHORG_SKIP_FORKS":
			os.Setenv(envVar, "false")
		case "GHORG_CLONE_WIKI":
			os.Setenv(envVar, "false")
		case "GHORG_CLONE_SNIPPETS":
			os.Setenv(envVar, "false")
		case "GHORG_NO_CLEAN":
			os.Setenv(envVar, "false")
		case "GHORG_CRON_TIMER_MINUTES":
			os.Setenv(envVar, "60")
		case "GHORG_RECLONE_SERVER_PORT":
			os.Setenv(envVar, ":8080")
		case "GHORG_FETCH_ALL":
			os.Setenv(envVar, "false")
		case "GHORG_DRY_RUN":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE_NO_CONFIRM":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE_UNTOUCHED":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE_UNTOUCHED_NO_CONFIRM":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_GITLAB_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_GITEA_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_BITBUCKET_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_SOURCEHUT_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_GITHUB_USER_OPTION":
			os.Setenv(envVar, "owner")
		case "GHORG_BACKUP":
			os.Setenv(envVar, "false")
		case "GHORG_PRESERVE_SCM_HOSTNAME":
			os.Setenv(envVar, "false")
		case "GHORG_NO_TOKEN":
			os.Setenv(envVar, "false")
		case "GHORG_NO_DIR_SIZE":
			os.Setenv(envVar, "false")
		case "GHORG_RECLONE_ENV_CONFIG_ONLY":
			os.Setenv(envVar, "false")
		case "GHORG_RECLONE_QUIET":
			os.Setenv(envVar, "false")
		case "GHORG_COLOR":
			os.Setenv(envVar, "disabled")
		case "GHORG_PRESERVE_DIRECTORY_STRUCTURE":
			os.Setenv(envVar, "false")
		case "GHORG_CONCURRENCY":
			os.Setenv(envVar, "25")
		case "GHORG_CLONE_DELAY_SECONDS":
			os.Setenv(envVar, "0")
		case "GHORG_QUIET":
			os.Setenv(envVar, "false")
		case "GHORG_STATS_ENABLED":
			os.Setenv(envVar, "false")
		case "GHORG_EXIT_CODE_ON_CLONE_INFOS":
			os.Setenv(envVar, "0")
		case "GHORG_EXIT_CODE_ON_CLONE_ISSUES":
			os.Setenv(envVar, "1")
		case "GHORG_GITHUB_TOKEN_FROM_GITHUB_APP":
			os.Setenv(envVar, "false")
		case "GHORG_SSH_HOSTNAME":
			os.Setenv(envVar, "")
		}
	} else {
		s := viper.GetString(envVar)
		// envs that need a trailing slash
		if envVar == "GHORG_SCM_BASE_URL" {
			os.Setenv(envVar, configs.EnsureTrailingSlashOnURL(s))
		} else if envVar == "GHORG_ABSOLUTE_PATH_TO_CLONE_TO" {
			os.Setenv(envVar, configs.EnsureTrailingSlashOnFilePath(s))
		} else {
			os.Setenv(envVar, s)
		}
	}

	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Printf("%s: %s\n", envVar, os.Getenv(envVar))
	}
}

func InitConfig() {
	curDir, _ := os.Getwd()
	localConfig := filepath.Join(curDir, "ghorg.yaml")

	if config != "" {
		viper.SetConfigFile(config)
		os.Setenv("GHORG_CONFIG", config)
	} else if os.Getenv("GHORG_CONFIG") != "" {
		// TODO maybe check if config is valid etc ...
		viper.SetConfigFile(os.Getenv("GHORG_CONFIG"))
	} else if _, err := os.Stat(localConfig); !errors.Is(err, os.ErrNotExist) {
		viper.SetConfigFile(localConfig)
		os.Setenv("GHORG_CONFIG", localConfig)
	} else {
		config = configs.DefaultConfFile()
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configs.GhorgConfDir())
		viper.SetConfigName("conf")
		os.Setenv("GHORG_CONFIG", configs.DefaultConfFile())
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.Setenv("GHORG_CONFIG", "none")
		} else {
			colorlog.PrintError(fmt.Sprintf("Something unexpected happened reading configuration file: %s, err: %s", os.Getenv("GHORG_CONFIG"), err))
			os.Exit(1)
		}
	}

	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Println("-------- Setting Default ENV values ---------")
		if os.Getenv("GHORG_CONCURRENCY_DEBUG") == "" {
			fmt.Println("Setting concurrency to 1, this can be overwritten by setting GHORG_CONCURRENCY_DEBUG; however when using concurrency with GHORG_DEBUG, not all debugging output will be printed in serial order.")
			os.Setenv("GHORG_CONCURRENCY", "1")
		}
	}

	getOrSetDefaults("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	getOrSetDefaults("GHORG_BRANCH")
	getOrSetDefaults("GHORG_CLONE_PROTOCOL")
	getOrSetDefaults("GHORG_CLONE_TYPE")
	getOrSetDefaults("GHORG_SCM_TYPE")
	getOrSetDefaults("GHORG_PRESERVE_SCM_HOSTNAME")
	getOrSetDefaults("GHORG_SKIP_ARCHIVED")
	getOrSetDefaults("GHORG_SKIP_FORKS")
	getOrSetDefaults("GHORG_NO_CLEAN")
	getOrSetDefaults("GHORG_NO_TOKEN")
	getOrSetDefaults("GHORG_NO_DIR_SIZE")
	getOrSetDefaults("GHORG_FETCH_ALL")
	getOrSetDefaults("GHORG_PRUNE")
	getOrSetDefaults("GHORG_PRUNE_NO_CONFIRM")
	getOrSetDefaults("GHORG_PRUNE_UNTOUCHED")
	getOrSetDefaults("GHORG_PRUNE_UNTOUCHED_NO_CONFIRM")
	getOrSetDefaults("GHORG_DRY_RUN")
	getOrSetDefaults("GHORG_GITHUB_USER_OPTION")
	getOrSetDefaults("GHORG_CLONE_WIKI")
	getOrSetDefaults("GHORG_CLONE_SNIPPETS")
	getOrSetDefaults("GHORG_INSECURE_GITLAB_CLIENT")
	getOrSetDefaults("GHORG_INSECURE_GITEA_CLIENT")
	getOrSetDefaults("GHORG_INSECURE_BITBUCKET_CLIENT")
	getOrSetDefaults("GHORG_INSECURE_SOURCEHUT_CLIENT")
	getOrSetDefaults("GHORG_BACKUP")
	getOrSetDefaults("GHORG_RECLONE_ENV_CONFIG_ONLY")
	getOrSetDefaults("GHORG_RECLONE_QUIET")
	getOrSetDefaults("GHORG_CONCURRENCY")
	getOrSetDefaults("GHORG_CLONE_DELAY_SECONDS")
	getOrSetDefaults("GHORG_INCLUDE_SUBMODULES")
	getOrSetDefaults("GHORG_EXIT_CODE_ON_CLONE_INFOS")
	getOrSetDefaults("GHORG_EXIT_CODE_ON_CLONE_ISSUES")
	getOrSetDefaults("GHORG_STATS_ENABLED")
	getOrSetDefaults("GHORG_CRON_TIMER_MINUTES")
	getOrSetDefaults("GHORG_RECLONE_SERVER_PORT")
	// Optionally set
	getOrSetDefaults("GHORG_TARGET_REPOS_PATH")
	getOrSetDefaults("GHORG_CLONE_DEPTH")
	getOrSetDefaults("GHORG_GITHUB_TOKEN")
	getOrSetDefaults("GHORG_GITHUB_TOKEN_FROM_GITHUB_APP")
	getOrSetDefaults("GHORG_GITHUB_FILTER_LANGUAGE")
	getOrSetDefaults("GHORG_COLOR")
	getOrSetDefaults("GHORG_TOPICS")
	getOrSetDefaults("GHORG_GITLAB_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_USERNAME")
	getOrSetDefaults("GHORG_BITBUCKET_APP_PASSWORD")
	getOrSetDefaults("GHORG_BITBUCKET_OAUTH_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_API_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_API_EMAIL")
	getOrSetDefaults("GHORG_SCM_BASE_URL")
	getOrSetDefaults("GHORG_PRESERVE_DIRECTORY_STRUCTURE")
	getOrSetDefaults("GHORG_OUTPUT_DIR")
	getOrSetDefaults("GHORG_MATCH_REGEX")
	getOrSetDefaults("GHORG_EXCLUDE_MATCH_REGEX")
	getOrSetDefaults("GHORG_MATCH_PREFIX")
	getOrSetDefaults("GHORG_EXCLUDE_MATCH_PREFIX")
	getOrSetDefaults("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX")
	getOrSetDefaults("GHORG_GITLAB_GROUP_MATCH_REGEX")
	getOrSetDefaults("GHORG_IGNORE_PATH")
	getOrSetDefaults("GHORG_RECLONE_PATH")
	getOrSetDefaults("GHORG_QUIET")
	getOrSetDefaults("GHORG_GIT_FILTER")
	getOrSetDefaults("GHORG_GITEA_TOKEN")
	getOrSetDefaults("GHORG_SOURCEHUT_TOKEN")
	getOrSetDefaults("GHORG_INSECURE_GITEA_CLIENT")
	getOrSetDefaults("GHORG_SSH_HOSTNAME")
	getOrSetDefaults("GHORG_GITHUB_APP_PEM_PATH")
	getOrSetDefaults("GHORG_GITHUB_APP_INSTALLATION_ID")
	getOrSetDefaults("GHORG_GITHUB_APP_ID")

	if os.Getenv("GHORG_DEBUG") != "" {
		viper.Debug()
		fmt.Println("Viper config file used:", viper.ConfigFileUsed())
		fmt.Printf("GHORG_CONFIG SET TO: %s\n", os.Getenv("GHORG_CONFIG"))
	}
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.PersistentFlags().StringVar(&color, "color", "", "GHORG_COLOR - Enable or disable colorful terminal output: 'enabled' or 'disabled'. Color improves readability of logs. (default: disabled)")
	rootCmd.PersistentFlags().StringVar(&config, "config", "", "GHORG_CONFIG - Path to a custom configuration file. Allows using multiple configs for different SCM providers or organizations")

	viper.SetDefault("config", configs.DefaultConfFile())
	viper.AutomaticEnv()

	_ = viper.BindPFlag("color", rootCmd.PersistentFlags().Lookup("color"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	cloneCmd.Flags().StringVar(&targetReposPath, "target-repos-path", "", "GHORG_TARGET_REPOS_PATH - Path to a file containing a list of specific repository names to clone (one per line). Useful for cloning a subset of repos from an org/user")
	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - Protocol to use for cloning: 'ssh' or 'https'. SSH requires proper SSH keys configured. (default: https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - Absolute path where all repos will be cloned. Directory will be created if it doesn't exist. Must start with / (default: $HOME/ghorg)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - Git branch to checkout after cloning each repository. Useful for cloning specific branches across all repos. (default: master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_GITEA_TOKEN/GHORG_BITBUCKET_APP_PASSWORD/GHORG_BITBUCKET_API_TOKEN/GHORG_BITBUCKET_OAUTH_TOKEN/GHORG_SOURCEHUT_TOKEN - Personal access token or API token for authentication with your SCM provider. Required for private repos")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - Bitbucket only: Username for legacy app password authentication. Required when using app passwords")
	cloneCmd.Flags().StringVarP(&bitbucketAPIEmail, "bitbucket-api-email", "", "", "GHORG_BITBUCKET_API_EMAIL - Bitbucket only: Email address for modern API token authentication. Use this instead of username for API tokens")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - Source code management platform to clone from: github, gitlab, gitea, bitbucket, or sourcehut (default: github)")
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - Target type to clone: 'org' for organization/group or 'user' for individual user repositories (default: org)")
	cloneCmd.Flags().BoolVar(&skipArchived, "skip-archived", false, "GHORG_SKIP_ARCHIVED - Skip archived/read-only repositories during cloning. Supported on GitHub, GitLab, and Gitea")
	cloneCmd.Flags().BoolVar(&noClean, "no-clean", false, "GHORG_NO_CLEAN - Only clone new repositories without running 'git clean' on existing ones. Use this to preserve local changes in already-cloned repos")
	cloneCmd.Flags().BoolVar(&prune, "prune", false, "GHORG_PRUNE - Remove local repositories that no longer exist remotely. When used with --skip-archived, also removes archived repos locally. Prompts before deletion unless combined with --prune-no-confirm")
	cloneCmd.Flags().BoolVar(&pruneNoConfirm, "prune-no-confirm", false, "GHORG_PRUNE_NO_CONFIRM - Skip confirmation prompts when pruning repositories. Use with caution as this will delete directories without asking")
	cloneCmd.Flags().BoolVar(&fetchAll, "fetch-all", false, "GHORG_FETCH_ALL - Fetch all remote branches for each repository using 'git fetch --all'. Useful for getting complete branch information")
	cloneCmd.Flags().BoolVar(&dryRun, "dry-run", false, "GHORG_DRY_RUN - Simulate the clone operation without actually cloning repositories. Shows what would be cloned for testing/verification")
	cloneCmd.Flags().BoolVar(&insecureGitlabClient, "insecure-gitlab-client", false, "GHORG_INSECURE_GITLAB_CLIENT - Skip TLS certificate verification for self-hosted GitLab instances. Use only for internal/trusted instances")
	cloneCmd.Flags().BoolVar(&insecureGiteaClient, "insecure-gitea-client", false, "GHORG_INSECURE_GITEA_CLIENT - Allow connections to Gitea instances using HTTP instead of HTTPS. Required for non-SSL Gitea servers")
	cloneCmd.Flags().BoolVar(&insecureBitbucketClient, "insecure-bitbucket-client", false, "GHORG_INSECURE_BITBUCKET_CLIENT - Allow connections to Bitbucket Server instances using HTTP. Required for non-SSL Bitbucket servers")
	cloneCmd.Flags().BoolVar(&insecureSourcehutClient, "insecure-sourcehut-client", false, "GHORG_INSECURE_SOURCEHUT_CLIENT - Allow connections to Sourcehut instances using HTTP. Required for non-SSL Sourcehut servers")
	cloneCmd.Flags().BoolVar(&cloneWiki, "clone-wiki", false, "GHORG_CLONE_WIKI - Additionally clone wiki pages associated with each repository if they exist")
	cloneCmd.Flags().BoolVar(&cloneSnippets, "clone-snippets", false, "GHORG_CLONE_SNIPPETS - Additionally clone all code snippets/gists. GitLab only")
	cloneCmd.Flags().BoolVar(&skipForks, "skip-forks", false, "GHORG_SKIP_FORKS - Skip repositories that are forks of other repositories. Supported on GitHub, GitLab, and Gitea")
	cloneCmd.Flags().BoolVar(&noToken, "no-token", false, "GHORG_NO_TOKEN - Run without authentication token. Only works if your SCM server allows unauthenticated API access (typically for public repos only)")
	cloneCmd.Flags().BoolVar(&noDirSize, "no-dir-size", false, "GHORG_NO_DIR_SIZE - Skip calculating total directory size at the end of cloning. Improves performance when cloning many/large repositories")
	cloneCmd.Flags().BoolVar(&preserveDir, "preserve-dir", false, "GHORG_PRESERVE_DIRECTORY_STRUCTURE - Preserve GitLab's group/subgroup hierarchy in local directory structure (e.g., company/unit/subunit/app). GitLab only")
	cloneCmd.Flags().BoolVar(&backup, "backup", false, "GHORG_BACKUP - Create bare mirror clones suitable for backups. No working directory, includes all refs. Ignores --branch flag")
	cloneCmd.Flags().BoolVar(&quietMode, "quiet", false, "GHORG_QUIET - Reduce output to only critical messages. Useful for scripting or when you don't want verbose logging")
	cloneCmd.Flags().BoolVar(&includeSubmodules, "include-submodules", false, "GHORG_INCLUDE_SUBMODULES - Initialize and update git submodules for all repositories. Applies to both clone and pull operations")
	cloneCmd.Flags().BoolVar(&ghorgStatsEnabled, "stats-enabled", false, "GHORG_STATS_ENABLED - Generate a CSV file (_ghorg_stats.csv) with statistics about each clone (commits, size, etc). Useful for tracking repository metrics over time")
	cloneCmd.Flags().BoolVar(&ghorgPreserveScmHostname, "preserve-scm-hostname", false, "GHORG_PRESERVE_SCM_HOSTNAME - Organize clones into subdirectories by SCM hostname (e.g., github.com/kubernetes, gitlab.com/myorg). Useful when cloning from multiple SCM providers")
	cloneCmd.Flags().BoolVar(&ghorgPruneUntouched, "prune-untouched", false, "GHORG_PRUNE_UNTOUCHED - Remove local repositories without uncommitted changes. See sample-conf.yaml for details. Prompts before deletion unless using --prune-untouched-no-confirm")
	cloneCmd.Flags().BoolVar(&ghorgPruneUntouchedNoConfirm, "prune-untouched-no-confirm", false, "GHORG_PRUNE_UNTOUCHED_NO_CONFIRM - Skip confirmation when pruning untouched repositories. Use with caution")
	cloneCmd.Flags().StringVarP(&baseURL, "base-url", "", "", "GHORG_SCM_BASE_URL - Base URL for self-hosted SCM instances. For GitHub Enterprise use format: https://github.example.com/api/v3. Required for self-hosted GitLab, Gitea, and GitHub")
	cloneCmd.Flags().StringVarP(&concurrency, "concurrency", "", "", "GHORG_CONCURRENCY - Maximum number of concurrent clone operations (goroutines). Higher values speed up cloning but use more resources. (default: 25)")
	cloneCmd.Flags().StringVarP(&cloneDelaySeconds, "clone-delay-seconds", "", "", "GHORG_CLONE_DELAY_SECONDS - Delay in seconds between each clone operation. Useful for rate limiting or reducing server load. Auto-sets concurrency to 1 when > 0 (default: 0)")
	cloneCmd.Flags().StringVarP(&cloneDepth, "clone-depth", "", "", "GHORG_CLONE_DEPTH - Create shallow clones with limited history (e.g., --clone-depth=1 for latest commit only). Reduces clone time and disk usage")
	cloneCmd.Flags().StringVarP(&topics, "topics", "", "", "GHORG_TOPICS - Comma-separated list of GitHub/Gitea topics to filter repositories (e.g., --topics=docker,kubernetes). Only clones repos with matching topics")
	cloneCmd.Flags().StringVarP(&outputDir, "output-dir", "", "", "GHORG_OUTPUT_DIR - Custom name for the directory where repositories will be cloned. (default: name of org/user being cloned)")
	cloneCmd.Flags().StringVarP(&matchPrefix, "match-prefix", "", "", "GHORG_MATCH_PREFIX - Only clone repositories with names starting with specified prefix(es). Comma-separated list supported (e.g., --match-prefix=frontend,backend)")
	cloneCmd.Flags().StringVarP(&excludeMatchPrefix, "exclude-match-prefix", "", "", "GHORG_EXCLUDE_MATCH_PREFIX - Exclude repositories with names starting with specified prefix(es). Comma-separated list supported")
	cloneCmd.Flags().StringVarP(&matchRegex, "match-regex", "", "", "GHORG_MATCH_REGEX - Only clone repositories whose names match the provided regular expression (e.g., --match-regex='^app-.*')")
	cloneCmd.Flags().StringVarP(&excludeMatchRegex, "exclude-match-regex", "", "", "GHORG_EXCLUDE_MATCH_REGEX - Exclude repositories whose names match the provided regular expression")
	cloneCmd.Flags().StringVarP(&gitlabGroupExcludeMatchRegex, "gitlab-group-exclude-match-regex", "", "", "GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX - GitLab only: Exclude groups/subgroups matching the regex pattern")
	cloneCmd.Flags().StringVarP(&gitlabGroupMatchRegex, "gitlab-group-match-regex", "", "", "GHORG_GITLAB_GROUP_MATCH_REGEX - GitLab only: Only clone from groups/subgroups matching the regex pattern")
	cloneCmd.Flags().StringVarP(&ghorgIgnorePath, "ghorgignore-path", "", "", "GHORG_IGNORE_PATH - Custom path to ghorgignore file (similar to .gitignore but for repos). Default: $HOME/.config/ghorg/ghorgignore")
	cloneCmd.Flags().StringVarP(&ghorgOnlyPath, "ghorgonly-path", "", "", "GHORG_ONLY_PATH - Custom path to ghorgonly file (whitelist of repos to clone). Default: $HOME/.config/ghorg/ghorgonly")
	cloneCmd.Flags().StringVarP(&exitCodeOnCloneInfos, "exit-code-on-clone-infos", "", "", "GHORG_EXIT_CODE_ON_CLONE_INFOS - Exit code when informational messages occur during cloning (non-critical issues). Useful for CI/CD pipelines (default: 0)")
	cloneCmd.Flags().StringVarP(&exitCodeOnCloneIssues, "exit-code-on-clone-issues", "", "", "GHORG_EXIT_CODE_ON_CLONE_ISSUES - Exit code when issues/errors occur during cloning. Useful for CI/CD failure detection (default: 1)")
	cloneCmd.Flags().StringVarP(&gitFilter, "git-filter", "", "", "GHORG_GIT_FILTER - Arguments to pass to git's --filter flag. Use --git-filter=blob:none to exclude binary objects and reduce clone size. Requires git 2.19+")
	cloneCmd.Flags().BoolVarP(&githubTokenFromGithubApp, "github-token-from-github-app", "", false, "GHORG_GITHUB_TOKEN_FROM_GITHUB_APP - GitHub only: Treat the provided token as a GitHub App token (when obtained outside ghorg). Use with pre-generated app tokens")
	cloneCmd.Flags().StringVarP(&githubAppPemPath, "github-app-pem-path", "", "", "GHORG_GITHUB_APP_PEM_PATH - GitHub only: Path to GitHub App private key (.pem file) for app-based authentication. Requires --github-app-id and --github-app-installation-id")
	cloneCmd.Flags().StringVarP(&githubAppInstallationID, "github-app-installation-id", "", "", "GHORG_GITHUB_APP_INSTALLATION_ID - GitHub only: Installation ID for GitHub App authentication. Find in org settings URL")
	cloneCmd.Flags().StringVarP(&githubFilterLanguage, "github-filter-language", "", "", "GHORG_GITHUB_FILTER_LANGUAGE - GitHub only: Filter repositories by programming language. Comma-separated values (e.g., --github-filter-language=go,python)")
	cloneCmd.Flags().StringVarP(&githubUserOption, "github-user-option", "", "", "GHORG_GITHUB_USER_OPTION - GitHub only: When using --clone-type=user, specify which repos to include: 'all', 'owner' (created by user), or 'member' (contributed to). (default: owner)")
	cloneCmd.Flags().StringVarP(&githubAppID, "github-app-id", "", "", "GHORG_GITHUB_APP_ID - GitHub only: GitHub App ID for app-based authentication. Required with --github-app-pem-path")
	cloneCmd.Flags().StringVar(&sshHostname, "ssh-hostname", "", "GHORG_SSH_HOSTNAME - Custom hostname to use in SSH clone URLs. Useful for SSH aliases in ~/.ssh/config (e.g., --ssh-hostname=my-github-alias creates git@my-github-alias:org/repo.git URLs)")

	reCloneCmd.Flags().StringVarP(&ghorgReClonePath, "reclone-path", "", "", "GHORG_RECLONE_PATH - If you want to set a path other than $HOME/.config/ghorg/reclone.yaml for your reclone configuration")
	reCloneCmd.Flags().StringVar(&sshHostname, "ssh-hostname", "", "GHORG_SSH_HOSTNAME - Hostname to use for SSH clone URLs (e.g., my-github-alias for git@my-github-alias:org/repo.git)")
	reCloneCmd.Flags().BoolVar(&ghorgReCloneQuiet, "quiet", false, "GHORG_RECLONE_QUIET - Quiet logging output")
	reCloneCmd.Flags().BoolVar(&ghorgReCloneList, "list", false, "Prints reclone commands and optional descriptions to stdout then will exit 0. Does not obsfucate tokens, and is only available as a commandline argument")
	reCloneCmd.Flags().BoolVar(&ghorgReCloneEnvConfigOnly, "env-config-only", false, "GHORG_RECLONE_ENV_CONFIG_ONLY - Only use environment variables to set the configuration for all reclones.")

	lsCmd.Flags().BoolP("long", "l", false, "Display detailed information about each clone directory, including size and number of repositories. Note: This may take longer depending on the number and size of the cloned organizations.")
	lsCmd.Flags().BoolP("total", "t", false, "Display total amounts of all repos cloned. Note: This may take longer depending on the number and size of the cloned organizations.")

	recloneCronCmd.Flags().StringVarP(&cronTimerMinutes, "minutes", "m", "", "GHORG_CRON_TIMER_MINUTES - Number of minutes to run the reclone command on a cron")

	recloneServerCmd.Flags().StringVarP(&recloneServerPort, "port", "p", "", "GHORG_RECLONE_SERVER_PORT - Specifiy the port the reclone server will run on.")

	rootCmd.AddCommand(lsCmd, versionCmd, cloneCmd, reCloneCmd, examplesCmd, recloneServerCmd, recloneCronCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
