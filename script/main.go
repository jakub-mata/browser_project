package main

import "os"

var printTokens bool = false
var printParser bool = false

func main() {
	body, err := httpClient("http://0.0.0.0:8080/") //run python3 http.server 8080
	if err != nil {
		panic(err)
	}
	//tokenizer
	tokenizer := NewHTMLTokenizer(body)
	tokenOutput := tokenizer.TokenizeHTML(printTokens)
	//parser
	root, err := buildParseTree(tokenOutput, printParser)
	if err != nil {
		panic(err)
	}
	//viewer
	os.Setenv("FYNE_THEME", "light")
	CreateViewer(root)
}
