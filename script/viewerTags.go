package main

import (
	"fmt"
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const DEFAULT_FONT_SIZE float32 = 16

type BOXTYPE int8

const (
	VBox BOXTYPE = iota
	HBox
)

var BORDER_COLOR color.Gray16 = color.Black
var TEXT_COLOR color.Gray16 = color.Black
var PAGE_TITLE string = "Hello World"
var CIRCLE_LIST_STROKEWIDTH int = 10

var headerSizes = map[string]float32{
	"h1": 32,
	"h2": 24,
	"h3": 18.72,
	"h4": 16,
	"h5": 13.28,
	"h6": 10.72,
}

var boxTypes = map[string]BOXTYPE{
	"h1":     HBox,
	"h2":     HBox,
	"h3":     HBox,
	"h4":     HBox,
	"h5":     HBox,
	"h6":     HBox,
	"p":      HBox,
	"li":     HBox,
	"a":      HBox,
	"img":    VBox,
	"div":    VBox,
	"span":   VBox,
	"ul":     VBox,
	"ol":     VBox,
	"body":   VBox,
	"header": VBox,
	"footer": VBox,
	"html":   VBox,
	"main":   VBox,
	"br":     VBox,
	"hr":     VBox,
}

func containerFactory(element *TreeVertex) (fyne.CanvasObject, bool) {

	var subObjects []fyne.CanvasObject

	if element.Token.Type == Character {
		switch element.Parent.Token.Name {
		case "h1", "h2", "h3", "h4", "h5", "h6", "p", "ul", "ol", "strong", "em":
			if element.Token.Content.Len() == 0 {
				break
			}
			label := canvas.NewText(element.Token.Content.String(), TEXT_COLOR)
			label.TextSize = getFontSize(element.Parent.Token.Name)
			if element.Parent.Token.Name == "strong" {
				label.TextStyle.Bold = true
			}
			if element.Parent.Token.Name == "em" {
				label.TextStyle.Italic = true
			}

			subObjects = append(subObjects, label)
		case "li":
			label := canvas.NewText(element.Token.Content.String(), TEXT_COLOR)
			label.TextSize = getFontSize(element.Token.Name)

			switch element.Parent.Parent.Token.Name {
			case "ul":
				circle := canvas.NewCircle(TEXT_COLOR)
				circle.StrokeWidth = float32(CIRCLE_LIST_STROKEWIDTH)
				subObjects = append(subObjects, circle, label)
			default:
				subObjects = append(subObjects, label)
			}
		case "a":
			linkValue, err := element.Parent.Token.findHref()
			if err != nil {
				break
			}
			hyperLink := widget.NewHyperlink(element.Token.Content.String(), linkValue)
			subObjects = append(subObjects, hyperLink)
		case "title":
			PAGE_TITLE = element.Token.Content.String()
		case "div", "body", "header", "footer", "html", "main", "span":
			if element.Token.Content.Len() == 0 {
				break
			}
			label := canvas.NewText(element.Token.Content.String(), TEXT_COLOR)
			subObjects = append(subObjects, label)
		}
	} else {
		switch element.Token.Name {
		case "img":
			imageURL, err := getURL(&element.Token)
			if err != nil {
				break
			}
			image := canvas.NewImageFromURI(imageURL)
			image.FillMode = canvas.ImageFillOriginal
			subObjects = append(subObjects, image)
		case "br":
			newline := canvas.NewText("\n", TEXT_COLOR)
			subObjects = append(subObjects, newline)
		case "hr":
			line := canvas.NewLine(TEXT_COLOR)
			subObjects = append(subObjects, line)
		}
	}
	/*
		//handling root
		if isMeta(element.Token.Name) {
			if element.Token.Name == "title" {
				PAGE_TITLE = element.Text.String()
			}
		} else if element.Token.Type == Character {
			label := canvas.NewText(element.Token.Content.String(), TEXT_COLOR)
			subObjects = append(subObjects, label)
		} else {

			switch element.Token.Name {
			case "h1", "h2", "h3", "h4", "h5", "h6", "p", "ul", "ol":
				if len(element.Text.String()) == 0 {
					break
				}
				label := canvas.NewText(element.Text.String(), TEXT_COLOR)
				label.TextSize = getFontSize(element.Token.Name)
				subObjects = append(subObjects, label)
			case "li":
				label := canvas.NewText(element.Text.String(), TEXT_COLOR)
				label.TextSize = getFontSize(element.Token.Name)

				switch element.Parent.Token.Name {
				case "ul":
					circle := canvas.NewCircle(TEXT_COLOR)
					circle.StrokeWidth = float32(CIRCLE_LIST_STROKEWIDTH)
					subObjects = append(subObjects, circle, label)
				default:
					subObjects = append(subObjects, label)
				}
			case "a":
				linkValue, err := element.Token.findHref()
			if err != nil {
				break
			}
			hyperLink := widget.NewHyperlink(element.Text.String(), linkValue)
			subObjects = append(subObjects, hyperLink)
			case "div", "body", "header", "footer", "html", "main", "span":
				if element.Text.Len() == 0 {
					break
				}
				label := canvas.NewText(element.Text.String(), TEXT_COLOR)
				subObjects = append(subObjects, label)
			case "br":
				break
			case "hr":
				line := canvas.NewLine(TEXT_COLOR)
				subObjects = append(subObjects, line)
			case "img":
				imageURL, err := getURL(&element.Token)
				if err != nil {
					break
				}
				image := canvas.NewImageFromURI(imageURL)
				image.FillMode = canvas.ImageFillOriginal
				subObjects = append(subObjects, image)
			default:
				w := widget.NewLabel(element.Token.Name)
				subObjects = append(subObjects, w)
			}
		}
	*/

	//recursion
	for _, child := range element.Children {
		subContainer, ok := containerFactory(child)
		if !ok {
			return container.NewWithoutLayout(), false
		}
		subObjects = append(subObjects, subContainer)
	}

	//returning
	boxType, ok := boxTypes[element.Token.Name]
	if !ok {
		base := container.NewHBox(subObjects...)
		return base, true
	}
	if boxType == VBox {
		base := container.NewVBox(subObjects...)
		return base, true
	} else {
		base := container.NewHBox(subObjects...)
		return base, true
	}

}

//HELPER FUNCTIONS

func (token *HTMLToken) findHref() (*url.URL, error) {

	for _, attr := range token.Attributes {
		if attr.Name == "href" {
			urlValue, err := url.Parse(attr.Value)
			if err != nil {
				return &url.URL{}, fmt.Errorf("couldn't parse given url")
			}
			return urlValue, nil
		}
	}

	return &url.URL{}, fmt.Errorf("anchor tag does not contain a reference")
}

func getURL(imageToken *HTMLToken) (fyne.URI, error) {
	for _, attribute := range imageToken.Attributes {
		if attribute.Name == "src" {
			source, err := storage.ParseURI(attribute.Value)
			if err != nil {
				return source, err
			}
			return source, nil
		}
	}
	panic("No url available")
}

func isMeta(tag string) bool {
	metaTags := [...]string{
		"meta", "head", "title",
	}
	for _, metaTag := range metaTags {
		if metaTag == tag {
			return true
		}
	}
	return false
}

func getFontSize(name string) float32 {
	res, ok := headerSizes[name]
	if ok {
		return res
	} else {
		return DEFAULT_FONT_SIZE
	}
}
