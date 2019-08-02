package configs

import (
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

func init() {
	initConfig()
}

func initConfig() {

	viper.SetConfigType("yaml")
	viper.AddConfigPath(ghorgDir())
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
		}
	} else {
		os.Setenv(envVar, viper.GetString(envVar))
	}
}

func ghorgDir() string {
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
	if isZero(os.Getenv("GHORG_GITLAB_TOKEN")) || len(os.Getenv("GHORG_GITLAB_TOKEN")) != 40 {
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
func VerifyTokenSet() {
	var tokenLength int
	if os.Getenv("GHORG_CLONE_PROTOCOL") != "https" {
		return
	}

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
			colorlog.PrintError("GHORG_BITBUCKET_USERNAME or --bitbucket_username must be set to clone repos from bitbucket, see BitBucket Setup in Readme")
			os.Exit(1)
		}
	}

	if len(token) != tokenLength {
		colorlog.PrintError("Could not find a " + scmProvider + " token in keychain. You should create a personal access token from " + scmProvider + " , then set the correct in your $HOME/ghorg/conf.yaml...or swtich to cloning via SSH also done by updating your $HOME/ghorg/conf.yaml. Or read the troubleshooting section of Readme.md https://github.com/gabrie30/ghorg to store your token in your osx keychain. Or set manually with -t flag")
		os.Exit(1)
	}
}

// VerifyConfigsSetCorrectly makes sure flags are set to appropriate values
func VerifyConfigsSetCorrectly() {
	scmType := os.Getenv("GHORG_SCM_TYPE")
	cloneType := os.Getenv("GHORG_CLONE_TYPE")
	protocol := os.Getenv("GHORG_CLONE_PROTOCOL")

	if scmType != "github" && scmType != "gitlab" && scmType != "bitbucket" {
		colorlog.PrintError("GHORG_SCM_TYPE or --scm must be one of github, gitlab, or bitbucket")
		os.Exit(1)
	}

	if cloneType != "user" && cloneType != "org" {
		colorlog.PrintError("GHORG_CLONE_TYPE or --clone-type must be one of org or user")
		os.Exit(1)
	}

	if protocol != "ssh" && protocol != "https" {
		colorlog.PrintError("GHORG_CLONE_PROTOCOL or --protocol must be one of https or ssh")
		os.Exit(1)
	}
}
