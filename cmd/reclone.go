package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var reCloneCmd = &cobra.Command{
	Use:   "reclone",
	Short: "Reruns one, multiple, or all preconfigured clones from configuration set in $HOME/.config/ghorg/reclone.yaml",
	Long:  `Allows you to set preconfigured clone commands for handling multiple users/orgs at once. See README.md at https://github.com/gabrie30/ghorg#cloning-multiple-usersorgsconfigurations for setup information.`,
	Run:   reCloneFunc,
}

type ReClone struct {
	Cmd string `yaml:"cmd"`
}

func reCloneFunc(cmd *cobra.Command, argz []string) {

	if cmd.Flags().Changed("reclone-path") {
		path := cmd.Flag("reclone-path").Value.String()
		os.Setenv("GHORG_RECLONE_PATH", path)
	}

	path := configs.GhorgReCloneLocation()
	yamlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		e := fmt.Sprintf("ERROR: parsing reclone.yaml, error: %v", err)
		colorlog.PrintErrorAndExit(e)
	}

	mapOfReClones := make(map[string]ReClone)

	err = yaml.Unmarshal(yamlBytes, &mapOfReClones)
	if err != nil {
		e := fmt.Sprintf("ERROR: unmarshaling reclone.yaml, error:%v", err)
		colorlog.PrintErrorAndExit(e)
	}

	if len(argz) == 0 {
		for _, key := range mapOfReClones {
			runReClone(key)
		}
	} else {
		for _, arg := range argz {
			if _, ok := mapOfReClones[arg]; !ok {
				e := fmt.Sprintf("ERROR: The key %v was not found in reclone.yaml", arg)
				colorlog.PrintErrorAndExit(e)
			}
			runReClone(mapOfReClones[arg])
		}
	}

	printFinalOutput(argz, mapOfReClones)
}

func printFinalOutput(argz []string, reCloneMap map[string]ReClone) {
	fmt.Println("")
	fmt.Println("Completed! The following reclones were ran...")
	if len(argz) == 0 {
		for key, _ := range reCloneMap {
			fmt.Printf("  * %v\n", key)
		}
	} else {
		for _, arg := range argz {
			fmt.Printf("  * %v\n", arg)
		}
	}
}

func runReClone(rc ReClone) {
	// make sure command starts with ghorg clone
	splitCommand := strings.Split(rc.Cmd, " ")
	ghorg, clone, remainingCommand := splitCommand[0], splitCommand[1], splitCommand[1:]

	if ghorg != "ghorg" || clone != "clone" {
		colorlog.PrintErrorAndExit("ERROR: Only ghorg clone commands are permitted in your reclone.yaml")
	}
	ghorgClone := exec.Command("ghorg", remainingCommand...)

	// have to unset all ghorg envs because root command will set them on initialization of ghorg cmd
	for _, e := range os.Environ() {
		keyValue := strings.SplitN(e, "=", 2)
		ghorgEnv := strings.HasPrefix(keyValue[0], "GHORG_")
		if ghorgEnv {
			os.Unsetenv(keyValue[0])
		}
	}

	stdout, err := ghorgClone.StdoutPipe()
	if err != nil {
		fmt.Printf("ERROR: Problem with piping to stdout, err: %v", err)
	}

	err = ghorgClone.Start()

	if err != nil {
		fmt.Printf("ERROR: Starting ghorg clone cmd: %v, err: %v", rc.Cmd, err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	err = ghorgClone.Wait()
	if err != nil {
		fmt.Printf("ERROR: Running ghorg clone cmd: %v, err: %v", rc.Cmd, err)
	}
}
