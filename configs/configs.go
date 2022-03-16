// Package configs sets up the environment. First it sets a number of default envs, then looks in the $HOME/ghorg/conf.yaml to overwrite the defaults. These values will be superseded by any command line flags used
package configs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/scm"
	"github.com/gabrie30/ghorg/utils"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// ErrNoGitHubToken error message when token is not found
	ErrNoGitHubToken = errors.New("Could not find a valid github token. GHORG_GITHUB_TOKEN or (--token, -t) flag must be set. Create a personal access token, then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag, see 'GitHub Setup' in README.md")

	// ErrNoGitLabToken error message when token is not found
	ErrNoGitLabToken = errors.New("Could not find a valid gitlab token. GHORG_GITLAB_TOKEN or (--token, -t) flag must be set. Create a token from gitlab then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag, see 'GitLab Setup' in README.md")

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

// EnsureTrailingSlashOnURL takes a url and ensures a single / is appened
func EnsureTrailingSlashOnURL(s string) string {
	trailing := "/"

	if !strings.HasSuffix(s, trailing) {
		s = s + trailing
	}

	return s
}

func GetAbsolutePathToCloneTo() string {
	path := HomeDir()
	path = filepath.Join(path, "ghorg")
	return EnsureTrailingSlashOnFilePath(path)
}

// EnsureTrailingSlashOnFilePath takes a filepath and ensures a single / is appened
func EnsureTrailingSlashOnFilePath(s string) string {
	trailing := GetCorrectFilePathSeparator()

	if !strings.HasSuffix(s, trailing) {
		s = s + trailing
	}

	return s
}

// GetCorrectFilePathSeparator returns the correct trailing slash based on os
func GetCorrectFilePathSeparator() string {
	trailing := "/"

	if runtime.GOOS == "windows" {
		trailing = "\\"
	}

	return trailing
}

// GhorgIgnoreLocation returns the path of users ghorgignore
func GhorgIgnoreLocation() string {
	ignoreLocation := os.Getenv("GHORG_IGNORE_PATH")
	if ignoreLocation != "" {
		return ignoreLocation
	}

	return filepath.Join(GhorgDir(), "ghorgignore")
}

// GhorgIgnoreDetected returns true if a ghorgignore file exists.
func GhorgIgnoreDetected() bool {
	_, err := os.Stat(GhorgIgnoreLocation())
	return !os.IsNotExist(err)
}

// GhorgDir returns the ghorg directory path
func GhorgDir() string {
	if XConfigHomeSet() {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		return filepath.Join(xdg, "ghorg")
	}

	return filepath.Join(HomeDir(), ".config", "ghorg")
}

// XConfigHomeSet checks for XDG_CONFIG_HOME env set
func XConfigHomeSet() bool {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return true
	}

	return false
}

func DefaultConfFile() string {
	return filepath.Join(GhorgDir(), "conf.yaml")
}

// HomeDir finds the users home directory
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error trying to find users home directory")
	}

	return home
}

func GhorgQuiet() bool {
	return os.Getenv("GHORG_QUIET") != ""
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
		if runtime.GOOS == "windows" {
			return
		}
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

	token = os.Getenv("GHORG_GITLAB_TOKEN")

	if strings.HasPrefix(token, "glpat-") && len(token) == 26 {
		return
	}

	if isZero(token) || len(token) != 20 {
		if runtime.GOOS == "windows" {
			return
		}
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
	if isZero(os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")) && isZero(os.Getenv("GHORG_BITBUCKET_OAUTH_TOKEN")) {
		if runtime.GOOS == "windows" {
			return
		}
		cmd := `security find-internet-password -s bitbucket.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		token = strings.TrimSuffix(string(out), "\n")

		if !isZero(os.Getenv("GHORG_BITBUCKET_USERNAME")) {
			os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", token)
		} else {
			os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", token)
		}
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
		token = os.Getenv("GHORG_GITLAB_TOKEN")
		if strings.HasPrefix(token, "glpat-") {
			tokenLength = 26
		} else if len(token) > 0 {
			// gitlab admins can change token prefixes so we dont know the exact length
			tokenLength = len(token)
		} else {
			tokenLength = -1
		}
	}

	if scmProvider == "bitbucket" {
		tokenLength = 20
		if os.Getenv("GHORG_BITBUCKET_USERNAME") == "" && len(os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")) == 20 {
			return ErrNoBitbucketUsername
		}

		if isZero(os.Getenv("GHORG_BITBUCKET_USERNAME")) {
			// todo not sure how long this is so, so just make it pass for now
			tokenLength = 0
			token = ""
		} else {
			token = os.Getenv("GHORG_BITBUCKET_APP_PASSWORD")
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
