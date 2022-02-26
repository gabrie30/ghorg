package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Ghorg",
	Long:  `All software has versions. This is Ghorg's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v1.7.9")
	},
}
