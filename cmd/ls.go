package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "List contents of your ghorg home or ghorg directories",
	Long:  `If no dir is specified it will list contents of GHORG_ABSOLUTE_PATH_TO_CLONE_TO`,
	Run:   lsFunc,
}

func lsFunc(cmd *cobra.Command, argz []string) {

	if cmd.Flags().Changed("color") {
		colorToggle := cmd.Flag("color").Value.String()
		if colorToggle == "on" {
			os.Setenv("GHORG_COLOR", colorToggle)
		} else {
			os.Setenv("GHORG_COLOR", "off")
		}

	}

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
	ghorgDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	files, err := ioutil.ReadDir(ghorgDir)
	if err != nil {
		colorlog.PrintError("No clones found. Please clone some and try again.")
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintInfo(ghorgDir + f.Name())
		}
	}
}

func listGhorgDir(arg string) {

	arg = strings.ReplaceAll(arg, "-", "_")

	ghorgDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + arg

	files, err := ioutil.ReadDir(ghorgDir)
	if err != nil {
		colorlog.PrintError("No clone found with that name. Please check spelling or reclone.")
	}

	for _, f := range files {
		if f.IsDir() {
			str := filepath.Join(ghorgDir, configs.GetCorrectFilePathSeparator(), f.Name())
			colorlog.PrintSubtleInfo(str)
		}
	}
}
