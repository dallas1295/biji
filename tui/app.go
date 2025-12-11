package tui

import (
	"codeberg.org/dallas1295/biji/local"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(store *local.Store) error {
	m := NewModel(store)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
