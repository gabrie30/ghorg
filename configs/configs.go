// Package configs sets up the environment. First it sets a number of default envs, then looks in the $HOME/ghorg/conf.yaml to overwrite the defaults. These values will be superseded by any command line flags used
package configs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/viper"
)

var (
	// ErrNoGitHubToken error message when token is not found
	ErrNoGitHubToken = errors.New("Could not find a valid github token. GHORG_TOKEN or (--token, -t) flag must be set. Create a personal access token, then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoGitLabToken error message when token is not found
	ErrNoGitLabToken = errors.New("Could not find a valid gitlab token. GHORG_TOKEN or (--token, -t) flag must be set. Create a token from gitlab then set it in your $HOME/.config/ghorg/conf.yaml or use the (--token, -t) flag...For best results read the troubleshooting section of README.md https://github.com/gabrie30/ghorg to properly store your token in the osx keychain")

	// ErrNoBitbucketUsername error message when no username found
	ErrNoBitbucketUsername = errors.New("Could not find bitbucket username. GHORG_BITBUCKET_USERNAME or (--bitbucket-username) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrNoBitbucketAppPassword error message when no app password found
	ErrNoBitbucketAppPassword = errors.New("Could not find a valid bitbucket app password. GHORG_TOKEN or (--token, -t) must be set to clone repos from bitbucket, see 'BitBucket Setup' in README.md")

	// ErrIncorrectCloneType indicates an unsupported clone type being used
	ErrIncorrectCloneType = errors.New("GHORG_CLONE_TYPE or --clone-type must be one of org or user")

	// ErrIncorrectProtocolType indicates an unsupported protocol type being used
	ErrIncorrectProtocolType = errors.New("GHORG_PROTOCOL or --protocol must be one of https or ssh")
)

type Config struct {
	Token                      string   `mapstructure:"token"`
	PreserveDirectoryStructure bool     `mapstructure:"preserve-dir"`
	BitbucketUsername          string   `mapstructure:"bitbucket-username"`
	ScmBaseUrl                 string   `mapstructure:"base-url"`
	ScmType                    string   `mapstructure:"scm"`
	CloneProtocol              string   `mapstructure:"protocol"`
	Path                       string   `mapstructure:"path"`
	Branch                     string   `mapstructure:"branch"`
	CloneType                  string   `mapstructure:"clone-type"`
	Color                      bool     `mapstructure:"color"`
	Topics                     []string `mapstructure:"topics"`
	SkipArchived               bool     `mapstructure:"skip-archived"`
	SkipForks                  bool     `mapstructure:"skip-forks"`
	Backup                     bool     `mapstructure:"backup"`
	Concurrency                int      `mapstructure:"concurrency"`
	MatchPrefix                string   `mapstructure:"match-prefix"`
}

// Load triggers the configs to load first
func Load(argz []string) (*Config, error) {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetEnvPrefix("GHORG")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()


	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			colorlog.PrintError(fmt.Sprintf("Error reading configuration file, err: %s", err))
		}
	}

	config := Config{}
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	config.parseToken()
	config.ScmBaseUrl = ensureTrailingSlash(config.ScmBaseUrl)
	config.parseOutputDir(argz)
	config.parseScmType()
	colorlog.UseColor = config.Color

	err = config.VerifyToken()
	if err != nil {
		return nil, err
	}

	err = config.VerifyClone()
	if err != nil {
		return nil, err
	}

	return &config, err
}

// ensureTrailingSlash takes a string and ensures a single / is appened
func ensureTrailingSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		s = s + "/"
	}

	return s
}

func (config *Config) parseScmType() {
	if len(config.ScmType) == 0 {
		colorlog.PrintError("GHORG_SCM_TYPE not set")
		os.Exit(1)
	}
	config.ScmType = strings.ToLower(config.ScmType)
}

// parseToken will set token based on scm
func (config *Config) parseToken() {
	switch config.ScmType {
	case "github":
		config.getOrSetGitHubToken()
	case "gitlab":
		config.getOrSetGitLabToken()
	case "bitbucket":
		config.getOrSetBitBucketToken()
	}
}

func (config *Config) getOrSetGitHubToken() {
	if len(config.Token) != 40 {
		cmd := `security find-internet-password -s github.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		config.Token = strings.TrimSuffix(string(out), "\n")
	}
}

func (config *Config) getOrSetGitLabToken() {
	if len(config.Token) != 20 {
		cmd := `security find-internet-password -s gitlab.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		config.Token = strings.TrimSuffix(string(out), "\n")
	}
}

func (config *Config) getOrSetBitBucketToken() {
	if len(config.Token) != 20 {
		cmd := `security find-internet-password -s bitbucket.com | grep "acct" | awk -F\" '{ print $4 }'`
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			colorlog.PrintError(fmt.Sprintf("Failed to execute command: %s", cmd))
		}

		config.Token = strings.TrimSuffix(string(out), "\n")
	}
}

// VerifyToken checks to make sure env is set for the correct scm provider
func (config *Config) VerifyToken() error {
	switch config.ScmType {
	case "github":
		if len(config.Token) != 40 {
			return ErrNoGitHubToken
		}
	case "gitlab":
		if len(config.Token) != 20 {
			return ErrNoGitLabToken
		}
	case "bitbucket":
		if len(config.Token) != 20 {
			return ErrNoBitbucketAppPassword
		}
		if config.BitbucketUsername == "" {
			return ErrNoBitbucketUsername
		}
	}
	return nil
}

// VerifyClone makes sure flags are set to appropriate values
func (config *Config) VerifyClone() error {
	if config.CloneType != "user" && config.CloneType != "org" {
		return ErrIncorrectCloneType
	}

	if config.CloneProtocol != "ssh" && config.CloneProtocol != "https" {
		return ErrIncorrectProtocolType
	}

	return nil
}

// Print shows the user what is set before cloning
func (config *Config) Print() {
	colorlog.PrintInfo("*************************************")
	colorlog.PrintInfo("* SCM           : " + config.ScmType)
	colorlog.PrintInfo("* Type          : " + config.CloneType)
	colorlog.PrintInfo("* Protocol      : " + config.CloneProtocol)
	colorlog.PrintInfo("* Location      : " + config.Path)
	colorlog.PrintInfo(fmt.Sprintf("* Concurrency   : %d", config.Concurrency))
	colorlog.PrintInfo("* Branch        : " + config.getGhorgBranch())
	colorlog.PrintInfo("* Base URL      : " + config.ScmBaseUrl)
	colorlog.PrintInfo(fmt.Sprintf("* Skip Archived : %t", config.SkipArchived))
	colorlog.PrintInfo(fmt.Sprintf("* Skip Forks    : %t", config.SkipForks))
	colorlog.PrintInfo(fmt.Sprintf("* Backup        : %t", config.Backup))
	if ghorgIgnoreDetected() {
		colorlog.PrintInfo("* Ghorgignore   : true")
	}
	colorlog.PrintInfo("* Output Dir    : " + config.Path)

	colorlog.PrintInfo("*************************************")
	fmt.Println("")
}

func (config *Config) getGhorgBranch() string {
	if config.Branch == "" {
		return "default branch"
	}

	return config.Branch
}

func (config *Config) parseOutputDir(argz []string) {
	if config.Path != "" {
		return
	}

	config.Path = HomeDir() + "/Desktop/ghorg/" + strings.ToLower(strings.ReplaceAll(argz[0], "-", "_"))
}
