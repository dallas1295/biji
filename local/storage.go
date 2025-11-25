package local

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	Done       bool      `json:"done"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type Store struct {
	dataFile string
	notes    []Note       // In-memory cache
	mutex    sync.RWMutex // For multithreading
}

// Init initializes the storage directory. If the directory does not exist, it creates one.
func (s *Store) Init() error {
	var err error // For function level error handling.

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %w", err)
	}

	var bijiDir string
	if runtime.GOOS == "windows" {
		bijiDir = filepath.Join(os.Getenv("APPDATA"), "biji")
	} else {
		bijiDir = filepath.Join(usr.HomeDir, ".config", "biji")
	}
	if err = os.MkdirAll(bijiDir, 0o755); err != nil {
		return fmt.Errorf("error creating biji config directory: %w", err)
	}

	s.dataFile = filepath.Join(bijiDir, "biji.json")
	if _, err = os.Stat(s.dataFile); os.IsNotExist(err) {
		if err = os.WriteFile(s.dataFile, []byte("[]"), 0o644); err != nil {
			return fmt.Errorf("error creating biji.json: %w", err)
		}
	}

	notes, err := s.LoadNotes()
	if err != nil {
		return fmt.Errorf("error loading notes: %w", err)
	}
	s.notes = notes

	return nil
}

func (s *Store) LoadNotes() ([]Note, error) {
	var notes []Note

	notesJSON, err := os.ReadFile(s.dataFile)
	if err != nil {
		return nil, fmt.Errorf("error reading json file: %w", err)
	}

	err = json.Unmarshal(notesJSON, &notes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json file: %w", err)
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ModifiedAt.After(notes[j].ModifiedAt)
	})

	return notes, nil
}

func (s *Store) AddNote(name, content string) (*Note, error) {
	noteID := uuid.NewString()
	note := Note{
		ID:         noteID,
		Name:       name,
		Content:    content,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Done:       false,
	}

	s.notes = append(s.notes, note)

	jsonData, err := json.Marshal(s.notes)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json file: %w", err)
	}

	err = os.WriteFile(s.dataFile, jsonData, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error saving json file: %w", err)
	}

	return &note, nil
}

func (s *Store) DeleteNote(id string) error {
	indexToDelete := -1
	for i, note := range s.notes {
		if note.ID == id {
			indexToDelete = i
			break
		}
	}

	if indexToDelete != -1 {
		s.notes = append(s.notes[:indexToDelete], s.notes[indexToDelete+1:]...)
	} else {
		return nil
	}

	jsonData, err := json.Marshal(s.notes)
	if err != nil {
		return fmt.Errorf("error saving json file: %w", err)
	}

	err = os.WriteFile(s.dataFile, jsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error saving json file: %w", err)
	}

	return nil
}

func (s *Store) UpdateNoteName(id string, newName string) (*Note, error) {
	var updatedNote *Note

	for i, note := range s.notes {
		if note.ID == id {
			if s.notes[i].Name != newName {
				s.notes[i].Name = newName
				s.notes[i].ModifiedAt = time.Now()
				updatedNote = &s.notes[i]

				jsonData, err := json.Marshal(s.notes)
				if err != nil {
					return nil, fmt.Errorf("error saving json file: %w", err)
				}

				err = os.WriteFile(s.dataFile, jsonData, 0o644)
				if err != nil {
					return nil, fmt.Errorf("error saving json file: %w", err)
				}
			} else {
				return nil, nil
			}
		}
	}

	return updatedNote, nil
}

func (s *Store) UpdateNoteContent(id string, newContent string) (*Note, error) {
	var updatedNote *Note
	trimmedContent := strings.TrimSpace(newContent)

	for i, note := range s.notes {
		if note.ID == id {
			if s.notes[i].Content != trimmedContent {
				s.notes[i].Content = trimmedContent

				s.notes[i].ModifiedAt = time.Now()
				updatedNote = &s.notes[i]

				jsonData, err := json.Marshal(s.notes)
				if err != nil {
					return nil, fmt.Errorf("error saving json file: %w", err)
				}

				err = os.WriteFile(s.dataFile, jsonData, 0o644)
				if err != nil {
					return nil, fmt.Errorf("error saving json file: %w", err)
				}
			} else {
				return nil, nil
			}
		}
	}

	return updatedNote, nil
}
