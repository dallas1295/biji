package main

import (
	"log"

	"codeberg.org/dallas1295/biji/cmd"
	"codeberg.org/dallas1295/biji/local"
	"codeberg.org/dallas1295/biji/tui"
)

func main() {
	s := &local.Store{}
	if err := s.Init(); err != nil {
		log.Fatalf("failed to initialize store: %v:", err)
	}

	cmd.Execute()

	if err := tui.Run(s); err != nil {
		log.Fatalf("TUI exited with err: %v", err)
	}
}
