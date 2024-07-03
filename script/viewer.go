package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func CreateViewer(root *TreeRoot) {
	App := app.New()
	window := App.NewWindow("Hello World")
	window.SetContent(widget.NewLabel("HelloWorld!"))
	window.Show()
	App.Run()
}
