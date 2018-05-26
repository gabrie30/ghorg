package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gabrie30/ghorg/cmd"
	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/config"
	"github.com/joho/godotenv"
	homedir "github.com/mitchellh/go-homedir"
)

func init() {
	if len(os.Args) <= 1 {
		log.Fatal("You must provide an org to clone from")
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error trying to find users home directory")
	}

	err = godotenv.Load(home + "/.ghorg")
	if err != nil {
		fmt.Println()
		colorlog.PrintSubtleInfo("Could not find a $HOME/.ghorg proceeding with defaults")
	}

	config.GitHubToken = os.Getenv("GHORG_GITHUB_TOKEN")
	config.AbsolutePathToCloneTo = os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	config.GhorgBranch = os.Getenv("GHORG_BRANCH")

	if config.GhorgBranch == "" {
		config.GhorgBranch = "master"
	}

	if config.AbsolutePathToCloneTo == "" {
		config.AbsolutePathToCloneTo = home + "/Desktop/"
	}

	withTrailingSlash := ensureTrailingSlash(config.AbsolutePathToCloneTo)
	config.AbsolutePathToCloneTo = withTrailingSlash
}

func ensureTrailingSlash(path string) string {
	if string(path[len(path)-1]) == "/" {
		return path
	}

	return path + "/"
}

func asciiTime() {
	colorlog.PrintInfo(
		`
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
 |T|I|M|E| |T|O| |G|H|O|R|G|
 +-+-+-+-+ +-+-+ +-+-+-+-+-+
`)
}

func main() {
	asciiTime()
	cmd.CloneAllReposByOrg()
}
