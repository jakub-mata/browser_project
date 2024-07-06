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

/*
func traverseIter(root *TreeVertex) *fyne.Container {
	var elementStack Stack[*TreeVertex]
	elementStack.Push(root)
	var containerStack Stack[*fyne.Container]

	for elementStack.Len() > 0 {
		currentElement := elementStack.Pop()
		borderContainer := container.NewBorder(nil, nil, nil, nil)
		content := container.NewVBox()
		label := canvas.NewText(currentElement.Token.Name, color.Black)

		content.Add(label)
		for _, child := range currentElement.Children {
			childContainer := traverseIter(child)
			content.Add(childContainer)
		}

		borderContainer.Add(content)
		containerStack.Push(borderContainer)
	}

	return containerStack.content[0]
}

//STACK IMPLEMENTATION

type Stack[T any] struct {
	content []T
}

func (stack *Stack[T]) Len() int {
	return len(stack.content)
}

func (stack *Stack[T]) Push(e T) {
	stack.content = append(stack.content, e)
}

func (stack *Stack[T]) Pop() T {
	popped := stack.content[stack.Len()-1]
	stack.content = stack.content[:stack.Len()-1]
	return popped
}
*/
