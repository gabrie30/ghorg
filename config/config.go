// Package config sets environment variables to vars for simplified sharing of environment variables between packages
package config

var (
	// GitHubToken used to auth to github, either comes from keychain locally or from the .env
	GitHubToken string
	// AbsolutePathToCloneTo Path to which ghorg will create a new folder to place all cloned repos
	AbsolutePathToCloneTo string
	// GhorgBranch branch that ghorg will checkout
	GhorgBranch string
)
