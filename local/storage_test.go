package local

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func (s *Store) cleanup() error {
	if s.dataFile != "" {
		return os.Remove(s.dataFile)
	}
	return nil
}

func TestAddNote(t *testing.T) {
	store := &Store{}

	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	name := "Hello"
	content := "Biji"

	note, err := store.AddNote(name, content)
	if err != nil {
		t.Fatalf("Failed to add note: %v", err)
	}

	if note.Name != name {
		t.Errorf("Expected name %s, got %s", name, note.Name)
	}

	if note.Content != content {
		t.Errorf("Expected content %s, got %s", content, note.Content)
	}

	if note.ID == "" {
		t.Error("Expected note to have an ID")
	}

	if note.Done != false {
		t.Error("Expected note to be not done")
	}

	// Verify timestamps were set
	if note.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if note.ModifiedAt.IsZero() {
		t.Error("Expected ModifiedAt to be set")
	}

	fmt.Printf("Note JSON: %v", note)

	defer store.cleanup()
}

func TestDeleteNote(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	name := "test"
	content := ""

	_, err = store.AddNote(name, content)
	if err != nil {
		t.Fatalf("Failed to add note: %v", err)
	}
	fmt.Printf("the notes array: %v", store.notes)

	deleteID := store.notes[0].ID

	store.DeleteNote(deleteID)
	fmt.Printf("the new notes array: %v", store.notes)

	if len(store.notes) > 0 {
		t.Error("Expected note deletion and empty array")
	}

	defer store.cleanup()
}

func TestUpdateNote(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	firstName := "test note"
	firstContent := "first content"

	firstNote, err := store.AddNote(firstName, firstContent)
	if err != nil {
		t.Fatalf("Failed to add note: %v", err)
	}

	fmt.Printf("Created Note with\nname: %v\ncontent: %v\n", firstNote.Name, firstNote.Content)

	updatedName := "updated test note"
	updatedContent := "second content"

	_, err = store.UpdateNoteName(firstNote.ID, updatedName)
	if err != nil {
		t.Fatalf("Failed to update note name: %v", err)
	}

	_, err = store.UpdateNoteContent(firstNote.ID, updatedContent)
	if err != nil {
		t.Fatalf("Failed to update note name: %v", err)
	}

	store.GetNotes()

	if store.notes[0].Name != updatedName {
		t.Error("Expected update to note's name")
	}

	if store.notes[0].Content != updatedContent {
		t.Error("Expected update to note's content")
	}

	defer store.cleanup()
}

func TestDeleteNote_NonExistentID(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.cleanup()
	nonExistentID := "non-existent-id"
	err = store.DeleteNote(nonExistentID)
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent note, got: %v", err)
	}

	if len(store.notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.notes))
	}
}

func TestUpdateNoteName_NonExistentID(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.cleanup()
	nonExistentID := "non-existent-id"
	newName := "updated name"

	updatedNote, err := store.UpdateNoteName(nonExistentID, newName)
	if err != nil {
		t.Errorf("Expected no error when updating non-existent note, got: %v", err)
	}

	if updatedNote != nil {
		t.Errorf("Expected nil note when updating non-existent ID, got: %v", updatedNote)
	}

	if len(store.notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.notes))
	}
}

func TestUpdateNoteContent_NonExistentID(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.cleanup()
	nonExistentID := "non-existent-id"
	newContent := "updated content"

	updatedNote, err := store.UpdateNoteContent(nonExistentID, newContent)
	if err != nil {
		t.Errorf("Expected no error when updating non-existent note content, got: %v", err)
	}

	if updatedNote != nil {
		t.Errorf("Expected nil note when updating non-existent ID, got: %v", updatedNote)
	}

	if len(store.notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.notes))
	}
}

func TestConcurrentAddNotes(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i <= 4; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			currNote := i + 1
			currContent := fmt.Sprintf("Note # %v", currNote)
			store.AddNote("Hello", currContent)
		}(i)
	}

	wg.Wait()
	for i, note := range store.notes {
		fmt.Printf("Here's %v note with content: %s\n", i, note.Content)
	}

	fmt.Printf("Curr array length: %v", len(store.notes))

	if len(store.notes) <= 4 {
		t.Error("Expected four notes in in-memory cache")
	}
}

func TestConcurrentDeleteNotes(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.cleanup()

	var IDArr []string
	var wg sync.WaitGroup

	for i := range store.notes {
		IDArr = append(IDArr, store.notes[i].ID)
	}

	for i, id := range IDArr {
		wg.Add(1)
		go func(num int, noteID string) {
			defer wg.Done()
			store.DeleteNote(noteID)
			fmt.Printf("Deleted note %v from array\n", i)
		}(i, id)
	}

	wg.Wait()

	if len(store.notes) != 0 {
		t.Error("Expected empty note array")
	}
}

func TestConcurrentUpdates(t *testing.T) {
	store := &Store{}
	err := store.Init()
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.cleanup()

	var wg sync.WaitGroup

	for i := 0; i <= 4; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			currNote := i + 1
			currContent := fmt.Sprintf("Note # %v", currNote)
			store.AddNote("Hello", currContent)
		}(i)
	}

	wg.Wait()

	var IDArr []string
	for i := range store.notes {
		IDArr = append(IDArr, store.notes[i].ID)
	}

	expectedUpdates := make(map[string]struct {
		name    string
		content string
	})

	for i, id := range IDArr {
		expectedName := fmt.Sprintf("Note %v", i)
		expectedContent := fmt.Sprintf("New content %v", i)
		expectedUpdates[id] = struct {
			name    string
			content string
		}{name: expectedName, content: expectedContent}
	}

	for i, id := range IDArr {
		wg.Add(1)
		go func(num int, noteID string) {
			defer wg.Done()
			newName := fmt.Sprintf("Note %v", i)
			store.UpdateNoteName(noteID, newName)
			fmt.Printf("Update note %v's name to: %v\n", i, store.notes[i].Name)
			newContent := fmt.Sprintf("New content %v", i)
			store.UpdateNoteContent(noteID, newContent)
			fmt.Printf("Updated note %v's content to : %v", i, store.notes[i].Content)
		}(i, id)
	}

	wg.Wait()

	for id, expected := range expectedUpdates {
		note, err := store.FindByID(store.notes, id)
		if err != nil {
			t.Errorf("Failed to find note %s: %v", id, err)
			continue
		}

		if note.Name != expected.name {
			t.Error("Expected note name to be updated")
		}
		if note.Content != expected.content {
			t.Error("Expected note content to be updated")
		}
	}
}
