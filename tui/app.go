// Package tui is where all business and model logic for the bubbletea portion of the app is located.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dallas1295/biji/local"
)

func Run(store *local.Store) error {
	m := NewModel(store)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
