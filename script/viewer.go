package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

var BORDER_COLOR color.Gray16 = color.Black
var TEXT_COLOR color.Gray16 = color.Black
var PAGE_TITLE string = "Hello World"

func CreateViewer(root *TreeRoot) {
	viewerApp := app.New()
	mainContainer := root.Root.traverseParsingTree()

	window := viewerApp.NewWindow(PAGE_TITLE)
	window.Resize(fyne.NewSize(800, 800))
	window.SetContent(container.NewScroll(mainContainer))

	window.Show()
	viewerApp.Run()
}

func (root *TreeVertex) traverseParsingTree() *fyne.Container {
	baseContainer := container.NewVBox()
	containerFactory(root, baseContainer)
	return baseContainer
}

/*
func isAList(name string) bool {
	return (name == "ul") || (name == "ol")
}
*/
