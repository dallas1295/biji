// Package tui is where all business and model logic for the bubbletea portion of the app is located.
package tui

import (
	"codeberg.org/dallas1295/biji/local"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(store *local.Store) error {
	m := NewModel(store)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
