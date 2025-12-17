package tui

import (
	"log"

	"codeberg.org/dallas1295/biji/local"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	listView uint = iota
)

type model struct {
	state     uint
	store     *local.Store
	notes     []local.Note
	listIndex int
	textArea  textarea.Model
	textInput textinput.Model
	width     int
	height    int
}

func (m model) Init() tea.Cmd {
	return nil
}

func NewModel(s *local.Store) model {
	notesSlice, err := s.GetNotes()
	if err != nil {
		log.Fatalf("error loading notes: %v", err)
	}

	return model{
		state:     listView,
		store:     s,
		notes:     notesSlice,
		textArea:  textarea.New(),
		textInput: textinput.New(),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		key := msg.String()
		switch m.state {
		case listView:
			switch key {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "k", "up":
				if m.listIndex > 0 {
					m.listIndex--
				}
			case "j", "down":
				if m.listIndex < len(m.notes)-1 {
					m.listIndex++
				}
			}
		}

	}
	return m, tea.Batch(cmds...)
}
