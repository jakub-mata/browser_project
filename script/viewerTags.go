package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DefaultValues struct {
	fontSize       uint16
	textAlign      string
	display        string
	textDecoration string
}

var defaultValuesMap = map[string]DefaultValues{
	"p": {
		fontSize: 16,
		display:  "block",
	},
}

func getDefaults(name string) (DefaultValues, bool) {
	res, err := defaultValuesMap[name]
	return res, err
}

func containerFactory(e *TreeVertex) fyne.CanvasObject {
	defaultValues, included := getDefaults(e.Token.Name)
	if !included {
		return container.NewHBox(
			widget.NewLabel(e.Token.Name),
		)
	}
	switch e.Token.Name {
	case "p":
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		label.TextSize = float32(defaultValues.fontSize)
		fmt.Println(e.Text.String())
		return container.NewHBox(label)
	}

	return container.NewWithoutLayout()
}
