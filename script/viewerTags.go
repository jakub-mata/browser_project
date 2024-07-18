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

var BORDER_COLOR color.Gray16 = color.Black
var TEXT_COLOR color.Gray16 = color.Black
var PAGE_TITLE string = "Hello World"
var CIRLCE_LIST_STROKEWIDTH int = 5

type NumberGenerator struct {
	currNumber int
}

var generator NumberGenerator = NumberGenerator{1}

var headerSizes = map[string]float32{
	"h1": 32,
	"h2": 24,
	"h3": 18.72,
	"h4": 16,
	"h5": 13.28,
	"h6": 10.72,
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

func containerFactory(e *TreeVertex, parentContainer *fyne.Container) {

	if isMeta(e.Token.Name) {
		if e.Token.Name == "title" {
			PAGE_TITLE = e.Text.String()
		}
		return
	}

	var currContainer *fyne.Container

	switch e.Token.Name {
	case "h1", "h2", "h3", "h4", "h5", "h6", "p":
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		label.TextSize = getFontSize(e.Token.Name)
		currContainer = container.NewVBox(label)
	case "li":
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		label.TextSize = getFontSize(e.Token.Name)

		switch e.Parent.Token.Name {
		case "ul":
			circle := canvas.NewCircle(TEXT_COLOR)
			circle.StrokeWidth = float32(CIRLCE_LIST_STROKEWIDTH)
			currContainer = container.NewHBox(circle, label)
		case "ol":
			number := canvas.NewText(string(generator.currNumber), TEXT_COLOR)
			generator.currNumber++
			currContainer = container.NewHBox(number, label)
		default:
			currContainer = container.NewHBox(label)
		}
	case "a":
		linkValue, err := e.Token.findHref()
		if err == nil {
			hyperLink := widget.NewHyperlink(e.Text.String(), linkValue)
			currContainer = container.NewVBox(hyperLink)
		}
	case "div", "body", "header", "footer", "html", "main", "span":
		if e.Text.Len() == 0 {
			break
		}
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		currContainer = container.NewVBox(label)
	case "br":

	case "hr":
		line := canvas.NewLine(TEXT_COLOR)
		currContainer = container.NewVBox(line)
	case "img":
		imageURL, err := getURL(&e.Token)
		if err != nil {
			break
		}
		image := canvas.NewImageFromURI(imageURL)
		image.FillMode = canvas.ImageFillOriginal
		currContainer = container.NewVBox(image)
	case "ul", "ol":
		currContainer = container.NewVBox()
	default:
		w := widget.NewLabel(e.Token.Name)
		currContainer = container.NewVBox(w)
	}

	if currContainer == nil {
		currContainer = container.NewVBox()
	}

	for _, child := range e.Children {
		containerFactory(child, currContainer)
	}

	parentContainer.Add(currContainer)
}

func (token *HTMLToken) findHref() (*url.URL, error) {
	if token.Name != "a" {
		return &url.URL{}, fmt.Errorf("not an anchor tag")
	}

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
