// Package cmd is for use with the Cobra CLI tool to develop a CLI interface for biji notes.
// it will include most if not all funcitonality of main GUI application.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "biji",
	// Short: "biji is a note taking application designed for both TUI and GUI environments.",
	// Long: `biji is a not taking application designed for both TUI and GUI environments.
	// 	It's being used to learn Go lang for both scripting and normal  app development;
	// 	It will eventually grow into having Cloud Sync and storage and an Android application.`,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.biji.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(newNote)
}
