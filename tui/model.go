package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/dallas1295/biji/local"
)

type model struct {
	list      list.Model
	textarea  textarea.Model
	view      viewport.Model
	nameInput textinput.Model

	store           *local.Store
	notes           []list.Item
	currNote        *local.Note
	originalContent string

	showlist  bool
	isEditing bool
	focused   focusState
	unsaved   bool
}

type focusState int

const (
	focusList focusState = iota
	focusEditor
	focusName
)
