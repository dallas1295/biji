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
	fmt.Printf("the notes array: %v", store.Notes)

	deleteID := store.Notes[0].ID

	store.DeleteNote(deleteID)
	fmt.Printf("the new notes array: %v", store.Notes)

	if len(store.Notes) > 0 {
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

	if store.Notes[0].Name != updatedName {
		t.Error("Expected update to note's name")
	}

	if store.Notes[0].Content != updatedContent {
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

	if len(store.Notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.Notes))
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

	_, err = store.UpdateNoteName(nonExistentID, newName)
	if err != nil {
		fmt.Printf("Passed does not accept invalid IDs: %v", err)
	}

	if len(store.Notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.Notes))
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

	_, err = store.UpdateNoteContent(nonExistentID, newContent)
	if err != nil {
		fmt.Printf("Passed does not accept invalid IDs: %v", err)
	}

	if len(store.Notes) != 0 {
		t.Errorf("Expected store to remain empty, but has %d notes", len(store.Notes))
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
			currName := fmt.Sprintf("name-%d", i)
			currContent := fmt.Sprintf("Note # %v", currNote)
			store.AddNote(currName, currContent)
		}(i)
	}

	wg.Wait()
	for i, note := range store.Notes {
		fmt.Printf("Here's %v note with content: %s\n", i, note.Content)
	}

	fmt.Printf("Curr array length: %v", len(store.Notes))

	if len(store.Notes) <= 4 {
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

	for i := range store.Notes {
		IDArr = append(IDArr, store.Notes[i].ID)
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

	if len(store.Notes) != 0 {
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
			currName := fmt.Sprintf("InitialName-%d", index)
			currContent := fmt.Sprintf("InitialContent-%d", index)
			store.AddNote(currName, currContent)
		}(i)
	}

	wg.Wait()

	var IDArr []string
	for i := range store.Notes {
		IDArr = append(IDArr, store.Notes[i].ID)
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

			newName := fmt.Sprintf("Note %v", num)
			updatedNameNote, _ := store.UpdateNoteName(noteID, newName)
			fmt.Printf("Update note %v's name to: %v\n", num, updatedNameNote.Name)

			newContent := fmt.Sprintf("New content %v", num)
			updatedContentNote, _ := store.UpdateNoteContent(noteID, newContent)
			fmt.Printf("Updated note %v's content to : %v\n", num, updatedContentNote.Content)
		}(i, id)
	}

	wg.Wait()

	for id, expected := range expectedUpdates {
		note, err := store.GetNoteFromID(id)
		if err != nil {
			t.Errorf("Failed to find note %s: %v", id, err)
			continue
		}

		if note.Name != expected.name {
			t.Errorf("Name mismatch for %s: expected %s, got %s", id, expected.name, note.Name)
		}
		if note.Content != expected.content {
			t.Errorf("Content mismatch for %s: expected %s, got %s", id, expected.content, note.Content)
		}
	}
}
