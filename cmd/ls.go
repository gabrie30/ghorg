package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "List contents of your ghorg home or ghorg directories",
	Long:  `If no dir is specified it will list contents of GHORG_ABSOLUTE_PATH_TO_CLONE_TO`,
	Run:   lsFunc,
}

func lsFunc(cmd *cobra.Command, argz []string) {
	if len(argz) == 0 {
		listGhorgHome()
	}

	if len(argz) >= 1 {
		for _, arg := range argz {
			listGhorgDir(arg)
		}
	}

}

func listGhorgHome() {
	path := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		colorlog.PrintError("No clones found. Please clone some and try again.")
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintInfo(path + f.Name())
		}
	}
}

func listGhorgDir(arg string) {

	arg = strings.ReplaceAll(arg, "-", "_")

	path := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + arg

	files, err := ioutil.ReadDir(path)
	if err != nil {
		colorlog.PrintError("No clone found with that name. Please check spelling or reclone.")
	}

	for _, f := range files {
		if f.IsDir() {
			str := filepath.Join(path, f.Name())
			colorlog.PrintSubtleInfo(str)
		}
	}
}
