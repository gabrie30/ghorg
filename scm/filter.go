package scm

import (
	"os"
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
