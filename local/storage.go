package local

import (
	"encoding/json"
	"fmt"
	"log"
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
	Notes    []Note       // In-memory cache
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

	notes, err := s.GetNotes()
	if err != nil {
		return fmt.Errorf("error loading notes: %w", err)
	}
	s.Notes = notes

	return nil
}

func (s *Store) GetNoteFromID(id string) (Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, note := range s.Notes {
		if note.ID == id {
			return note, nil
		}
	}

	return Note{}, fmt.Errorf("could not find note with ID: %v", id)
}

func (s *Store) GetNotes() ([]Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	trimmedName := strings.TrimSpace(name)
	id, _ := s.FindNoteID(s.Notes, trimmedName)
	if id == "" {
		name = trimmedName
	} else {
		log.Fatalf("name: %v, is already taken", trimmedName)
	}

	noteID := uuid.NewString()
	note := Note{
		ID:         noteID,
		Name:       name,
		Content:    content,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Done:       false,
	}

	s.Notes = append(s.Notes, note)

	jsonData, err := json.Marshal(s.Notes)
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	indexToDelete := -1
	for i, note := range s.Notes {
		if note.ID == id {
			indexToDelete = i
			break
		}
	}

	if indexToDelete != -1 {
		s.Notes = append(s.Notes[:indexToDelete], s.Notes[indexToDelete+1:]...)
	} else {
		return nil
	}

	jsonData, err := json.Marshal(s.Notes)
	if err != nil {
		return fmt.Errorf("error saving json file: %w", err)
	}

	err = os.WriteFile(s.dataFile, jsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error saving json file: %w", err)
	}

	return nil
}

func (s *Store) UpdateNoteName(id string, newName string) (Note, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newName = strings.TrimSpace(newName)
	for i, note := range s.Notes {
		if note.ID == id {
			if s.Notes[i].Name == newName {
				return s.Notes[i], nil
			}

			s.Notes[i].Name = newName
			s.Notes[i].ModifiedAt = time.Now()

			jsonData, err := json.Marshal(s.Notes)
			if err != nil {
				return Note{}, fmt.Errorf("error marshalling new JSON file: %v", err)
			}
			if err := os.WriteFile(s.dataFile, jsonData, 0o644); err != nil {
				return Note{}, fmt.Errorf("error saving JSON file: %v", err)
			}

			return s.Notes[i], nil
		}
	}
	return Note{}, fmt.Errorf("could not find note with provided ID")
}

func (s *Store) UpdateNoteContent(id string, newContent string) (Note, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	trimmedContent := strings.TrimSpace(newContent)

	for i, note := range s.Notes {
		if note.ID == id {
			// Only update if the content is actually different
			if s.Notes[i].Content == trimmedContent {
				// Return a copy of the unchanged note
				return s.Notes[i], nil
			}

			// Modify the note in-place within the locked section
			s.Notes[i].Content = trimmedContent
			s.Notes[i].ModifiedAt = time.Now()

			// Persist the change
			jsonData, err := json.Marshal(s.Notes)
			if err != nil {
				return Note{}, fmt.Errorf("error saving json file: %w", err)
			}
			if err := os.WriteFile(s.dataFile, jsonData, 0o644); err != nil {
				return Note{}, fmt.Errorf("error saving json file: %w", err)
			}

			// Return a COPY of the newly updated note
			return s.Notes[i], nil
		}
	}

	return Note{}, fmt.Errorf("could not find note with ID: %s", id)
}

func (s *Store) FindNoteID(notes []Note, name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	for i := range notes {
		if notes[i].Name == trimmedName {
			return notes[i].ID, nil
		}
	}

	return "", fmt.Errorf("could not find note with name: %s", trimmedName)
}

func (s *Store) GetNoteNames() []string {
	notesArr, err := s.GetNotes()
	if err != nil {
		log.Fatalf("error loading notes: %v", err)
	}

	var noteNames []string

	for _, note := range notesArr {
		name := note.Name
		noteNames = append(noteNames, name)

	}

	return noteNames
}
