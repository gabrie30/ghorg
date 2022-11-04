package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	color                        string
	baseURL                      string
	concurrency                  string
	exitCodeOnCloneInfos         string
	exitCodeOnCloneIssues        string
	outputDir                    string
	topics                       string
	gitFilter                    string
	includeSubmodules            bool
	skipArchived                 bool
	skipForks                    bool
	backup                       bool
	noClean                      bool
	dryRun                       bool
	prune                        bool
	pruneNoConfirm               bool
	cloneWiki                    bool
	preserveDir                  bool
	insecureGitlabClient         bool
	insecureGiteaClient          bool
	fetchAll                     bool
	ghorgReCloneVerbose          bool
	ghorgReCloneQuiet            bool
	args                         []string
	cloneErrors                  []string
	cloneInfos                   []string
	targetCloneSource            string
	matchPrefix                  string
	excludeMatchPrefix           string
	matchRegex                   string
	excludeMatchRegex            string
	config                       string
	gitlabGroupExcludeMatchRegex string
	ghorgIgnorePath              string
	ghorgReClonePath             string
	quietMode                    bool
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
		case "GHORG_NO_CLEAN":
			os.Setenv(envVar, "false")
		case "GHORG_FETCH_ALL":
			os.Setenv(envVar, "false")
		case "GHORG_DRY_RUN":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE":
			os.Setenv(envVar, "false")
		case "GHORG_PRUNE_NO_CONFIRM":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_GITLAB_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_INSECURE_GITEA_CLIENT":
			os.Setenv(envVar, "false")
		case "GHORG_BACKUP":
			os.Setenv(envVar, "false")
		case "GHORG_RECLONE_VERBOSE":
			os.Setenv(envVar, "false")
		case "GHORG_RECLONE_QUIET":
			os.Setenv(envVar, "false")
		case "GHORG_COLOR":
			os.Setenv(envVar, "disabled")
		case "GHORG_PRESERVE_DIRECTORY_STRUCTURE":
			os.Setenv(envVar, "false")
		case "GHORG_CONCURRENCY":
			os.Setenv(envVar, "25")
		case "GHORG_QUIET":
			os.Setenv(envVar, "false")
		case "GHORG_EXIT_CODE_ON_CLONE_INFOS":
			os.Setenv(envVar, "0")
		case "GHORG_EXIT_CODE_ON_CLONE_ISSUES":
			os.Setenv(envVar, "1")
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
	}

	getOrSetDefaults("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	getOrSetDefaults("GHORG_BRANCH")
	getOrSetDefaults("GHORG_CLONE_PROTOCOL")
	getOrSetDefaults("GHORG_CLONE_TYPE")
	getOrSetDefaults("GHORG_SCM_TYPE")
	getOrSetDefaults("GHORG_SKIP_ARCHIVED")
	getOrSetDefaults("GHORG_SKIP_FORKS")
	getOrSetDefaults("GHORG_NO_CLEAN")
	getOrSetDefaults("GHORG_FETCH_ALL")
	getOrSetDefaults("GHORG_PRUNE")
	getOrSetDefaults("GHORG_PRUNE_NO_CONFIRM")
	getOrSetDefaults("GHORG_DRY_RUN")
	getOrSetDefaults("GHORG_CLONE_WIKI")
	getOrSetDefaults("GHORG_INSECURE_GITLAB_CLIENT")
	getOrSetDefaults("GHORG_INSECURE_GITEA_CLIENT")
	getOrSetDefaults("GHORG_BACKUP")
	getOrSetDefaults("GHORG_RECLONE_VERBOSE")
	getOrSetDefaults("GHORG_RECLONE_QUIET")
	getOrSetDefaults("GHORG_CONCURRENCY")
	getOrSetDefaults("GHORG_INCLUDE_SUBMODULES")
	getOrSetDefaults("GHORG_EXIT_CODE_ON_CLONE_INFOS")
	getOrSetDefaults("GHORG_EXIT_CODE_ON_CLONE_ISSUES")
	// Optionally set
	getOrSetDefaults("GHORG_GITHUB_TOKEN")
	getOrSetDefaults("GHORG_COLOR")
	getOrSetDefaults("GHORG_TOPICS")
	getOrSetDefaults("GHORG_GITLAB_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_USERNAME")
	getOrSetDefaults("GHORG_BITBUCKET_APP_PASSWORD")
	getOrSetDefaults("GHORG_BITBUCKET_OAUTH_TOKEN")
	getOrSetDefaults("GHORG_SCM_BASE_URL")
	getOrSetDefaults("GHORG_PRESERVE_DIRECTORY_STRUCTURE")
	getOrSetDefaults("GHORG_OUTPUT_DIR")
	getOrSetDefaults("GHORG_MATCH_REGEX")
	getOrSetDefaults("GHORG_EXCLUDE_MATCH_REGEX")
	getOrSetDefaults("GHORG_MATCH_PREFIX")
	getOrSetDefaults("GHORG_EXCLUDE_MATCH_PREFIX")
	getOrSetDefaults("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX")
	getOrSetDefaults("GHORG_IGNORE_PATH")
	getOrSetDefaults("GHORG_RECLONE_PATH")
	getOrSetDefaults("GHORG_QUIET")
	getOrSetDefaults("GHORG_GIT_FILTER")

	if os.Getenv("GHORG_DEBUG") != "" {
		viper.Debug()
		fmt.Println("Viper config file used:", viper.ConfigFileUsed())
		fmt.Printf("GHORG_CONFIG SET TO: %s\n", os.Getenv("GHORG_CONFIG"))
	}
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.PersistentFlags().StringVar(&color, "color", "", "GHORG_COLOR - toggles colorful output, enabled/disabled (default: disabled)")
	rootCmd.PersistentFlags().StringVar(&config, "config", "", "GHORG_CONFIG - manually set the path to your config file")

	viper.SetDefault("config", configs.DefaultConfFile())
	viper.AutomaticEnv()

	_ = viper.BindPFlag("color", rootCmd.PersistentFlags().Lookup("color"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - Protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - Absolute path the ghorg_* directory will be created. Must end with / (default $HOME/Desktop/ghorg)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - Branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_GITEA_TOKEN/GHORG_BITBUCKET_APP_PASSWORD/GHORG_BITBUCKET_OAUTH_TOKEN - scm token to clone with")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - Bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - Type of scm used, github, gitlab, gitea or bitbucket (default github)")
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - Clone target type, user or org (default org)")
	cloneCmd.Flags().BoolVar(&skipArchived, "skip-archived", false, "GHORG_SKIP_ARCHIVED - Skips archived repos, github/gitlab/gitea only")
	cloneCmd.Flags().BoolVar(&noClean, "no-clean", false, "GHORG_NO_CLEAN - Only clones new repos and does not perform a git clean on existing repos")
	cloneCmd.Flags().BoolVar(&prune, "prune", false, "GHORG_PRUNE - Deletes all files/directories found in your local clone directory that are not found on the remote (e.g., after remote deletion).  With GHORG_SKIP_ARCHIVED set, archived repositories will also be pruned from your local clone.  Will prompt before deleting any files unless used in combination with --prune-no-confirm")
	cloneCmd.Flags().BoolVar(&pruneNoConfirm, "prune-no-confirm", false, "GHORG_PRUNE_NO_CONFIRM - Don't prompt on every prune candidate, just delete")
	cloneCmd.Flags().BoolVar(&fetchAll, "fetch-all", false, "GHORG_FETCH_ALL - Fetches all remote branches for each repo by running a git fetch --all")
	cloneCmd.Flags().BoolVar(&dryRun, "dry-run", false, "GHORG_DRY_RUN - Perform a dry run of the clone; fetches repos but does not clone them")
	cloneCmd.Flags().BoolVar(&insecureGitlabClient, "insecure-gitlab-client", false, "GHORG_INSECURE_GITLAB_CLIENT - Skip TLS certificate verification for hosted gitlab instances")
	cloneCmd.Flags().BoolVar(&insecureGiteaClient, "insecure-gitea-client", false, "GHORG_INSECURE_GITEA_CLIENT - Must be set to clone from a Gitea instance using http")
	cloneCmd.Flags().BoolVar(&cloneWiki, "clone-wiki", false, "GHORG_CLONE_WIKI - Additionally clone the wiki page for repo")
	cloneCmd.Flags().BoolVar(&skipForks, "skip-forks", false, "GHORG_SKIP_FORKS - Skips repo if its a fork, github/gitlab/gitea only")
	cloneCmd.Flags().BoolVar(&preserveDir, "preserve-dir", false, "GHORG_PRESERVE_DIRECTORY_STRUCTURE - Clones repos in a directory structure that matches gitlab namespaces eg company/unit/subunit/app would clone into ghorg/unit/subunit/app, gitlab only")
	cloneCmd.Flags().BoolVar(&backup, "backup", false, "GHORG_BACKUP - Backup mode, clone as mirror, no working copy (ignores branch parameter)")
	cloneCmd.Flags().BoolVar(&quietMode, "quiet", false, "GHORG_QUIET - Emit critical output only")
	cloneCmd.Flags().BoolVar(&includeSubmodules, "include-submodules", false, "GHORG_INCLUDE_SUBMODULES - Include submodules in all clone and pull operations.")
	cloneCmd.Flags().StringVarP(&baseURL, "base-url", "", "", "GHORG_SCM_BASE_URL - Change SCM base url, for on self hosted instances (currently gitlab, gitea and github (use format of https://git.mydomain.com/api/v3))")
	cloneCmd.Flags().StringVarP(&concurrency, "concurrency", "", "", "GHORG_CONCURRENCY - Max goroutines to spin up while cloning (default 25)")
	cloneCmd.Flags().StringVarP(&topics, "topics", "", "", "GHORG_TOPICS - Comma separated list of github/gitea topics to filter for")
	cloneCmd.Flags().StringVarP(&outputDir, "output-dir", "", "", "GHORG_OUTPUT_DIR - Name of directory repos will be cloned into (default name of org/repo being cloned")
	cloneCmd.Flags().StringVarP(&matchPrefix, "match-prefix", "", "", "GHORG_MATCH_PREFIX - Only clone repos with matching prefix, can be a comma separated list")
	cloneCmd.Flags().StringVarP(&excludeMatchPrefix, "exclude-match-prefix", "", "", "GHORG_EXCLUDE_MATCH_PREFIX - Exclude cloning repos with matching prefix, can be a comma separated list")
	cloneCmd.Flags().StringVarP(&matchRegex, "match-regex", "", "", "GHORG_MATCH_REGEX - Only clone repos that match name to regex provided")
	cloneCmd.Flags().StringVarP(&excludeMatchRegex, "exclude-match-regex", "", "", "GHORG_EXCLUDE_MATCH_REGEX - Exclude cloning repos that match name to regex provided")
	cloneCmd.Flags().StringVarP(&gitlabGroupExcludeMatchRegex, "gitlab-group-exclude-match-regex", "", "", "GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX - Exclude cloning gitlab groups that match name to regex provided")
	cloneCmd.Flags().StringVarP(&ghorgIgnorePath, "ghorgignore-path", "", "", "GHORG_IGNORE_PATH - If you want to set a path other than $HOME/.config/ghorg/ghorgignore for your ghorgignore")
	cloneCmd.Flags().StringVarP(&exitCodeOnCloneInfos, "exit-code-on-clone-infos", "", "", "GHORG_EXIT_CODE_ON_CLONE_INFOS - Allows you to control the exit code when ghorg runs into a problem (info level message) cloning a repo from the remote. Info messages will appear after a clone is complete, similar to success messages. (default 0)")
	cloneCmd.Flags().StringVarP(&exitCodeOnCloneIssues, "exit-code-on-clone-issues", "", "", "GHORG_EXIT_CODE_ON_CLONE_ISSUES - Allows you to control the exit code when ghorg runs into a problem (issue level message) cloning a repo from the remote. Issue messages will appear after a clone is complete, similar to success messages (default 1)")
	cloneCmd.Flags().StringVarP(&gitFilter, "git-filter", "", "", "GHORG_GIT_FILTER - Allows you to pass arguments to git's filter flag. Useful for filtering out binary objects from repos with --git-filter=blob:none, this requires git version 2.19 or greater.")

	reCloneCmd.Flags().StringVarP(&ghorgReClonePath, "reclone-path", "", "", "GHORG_RECLONE_PATH - If you want to set a path other than $HOME/.config/ghorg/reclone.yaml for your reclone configuration")
	reCloneCmd.Flags().BoolVar(&ghorgReCloneVerbose, "verbose", false, "GHORG_RECLONE_VERBOSE - Verbose logging output")
	reCloneCmd.Flags().BoolVar(&ghorgReCloneQuiet, "quiet", false, "GHORG_RECLONE_QUIET - Quiet logging output")

	rootCmd.AddCommand(lsCmd, versionCmd, cloneCmd, reCloneCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
