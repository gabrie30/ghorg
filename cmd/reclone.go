package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
	Long:  `Allows you to set preconfigured clone commands for handling multiple users/orgs at once. See README.md at https://github.com/gabrie30/ghorg for setup information.`,
	Run:   reCloneFunc,
}

type ReClone struct {
	Cmd string `yaml:"cmd"`
}

func reCloneFunc(cmd *cobra.Command, argz []string) {
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
	if len(argz) == 0 {
		for key, value := range reCloneMap {
			fmt.Printf("%v: %v\n", key, value)
		}
	} else {
		for _, arg := range argz {
			key, value := reCloneMap[arg]
			fmt.Printf("%v: %v\n", key, value)
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

	stdout, err := ghorgClone.StdoutPipe()
	if err != nil {
		e := fmt.Sprintf("ERROR: Problem with piping to stdout, err: %v", err)
		colorlog.PrintErrorAndExit(e)
	}

	ghorgClone.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	err = ghorgClone.Wait()
	if err != nil {
		e := fmt.Sprintf("ERROR: Running ghorg clone cmd: %v, err: %v", rc.Cmd, err)
		colorlog.PrintErrorAndExit(e)
	}
}
