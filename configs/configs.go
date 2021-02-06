// Package configs sets up the environment. First it sets a number of default envs, then looks in the $HOME/ghorg/conf.yaml to overwrite the defaults. These values will be superseded by any command line flags used
package configs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/scm"
	"github.com/gabrie30/ghorg/utils"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// ErrNoGitHubToken error message when token is not found
	ErrNoGitHubToken = errors.New("Could not find a valid github token. GHORG_GITHUB_TOKEN or (--token, -t) flag must be set. Create a personal access token, then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoGitLabToken error message when token is not found
	ErrNoGitLabToken = errors.New("Could not find a valid gitlab token. GHORG_GITLAB_TOKEN or (--token, -t) flag must be set. Create a token from gitlab then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoBitbucketUsername error message when no username found
	ErrNoBitbucketUsername = errors.New("Could not find bitbucket username. GHORG_BITBUCKET_USERNAME or (--bitbucket-username) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrNoBitbucketAppPassword error message when no app password found
	ErrNoBitbucketAppPassword = errors.New("Could not find a valid bitbucket app password. GHORG_BITBUCKET_APP_PASSWORD or (--token, -t) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrIncorrectScmType indicates an unsupported scm type being used
	ErrIncorrectScmType = errors.New("GHORG_SCM_TYPE or --scm must be one of " + strings.Join(scm.SupportedClients(), ", "))

	// ErrIncorrectCloneType indicates an unsupported clone type being used
	ErrIncorrectCloneType = errors.New("GHORG_CLONE_TYPE or --clone-type must be one of org or user")

	// ErrIncorrectProtocolType indicates an unsupported protocol type being used
	ErrIncorrectProtocolType = errors.New("GHORG_CLONE_PROTOCOL or --protocol must be one of https or ssh")
)

// Load triggers the configs to load first
func Load(config string) {
	if config != "" {
		viper.SetConfigFile(config)
	} else {
		viper.SetConfigType("yaml")
		viper.AddConfigPath(GhorgDir())
		viper.SetConfigName("conf")
	}

	ghorgDir := GhorgDir()

	if err := viper.ReadInConfig(); err != nil {

		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired

			if XConfigHomeSet() {
				colorlog.PrintError("Found XDG_CONFIG_HOME set")
			}

			fmt.Println("")
			colorlog.PrintError(err)

			colorlog.PrintError(fmt.Sprintf("Could not find %s/conf.yaml file, add one by running the following \n \n $ mkdir -p %s \n $ curl https://raw.githubusercontent.com/gabrie30/ghorg/master/sample-conf.yaml > %s/conf.yaml \n", ghorgDir, ghorgDir, ghorgDir))
			log.Fatal("Exiting due to improper configuration")

		} else {
			// Config file was found but another error was produced
			colorlog.PrintError(fmt.Sprintf("Something unexpected happened reading configuration file %s/conf.yaml, err: %s", ghorgDir, err))
		}
	}

	// Set With Default values
	getOrSetDefaults("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	getOrSetDefaults("GHORG_BRANCH")
	getOrSetDefaults("GHORG_CLONE_PROTOCOL")
	getOrSetDefaults("GHORG_CLONE_TYPE")
	getOrSetDefaults("GHORG_SCM_TYPE")
	getOrSetDefaults("GHORG_COLOR")
	getOrSetDefaults("GHORG_SKIP_ARCHIVED")
	getOrSetDefaults("GHORG_SKIP_FORKS")
	getOrSetDefaults("GHORG_BACKUP")
	getOrSetDefaults("GHORG_CONCURRENCY")
	getOrSetDefaults("GHORG_MATCH_PREFIX")
	// Optionally set
	getOrSetDefaults("GHORG_GITHUB_TOKEN")
	getOrSetDefaults("GHORG_TOPICS")
	getOrSetDefaults("GHORG_GITLAB_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_USERNAME")
	getOrSetDefaults("GHORG_BITBUCKET_APP_PASSWORD")
	getOrSetDefaults("GHORG_SCM_BASE_URL")
	getOrSetDefaults("GHORG_PRESERVE_DIRECTORY_STRUCTURE")
	getOrSetDefaults("GHORG_OUTPUT_DIR")
}

// GetRequiredString verifies env is set
func GetRequiredString(key string) string {
	value := viper.GetString(key)

	if isZero(value) {
		log.Fatalf("Fatal: '%s' ENV VAR is required", key)
	}

	return value
}

func isZero(value interface{}) bool {
	return value == reflect.Zero(reflect.TypeOf(value)).Interface()
}

func getOrSetDefaults(envVar string) {
	fmt.Println("getOrSetDefaults", envVar)
	// When a user does not set value in $HOME/.config/ghorg/conf.yaml set the default values, else set env to what they have added to the file.
	if viper.GetString(envVar) == "" {
		switch envVar {
		case "GHORG_ABSOLUTE_PATH_TO_CLONE_TO":
			os.Setenv(envVar, HomeDir()+"/Desktop/ghorg/")
		case "GHORG_CLONE_PROTOCOL":
			os.Setenv(envVar, "https")
		case "GHORG_CLONE_TYPE":
			os.Setenv(envVar, "org")
		case "GHORG_SCM_TYPE":
			os.Setenv(envVar, "github")
		case "GHORG_COLOR":
			os.Setenv(envVar, "on")
		case "GHORG_SKIP_ARCHIVED":
			os.Setenv(envVar, "false")
		case "GHORG_SKIP_FORKS":
			os.Setenv(envVar, "false")
		case "GHORG_BACKUP":
			os.Setenv(envVar, "false")
		case "GHORG_PRESERVE_DIRECTORY_STRUCTURE":
			os.Setenv(envVar, "false")
		case "GHORG_CONCURRENCY":
			os.Setenv(envVar, "25")
		}
	} else {
		s := viper.GetString(envVar)

		// envs that need a trailing slash
		if envVar == "GHORG_SCM_BASE_URL" || envVar == "GHORG_ABSOLUTE_PATH_TO_CLONE_TO" {
			os.Setenv(envVar, EnsureTrailingSlash(s))
		} else {
			os.Setenv(envVar, s)
		}
	}
}

// EnsureTrailingSlash takes a string and ensures a single / is appened
func EnsureTrailingSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		s = s + "/"
	}

	return s
}

// GhorgIgnoreLocation returns the path of users ghorgignore
func GhorgIgnoreLocation() string {
	return GhorgDir() + "/ghorgignore"
}

// GhorgIgnoreDetected identify if a ghorgignore file exists in users .config/ghorg directory
func GhorgIgnoreDetected() bool {
	_, err := os.Stat(GhorgIgnoreLocation())
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// GhorgDir returns the ghorg directory path
func GhorgDir() string {
	if XConfigHomeSet() {
		return os.Getenv("XDG_CONFIG_HOME") + "/ghorg"
	}

	return HomeDir() + "/.config/ghorg"
}

// XConfigHomeSet checks for XDG_CONFIG_HOME env set
func XConfigHomeSet() bool {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return true
	}

	return false
}

// HomeDir finds the users home directory
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error trying to find users home directory")
	}

	return home
}

// GetOrSetToken will set token based on scm
func GetOrSetToken() {
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		getOrSetGitHubToken()
	case "gitlab":
		getOrSetGitLabToken()
	case "bitbucket":
		getOrSetBitBucketToken()
	}
}

func getOrSetGitHubToken() {
	var token string
	if isZero(os.Getenv("GHORG_GITHUB_TOKEN")) || len(os.Getenv("GHORG_GITHUB_TOKEN")) != 40 {
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token = strings.TrimSuffix(string(out), "\n")

		os.Setenv("GHORG_GITHUB_TOKEN", token)
	}
}

func getOrSetGitLabToken() {
	var token string
	if isZero(os.Getenv("GHORG_GITLAB_TOKEN")) || len(os.Getenv("GHORG_GITLAB_TOKEN")) != 20 {
		cmd := `security find-internet-password -s gitlab.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token = strings.TrimSuffix(string(out), "\n")

		os.Setenv("GHORG_GITLAB_TOKEN", token)
	}
}

func getOrSetBitBucketToken() {
	var token string
	if isZero(os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")) || len(os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")) != 20 {
		cmd := `security find-internet-password -s bitbucket.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token = strings.TrimSuffix(string(out), "\n")

		os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", token)
	}
}

// VerifyTokenSet checks to make sure env is set for the correct scm provider
func VerifyTokenSet() error {
	var tokenLength int
	var token string
	scmProvider := os.Getenv("GHORG_SCM_TYPE")

	if scmProvider == "github" {
		tokenLength = 40
		token = os.Getenv("GHORG_GITHUB_TOKEN")
	}

	if scmProvider == "gitlab" {
		tokenLength = 20
		token = os.Getenv("GHORG_GITLAB_TOKEN")
	}

	if scmProvider == "bitbucket" {
		tokenLength = 20
		token = os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
		if os.Getenv("GHORG_BITBUCKET_USERNAME") == "" {
			return ErrNoBitbucketUsername
		}
	}

	if len(token) != tokenLength {
		if scmProvider == "github" {
			return ErrNoGitHubToken
		}

		if scmProvider == "gitlab" {
			return ErrNoGitLabToken
		}

		if scmProvider == "bitbucket" {
			return ErrNoBitbucketAppPassword
		}

	}

	return nil
}

// VerifyConfigsSetCorrectly makes sure flags are set to appropriate values
func VerifyConfigsSetCorrectly() error {
	scmType := os.Getenv("GHORG_SCM_TYPE")
	cloneType := os.Getenv("GHORG_CLONE_TYPE")
	protocol := os.Getenv("GHORG_CLONE_PROTOCOL")

	if !utils.IsStringInSlice(scmType, scm.SupportedClients()) {
		return ErrIncorrectScmType
	}

	if cloneType != "user" && cloneType != "org" {
		return ErrIncorrectCloneType
	}

	if protocol != "ssh" && protocol != "https" {
		return ErrIncorrectProtocolType
	}

	return nil
}
