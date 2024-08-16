package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

// Define the new segment struct
type MyCustomSegment struct {
	Text       string
	textObject *canvas.Text
	selected   bool
}

// Ensure MyCustomSegment implements the RichTextSegment interface
func (m *MyCustomSegment) Inline() bool {
	return true // Return true if this segment should be inline with others
}

func (m *MyCustomSegment) Textual() string {
	return m.Text // Return the text content for this segment
}

func (m *MyCustomSegment) Update(obj fyne.CanvasObject) {
	m.textObject = obj.(*canvas.Text)
	m.textObject.Text = m.Text
	if m.selected {
		m.textObject.Color = theme.PrimaryColor()
	} else {
		m.textObject.Color = theme.ForegroundColor()
	}
	m.textObject.Refresh()
}

func (m *MyCustomSegment) Visual() fyne.CanvasObject {
	if m.textObject == nil {
		m.textObject = canvas.NewText(m.Text, theme.ForegroundColor())
	}
	return m.textObject
}

func (m *MyCustomSegment) Select(pos1, pos2 fyne.Position) {
	m.selected = true
	m.Update(m.Visual())
}

func (m *MyCustomSegment) SelectedText() string {
	if m.selected {
		return m.Text
	}
	return ""
}

func (m *MyCustomSegment) Unselect() {
	m.selected = false
	m.Update(m.Visual())
}
