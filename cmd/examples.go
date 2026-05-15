package cmd

import (
	"embed"
	"fmt"
	"strings"

	gtm "github.com/Klaus-Tockloth/go-term-markdown"
	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var (
	//go:embed examples-copy/*
	examples embed.FS
)

// availableExamples lists the documentation topics available via the
// `ghorg examples` command. Each entry corresponds to a markdown file in
// cmd/examples-copy/ (mirrored from the top-level examples/ directory).
var availableExamples = []string{
	"github",
	"gitlab",
	"bitbucket",
	"gitea",
	"sourcehut",
	"reclone-server",
	"reclone-cron",
}

var examplesCmd = &cobra.Command{
	Use:   fmt.Sprintf("examples [%s]", strings.Join(availableExamples, ", ")),
	Short: "Documentation and examples for ghorg features and SCM providers",
	Long:  `Get documentation and examples for ghorg features and SCM providers in the terminal`,
	Run:   examplesFunc,
}

func examplesFunc(cmd *cobra.Command, argz []string) {
	if len(argz) == 0 {
		colorlog.PrintErrorAndExit("Please additionally provide a topic, one of: " + strings.Join(availableExamples, ", "))
	}

	// TODO: fix the examples-copy directory mess
	filePath := fmt.Sprintf("examples-copy/%s.md", argz[0])
	input := getFileContents(filePath)
	result := gtm.Render(string(input), 80, 6)
	fmt.Println(string(result))
}

func getFileContents(filepath string) []byte {

	contents, err := examples.ReadFile(filepath)
	if err != nil {
		colorlog.PrintErrorAndExit("Unknown examples topic, please use one of the following: " + strings.Join(availableExamples, ", "))
	}

	return contents

}
