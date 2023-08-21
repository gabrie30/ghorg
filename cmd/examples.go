package cmd

import (
	"fmt"
	"io/ioutil"

	gtm "github.com/MichaelMure/go-term-markdown"
	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples [github, gitlab, bitbucket, gitea]",
	Short: "Documentation and examples for each SCM provider",
	Long:  `Get documentation and examples for each SCM provider in the terminal`,
	Run:   examplesFunc,
}

func examplesFunc(cmd *cobra.Command, argz []string) {
	filePath := fmt.Sprintf("examples/%s.md", argz[0])
	input := getFileContents(filePath)
	result := gtm.Render(string(input), 80, 6)
	fmt.Println(string(result))
}

func getFileContents(filepath string) []byte {

	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		colorlog.PrintErrorAndExit("Only supported SCM providers are available for examples, please use one of the following: github, gitlab, bitbucket, or gitea")
	}

	return contents

}
