package local

import (
	"fmt"
	"os"
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

	store.LoadNotes()

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
