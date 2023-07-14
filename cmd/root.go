/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "glfast",
	Short:   "GitLab project scaffold initialization.",
	Long:    `Specify the project scaffold template, initialize the GitLab project configuration and pipeline.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// rootCmd.PersistentFlags().String("baseurl", "https://gitlab.com", "base URL for the GitLab instance")
	// rootCmd.PersistentFlags().String("token", "", "token for GitLab")

	// viper.BindPFlag("gitlab.baseurl", rootCmd.PersistentFlags().Lookup("baseurl"))
	// viper.BindPFlag("gitlab.token", rootCmd.PersistentFlags().Lookup("token"))

}
