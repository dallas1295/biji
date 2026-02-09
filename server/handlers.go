package server

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/dallas1295/biji/local"
)

type SyncRequest struct {
	Notes    []local.Note `json:"notes"`
	LastSync string       `json:"lastSync"`
}

func (s *Server) SyncHandler(w http.ResponseWriter, r *http.Request) {
	// TODO implement real sync logic

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	syncCode := r.Header.Get("X-Sync-Code")
	if syncCode == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Notes    []local.Note `json:"notes"`
		LastSync string       `json:"lastSync"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func generateSyncCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const segments = 4
	const segmentLength = 4

	const totalLen = segments*segmentLength + (segments - 1)
	b := make([]byte, totalLen)

	pos := 0
	for i := range segments {
		if i > 0 {
			b[pos] = ' '
			pos++
		}

		for range segmentLength {
			b[pos] = charset[rand.N(len(charset))]
			pos++
		}
	}

	return string(b)
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate a new sync code for user then create a user struct
	// with the generated code that is unique
	const maxAttempts = 10
	var syncCode string
	var user *User

	// Set a maximum number of attempts, good for non-infinite loop
	for range maxAttempts {
		syncCode = generateSyncCode()
		s.mu.Lock()
		_, exists := s.users[syncCode]
		if !exists {
			user = &User{
				SyncCode:  syncCode,
				Notes:     []local.Note{},
				CreatedAt: time.Now(),
				LastSync:  time.Now(),
			}
			s.users[syncCode] = user
			s.userLocks[syncCode] = &sync.RWMutex{}
			s.mu.Unlock()
			break
		}
		s.mu.Unlock()
	}

	if user == nil {
		http.Error(w, "Failed to generate unique sync code, please try again", http.StatusInternalServerError)
		return
	}

	if err := s.saveUserToDisk(syncCode); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	res := map[string]string{"syncCode": syncCode}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check the header for a sync code
	syncCode := r.Header.Get("X-Sync-Code")
	if syncCode == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Lock Master Lock and Release it when we're done checking
	s.mu.RLock()
	user, exists := s.users[syncCode]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Map the bytes from the server into a user readable json output
	res := map[string]any{"notes": user.Notes}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) GetUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// just returns the number of users stored
	return len(s.users)
}
