package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type ImageFetch struct {
	url             fyne.URI
	parentContainer *fyne.Container
	placeholder     *canvas.Rectangle
}

var imagesToFetch []ImageFetch

func CreateViewer(root *TreeRoot) {
	viewerApp := app.New()
	mainContainer, ok := root.Root.traverseParsingTree()

	if !ok {
		panic("no content")
	}

	window := viewerApp.NewWindow(PAGE_TITLE)
	window.Resize(fyne.NewSize(800, 800))
	window.SetContent(container.NewScroll(mainContainer))
	refreshImages()
	window.ShowAndRun()
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

func refreshImages() {
	for _, image := range imagesToFetch {
		fetched := canvas.NewImageFromURI(image.url)
		fetched.FillMode = canvas.ImageFillOriginal
		image.parentContainer.Objects = []fyne.CanvasObject{fetched}
		image.parentContainer.Refresh()
	}
}
