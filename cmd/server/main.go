package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/dallas1295/biji/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		if runtime.GOOS == "windows" {
			dataDir = filepath.Join(os.Getenv("APPDATA"), "biji-server")
		} else {
			// I don't like this a default but unsure where
			dataDir = filepath.Join(usr.HomeDir, "biji-server")
		}
	}

	log.Printf("Starting biji server on port %s", port)
	log.Printf("Data Directory: %s", dataDir)

	srv, err := server.NewServer(dataDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", srv.RegisterHandler)
	mux.HandleFunc("/api/sync", srv.SyncHandler)
	mux.HandleFunc("/api/notes", srv.GetNotesHandler)

	log.Println("Available Routes:")
	log.Println("  POST  /api/register")
	log.Println("  POST  /api/sync")
	log.Println("  GET   /api/notes")

	httpServer := &http.Server{
		Addr:         "127.0.0.1:" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Recieve signal %v, shutting server down", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server stopped")
}
