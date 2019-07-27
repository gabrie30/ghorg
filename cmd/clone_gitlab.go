package cmd

import (
	"os"

	"github.com/gabrie30/ghorg/colorlog"
)

func getGitLabOrgCloneUrls() ([]string, error) {
	colorlog.PrintError("Gitlab org clone not available yet")
	os.Exit(1)
	// todo
	cloneUrls := []string{}
	return cloneUrls, nil
}

func getGitLabUserCloneUrls() ([]string, error) {
	colorlog.PrintError("Gitlab user clone not available yet")
	os.Exit(1)
	// todo
	cloneUrls := []string{}
	return cloneUrls, nil
}
