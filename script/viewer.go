package main

import (
	"fmt"
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
	mainContainer, err := root.Root.traverseParsingTree()
	if err != nil {
		fmt.Println("No page to display")
	}

	window := viewerApp.NewWindow(PAGE_TITLE)
	window.Resize(fyne.NewSize(800, 800))
	window.SetContent(container.NewScroll(mainContainer))

	window.Show()
	viewerApp.Run()
}

func (root *TreeVertex) traverseParsingTree() (fyne.CanvasObject, error) {

	//Recursion
	var childContainers []fyne.CanvasObject
	for _, child := range root.Children {
		cc, err := child.traverseParsingTree()
		if err == nil {
			childContainers = append(childContainers, cc)
		}
	}
	//Node itself

	if root.Token.Name == "title" {
		PAGE_TITLE = root.Text.String()
	}

	rootContainer, err := containerFactory(root)
	if err != nil {
		return rootContainer, err
	}

	baseContainer := container.NewVBox(
		rootContainer,
		container.NewVBox(childContainers...),
	)
	return baseContainer, nil
}
