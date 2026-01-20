package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type status int

var models []tea.Model

const (
	model status = iota
)

type Model struct {
	focused  status
	noteList []list.Model
}

func New() *Model {
	return &Model{}
}

func (m Model) Initi() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return nil, tea.Quit
		}
	}

	var cmd tea.Cmd
	return nil, cmd
}
