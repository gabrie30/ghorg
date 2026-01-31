package scm

import (
	"os"
	"regexp"
	"strings"
)

func hasMatchingTopic(rpTopics []string) bool {
	envTopics := strings.Split(os.Getenv("GHORG_TOPICS"), ",")

	// If user defined a list of topics, check if any match with this repo
	if os.Getenv("GHORG_TOPICS") != "" {
		for _, rpTopic := range rpTopics {
			for _, envTopic := range envTopics {
				if rpTopic == envTopic {
					return true
				}
			}
		}
		return false
	}

	// If no user defined topics are specified, accept any topics
	return true
}

// ReplaceSSHHostname replaces the hostname in an SSH clone URL with a custom hostname.
// This allows users to leverage SSH configs with multiple host aliases.
// For example: git@gitlab.com:org/repo.git -> git@my-gitlab-alias:org/repo.git
func ReplaceSSHHostname(sshURL string, newHostname string) string {
	if newHostname == "" {
		return sshURL
	}

	// Match SSH URLs in the format git@hostname:path or git@hostname/path
	re := regexp.MustCompile(`^git@([^:/]+)([:\/])`)
	return re.ReplaceAllString(sshURL, "git@"+newHostname+"$2")
}
