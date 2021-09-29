package cmd

import (
	"fmt"
	"os"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	skipForks         bool
	backup            bool
	noClean           bool
	preserveDir       bool
	args              []string
	cloneErrors       []string
	cloneInfos        []string
	targetCloneSource string
	matchPrefix       string
	matchRegex        string
	config            string
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
		case "GHORG_CLONE_PROTOCOL":
			os.Setenv(envVar, "https")
		case "GHORG_CLONE_TYPE":
			os.Setenv(envVar, "org")
		case "GHORG_SCM_TYPE":
			os.Setenv(envVar, "github")
		case "GHORG_SKIP_ARCHIVED":
			os.Setenv(envVar, "false")
		case "GHORG_SKIP_FORKS":
			os.Setenv(envVar, "false")
		case "GHORG_NO_CLEAN":
			os.Setenv(envVar, "false")
		case "GHORG_BACKUP":
			os.Setenv(envVar, "false")
		case "GHORG_COLOR":
			os.Setenv(envVar, "disabled")
		case "GHORG_PRESERVE_DIRECTORY_STRUCTURE":
			os.Setenv(envVar, "false")
		case "GHORG_CONCURRENCY":
			os.Setenv(envVar, "25")
		}
	} else {
		s := viper.GetString(envVar)
		// envs that need a trailing slash
		if envVar == "GHORG_SCM_BASE_URL" || envVar == "GHORG_ABSOLUTE_PATH_TO_CLONE_TO" {
			os.Setenv(envVar, configs.EnsureTrailingSlash(s))
		} else {
			os.Setenv(envVar, s)
		}
	}

	if os.Getenv("GHORG_DEBUG") != "" {
		fmt.Printf("%s: %s\n", envVar, os.Getenv(envVar))
	}
}

func InitConfig() {
	if config != "" {
		viper.SetConfigFile(config)
		os.Setenv("GHORG_CONF", config)
	} else {
		config = configs.DefaultConfFile()
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configs.GhorgDir())
		viper.SetConfigName("conf")
		os.Setenv("GHORG_CONF", configs.DefaultConfFile())

	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.Setenv("GHORG_CONF", "none")
		} else {
			colorlog.PrintError(fmt.Sprintf("Something unexpected happened reading configuration file: %s, err: %s", os.Getenv("GHORG_CONF"), err))
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
	getOrSetDefaults("GHORG_BACKUP")
	getOrSetDefaults("GHORG_CONCURRENCY")
	getOrSetDefaults("GHORG_MATCH_PREFIX")
	// Optionally set
	getOrSetDefaults("GHORG_GITHUB_TOKEN")
	getOrSetDefaults("GHORG_COLOR")
	getOrSetDefaults("GHORG_TOPICS")
	getOrSetDefaults("GHORG_GITLAB_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_USERNAME")
	getOrSetDefaults("GHORG_BITBUCKET_APP_PASSWORD")
	getOrSetDefaults("GHORG_SCM_BASE_URL")
	getOrSetDefaults("GHORG_PRESERVE_DIRECTORY_STRUCTURE")
	getOrSetDefaults("GHORG_OUTPUT_DIR")
	getOrSetDefaults("GHORG_MATCH_REGEX")

	if os.Getenv("GHORG_DEBUG") != "" {
		viper.Debug()
		fmt.Println("Viper config file used:", viper.ConfigFileUsed())
		fmt.Printf("GHORG_CONF SET TO: %s\n", os.Getenv("GHORG_CONF"))
	}

}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.PersistentFlags().StringVar(&color, "color", "", "GHORG_COLOR - toggles colorful output, enabled/disabled (default: disabled)")
	rootCmd.PersistentFlags().StringVar(&config, "config", "", "manually set the path to your config file")

	viper.SetDefault("config", configs.DefaultConfFile())

	_ = viper.BindPFlag("color", rootCmd.PersistentFlags().Lookup("color"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	cloneCmd.Flags().StringVar(&protocol, "protocol", "", "GHORG_CLONE_PROTOCOL - protocol to clone with, ssh or https, (default https)")
	cloneCmd.Flags().StringVarP(&path, "path", "p", "", "GHORG_ABSOLUTE_PATH_TO_CLONE_TO - absolute path the ghorg_* directory will be created. Must end with / (default $HOME/Desktop/ghorg)")
	cloneCmd.Flags().StringVarP(&branch, "branch", "b", "", "GHORG_BRANCH - branch left checked out for each repo cloned (default master)")
	cloneCmd.Flags().StringVarP(&token, "token", "t", "", "GHORG_GITHUB_TOKEN/GHORG_GITLAB_TOKEN/GHORG_GITEA_TOKEN/GHORG_BITBUCKET_APP_PASSWORD - scm token to clone with")
	cloneCmd.Flags().StringVarP(&bitbucketUsername, "bitbucket-username", "", "", "GHORG_BITBUCKET_USERNAME - bitbucket only: username associated with the app password")
	cloneCmd.Flags().StringVarP(&scmType, "scm", "s", "", "GHORG_SCM_TYPE - type of scm used, github, gitlab, gitea or bitbucket (default github)")
	cloneCmd.Flags().StringVarP(&cloneType, "clone-type", "c", "", "GHORG_CLONE_TYPE - clone target type, user or org (default org)")
	cloneCmd.Flags().BoolVar(&skipArchived, "skip-archived", false, "GHORG_SKIP_ARCHIVED - skips archived repos, github/gitlab/gitea only")
	cloneCmd.Flags().BoolVar(&noClean, "no-clean", false, "GHORG_NO_CLEAN - only clones new repos and does not perform a git clean on existing repos")
	cloneCmd.Flags().BoolVar(&skipForks, "skip-forks", false, "GHORG_SKIP_FORKS - skips repo if its a fork, github/gitlab/gitea only")
	cloneCmd.Flags().BoolVar(&preserveDir, "preserve-dir", false, "GHORG_PRESERVE_DIRECTORY_STRUCTURE - clones repos in a directory structure that matches gitlab namespaces eg company/unit/subunit/app would clone into ghorg/unit/subunit/app, gitlab only")
	cloneCmd.Flags().BoolVar(&backup, "backup", false, "GHORG_BACKUP - backup mode, clone as mirror, no working copy (ignores branch parameter)")
	cloneCmd.Flags().StringVarP(&baseURL, "base-url", "", "", "GHORG_SCM_BASE_URL - change SCM base url, for on self hosted instances (currently gitlab, gitea and github (use format of https://git.mydomain.com/api/v3))")
	cloneCmd.Flags().StringVarP(&concurrency, "concurrency", "", "", "GHORG_CONCURRENCY - max goroutines to spin up while cloning (default 25)")
	cloneCmd.Flags().StringVarP(&topics, "topics", "", "", "GHORG_TOPICS - comma separated list of github/gitea topics to filter for")
	cloneCmd.Flags().StringVarP(&outputDir, "output-dir", "", "", "GHORG_OUTPUT_DIR - name of directory repos will be cloned into (default name of org/repo being cloned")
	cloneCmd.Flags().StringVarP(&matchPrefix, "match-prefix", "", "", "GHORG_MATCH_PREFIX - only clone repos with matching prefix, can be a comma separated list (default \"\")")
	cloneCmd.Flags().StringVarP(&matchRegex, "match-regex", "", "", "GHORG_MATCH_REGEX - only clone repos that match name to regex provided")

	rootCmd.AddCommand(lsCmd, versionCmd, cloneCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
