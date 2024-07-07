package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var BORDER_COLOR color.Gray16 = color.Black

func CreateViewer(root *TreeRoot) {
	viewerApp := app.New()
	window := viewerApp.NewWindow("Hello World")
	window.Resize(fyne.NewSize(1000, 1000))

	mainContainer := root.Root.traverseParsingTree()
	window.SetContent(mainContainer)

	window.Show()
	viewerApp.Run()
}

func (root *TreeVertex) traverseParsingTree() fyne.CanvasObject {
	fmt.Println(root.Token.Name)
	label := widget.NewLabel(root.Token.Name)
	var childContainers []fyne.CanvasObject
	for _, child := range root.Children {
		childContainers = append(childContainers, child.traverseParsingTree())
	}

	subContainer := container.NewVBox(childContainers...)
	baseContainer := container.NewVBox(label, subContainer)
	return baseContainer
}
