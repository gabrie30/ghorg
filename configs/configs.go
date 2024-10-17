// Package configs sets up the environment. First it sets a number of default envs, then looks in the $HOME/.config/ghorg/conf.yaml to overwrite the defaults. These values will be superseded by any command line flags used
package configs

import (
	"errors"
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

	// ErrNoGiteaToken error message when token is not found
	ErrNoGiteaToken = errors.New("Could not find a valid gitea token. GHORG_GITEA_TOKEN or (--token, -t) flag must be set. Create a token from gitea then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag, see 'Gitea Setup' in README.md")

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

	// ErrIncorrectGithubUserOptionValue indicates an incorrectly set GHORG_GITHUB_USER_OPTION value
	ErrIncorrectGithubUserOptionValue = errors.New("GHORG_GITHUB_USER_OPTION or --github-user-option must be one of 'owner', 'member', or 'all' and is only available to be used when GHORG_CLONE_TYPE: user or --clone-type=user is set")
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
	trailing := string(os.PathSeparator)

	if !strings.HasSuffix(s, trailing) {
		s = s + trailing
	}

	return s
}

// GhorgIgnoreLocation returns the path of users ghorgignore
func GhorgIgnoreLocation() string {
	ignoreLocation := os.Getenv("GHORG_IGNORE_PATH")
	if ignoreLocation != "" {
		return ignoreLocation
	}

	return filepath.Join(GhorgConfDir(), "ghorgignore")
}

// GhorgReCloneLocation returns the path of users ghorgignore
func GhorgReCloneLocation() string {
	recloneConfLocation := os.Getenv("GHORG_RECLONE_PATH")
	if recloneConfLocation != "" {
		return recloneConfLocation
	}

	return filepath.Join(GhorgConfDir(), "reclone.yaml")
}

// GhorgIgnoreDetected returns true if a ghorgignore file exists.
func GhorgIgnoreDetected() bool {
	_, err := os.Stat(GhorgIgnoreLocation())
	return !os.IsNotExist(err)
}

// GhorgReCloneDetected returns true if a reclone.yaml file exists.
func GhorgReCloneDetected() bool {
	_, err := os.Stat(GhorgReCloneLocation())
	return !os.IsNotExist(err)
}

// GhorgConfDir returns the ghorg directory path
func GhorgConfDir() string {
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
	return filepath.Join(GhorgConfDir(), "conf.yaml")
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

func IsFilePath(path string) bool {
	pathValue, err := homedir.Expand(path)
	if err != nil {
		log.Fatal("Error while expanding tilde to user home directory")
	}
	info, err := os.Stat(pathValue)
	if err != nil {
		return false
	}
	// Check if it's a regular file (not a directory or a symbolic link)
	if !info.IsDir() && (info.Mode()&os.ModeType == 0) {
		return true
	}
	return false
}

func GetTokenFromFile(path string) string {
	expandedPath, _ := homedir.Expand(path)
	fileContents, err := os.ReadFile(expandedPath)
	if err != nil {
		log.Fatal("Error while reading file")
	}
	return strings.TrimSpace(string(fileContents))
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
	case "gitea":
		getOrSetGiteaToken()
	}
}

func getOrSetGitHubToken() {
	var token = os.Getenv("GHORG_GITHUB_TOKEN")
	if IsFilePath(token) {
		os.Setenv("GHORG_GITHUB_TOKEN", GetTokenFromFile(token))
	}

	if isZero(token) {
		if runtime.GOOS == "windows" {
			return
		}
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, _ := exec.Command("bash", "-c", cmd).Output()

		token = strings.TrimSuffix(string(out), "\n")

		os.Setenv("GHORG_GITHUB_TOKEN", token)
	}
}

func getOrSetGitLabToken() {
	token := os.Getenv("GHORG_GITLAB_TOKEN")

	if IsFilePath(token) {
		os.Setenv("GHORG_GITLAB_TOKEN", GetTokenFromFile(token))
	}

	if isZero(token) {
		if runtime.GOOS == "windows" {
			return
		}
		cmd := `security find-internet-password -s gitlab.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, _ := exec.Command("bash", "-c", cmd).Output()

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
		out, _ := exec.Command("bash", "-c", cmd).Output()

		token = strings.TrimSuffix(string(out), "\n")

		if !isZero(os.Getenv("GHORG_BITBUCKET_USERNAME")) {
			os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", token)
		} else {
			os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", token)
		}
	}
}

func getOrSetGiteaToken() {
	token := os.Getenv("GHORG_GITEA_TOKEN")

	if IsFilePath(token) {
		os.Setenv("GHORG_GITEA_TOKEN", GetTokenFromFile(token))
	}

	if isZero(token) {
		if runtime.GOOS == "windows" {
			return
		}
		os.Setenv("GHORG_GITEA_TOKEN", token)
	}
}

// VerifyTokenSet checks to make sure env is set for the correct scm provider
func VerifyTokenSet() error {

	if os.Getenv("GHORG_NO_TOKEN") == "true" {
		return nil
	}

	scmProvider := os.Getenv("GHORG_SCM_TYPE")

	if scmProvider == "github" && os.Getenv("GHORG_GITHUB_TOKEN") == "" {
		if os.Getenv("GHORG_GITHUB_APP_PEM_PATH") != "" {
			return nil
		}
		return ErrNoGitHubToken
	}

	if scmProvider == "gitlab" && os.Getenv("GHORG_GITLAB_TOKEN") == "" {
		return ErrNoGitLabToken
	}

	if scmProvider == "gitea" && os.Getenv("GHORG_GITEA_TOKEN") == "" {
		return ErrNoGiteaToken
	}

	if scmProvider == "bitbucket" {
		if os.Getenv("GHORG_BITBUCKET_OAUTH_TOKEN") == "" {

			if os.Getenv("GHORG_BITBUCKET_USERNAME") == "" {
				return ErrNoBitbucketUsername
			}

			if os.Getenv("GHORG_BITBUCKET_APP_PASSWORD") == "" {
				return ErrNoBitbucketAppPassword
			}
		}
	}

	return nil
}

func GetCloudScmTypeHostnames() string {
	switch os.Getenv("GHORG_SCM_TYPE") {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "gitea":
		return "gitea.com"
	case "bitbucket":
		return "bitbucket.com"
	default:
		colorlog.PrintErrorAndExit("Unsupported GHORG_SCM_TYPE")
		return ""
	}
}

// VerifyConfigsSetCorrectly makes sure flags are set to appropriate values
func VerifyConfigsSetCorrectly() error {
	scmType := os.Getenv("GHORG_SCM_TYPE")
	cloneType := os.Getenv("GHORG_CLONE_TYPE")
	protocol := os.Getenv("GHORG_CLONE_PROTOCOL")
	githubUserOption := os.Getenv("GHORG_GITHUB_USER_OPTION")

	if !utils.IsStringInSlice(scmType, scm.SupportedClients()) {
		return ErrIncorrectScmType
	}

	if cloneType != "user" && cloneType != "org" {
		return ErrIncorrectCloneType
	}

	if scmType == "github" && cloneType == "user" {
		if githubUserOption != "owner" && githubUserOption != "all" && githubUserOption != "member" {
			return ErrIncorrectGithubUserOptionValue
		}
	}

	if protocol != "ssh" && protocol != "https" {
		return ErrIncorrectProtocolType
	}

	return nil
}
