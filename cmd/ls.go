package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "List contents of your ghorg home or ghorg directories",
	Long:  `If no dir is specified it will list contents of GHORG_ABSOLUTE_PATH_TO_CLONE_TO. When specifying a dir you can omit _ghorg`,
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
		log.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintInfo(ghorgDir + f.Name())
		}
	}
}

func listGhorgDir(arg string) {

	if !strings.HasSuffix(arg, "_ghorg") {
		arg = arg + "_ghorg"
	}

	ghorgDir := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + arg

	files, err := ioutil.ReadDir(ghorgDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintInfo(ghorgDir + "/" + f.Name())
		}
	}
}
