package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dallas1295/biji/local"
)

type User struct {
	SyncCode  string       `json:"syncCode"`
	Notes     []local.Note `json:"notes"`
	CreatedAt time.Time    `json:"createdAt"`
	LastSync  time.Time    `json:"lastSync"`
}
type Server struct {
	users     map[string]*User
	userLocks map[string]*sync.RWMutex
	dataDir   string
	mu        sync.RWMutex
}

func NewServer(dataDir string) (*Server, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	server := &Server{
		users:     make(map[string]*User),
		userLocks: make(map[string]*sync.RWMutex),
		dataDir:   dataDir,
	}

	if err := server.loadAllUsers(); err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	return server, nil
}

func (s *Server) loadAllUsers() error {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		syncCode := strings.TrimSuffix(entry.Name(), ".json")
		if err := s.loadUserFromDisk(syncCode); err != nil {
			fmt.Printf("Warning: failed to load user %s: %v\n", syncCode, err)
		}
	}

	return nil
}

func (s *Server) loadUserFromDisk(syncCode string) error {
	filePath := filepath.Join(s.dataDir, syncCode+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return err
	}

	s.mu.Lock()
	s.users[syncCode] = &user
	s.userLocks[syncCode] = &sync.RWMutex{}
	s.mu.Unlock()

	return nil
}

func (s *Server) saveUserToDisk(syncCode string) error {
	s.mu.RLock()
	user, userExists := s.users[syncCode]
	userLock, lockExists := s.userLocks[syncCode]

	if !userExists || !lockExists {
		return fmt.Errorf("user %s not found", syncCode)
	}

	userLock.Lock()
	defer userLock.Unlock()

	user.LastSync = time.Now()

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.dataDir, syncCode+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}
