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
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`

	Version  int       `json:"version"`
	LastSync time.Time `json:"lastSync"`
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
	// Get OS and set default dir based on results
	// NOTE currently this is set up to save to a JSON file. However,
	// in the future I'll consider switching to a SQLite database.
	// The current implementation is easier for cloud sync learning.
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

// GetNoteFromID get's the notes ID in the JSON file from the title.
func (s *Store) GetNoteFromID(id string) (Note, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, note := range s.Notes {
		if note.ID == id {
			return note, nil
		}
	}

	return Note{}, fmt.Errorf("could not find note with ID: %s", id)
}

// GetNotes reads the JSON file and loads them into memory via an vector of notes.
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

// AddNote takes a name and some content and trims and preps them.
// It creates an in memory Note and adds it to the JSON file.
func (s *Store) AddNote(name, content string) (*Note, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	trimmedName := strings.TrimSpace(name)
	id, _ := s.FindNoteID(s.Notes, trimmedName)
	if id == "" {
		name = trimmedName
	} else {
		return &Note{}, fmt.Errorf("name: %s, is already taken", trimmedName)
	}

	noteID := uuid.NewString()
	note := Note{
		ID:         noteID,
		Name:       name,
		Content:    content,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
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

// DeleteNote searches the in memory notes marks it's position in the vector,
// and recompiles the in memory vector. After it's changed in memory it resaves the json.
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

// UpdateNoteName takes the notes ID and a new name. and returns a changed note in memory and then resaves the JSON.
// It returns an error if no note is found
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
				return Note{}, fmt.Errorf("error marshalling new JSON file: %w", err)
			}
			if err := os.WriteFile(s.dataFile, jsonData, 0o644); err != nil {
				return Note{}, fmt.Errorf("error saving JSON file: %w", err)
			}

			return s.Notes[i], nil
		}
	}
	return Note{}, fmt.Errorf("could not find note with provided ID")
}

// UpdateNoteContent operates the same as UpdateNoteName. changes the in memory slices then writes to the JSON
// It returns an error if no note is found
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

// FindNoteID takes the in memory notes array and a name, it iterates over the array until it matches the name.
// It returns an error if no note is found
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
	notesArr, _ := s.GetNotes()

	var noteNames []string

	for _, note := range notesArr {
		name := note.Name
		noteNames = append(noteNames, name)

	}

	return noteNames
}

// ExportNote takes a note ID, GetNoteFromID and preps and trims into a .md file to Documents.
// It returns an error if no note is found, or if there is an error writing the new .md file.
func (s *Store) ExportNote(id string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find user's home directory: %w", err)
	}

	docs := filepath.Join(home, "Documents")

	noteJSON, err := s.GetNoteFromID(id)
	if err != nil {
		return fmt.Errorf("could not get note with id: %s", id)
	}

	prepName := strings.ReplaceAll(noteJSON.Name, " ", "_")
	fileName := prepName + ".md"
	filePath := filepath.Join(docs, fileName)

	data := []byte(noteJSON.Content)

	err = os.WriteFile(filePath, data, 0o666)
	if err != nil {
		return fmt.Errorf("could not export note: %w", err)
	}

	return nil
}

// ExportAll takes the entire saved JSON file compiles multiple .md files and zips them.
// It places it in Documents under biji-export.zip
func (s *Store) ExportAll() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find user's home directory: %w", err)
	}

	exportPath := filepath.Join(home, "Documents", "biji-export.zip")

	tempDir, err := os.MkdirTemp("", "biji-export")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}

	defer os.RemoveAll(tempDir)

	notes, err := s.GetNotes()
	if err != nil {
		return fmt.Errorf("could not retrive notes: %w", err)
	}

	for _, note := range notes {
		cleanName := strings.ReplaceAll(note.Name, " ", "_") + ".md"
		tmpFile := filepath.Join(tempDir, cleanName)

		data := []byte(note.Content)

		if err = os.WriteFile(tmpFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to write temp file %s: %w", cleanName, err)
		}

	}

	err = createZip(tempDir, exportPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	fmt.Printf("Export Complete!\nCheck your ~/Documents directory.")

	return nil
}
