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
	Long:  `Allows you to set preconfigured clone commands for handling multiple users/orgs at once. See https://github.com/gabrie30/ghorg#reclone-command for setup and additional information.`,
	Run:   reCloneFunc,
}

type ReClone struct {
	Cmd string `yaml:"cmd"`
}

func isVerboseReClone() bool {
	return os.Getenv("GHORG_RECLONE_VERBOSE") == "true"
}

func isQuietReClone() bool {
	return os.Getenv("GHORG_RECLONE_QUIET") == "true"
}

func reCloneFunc(cmd *cobra.Command, argz []string) {

	if cmd.Flags().Changed("reclone-path") {
		path := cmd.Flag("reclone-path").Value.String()
		os.Setenv("GHORG_RECLONE_PATH", path)
	}

	if cmd.Flags().Changed("verbose") {
		os.Setenv("GHORG_RECLONE_VERBOSE", "true")
	}

	if cmd.Flags().Changed("quiet") {
		os.Setenv("GHORG_RECLONE_QUIET", "true")
	}

	path := configs.GhorgReCloneLocation()
	yamlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: parsing reclone.yaml, error: %v", err))
	}

	mapOfReClones := make(map[string]ReClone)

	err = yaml.Unmarshal(yamlBytes, &mapOfReClones)
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: unmarshaling reclone.yaml, error:%v", err))
	}

	if !isVerboseReClone() && !isQuietReClone() {
		asciiTime()
	}

	if len(argz) == 0 {
		for _, key := range mapOfReClones {
			runReClone(key)
		}
	} else {
		for _, arg := range argz {
			if _, ok := mapOfReClones[arg]; !ok {
				colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: The key %v was not found in reclone.yaml", arg))
			}
			runReClone(mapOfReClones[arg])
		}
	}

	printFinalOutput(argz, mapOfReClones)
}

func printFinalOutput(argz []string, reCloneMap map[string]ReClone) {
	fmt.Println("")
	colorlog.PrintSuccess("Completed! The following reclones were ran successfully...")
	if len(argz) == 0 {
		for key, _ := range reCloneMap {
			colorlog.PrintSuccess(fmt.Sprintf("  * %v", key))
		}
	} else {
		for _, arg := range argz {
			colorlog.PrintSuccess(fmt.Sprintf("  * %v", arg))
		}
	}
}

func sanitizeCmd(cmd string) string {
	if strings.Contains(cmd, "-t=") {
		splitCmd := strings.Split(cmd, "-t=")
		firstHalf := splitCmd[0]
		secondHalf := strings.Split(splitCmd[1], " ")[1:]
		return firstHalf + "-t=XXXXXXX " + strings.Join(secondHalf, " ")
	}
	if strings.Contains(cmd, "-t ") {
		splitCmd := strings.Split(cmd, "-t ")
		firstHalf := splitCmd[0]
		secondHalf := strings.Split(splitCmd[1], " ")[1:]
		return firstHalf + "-t XXXXXXX " + strings.Join(secondHalf, " ")
	}
	if strings.Contains(cmd, "--token=") {
		splitCmd := strings.Split(cmd, "--token=")
		firstHalf := splitCmd[0]
		secondHalf := strings.Split(splitCmd[1], " ")[1:]
		return firstHalf + "--token=XXXXXXX " + strings.Join(secondHalf, " ")
	}
	if strings.Contains(cmd, "--token ") {
		splitCmd := strings.Split(cmd, "--token ")
		firstHalf := splitCmd[0]
		secondHalf := strings.Split(splitCmd[1], " ")[1:]
		return firstHalf + "--token XXXXXXX " + strings.Join(secondHalf, " ")
	}
	return cmd
}

func runReClone(rc ReClone) {
	// make sure command starts with ghorg clone
	splitCommand := strings.Split(rc.Cmd, " ")
	ghorg, clone, remainingCommand := splitCommand[0], splitCommand[1], splitCommand[1:]

	if ghorg != "ghorg" || clone != "clone" {
		colorlog.PrintErrorAndExit("ERROR: Only ghorg clone commands are permitted in your reclone.yaml")
	}

	safeToLogCmd := sanitizeCmd(strings.Clone(rc.Cmd))

	if !isQuietReClone() {
		colorlog.PrintInfo(fmt.Sprintf("$ %v", safeToLogCmd))
	}

	ghorgClone := exec.Command("ghorg", remainingCommand...)

	if os.Getenv("GHORG_CONFIG") == "none" {
		os.Setenv("GHORG_CONFIG", "")
	}

	os.Setenv("GHORG_RECLONE_RUNNING", "true")
	defer os.Setenv("GHORG_RECLONE_RUNNING", "false")

	// have to unset all ghorg envs because root command will set them on initialization of ghorg cmd
	for _, e := range os.Environ() {
		keyValue := strings.SplitN(e, "=", 2)
		env := keyValue[0]
		ghorgEnv := strings.HasPrefix(env, "GHORG_")

		// skip global flags and reclone flags which are set in the conf.yaml
		if env == "GHORG_COLOR" || env == "GHORG_CONFIG" || env == "GHORG_RECLONE_VERBOSE" || env == "GHORG_RECLONE_QUIET" || env == "GHORG_RECLONE_PATH" || env == "GHORG_RECLONE_RUNNING" {
			continue
		}
		if ghorgEnv {
			os.Unsetenv(env)
		}
	}

	stdout, err := ghorgClone.StdoutPipe()
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: Problem with piping to stdout, err: %v", err))
	}

	err = ghorgClone.Start()

	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: Starting ghorg clone cmd: %v, err: %v", safeToLogCmd, err))
	}

	if isVerboseReClone() && !isQuietReClone() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}

	err = ghorgClone.Wait()
	if err != nil {
		colorlog.PrintErrorAndExit(fmt.Sprintf("ERROR: Running ghorg clone cmd: %v, err: %v", safeToLogCmd, err))
	}
}
