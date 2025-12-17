package cmd

import (
	"fmt"
	"log"

	"codeberg.org/dallas1295/biji/local"
	"github.com/spf13/cobra"
)

var newNote = cobra.Command{
	Use:   "new [name] [content]",
	Short: "Create new note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		content := args[1]
		_, err := (*local.Store).AddNote(name, content)
		if err != nil {
			log.Fatalf("note creation failed: %v", err)
		}

		fmt.Println("note created")

		return nil
	},
}
