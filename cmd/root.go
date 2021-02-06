package cmd

import (
	"fmt"
	"github.com/gabrie30/ghorg/configs"
	"github.com/spf13/viper"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "ghorg",
		Short: "Ghorg is a fast way to clone multiple repos into a single directory",
		Long:  `Ghorg is a fast way to clone multiple repos into a single directory`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("For help run: ghorg clone --help")
		},
	}
	rootCmd.PersistentFlags().Bool("color", true, "GHORG_COLOR - toggles colorful output (default true)")
	rootCmd.PersistentFlags().String("config", filepath.Join(configs.GhorgDir(), "conf.yaml"), "GHORG_CONFIG Set the path of the conf.yml file (default $HOME/ghorg/conf.yaml)")

	_ = viper.BindPFlag("color", rootCmd.PersistentFlags().Lookup("color"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(lsCmd(), versionCmd(), cloneCmd())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
