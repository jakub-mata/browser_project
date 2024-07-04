package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateViewer(root TreeRoot) {
	App := app.New()
	window := App.NewWindow("Hello World")
	window.Resize(fyne.NewSize(1000, 1000))

	baseContainer := traverseParsingTree(&root.Root)
	window.SetContent(baseContainer)

	window.Show()
	App.Run()
}

func traverseParsingTree(element *TreeVertex) *fyne.Container {
	currContainer := container.NewWithoutLayout()

	label := widget.NewLabel(element.Token.Name)
	currContainer.Add(label)

	for _, child := range element.Children {
		currContainer.Add(traverseParsingTree(child))
	}

	return currContainer
}
