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

func (root *TreeVertex) traverseParsingTree() fyne.CanvasObject {

	//Recursion
	var childContainers []fyne.CanvasObject
	for _, child := range root.Children {
		childContainers = append(childContainers, child.traverseParsingTree())
	}
	//Node itself

	if root.Token.Name == "head" {
		findTitle(root)
	}

	rootContainer, err := containerFactory(root)
	if err != nil {
		return rootContainer
	}
	baseContainer := container.NewVBox(
		rootContainer,
		container.NewVBox(childContainers...),
	)
	return baseContainer
}

func findTitle(headElement *TreeVertex) {
	for _, child := range headElement.Children {
		if child.Token.Name == "title" {
			PAGE_TITLE = child.Text.String()
		}
	}
}
