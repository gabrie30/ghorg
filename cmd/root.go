package cmd

import (
	"fmt"
	"os"

	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghorg",
	Short: "Ghorg is a fast way to clone multiple repos into a single directory",
	Long:  `Ghorg is a fast way to clone multiple repos into a single directory`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		configs.Load(cmd.Flag("config").Value.String())
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("For help run: ghorg clone --help")
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
