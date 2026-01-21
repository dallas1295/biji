package main

import (
	"log"

	"github.com/dallas1295/biji/cmd"
	"github.com/dallas1295/biji/local"
)

func main() {
	s := &local.Store{}
	if err := s.Init(); err != nil {
		log.Fatalf("failed to initialize store: %v:", err)
	}

	cmd.Execute(s)
}
