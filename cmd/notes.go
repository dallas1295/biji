package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dallas1295/biji/local"
	"github.com/spf13/cobra"
)

// TODO: This logic is broken... need to create a switch case to get desired outcome.
// it can probably get done via the if or else logic but it's lazy and not pretty and
// easy to read.
// DO THE SAME FOR DELETION MAYBE????
func newNote(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "new [name] [content]",
		Short: "Create a new note",
		RunE: func(cmd *cobra.Command, args []string) error {
			var name, content string
			reader := bufio.NewReader(os.Stdin)

			if len(args) == 2 {
				name = args[0]

				content = strings.TrimSpace(strings.ReplaceAll(args[1], "\\n", "\n"))
			} else {
				fmt.Print("Name: ")

				rawInput, err := reader.ReadString('\n')
				if err != nil {
					log.Fatalf("Failed to read name: %v", err)
				}
				name = strings.TrimSpace(rawInput)

				fmt.Print("Content: ")
				rawInput, err = reader.ReadString('\n')
				if err != nil {
					log.Fatalf("Failed to read content: %v", err)
				}
				content = strings.ReplaceAll(rawInput, "\\n", "\n")
				content = strings.TrimSpace(content)
			}

			_, err := s.AddNote(name, content)
			if err != nil {
				log.Fatalf("Note creation failed: %v", err)
			}

			fmt.Println("Note created sucessfully")

			return nil
		},
	}

	return &cmd
}

func deleteNote(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete [name], [name], ...",
		Short: "Delete a note by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("No note names provided")
				return nil
			}
			for i := range args {
				name := strings.TrimSpace(args[i])
				id, err := s.FindNoteID(s.Notes, name)
				if err != nil {
					log.Fatalf("Error deleting note: %v", err)
				}

				err = s.DeleteNote(id)
				if err != nil {
					log.Fatalf("Error deleting note: %v", err)
				}
			}

			fmt.Println("Note deleted successfully")

			return nil
		},
	}

	return &cmd
}

func updateNoteName(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "rename [currName] [newName]",
		Short: "rename note with current name [name]",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			currName := strings.TrimSpace(args[0])
			var newName string

			if len(args) > 1 {
				newName = strings.TrimSpace(args[1])
			} else {
				fmt.Print("New Name: ")
				rawInput, err := reader.ReadString('\n')
				if err != nil {
					log.Fatalf("Failed to read new name input: %v", err)
				}
				newName = strings.TrimSpace(rawInput)
			}

			id, err := s.FindNoteID(s.Notes, currName)
			if err != nil {
				log.Fatalf("Could not find note: %v", err)
			}

			_, err = s.UpdateNoteName(id, newName)
			if err != nil {
				log.Fatalf("Failed to update note name: %v", err)
			}

			return nil
		},
	}

	return &cmd
}

func viewNote(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "view [name]",
		Short: "View note contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])

			id, err := s.FindNoteID(s.Notes, name)
			if err != nil {
				log.Fatalf("Could not find note: %v", err)
			}

			note, err := s.GetNoteFromID(id)
			if err != nil {
				log.Fatalf("Failed to retrieve note: %v", err)
			}

			fmt.Printf(
				"\n\nNote: %s\nContent: %s\n\n",
				note.Name,
				note.Content,
			)

			return nil
		},
	}

	return &cmd
}

func listNotes(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "list currently saved notes",
		RunE: func(cmd *cobra.Command, args []string) error {
			names := s.GetNoteNames()
			if len(names) == 0 {
				fmt.Printf("	no notes\n")
				return nil
			}

			fmt.Println("Notes:")
			for _, name := range names {
				fmt.Printf("	%s\n", name)
			}

			return nil
		},
	}

	return &cmd
}

func export(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "export [name] [name] ...",
		Short: "Export designated note(s) to .md file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("No note names provided")
				return nil
			}
			for i := range args {
				name := strings.TrimSpace(args[i])
				id, err := s.FindNoteID(s.Notes, name)
				if err != nil {
					log.Fatalf("Error exporting note: %v", err)
				}

				err = s.ExportNote(id)
				if err != nil {
					log.Fatalf("Error exporting note: %v", err)
				}

				fmt.Printf("%s successfully exported\n", name)
			}

			return nil
		},
	}

	return &cmd
}

func migrate(s *local.Store) *cobra.Command {
	cmd := cobra.Command{
		Use:   "migrate",
		Short: "Export all notes for migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := s.ExportAll()
			if err != nil {
				log.Fatalf("error exporting notes: %v", err)
			}
			return nil
		},
	}

	return &cmd
}
