// Package cmd is for use with the Cobra CLI tool to develop a CLI interface for biji notes.
// it will include most if not all funcitonality of main GUI application.
package cmd

import (
	"log"
	"os"

	"codeberg.org/dallas1295/biji/local"
	"codeberg.org/dallas1295/biji/tui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "biji",
	Short: "biji is a note taking application designed for both TUI and GUI environments.",
	Long: `biji is a not taking application designed for both TUI and GUI environments.
		It's being used to learn Go lang for both scripting and normal app development;`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// It takes local.Store as a param for use by commands
func Execute(s *local.Store) {
	rootCmd.AddCommand(newNote(s))
	rootCmd.AddCommand(deleteNote(s))
	rootCmd.AddCommand(updateNoteName(s))
	rootCmd.AddCommand(listNotes(s))
	rootCmd.AddCommand(viewNote(s))
	rootCmd.AddCommand(export(s))
	rootCmd.AddCommand(migrate(s))

	// Defining absent subcommands to launch the tui environment.

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		// initialize tui
		if err := tui.Run(s); err != nil {
			log.Fatalf("TUI exited with err: %v", err)
		}
	}

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
}
