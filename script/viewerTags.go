package main

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DisplayValue int8

const (
	block DisplayValue = iota
	inline
	inlineBlock
	none
)

type DefaultValues struct {
	fontSize  float32
	fontStyle fyne.TextStyle
	textAlign fyne.TextAlign
	display   DisplayValue
}

var defaultValuesMap = map[string]DefaultValues{
	"p": {
		fontSize:  16,
		textAlign: fyne.TextAlignLeading,
		display:   block,
	},
	"h1": {
		fontSize: 32,
		display:  block,
	},
	"h2": {
		fontSize: 24,
		display:  block,
	},
	"h3": {
		fontSize: 18.72,
		display:  block,
	},
	"h4": {
		fontSize: 16,
		display:  block,
	},
	"h5": {
		fontSize: 13.28,
		display:  block,
	},
	"h6": {
		fontSize: 10.72,
		display:  block,
	},
	"em": {
		fontStyle: fyne.TextStyle{Bold: true},
		display:   inline,
	},
	"a": {
		display: inline,
	},
	"div": {
		display: block,
	},
	"head": {
		display: none,
	},
	"title": {
		display: none,
	},
	"link": {
		display: none,
	},
	"meta": {
		display: none,
	},
	"br": {
		display: block,
	},
	"hr": {
		display: block,
	},
	"body": {
		display: block,
	},
	"footer": {
		display: block,
	},
	"header": {
		display: block,
	},
	"html": {
		display: block,
	},
}

func getDefaults(name string) (DefaultValues, bool) {
	res, err := defaultValuesMap[name]
	return res, err
}

func containerFactory(e *TreeVertex) (fyne.CanvasObject, error) {
	defaultValues, included := getDefaults(e.Token.Name)

	if !included {
		return container.NewHBox(
			widget.NewLabel(e.Token.Name),
		), nil
	}

	if defaultValues.display == none {
		return container.NewWithoutLayout(), fmt.Errorf("no display value")
	}

	switch e.Token.Name {
	case "h1", "h2", "h3", "h4", "h5", "h6", "p":
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		label.TextSize = defaultValues.fontSize
		return container.NewHBox(label), nil
	case "a":
		linkValue, err := e.Token.findHref()
		if err != nil {
			return container.NewWithoutLayout(), err
		}

		hyperLink := widget.NewHyperlink(e.Text.String(), linkValue)
		return container.NewHBox(hyperLink), nil
	case "div", "body", "header", "footer", "html":
		if e.Text.Len() == 0 {
			return container.NewVBox(), nil
		}
		label := canvas.NewText(e.Text.String(), TEXT_COLOR)
		return container.NewVBox(label), nil
	case "br":
		return container.NewVBox(), nil
	case "hr":
		line := canvas.NewLine(TEXT_COLOR)
		return container.NewHBox(line), nil
	default:
		return container.NewWithoutLayout(), nil
	}
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
