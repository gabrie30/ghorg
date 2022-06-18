package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const ghorgVersion = "v1.7.17"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Ghorg",
	Long:  `All software has versions. This is Ghorg's`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVersion()
	},
}

func PrintVersion() {
	fmt.Println(ghorgVersion)
}

func GetVersion() string {
	return ghorgVersion
}
