package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

var BORDER_COLOR color.Gray16 = color.Black
var TEXT_COLOR color.Gray16 = color.Black

func CreateViewer(root *TreeRoot) {
	viewerApp := app.New()
	window := viewerApp.NewWindow("Hello World")
	window.Resize(fyne.NewSize(800, 800))

	mainContainer := root.Root.traverseParsingTree()
	window.SetContent(container.NewScroll(mainContainer))

	window.Show()
	viewerApp.Run()
}

func (root *TreeVertex) traverseParsingTree() fyne.CanvasObject {

	var childContainers []fyne.CanvasObject
	for _, child := range root.Children {
		childContainers = append(childContainers, child.traverseParsingTree())
	}

	rootContainer := containerFactory(root)
	baseContainer := container.NewVBox(
		rootContainer,
		container.NewVBox(childContainers...),
	)
	return baseContainer
}
