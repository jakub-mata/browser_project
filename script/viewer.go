package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func CreateViewer(root *TreeRoot) {
	viewerApp := app.New()
	mainContainer, ok := root.Root.traverseParsingTree()

	if !ok {
		panic("no content")
	}

	window := viewerApp.NewWindow(PAGE_TITLE)
	window.Resize(fyne.NewSize(800, 800))
	window.SetContent(container.NewScroll(mainContainer))

	window.Show()
	viewerApp.Run()
}

func (root *TreeVertex) traverseParsingTree() (*fyne.Container, bool) {
	var subObjects []fyne.CanvasObject
	for _, child := range root.Children {
		subContainer, ok := containerFactory(child)
		if !ok {
			continue
		}
		subObjects = append(subObjects, subContainer)
	}
	if len(subObjects) == 0 {
		return container.NewWithoutLayout(), false
	}

	baseContainer := container.NewVBox(subObjects...)
	return baseContainer, true
}
