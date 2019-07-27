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

// VerifyTokenSet checks to make sure env is set for the correct scm provider
func VerifyTokenSet() {

	if os.Getenv("GHORG_CLONE_PROTOCOL") != "https" {
		return
	}

	var token string
	scmProvider := os.Getenv("GHORG_SCM_TYPE")

	if scmProvider == "github" {
		token = os.Getenv("GHORG_GITHUB_TOKEN")
	}

	if scmProvider == "gitlab" {
		token = os.Getenv("GHORG_GITLAB_TOKEN")
	}

	if len(token) != 40 {
		colorlog.PrintError("Could not find a " + scmProvider + " token in keychain. You should create a personal access token from " + scmProvider + " , then set the correct in your $HOME/ghorg/conf.yaml...or swtich to cloning via SSH also done by updating your $HOME/ghorg/conf.yaml. Or read the troubleshooting section of Readme.md https://github.com/gabrie30/ghorg to store your token in your osx keychain. Or set manually with -t flag")
	}
}
