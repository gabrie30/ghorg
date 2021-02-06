package cmd

import (
	"github.com/gabrie30/ghorg/configs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

func lsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls [dir]",
		Short: "List contents of your ghorg home or ghorg directories",
		Long:  `If no dir is specified it will list contents of GHORG_PATH. When specifying a dir you can omit _ghorg`,
		Run:   lsFunc,
	}
}

func lsFunc(_ *cobra.Command, argz []string) {
	config, err := configs.Load(argz)
	if err != nil {
		colorlog.PrintError("Loading config failed")
		os.Exit(1)
	}

	if len(argz) == 0 {
		listGhorgHome(config)
	}

	if len(argz) >= 1 {
		for _, arg := range argz {
			listGhorgDir(config, arg)
		}
	}
}

func listGhorgHome(config *configs.Config) {
	files, err := ioutil.ReadDir(config.Path)
	if err != nil {
		colorlog.PrintError("No clones found. Please clone some and try again.")
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintInfo(config.Path + f.Name())
		}
	}
}

func listGhorgDir(config *configs.Config, arg string) {
	if !strings.HasSuffix(arg, "_ghorg") {
		arg = arg + "_ghorg"
	}

	arg = strings.ReplaceAll(arg, "-", "_")

	ghorgDir := config.Path + arg

	files, err := ioutil.ReadDir(ghorgDir)
	if err != nil {
		colorlog.PrintError("No clone found with that name. Please check spelling or reclone.")
	}

	for _, f := range files {
		if f.IsDir() {
			colorlog.PrintSubtleInfo(ghorgDir + "/" + f.Name())
		}
	}
}
