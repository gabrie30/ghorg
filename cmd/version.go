package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Ghorg",
	Long:  `All software has versions. This is Ghorg's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("1.3.2")
	},
}
