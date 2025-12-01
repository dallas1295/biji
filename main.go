package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dallas1295/biji/local"
)

func main() {
	store := &local.Store{}
	store.Init()

	a := app.New()
	w := a.NewWindow("Biji Notes")

	nameBox := widget.NewEntry()
	contentBox := widget.NewEntry()
	button := widget.NewButton("Create Note", func() {
		go store.AddNote(nameBox.Text, contentBox.Text)
	})

	label := widget.NewLabel("Hello Biji!")
	content := container.NewVBox(button, label, container.NewHBox(nameBox, contentBox))

	w.SetContent(content)
	w.ShowAndRun()
}
