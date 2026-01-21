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
	// get port from env file, if not set auto to 420420
	port := os.Getenv("PORT")
	if port == "" {
		port = "420420"
	}

	// get the current user
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	// get data directory from env file
	// if not appdata/biji-server or directly into home directory in Unix-like
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

	// creates a new server with the preexisting data's users
	srv, err := server.NewServer(dataDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// create a new mux and run my handlers via it
	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", srv.RegisterHandler)
	mux.HandleFunc("/api/sync", srv.SyncHandler)
	mux.HandleFunc("/api/notes", srv.GetNotesHandler)

	log.Println("Available Routes:")
	log.Println("  POST  /api/register")
	log.Println("  POST  /api/sync")
	log.Println("  GET   /api/notes")

	// server configuration
	httpServer := &http.Server{
		Addr:         "127.0.0.1:" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// create a channel to recieve signal input from
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// embed the ListenAndServe into a go function to move off main thread
	go func() {
		log.Printf("Server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// when the sigterm is notified the channel begins shutdown
	sig := <-sigChan
	log.Printf("Recieve signal %v, shutting server down", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server stopped")
}
