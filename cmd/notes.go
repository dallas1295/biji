package cmd

import (
	"fmt"
	"log"
	"strings"

	"codeberg.org/dallas1295/biji/local"
	"github.com/spf13/cobra"
)

func newNote(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "new [name] [content]",
		Short: "Create a new note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			content := args[1]
			_, err := s.AddNote(name, content)
			if err != nil {
				log.Fatalf("Note creation failed: %v", err)
			}

			fmt.Println("note created sucessfully")

			return nil
		},
	}

	return &cmd
}

func deleteNote(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a note by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			id, err := s.FindNoteID(s.Notes, name)
			if err != nil {
				log.Fatalf("Could not find note: %v", err)
			}

			err = s.DeleteNote(id)
			if err != nil {
				log.Fatalf("Note deletion failed: %v", err)
			}
			fmt.Println("note deleted successfuly")

			return nil
		},
	}

	return &cmd
}

func updateNoteName(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "update name [currName] [newName]",
		Short: "Update name of note",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			currName := strings.TrimSpace(args[1])
			newName := strings.TrimSpace(args[2])

			id, err := s.FindNoteID(s.Notes, currName)
			if err != nil {
				log.Fatalf("Could not find note: %v", err)
			}

			_, err = s.UpdateNoteName(id, newName)
			if err != nil {
				log.Fatalf("failed to update note name: %v", err)
			}

			return nil
		},
	}

	return &cmd
}
