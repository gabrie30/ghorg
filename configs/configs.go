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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// ErrNoGitHubToken error message when token is not found
	ErrNoGitHubToken = errors.New("Could not find a valid github token. GHORG_GITHUB_TOKEN or (--token, -t) flag must be set. Create a personal access token, then set it in your $HOME/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoGitLabToken error message when token is not found
	ErrNoGitLabToken = errors.New("Could not find a valid gitlab token. GHORG_GITLAB_TOKEN or (--token, -t) flag must be set. Create a token from gitlab then set it in your $HOME/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoBitbucketUsername error message when no username found
	ErrNoBitbucketUsername = errors.New("Could not find bitbucket username. GHORG_BITBUCKET_USERNAME or (--bitbucket-username) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrNoBitbucketAppPassword error message when no app password found
	ErrNoBitbucketAppPassword = errors.New("Could not find a valid bitbucket app password. GHORG_BITBUCKET_APP_PASSWORD or (--token, -t) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrIncorrectScmType indicates an unsupported scm type being used
	ErrIncorrectScmType = errors.New("GHORG_SCM_TYPE or --scm must be one of github, gitlab, or bitbucket")

	// ErrIncorrectCloneType indicates an unsupported clone type being used
	ErrIncorrectCloneType = errors.New("GHORG_CLONE_TYPE or --clone-type must be one of org or user")

	// ErrIncorrectProtocolType indicates an unsupported protocol type being used
	ErrIncorrectProtocolType = errors.New("GHORG_CLONE_PROTOCOL or --protocol must be one of https or ssh")
)

func init() {
	initConfig()
}

func initConfig() {

	viper.SetConfigType("yaml")
	viper.AddConfigPath(GhorgDir())
	viper.SetConfigName("conf")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println(err)
			fmt.Println("Could not find $HOME/ghorg/conf.yaml file, please add one")
		} else {
			// Config file was found but another error was produced
			fmt.Println(err)
			fmt.Println("Something unexpected happened")
		}
	}

	getOrSetDefaults("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	getOrSetDefaults("GHORG_BRANCH")
	getOrSetDefaults("GHORG_CLONE_PROTOCOL")
	getOrSetDefaults("GHORG_CLONE_TYPE")
	getOrSetDefaults("GHORG_SCM_TYPE")
	getOrSetDefaults("GHORG_GITLAB_DEFAULT_NAMESPACE")
	getOrSetDefaults("GHORG_COLOR")
	getOrSetDefaults("GHORG_SKIP_ARCHIVED")
	getOrSetDefaults("GHORG_BACKUP")
	// Optionally set
	getOrSetDefaults("GHORG_GITHUB_TOKEN")
	getOrSetDefaults("GHORG_GITLAB_TOKEN")
	getOrSetDefaults("GHORG_BITBUCKET_USERNAME")
	getOrSetDefaults("GHORG_BITBUCKET_APP_PASSWORD")
	getOrSetDefaults("GHORG_SCM_BASE_URL")
	getOrSetDefaults("GHORG_PRESERVE_DIRECTORY_STRUCTURE")
}

// Load triggers the configs to load first, not sure if this is actually needed
func Load() {}

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

	// When a user does not set value in $HOME/ghorg/conf.yaml set the default values, else set env to what they have added to the file.
	if viper.GetString(envVar) == "" {
		switch envVar {
		case "GHORG_ABSOLUTE_PATH_TO_CLONE_TO":
			os.Setenv(envVar, HomeDir()+"/Desktop/")
		case "GHORG_BRANCH":
			os.Setenv(envVar, "master")
		case "GHORG_CLONE_PROTOCOL":
			os.Setenv(envVar, "https")
		case "GHORG_CLONE_TYPE":
			os.Setenv(envVar, "org")
		case "GHORG_SCM_TYPE":
			os.Setenv(envVar, "github")
		case "GHORG_GITLAB_DEFAULT_NAMESPACE":
			os.Setenv(envVar, "unset")
		case "GHORG_COLOR":
			os.Setenv(envVar, "on")
		case "GHORG_SKIP_ARCHIVED":
			os.Setenv(envVar, "false")
		case "GHORG_BACKUP":
			os.Setenv(envVar, "false")
		case "GHORG_PRESERVE_DIRECTORY_STRUCTURE":
			os.Setenv(envVar, "false")
		}
	} else {
		os.Setenv(envVar, viper.GetString(envVar))
	}
}

// GhorgIgnoreLocation returns the path of users ghorgignore
func GhorgIgnoreLocation() string {
	return GhorgDir() + "/ghorgignore"
}

// GhorgDir returns the ghorg directory path
func GhorgDir() string {
	return HomeDir() + "/ghorg"
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

	if scmType != "github" && scmType != "gitlab" && scmType != "bitbucket" {
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
