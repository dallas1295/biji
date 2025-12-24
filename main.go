package main

import (
	"log"

	"codeberg.org/dallas1295/biji/cmd"
	"codeberg.org/dallas1295/biji/local"
)

func main() {
	s := &local.Store{}
	if err := s.Init(); err != nil {
		log.Fatalf("failed to initialize store: %w:", err)
	}

	cmd.Execute(s)
}
