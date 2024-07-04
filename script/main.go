package main

import "os"

var printTokens bool = false
var printParser bool = true

func main() {
	body, err := httpClient("http://0.0.0.0:8080/")
	if err != nil {
		panic(err)
	}

	tokenizer := NewHTMLTokenizer(body)
	tokenOutput := tokenizer.TokenizeHTML(printTokens)

	root, err := buildParseTree(tokenOutput, printParser)
	if err != nil {
		panic(err)
	}
	os.Setenv("FYNE_THEME", "light")
	CreateViewer(*root)
}
